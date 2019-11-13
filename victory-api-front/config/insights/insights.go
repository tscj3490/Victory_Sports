package insights

/*
All libs and connector for any insight lib
*/

import (
	"github.com/getsentry/raven-go"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/config"
	"log"
	"net/http"
	"runtime/debug"
	"fmt"
	"github.com/go-chi/chi/middleware"
	"os"
	"errors"
)

var Sentry *raven.Client
var RecoveryHandler = raven.RecoveryHandler

func init() {
	var err error
	sentryConfig := config.ENVSentry()
	Sentry, err = raven.New(sentryConfig.DSN)
	if err != nil {
		log.Printf("config.insight.init failed: %v", err)
	}
	if sentryConfig.Environment != "" {
		Sentry.SetEnvironment(sentryConfig.Environment)
		log.Printf("config.insight.init Sentry.Environment: %v", sentryConfig.Environment)
	}
	if sentryConfig.Release != "" {
		Sentry.SetRelease(sentryConfig.Release)
		log.Printf("config.insight.init Sentry.Release: %v", sentryConfig.Release)
	}
}

func GenericRecoverer(rvr interface{}) {
	// BEGIN SENTRY
	var packet *raven.Packet
	rvalStr := fmt.Sprint(rvr)
	if err, ok := rvr.(error); ok {
		packet = raven.NewPacket(rvalStr, raven.NewException(errors.New(rvalStr), raven.GetOrNewStacktrace(err, 2, 3, nil)))
	} else {
		packet = raven.NewPacket(rvalStr, raven.NewException(errors.New(rvalStr), raven.NewStacktrace(2, 3, nil)))
	}
	Sentry.Capture(packet, nil)
	// END SENTRY
	fmt.Fprintf(os.Stderr, "Panic: %+v\n", rvr)
	debug.PrintStack()
}

func Recoverer(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {

				// BEGIN SENTRY
				var packet *raven.Packet
				rvalStr := fmt.Sprint(rvr)
				if err, ok := rvr.(error); ok {
					packet = raven.NewPacket(rvalStr, raven.NewException(errors.New(rvalStr), raven.GetOrNewStacktrace(err, 2, 3, nil)), raven.NewHttp(r))
				} else {
					packet = raven.NewPacket(rvalStr, raven.NewException(errors.New(rvalStr), raven.NewStacktrace(2, 3, nil)), raven.NewHttp(r))
				}
				Sentry.Capture(packet, nil)
				// END SENTRY

				logEntry := middleware.GetLogEntry(r)
				if logEntry != nil {
					logEntry.Panic(rvr, debug.Stack())
				} else {
					fmt.Fprintf(os.Stderr, "Panic: %+v\n", rvr)
					debug.PrintStack()
				}

				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

