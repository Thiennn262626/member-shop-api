package authjwt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"member-shop-api/pkg/logger"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenClaims chứa các trường thông tin chuẩn trong Access Token JWT
type TokenClaims struct {
	JTI      string `json:"jti"`
	UserID   string `json:"sub"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Name     string `json:"name"`
	jwt.RegisteredClaims
}

// ECDSAKeyManager quản lý cặp khóa ECDSA (thuật toán ES256) để ký và xác thực AccessToken.
type ECDSAKeyManager struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
}

// NewECDSAKeyManager khởi tạo Key Manager:
// - Nếu đã cung cấp PEM trong biến môi trường -> nạp khóa.
// - Nếu chưa cung cấp (môi trường dev) -> tự động sinh cặp khóa ECDSA P-256 an toàn trong bộ nhớ.
func NewECDSAKeyManager(privateKeyPEM, publicKeyPEM string) (*ECDSAKeyManager, error) {
	if strings.TrimSpace(privateKeyPEM) != "" {
		priv, err := ParseECPrivateKeyFromPEM(privateKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ECDSA private key: %w", err)
		}
		var pub *ecdsa.PublicKey
		if strings.TrimSpace(publicKeyPEM) != "" {
			pub, err = ParseECPublicKeyFromPEM(publicKeyPEM)
			if err != nil {
				return nil, fmt.Errorf("failed to parse ECDSA public key: %w", err)
			}
		} else {
			pub = &priv.PublicKey
		}
		return &ECDSAKeyManager{
			PrivateKey: priv,
			PublicKey:  pub,
		}, nil
	}

	// Tự động sinh cặp khóa ECDSA P-256 an toàn
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to auto-generate ECDSA key: %w", err)
	}

	if logger.Log != nil {
		logger.Log.Info().Msg("🔐 Auto-generated ECDSA P-256 keypair for ES256 JWT signing")
	}

	return &ECDSAKeyManager{
		PrivateKey: priv,
		PublicKey:  &priv.PublicKey,
	}, nil
}

// SignAccessToken tạo và ký một Access Token mới bằng thuật toán ECDSA ES256,
// tự động gắn mã jti độc nhất để phục vụ thu hồi token (blacklist).
func (km *ECDSAKeyManager) SignAccessToken(userID, username, role, name string, ttl time.Duration) (tokenString, jti string, exp time.Time, err error) {
	now := time.Now().UTC()
	exp = now.Add(ttl)
	jti = uuid.New().String()

	claims := TokenClaims{
		JTI:      jti,
		UserID:   userID,
		Username: username,
		Role:     role,
		Name:     name,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			Issuer:    "member-shop-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, err = token.SignedString(km.PrivateKey)
	if err != nil {
		return "", "", exp, fmt.Errorf("sign ES256 token failed: %w", err)
	}

	return tokenString, jti, exp, nil
}

// ValidateAccessToken kiểm tra tính hợp lệ và chữ ký ECDSA của Access Token.
func (km *ECDSAKeyManager) ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing algorithm: %v", token.Header["alg"])
		}
		return km.PublicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := parsedToken.Claims.(*TokenClaims); ok && parsedToken.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

// GenerateRandomRefreshToken tạo ra một chuỗi ngẫu nhiên bảo mật 64 ký tự hex (256 bits entropy).
// Chuỗi này không phải là JWT, dùng làm Opaque Refresh Token lưu trong HTTP-Only Cookie.
func GenerateRandomRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate secure random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// ParseECPrivateKeyFromPEM giải mã chuỗi PEM thành *ecdsa.PrivateKey (hỗ trợ cả SEC1 EC và PKCS#8).
func ParseECPrivateKeyFromPEM(pemStr string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing private key")
	}

	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	keyAny, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EC private key: %w", err)
	}

	ecKey, ok := keyAny.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not an ECDSA private key")
	}
	return ecKey, nil
}

// ParseECPublicKeyFromPEM giải mã chuỗi PEM thành *ecdsa.PublicKey (PKIX format).
func ParseECPublicKeyFromPEM(pemStr string) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing public key")
	}

	pubAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EC public key: %w", err)
	}

	ecPub, ok := pubAny.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not an ECDSA public key")
	}
	return ecPub, nil
}

// ExportECPrivateKeyPEM xuất *ecdsa.PrivateKey thành chuỗi định dạng PEM.
func ExportECPrivateKeyPEM(priv *ecdsa.PrivateKey) (string, error) {
	der, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return "", err
	}
	block := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: der,
	}
	return string(pem.EncodeToMemory(block)), nil
}

// ExportECPublicKeyPEM xuất *ecdsa.PublicKey thành chuỗi định dạng PEM.
func ExportECPublicKeyPEM(pub *ecdsa.PublicKey) (string, error) {
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", err
	}
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	}
	return string(pem.EncodeToMemory(block)), nil
}
