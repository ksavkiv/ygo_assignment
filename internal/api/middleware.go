package api

import (
	"log/slog"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

func RateLimiter(rateLimit rate.Limit, burst int) func(http.Handler) http.Handler {
	var mu sync.Mutex
	limiters := make(map[string]*rate.Limiter)

	getLimiter := func(ip string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()
		if lim, ok := limiters[ip]; ok {
			return lim
		}
		lim := rate.NewLimiter(rateLimit, burst)
		limiters[ip] = lim
		return lim
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			lim := getLimiter(ip)
			if !lim.Allow() {
				retryAfter := math.Ceil(1 / float64(rateLimit))
				w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter)))
				slog.Info("rate limit exceeded", "ip", ip, "method", r.Method, "path", r.URL.Path)
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func BearerAuth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				slog.Info("auth: missing authorization header", "method", r.Method, "path", r.URL.Path)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(auth, "Bearer ") {
				slog.Info("auth: invalid authorization scheme", "method", r.Method, "path", r.URL.Path)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if strings.TrimPrefix(auth, "Bearer ") != token {
				slog.Info("auth: invalid bearer token", "method", r.Method, "path", r.URL.Path)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
