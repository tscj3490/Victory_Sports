package controllers

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/go-chi/chi/middleware"
	"net/http"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/models"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/db"
	"fmt"
	"context"
	"github.com/jinzhu/gorm"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/i18n/l10n"
	"log"
)

// MARK: League
const StatsLeagueAdminResourceContextKey = "statsLeagueAdminCtx"
type StatsLeagueAdminRequest struct {
	*models.League
}
func (t *StatsLeagueAdminRequest) Bind(r *http.Request) error {
	// post-processing after decode
	return nil
}
type StatsLeagueAdminResponse struct {
	*models.League
}
func (t *StatsLeagueAdminResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// pre-processing before sending it out
	return nil
}
type StatsLeagueAdminResource struct {
	BaseURL string
}
func (s *StatsLeagueAdminResource) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(render.SetContentType(render.ContentTypeJSON))
	// URLFormat is a middleware that parses the url extension from a request path and stores it
	// on the context as a string under the key `middleware.URLFormatCtxKey`. The middleware will
	// trim the suffix from the routing path and continue routing.
	r.Use(middleware.URLFormat)

	r.Get("/", s.ListView)
	r.Post("/", s.CreateView)

	r.Route("/{id}", func(in chi.Router) {
		in.Use(StatsLeagueAdminCtx)
		in.Get("/", s.ShowView)
		//in.Put("/", s.EditView)
		//in.Delete("/", s.DeleteView)
	})
	return r
}
func StatsLeagueAdminCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			league = &models.League{}
			err    error
			objID  = chi.URLParam(r, "id")
			tx     = db.GetDBFromRequestContext(r)
		)

		if objID == "" {
			render.Render(w, r, ErrNotFound)
			return
		}

		err = tx.Model(league).Where("id = ?", objID).Find(league).Error
		if err == gorm.ErrRecordNotFound {
			render.Render(w, r, ErrNotFound)
			return
		}
		if err != nil {
			fmt.Errorf("StatsLeagueAdminCtx Err: %v", err)
			e := ErrInternalServerError(err)
			render.Render(w, r, e)
			return
		}

		ctx := context.WithValue(r.Context(), StatsLeagueAdminResourceContextKey, league)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func GetStatsLeagueCtx(r *http.Request) *models.League {
	obj, _ := r.Context().Value(StatsLeagueAdminResourceContextKey).(*models.League)
	return obj
}
func (s *StatsLeagueAdminResource) ListView(w http.ResponseWriter, r *http.Request) {
	var (
		tx = db.GetDBFromRequestContext(r)
		leagues = []models.League{}
	)

	if err := tx.Order("created_at").Find(&leagues).Error; err != nil {
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}


	var localizedLeagues []map[string]interface{}
	var tmpLeagues []l10n.L10NModel
	for _, t := range leagues {
		tmpLeagues = append(tmpLeagues, l10n.L10NModel(t))
	}

	localizedLeagues = l10n.List(&models.League{}, tmpLeagues)

	render.JSON(w, r, localizedLeagues)

	//smApi := &sportmonks_api.SportmonksAPI{
	//	Ctx: ctx,
	//	DontTriggerCacheDBSave: true,
	//}
	//
	//render.JSON(w, r, smApi.ListCompetitions())
	return
}
func (s *StatsLeagueAdminResource) CreateView(w http.ResponseWriter, r *http.Request) {
	var (
		tx = db.GetDBFromRequestContext(r)
	)
	data := &StatsLeagueAdminRequest{}
	// serialize the payload
	if err := render.Bind(r, data); err != nil {
		fmt.Errorf("bind failed %v", err)
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	league := data.League;

	if err := tx.Create(&league).Error; err != nil {
		log.Printf("create failed %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
	render.Status(r, http.StatusCreated)
	render.Render(w, r, &StatsLeagueAdminResponse{league})
}
func (s *StatsLeagueAdminResource) ShowView(w http.ResponseWriter, r *http.Request) {
	var (
		obj = GetStatsLeagueCtx(r)
	)
	if obj == nil {
		render.Render(w, r, ErrNotFound)
		return
	}

	if err := render.Render(w, r, &StatsLeagueAdminResponse{obj}); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

