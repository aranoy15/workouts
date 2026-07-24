package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimit limits requests per client IP within a sliding window.
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	type clientHits struct {
		times []time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*clientHits)
	)

	return func(c *gin.Context) {
		key := c.ClientIP()
		now := time.Now()
		cutoff := now.Add(-window)

		mu.Lock()
		hits, ok := clients[key]
		if !ok {
			hits = &clientHits{}
			clients[key] = hits
		}

		kept := hits.times[:0]
		for _, t := range hits.times {
			if t.After(cutoff) {
				kept = append(kept, t)
			}
		}
		hits.times = kept

		if len(hits.times) >= limit {
			mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}

		hits.times = append(hits.times, now)
		mu.Unlock()

		c.Next()
	}
}
