package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo"
)

type (
	// APIStats is the middleware that collects the API execution stats and provides APIs
	// to retrieve the stats
	APIStats interface {
		Middleware() echo.MiddlewareFunc
		GetStatsHandler() echo.HandlerFunc
	}

	stats struct {
		Uptime       time.Time      `json:"uptime"`
		RequestCount uint64         `json:"requestCount"`
		Statuses     map[string]int `json:"statuses"`
		mutex        sync.RWMutex
	}
)

// NewAPIStats constucts an instance of APIStats
func NewAPIStats() APIStats {
	return &stats{
		Uptime:   time.Now(),
		Statuses: map[string]int{},
	}
}

// Middleware returns the middleware function.
func (s *stats) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if err := next(c); err != nil {
				c.Error(err)
			}
			s.mutex.Lock()
			defer s.mutex.Unlock()
			s.RequestCount++
			status := strconv.Itoa(c.Response().Status)
			s.Statuses[status]++
			return nil
		}
	}
}

// AddRoutes registers all APIs related to the stats
func (s *stats) GetStatsHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		s.mutex.RLock()
		defer s.mutex.RUnlock()
		return c.JSON(http.StatusOK, s)
	}
}
