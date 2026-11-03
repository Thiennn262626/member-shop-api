package config

import (
	"member-shop-api/internal/utils"
	"strconv"
)

type Config struct {
	AppEnv string
	Port   string

	// MongoDB
	MongoURI string
	MongoDB  string

	// Redis
	RedisHost     string
	RedisDB       string
	RedisUsername string
	RedisPassword string

	// JWT Auth (ECDSA ES256 & Opaque Refresh Tokens)
	JWTAccessSecret     string
	JWTRefreshSecret    string
	JWTECDSAPrivateKey  string
	JWTECDSAPublicKey   string
	AccessTokenExpires  int // seconds (default 900 = 15m)
	RefreshTokenExpires int // seconds (default 604800 = 7d)
}

func NewConfig() *Config {
	accessExp, _ := strconv.Atoi(utils.GetEnv("ACCESS_TOKEN_EXPIRES", "900"))
	refreshExp, _ := strconv.Atoi(utils.GetEnv("REFRESH_TOKEN_EXPIRES", "604800"))

	return &Config{
		AppEnv:              utils.GetEnv("APP_ENV", "development"),
		Port:                utils.GetEnv("PORT", "8088"),
		MongoURI:            utils.GetEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDB:             utils.GetEnv("MONGODB_DATABASE", "member-shop"),
		RedisHost:           utils.GetEnv("REDIS_HOST", "localhost:6379"),
		RedisDB:             utils.GetEnv("REDIS_DATABASE", "0"),
		RedisUsername:       utils.GetEnv("REDIS_USERNAME", "default"),
		RedisPassword:       utils.GetEnv("REDIS_PASSWORD", ""),
		JWTAccessSecret:     utils.GetEnv("JWT_SECRET_ACCESS", "member-shop-access-secret"),
		JWTRefreshSecret:    utils.GetEnv("JWT_SECRET_REFRESH", "member-shop-refresh-secret"),
		JWTECDSAPrivateKey:  utils.GetEnv("JWT_ECDSA_PRIVATE_KEY", ""),
		JWTECDSAPublicKey:   utils.GetEnv("JWT_ECDSA_PUBLIC_KEY", ""),
		AccessTokenExpires:  accessExp,
		RefreshTokenExpires: refreshExp,
	}
}
