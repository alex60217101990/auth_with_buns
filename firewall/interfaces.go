package firewall

import (
	"context"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// Create a custom visitor struct which holds the rate limiter for each
// visitor and the last time that the visitor was seen.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type bun struct {
	counter  int
	isBun    bool
	startBun time.Time
}

type Firewall interface {
	// Retrieve and return the rate limiter for the current visitor if it
	// already exists. Otherwise call the addVisitor function to add a
	// new entry to the map.
	GetVisitor(ip string) *rate.Limiter
	// Return request emote Address.
	// Remote Address - last resort
	// (usually won't be reliable as this might be the last ip or if it is a
	// naked http request to server ie no load balancer)
	GetCurrentIp(r *http.Request) string
	// Every (N) minutes check the map for visitors that haven't been seen for
	// more than (N) minutes and delete the entries.
	CleanupVisitors(ctx context.Context)
	// Middleware for detect too many requests status.
	LimitHttpMiddleware(next http.HandlerFunc) http.HandlerFunc
	// Middleware for detect request from ip with "bun" status.
	BunHttpMiddleware(next http.HandlerFunc) http.HandlerFunc
}
