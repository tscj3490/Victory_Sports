package routes

import (
	"net/http"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"path"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/config"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/controllers"
	"os"
	"path/filepath"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/db"
	"log"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/insights"
)

var rootMux *http.ServeMux

func Router() *http.ServeMux {
	if rootMux != nil {
		return rootMux
	}

	// build the r
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(insights.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.GetHead)
	r.Use(middleware.Compress(2, "gzip"))

	// add database handle to request context
	r.Use(db.DBRequestContext)

	//r.Get("/", controllers.Homepage)
	r.Mount("/", controllers.WebResource{}.Routes())

	// connect the routes and the http handler
	rootMux = http.NewServeMux()
	rootMux.Handle("/", r)
	rootMux.Handle("/public/", http.StripPrefix("/public/",
		FileServer(http.Dir(config.ENVPublicDir()))))
	rootMux.Handle("/assets/", http.StripPrefix("/assets/",
		FileServer(http.Dir(filepath.Join(config.ENVPublicDir(), "static")))))
	rootMux.Handle(config.ENVFileUploadURI, http.StripPrefix("/uploads",
		FileServer(http.Dir(filepath.Join(config.ENVFileUploadDir())))))

	return rootMux
}

// FileServer file server that disabled file listing
func FileServer(dir http.Dir) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := path.Join(string(dir), r.URL.Path)
		if f, err := os.Stat(p); err == nil && !f.IsDir() {
			http.ServeFile(w, r, p)
			return
		}
		log.Printf("static file not found %v", p)

		http.NotFound(w, r)
	})
}
