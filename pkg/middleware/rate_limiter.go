package middleware

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/fredericobenevides/pge-rate-limiter/internal/storage"
)

type RateLimitMiddleware struct {
	storageStrategy    storage.StorageStrategy
	requestIpLimit     int
	requestTokenLimit  int
	windowDuration     time.Duration
	blockIpDuration    time.Duration
	blockTokenDuration time.Duration
}

func NewRateLimitMiddleware(storageStrategy storage.StorageStrategy, requestIpLimit int, requestTokenLimit int,
	windowDuration time.Duration,
	blockIpDuration time.Duration, blockTokenDuration time.Duration) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		storageStrategy:    storageStrategy,
		requestIpLimit:     requestIpLimit,
		requestTokenLimit:  requestTokenLimit,
		windowDuration:     windowDuration,
		blockIpDuration:    blockIpDuration,
		blockTokenDuration: blockTokenDuration,
	}
}

func (rl *RateLimitMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var id string
		var requestLimit int
		var blockDuration time.Duration

		token := r.Header.Get("API_KEY")

		if token != "" {
			id = token
			requestLimit = rl.requestTokenLimit
			blockDuration = rl.blockTokenDuration

		} else {
			ip, err := rl.getClientIP(r)
			if err != nil {
				http.Error(w, "Invalid remote address", http.StatusInternalServerError)
				return
			}

			id = ip
			requestLimit = rl.requestIpLimit
			blockDuration = rl.blockIpDuration
		}

		allowed, err := rl.storageStrategy.AllowRequest(r.Context(), id, requestLimit, rl.windowDuration, blockDuration)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !allowed {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(blockDuration.Seconds())))
			http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimitMiddleware) getClientIP(r *http.Request) (string, error) {
	if xForwardedFor := r.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0]), nil
		}
	}

	if xRealIP := r.Header.Get("X-Real-IP"); xRealIP != "" {
		log.Println("X-Real-IP:", xRealIP)
		return xRealIP, nil
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}

	if ip == "::1" {
		ip = "127.0.0.1"
	}

	return ip, nil
}
