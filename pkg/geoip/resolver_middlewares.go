package geoip

import (
	"context"
	"errors"
	"net"
)

func DefaultProxyResolver() ResolverMiddleware {
	return func(next Resolver) Resolver {
		return ResolverFunc(func(c context.Context, i net.IP) (ISOCountryCode, error) {
			if i.IsUnspecified() {
				return "", errors.New("ip is unspecified")
			}
			return next.ResolveCountryCode(c, i)
		})
	}
}
