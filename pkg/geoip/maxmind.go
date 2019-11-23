package geoip

import (
	"context"
	"net"

	"github.com/oschwald/geoip2-golang"
)

type MaxMindResolver struct {
	db *geoip2.Reader
}

// NewMaxMindResolver returns a struct that implements CountryCodeResolver
func NewMaxMindResolver(db *geoip2.Reader) *MaxMindResolver {
	return &MaxMindResolver{
		db: db,
	}
}

func (r *MaxMindResolver) ResolveCountryCode(_ context.Context, ip net.IP) (code ISOCountryCode, err error) {
	record, err := r.db.Country(ip)
	if err != nil {
		return "", err
	}
	return ParseISOCode(record.Country.IsoCode), nil
}
