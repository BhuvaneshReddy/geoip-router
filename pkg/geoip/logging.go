package geoip

import (
	"context"
	"net"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// ResolverMiddleware supports error logging
type ResolverMiddleware func(Resolver) Resolver

// ErrorLoggingResolverMiddleware returns a ResolverMiddleware with error logging
func ErrorLoggingResolverMiddleware(l log.Logger) ResolverMiddleware {
	return func(r Resolver) Resolver {
		return &errorLoggingMiddleware{
			logger: l,
			next:   r,
		}
	}
}

type errorLoggingMiddleware struct {
	logger log.Logger
	next   Resolver
}

func (mw *errorLoggingMiddleware) ResolveCountryCode(ctx context.Context, ip net.IP) (code ISOCountryCode, err error) {
	defer func() {
		if err != nil {
			level.Error(mw.logger).Log(
				"method", "geoip.ResolveCountryCode",
				"ip", ip.String(),
				"err", err,
			)
		}
	}()
	return mw.next.ResolveCountryCode(ctx, ip)
}
