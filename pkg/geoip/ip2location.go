package geoip

import (
	"context"
	"fmt"
	"net"

	"github.com/ip2location/ip2location-go"
)

type IP2LocationResolver struct{}

// NewIP2LocationResolver returns a struct that implements CountryCodeResolver
func NewIP2LocationResolver() *IP2LocationResolver {
	return &IP2LocationResolver{}
}

func (r *IP2LocationResolver) ResolveCountryCode(_ context.Context, ip net.IP) (code ISOCountryCode, err error) {
	record := ip2location.Get_country_short(ip.String())
	if record.Country_short == "" {
		return "", fmt.Errorf("ip2location: could not resolve IP: %s to a country_short", ip)
	}
	return ParseISOCode(record.Country_short), nil
}
