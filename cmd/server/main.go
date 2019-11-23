package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-kit/kit/log"
	"github.com/ip2location/ip2location-go"

	stdlog "log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	zaplogger "github.com/go-kit/kit/log/zap"

	"github.com/etherlabsio/pkg/version"
	"github.com/heartfulnessinstitute/geoip-router/pkg/geoip"
	"github.com/oklog/run"
	"github.com/oschwald/geoip2-golang"
	"github.com/peterbourgon/ff"
)

type Configuration struct {
	resolver   *string
	dbFilePath *string
	httpAddr   *string
}

func exitOnErr(err error) {
	if err == nil {
		return
	}
	stdlog.Fatalf("received err: %+v", err)
}

func main() {
	fs := flag.NewFlagSet("geo-router", flag.ExitOnError)
	config := Configuration{
		resolver:   fs.String("resolver", "ip2location", "resolver to use for resolving IP to a Country Code"),
		dbFilePath: fs.String("db_file_path", "resources/GeoLite2-Country.mmdb", "base url for the API server"),
		httpAddr:   fs.String("http_addr", ":8080", "address to bind the http listener"),
	}
	err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarNoPrefix())
	exitOnErr(err)

	zaplog, err := zap.NewDevelopment()
	exitOnErr(err)

	var logger log.Logger
	{
		logger = zaplogger.NewZapSugarLogger(zaplog, zapcore.InfoLevel)
	}

	var resolver geoip.Resolver
	if *config.resolver == "ip2location" {
		db, err := geoip2.Open(*config.dbFilePath)
		exitOnErr(err)
		defer db.Close()

		resolver = geoip.NewMaxMindResolver(db)
	} else {
		ip2location.Open(*config.dbFilePath)
		defer ip2location.Close()

		resolver = geoip.NewIP2LocationResolver()
	}

	// applying common middlewares
	resolver = geoip.DefaultProxyResolver()(resolver)
	resolver = geoip.ErrorLoggingResolverMiddleware(logger)(resolver)

	rules := geoip.CountryLocationRoutingRules{
		geoip.ISOCountryCodeUS: "/us",
		geoip.ISOCountryCodeIN: "/in",
	}
	geoRouter := geoip.HTTPResolverHandler(resolver, rules)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/liveness_check"))
	r.Use(middleware.Heartbeat("/readiness_check"))
	r.Mount("/debug", middleware.Profiler())
	r.Handle("/us", http.NotFoundHandler())
	r.Handle("/in", http.NotFoundHandler())
	r.Handle("/", geoRouter)

	// Now we're to the part of the func main where we want to start actually
	// running things, like servers bound to listeners to receive connections.
	//
	// The method is the same for each component: add a new actor to the group
	// struct, which is a combination of 2 anonymous functions: the first
	// function actually runs the component, and the second function should
	// interrupt the first function and cause it to return. It's in these
	// functions that we actually bind the server/handler structs to the
	// concrete transports and run them.
	//
	// Putting each component into its own block is mostly for aesthetics: it
	// clearly demarcates the scope in which each listener/socket may be used.
	var g run.Group
	{
		// The HTTP listener mounts the HTTP handler we created.
		server := &http.Server{Addr: *config.httpAddr, Handler: r}
		g.Add(func() error {
			logger.Log("transport", "HTTP", "addr", *config.httpAddr, "version", version.Version())
			return server.ListenAndServe()
		}, func(error) {
			ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
			server.Shutdown(ctx)
		})
	}
	{
		// This function just sits and waits for ctrl-C.
		cancelInterrupt := make(chan struct{})
		g.Add(func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				return fmt.Errorf("received signal %s", sig)
			case <-cancelInterrupt:
				return nil
			}
		}, func(error) {
			close(cancelInterrupt)
		})
	}
	logger.Log("exit", g.Run())
}
