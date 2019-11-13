package i18n

import (
	"path/filepath"

	"github.com/qor/i18n"
	"github.com/qor/i18n/backends/yaml"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/config"
	"log"
	"net/http"
	"context"
	"github.com/go-chi/chi"
)

var I18n *i18n.I18n
var (
	Locales = map[string]string{
		"en": "en-US",
		"ar": "ar-AE",
		"en-US": "en",
		"ar-AE": "ar",
	}
	LocaleLength = 2
)

const (
	I18NCookieName = "locale"
	I18NHeaderName = "Locale"
	I18NContextName = "CTXLocale"
	I18NDefault = "en-US"
)

func init() {
	log.Printf("I18N - Loading translation from the database and %v", filepath.Join(config.Root, "config/locales"))
	I18n = i18n.New(yaml.New(filepath.Join(config.Root, "config/locales")))
}

func GetLocale(r *http.Request) (string, string) {
	// check first url path fragment ...
	log.Printf("I18N.GetLocale %v %v", r.URL.Path, len(r.URL.Path))
	if len(r.URL.Path) > (LocaleLength+1) {
		pathFragment := r.URL.Path[1:(LocaleLength+1)]
		if locale, ok := Locales[pathFragment]; ok {
			return locale, "urlFragment"
		}
	}

	if locale := r.Header.Get(I18NHeaderName); locale != "" {
		return locale, "header"
	}
	if locale := r.URL.Query().Get(I18NCookieName); locale != "" {
		return locale, "url"
	}
	if locale, err := r.Cookie(I18NCookieName); err == nil {
		return locale.Value, "cookie"
	}

	return "", ""
}
func GetLocaleContext(r *http.Request) string {
	return r.Context().Value(I18NContextName).(string)
}
func SetLocale(w http.ResponseWriter, locale string) {
	http.SetCookie(w, &http.Cookie{
		Name: I18NCookieName,
		Value: locale,
		Secure: true,
		MaxAge: 86400, // keep the cookie one day
	})
}
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locale, locSource := GetLocale(r)
		if locSource == "urlFragment" {
			var path string
			rctx := chi.RouteContext(r.Context())
			if rctx.RoutePath != "" {
				path = rctx.RoutePath
			} else {
				path = r.URL.Path
			}
			log.Printf("I18N.Middleware locale: %v path: %v - %v", locale, path, path[(LocaleLength+1):])
			rctx.RoutePath = path[(LocaleLength+1):]
		}
		if locale == "" {
			locale = I18NDefault
		}
		ctx := context.WithValue(r.Context(), I18NContextName, locale)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
