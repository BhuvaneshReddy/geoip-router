package geoip

import (
	"context"
	"fmt"
	"net"

	"github.com/oschwald/geoip2-golang"
)

type DatabaseResolver struct {
	fallback ISOCountryCode
	db       DatabaseReader
}

type DatabaseReader interface {
	Country(net.IP) (*geoip2.Country, error)
}

// NewDatabaseResolver returns a struct that implements CountryCodeResolver
func NewDatabaseResolver(db DatabaseReader, fallback ISOCountryCode) *DatabaseResolver {
	return &DatabaseResolver{
		fallback: fallback,
		db:       db,
	}
}

func (r *DatabaseResolver) ResolveCountryCode(_ context.Context, ip net.IP) (code ISOCountryCode, err error) {
	if ip.IsUnspecified() {
		return r.fallback, err
	}
	record, err := r.db.Country(ip)
	if err != nil {
		return code, err
	}
	fmt.Println("ISO country code: ", record.Country.IsoCode)
	return ParseISOCode(record.Country.IsoCode), nil
}
