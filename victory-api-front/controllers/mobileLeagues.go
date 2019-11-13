package controllers

import (
	"net/http"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/controllers/filters"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/gosportmonks"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

//const StatsTeamAdminResourceContextKey = "statsTeamAdminCtx"
type StatsLeagueMobileRequest struct {
	*gosportmonks.League
}

func (t *StatsLeagueMobileRequest) Bind(r *http.Request) error {
	// post-processing after decode
	return nil
}

type StatsLeagueMobileResponse struct {
	League map[string]interface{}
}

func (t *StatsLeagueMobileResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// pre-processing before sending it out
	return nil
}

type StatsLeagueMobileResource struct {
	BaseURL string
}

func (s *StatsLeagueMobileResource) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(render.SetContentType(render.ContentTypeJSON))
	// URLFormat is a middleware that parses the url extension from a request path and stores it
	// on the context as a string under the key `middleware.URLFormatCtxKey`. The middleware will
	// trim the suffix from the routing path and continue routing.
	r.Use(middleware.URLFormat)

	r.Get("/", s.ListView)
	r.Post("/", s.CreateView)

	r.Route("/{id}", func(in chi.Router) {
		//in.Use(StatsTeamAdminCtx)
		in.Get("/", s.ShowView)
		in.Put("/", s.EditView)
		in.Delete("/", s.DeleteView)
	})
	return r
}

// func StatsTeamAdminCtx(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		var (
// 			League  = &models.League{}
// 			err   error
// 			objID = chi.URLParam(r, "id")
// 			tx    = db.GetDBFromRequestContext(r)
// 		)

// 		if objID == "" {
// 			render.Render(w, r, ErrNotFound)
// 			return
// 		}

// 		err = tx.Model(League).Where("id = ?", objID).Find(League).Error
// 		if err == gorm.ErrRecordNotFound {
// 			render.Render(w, r, ErrNotFound)
// 			return
// 		}
// 		if err != nil {
// 			render.Render(w, r, ErrInternalServerError)
// 			return
// 		}

// 		ctx := context.WithValue(r.Context(), StatsTeamAdminResourceContextKey, League)
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }
// func GetStatsTeamCtx(r *http.Request) *models.League {
// 	obj, _ := r.Context().Value(StatsTeamAdminResourceContextKey).(*models.League)
// 	return obj
// }

//type dynamicMobileType map[string]interface{}

func (s *StatsLeagueMobileResource) ListView(w http.ResponseWriter, r *http.Request) {
	var (
		leagues = []*gosportmonks.League{}
		ctx     = r.Context()
	)

	statsFilter := &filters.StatsFilter{
		Ctx: ctx,
	}

	leagues = statsFilter.ListCompetitions()

	render.JSON(w, r, leagues)
	return
}
func (s *StatsLeagueMobileResource) CreateView(w http.ResponseWriter, r *http.Request) {

}
func (s *StatsLeagueMobileResource) ShowView(w http.ResponseWriter, r *http.Request) {

}
func (s *StatsLeagueMobileResource) EditView(w http.ResponseWriter, r *http.Request) {
}
func (s *StatsLeagueMobileResource) DeleteView(w http.ResponseWriter, r *http.Request) {
}
