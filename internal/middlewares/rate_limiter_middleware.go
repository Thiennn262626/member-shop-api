package middlewares

import (
	"member-shop-api/internal/utils"
	"member-shop-api/pkg/apperror"
	"member-shop-api/pkg/response"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

type Client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	mu          sync.RWMutex
	clients     = make(map[string]*Client)
	cleanupOnce sync.Once
)

func getClientIP(ctx *gin.Context) string {
	ip := ctx.ClientIP()
	if ip == "" {
		ip = ctx.Request.RemoteAddr
	}

	return ip
}

func getRateLimiter(ip string) *rate.Limiter {
	mu.RLock()
	client, exists := clients[ip]
	mu.RUnlock()

	if exists {
		mu.Lock()
		client.lastSeen = time.Now()
		mu.Unlock()
		return client.limiter
	}

	mu.Lock()
	defer mu.Unlock()

	// Double check sau khi lấy Write Lock
	if client, exists := clients[ip]; exists {
		client.lastSeen = time.Now()
		return client.limiter
	}

	requestSec := utils.GetIntEnv("RATE_LIMITER_REQUEST_SEC", 5)
	brust := utils.GetIntEnv("RATE_LIMITER_REQUEST_BURST", 10)

	limiter := rate.NewLimiter(rate.Limit(requestSec), brust)
	newClient := &Client{limiter, time.Now()}
	clients[ip] = newClient
	return limiter
}

func CleanupClients() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		mu.Lock()
		now := time.Now()
		for ip, client := range clients {
			if now.Sub(client.lastSeen) > 3*time.Minute {
				delete(clients, ip)
			}
		}
		mu.Unlock()
	}
}

// ab -n 20 -c 1 -H "X-API-Key:894f5978-f86b-4b04-9bc0-328e7ac6166e" http://localhost:8080/api/v1/users
func RateLimiterMiddleware(rateLimterLogger *zerolog.Logger) gin.HandlerFunc {
	// Khởi chạy dọn dẹp bộ nhớ RAM định kỳ chống memory leak
	cleanupOnce.Do(func() {
		go CleanupClients()
	})

	return func(ctx *gin.Context) {
		ip := getClientIP(ctx)

		limiter := getRateLimiter(ip)

		if !limiter.Allow() {
			if shoudLogRateLimit(ip) {
				rateLimterLogger.Warn().
					Str("method", ctx.Request.Method).
					Str("path", ctx.Request.URL.Path).
					Str("query", ctx.Request.URL.RawQuery).
					Str("client_ip", ctx.ClientIP()).
					Str("user_agent", ctx.Request.UserAgent()).
					Str("referer", ctx.Request.Referer()).
					Str("protocol", ctx.Request.Proto).
					Str("host", ctx.Request.Host).
					Str("remote_addr", ctx.Request.RemoteAddr).
					Str("request_uri", ctx.Request.RequestURI).
					Interface("headers", ctx.Request.Header).
					Msg("rate limiter exceeded")
			}

			response.FailWithStatus(ctx, http.StatusTooManyRequests, apperror.CodeTooManyRequests, "Bạn đã gửi quá nhiều yêu cầu. Vui lòng thử lại sau ít phút.")
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

var rateLimitLogCache = sync.Map{}

const rateLimitLogTTL = 20 * time.Second

func shoudLogRateLimit(ip string) bool {
	now := time.Now()

	if val, ok := rateLimitLogCache.Load(ip); ok {
		if t, ok := val.(time.Time); ok && now.Sub(t) < rateLimitLogTTL {
			return false
		}
	}

	rateLimitLogCache.Store(ip, now)
	return true
}
