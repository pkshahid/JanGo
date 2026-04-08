package security

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	godjangohttp "github.com/godjango/godjango/http"
	"github.com/godjango/godjango/http/middleware"
)

type bucket struct {
	tokens int
	last   time.Time
}

// RateLimitMiddleware implements a token bucket rate limiter.
type RateLimitMiddleware struct {
	mu           sync.Mutex
	ipBuckets    map[string]*bucket
	userBuckets  map[string]*bucket
	rate         int           // tokens added per period
	capacity     int           // max tokens
	period       time.Duration // refill period
}

// NewRateLimitMiddleware creates a new rate limiter.
// rate is how many requests allowed per period.
func NewRateLimitMiddleware(rate, capacity int, period time.Duration) *RateLimitMiddleware {
	m := &RateLimitMiddleware{
		ipBuckets:   make(map[string]*bucket),
		userBuckets: make(map[string]*bucket),
		rate:        rate,
		capacity:    capacity,
		period:      period,
	}

	// Background cleanup/refill isn't strictly necessary if we calculate on access,
	// but requirement explicitly says "Uses goroutine + ticker to refill buckets".
	go m.refillLoop()
	return m
}

func (m *RateLimitMiddleware) refillLoop() {
	ticker := time.NewTicker(m.period)
	for range ticker.C {
		m.mu.Lock()
		// Refill IP buckets
		for k, b := range m.ipBuckets {
			b.tokens += m.rate
			if b.tokens > m.capacity {
				b.tokens = m.capacity
			}
			// Cleanup old buckets
			if time.Since(b.last) > m.period*10 {
				delete(m.ipBuckets, k)
			}
		}
		// Refill User buckets
		for k, b := range m.userBuckets {
			b.tokens += m.rate
			if b.tokens > m.capacity {
				b.tokens = m.capacity
			}
			if time.Since(b.last) > m.period*10 {
				delete(m.userBuckets, k)
			}
		}
		m.mu.Unlock()
	}
}

func (m *RateLimitMiddleware) Process(next middleware.Handler) middleware.Handler {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		var key string
		var buckets map[string]*bucket

		if req.User != nil && req.User.IsAuthenticated() {
			key = strconv.FormatUint(req.User.ID(), 10)
			buckets = m.userBuckets
		} else {
			key = req.META["REMOTE_ADDR"]
			buckets = m.ipBuckets
		}

		m.mu.Lock()
		b, exists := buckets[key]
		if !exists {
			b = &bucket{tokens: m.capacity, last: time.Now()}
			buckets[key] = b
		}

		b.last = time.Now()

		if b.tokens <= 0 {
			m.mu.Unlock()
			resp := godjangohttp.NewHttpResponse("429 Too Many Requests", http.StatusTooManyRequests)
			resp.Headers.Set("Retry-After", strconv.Itoa(int(m.period.Seconds())))
			return resp
		}

		b.tokens--
		m.mu.Unlock()

		return next(req)
	}
}
