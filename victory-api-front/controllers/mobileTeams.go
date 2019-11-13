package controllers

import (
	"net/http"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/controllers/filters"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/gosportmonks"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type formSelection struct {
	Name string
	ID   int
	//Href string
}

//const StatsTeamAdminResourceContextKey = "statsTeamAdminCtx"
type StatsTeamMobileRequest struct {
	*gosportmonks.Team
}

func (t *StatsTeamMobileRequest) Bind(r *http.Request) error {
	// post-processing after decode
	return nil
}

type StatsTeamMobileResponse struct {
	team map[string]interface{}
}

func (t *StatsTeamMobileResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// pre-processing before sending it out
	return nil
}

type StatsTeamMobileResource struct {
	BaseURL string
}

func (s *StatsTeamMobileResource) Routes() chi.Router {
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
// 			team  = &models.Team{}
// 			err   error
// 			objID = chi.URLParam(r, "id")
// 			tx    = db.GetDBFromRequestContext(r)
// 		)

// 		if objID == "" {
// 			render.Render(w, r, ErrNotFound)
// 			return
// 		}

// 		err = tx.Model(team).Where("id = ?", objID).Find(team).Error
// 		if err == gorm.ErrRecordNotFound {
// 			render.Render(w, r, ErrNotFound)
// 			return
// 		}
// 		if err != nil {
// 			render.Render(w, r, ErrInternalServerError)
// 			return
// 		}

// 		ctx := context.WithValue(r.Context(), StatsTeamAdminResourceContextKey, team)
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }
// func GetStatsTeamCtx(r *http.Request) *models.Team {
// 	obj, _ := r.Context().Value(StatsTeamAdminResourceContextKey).(*models.Team)
// 	return obj
// }

//type dynamicMobileType map[string]interface{}

func (s *StatsTeamMobileResource) ListView(w http.ResponseWriter, r *http.Request) {
	var (
		teams    = []gosportmonks.Team{}
		ctx      = r.Context()
		seasonID = 6397
	)

	statsFilter := &filters.StatsFilter{
		Ctx: ctx,
	}

	teams = statsFilter.ListTeamsBy(seasonID)

	render.JSON(w, r, teams)
	return
}
func (s *StatsTeamMobileResource) CreateView(w http.ResponseWriter, r *http.Request) {

}
func (s *StatsTeamMobileResource) ShowView(w http.ResponseWriter, r *http.Request) {

}
func (s *StatsTeamMobileResource) EditView(w http.ResponseWriter, r *http.Request) {
}
func (s *StatsTeamMobileResource) DeleteView(w http.ResponseWriter, r *http.Request) {
}
