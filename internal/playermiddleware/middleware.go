// Package playermiddleware provides the playback-chain middleware mechanism
// for go-musicfox. A Chain holds an ordered slice of URLMiddleware funcs that
// run between playable-source resolution (ResolvePlayableSource) and the final
// engine Play call, letting plugins observe or rewrite the resolved URLMusic.
package playermiddleware

import (
	"context"
	"errors"

	"github.com/go-musicfox/go-musicfox/internal/player"
)

// MiddlewareNext is the next handler in the middleware chain.
type MiddlewareNext func(ctx context.Context, urlMusic *player.URLMusic) error

// URLMiddleware is a playback-chain middleware. It receives the resolved
// URLMusic (URL, Song, Type), may rewrite it, and must call next to continue
// the chain. Returning an error aborts the chain.
type URLMiddleware func(ctx context.Context, urlMusic *player.URLMusic, next MiddlewareNext) error

// ErrBlockedTrack is returned by middleware to mark a track as intentionally
// unplayable (for example the UNM banned-path interception), so callers can
// distinguish "blocked" from generic failures.
var ErrBlockedTrack = errors.New("track blocked by playback middleware")

// Chain is an ordered collection of URLMiddleware.
type Chain struct {
	middlewares []URLMiddleware
}

// NewChain creates an empty middleware chain.
func NewChain() *Chain {
	return &Chain{}
}

// Use appends one or more middleware to the chain in order.
func (c *Chain) Use(mw ...URLMiddleware) *Chain {
	c.middlewares = append(c.middlewares, mw...)
	return c
}

// Len returns the number of registered middleware.
func (c *Chain) Len() int {
	return len(c.middlewares)
}

// Execute runs the middleware chain against urlMusic. An empty chain is a
// no-op returning nil. The first non-nil error returned by any middleware
// aborts the chain and is propagated to the caller.
func (c *Chain) Execute(ctx context.Context, urlMusic *player.URLMusic) error {
	var run func(i int) error
	run = func(i int) error {
		if i >= len(c.middlewares) {
			return nil
		}
		return c.middlewares[i](ctx, urlMusic, func(ctx context.Context, m *player.URLMusic) error {
			return run(i + 1)
		})
	}
	return run(0)
}
