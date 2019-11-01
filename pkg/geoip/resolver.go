package geoip

import (
	"context"
	"net"
)

// Resolver takes an IP address and resolves it to an ISO standardized country code
type Resolver interface {
	ResolveCountryCode(context.Context, net.IP) (ISOCountryCode, error)
}

// ResolverFunc is a helper function alias
type ResolverFunc func(context.Context, net.IP) (ISOCountryCode, error)

// ResolveCountryCode method for the function alias CountryCodeResolverFunc to comply with the interface CountryCodeResolver
func (f ResolverFunc) ResolveCountryCode(ctx context.Context, ip net.IP) (ISOCountryCode, error) {
	return f(ctx, ip)
}
