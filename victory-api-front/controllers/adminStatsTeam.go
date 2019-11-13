package controllers

import (
	"github.com/go-chi/chi"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/models"
	"github.com/jinzhu/gorm"
	"fmt"
	"net/http"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/db"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/i18n/l10n"
	"log"
	"github.com/go-chi/render"
	"github.com/go-chi/chi/middleware"
	"context"
	"strings"
	"strconv"
)

// MARK: Team:
const StatsTeamAdminResourceContextKey = "statsTeamAdminCtx"
type StatsTeamAdminRequest struct {
	*models.Team
	NameL10N               map[string]string `json:"Name"`
	LeaguesTmp             []int             `json:"Leagues"`
}
func (t *StatsTeamAdminRequest) Bind(r *http.Request) error {
	// post-processing after decode
	tx := db.GetDBFromRequestContext(r)

	t.Team.Name = l10n.SetAll(t.NameL10N)

	leagues := []models.League{}
	if err := tx.Where("id in (?)", t.LeaguesTmp).Find(&leagues).Error; err != nil {
		log.Printf("STAR.CreateResource.Bind cant find leagues. %v", err)
	}
	t.Team.Leagues = leagues


	if t.StatsTeamIDCombinedKey != "" {
		parts := strings.Split(t.StatsTeamIDCombinedKey, "-")
		if len(parts) == 3 {
			teamID, err := strconv.Atoi(parts[2])
			if err == nil {
				t.Team.StatsTeamID = teamID
			}
		}
	}

	return nil
}
type StatsTeamAdminResource struct {
	BaseURL string
}
type StatsTeamAdminResponse struct {
	*models.Team
	NameL10N map[string]string `json:"Name"`
	LeaguesTmp []int `json:"Leagues"`
}
func (t *StatsTeamAdminResponse) Render(w http.ResponseWriter, r *http.Request) error {
	t.NameL10N = l10n.GetAll(t.Name)

	leagueIds := []int{}
	for _, l := range t.Leagues {
		leagueIds = append(leagueIds, int(l.ID))
	}
	t.LeaguesTmp = leagueIds
	return nil
}

type StatsTeamAdminListResponse struct {
	//TeamsL10N []*StatsTeamAdminResponse `json:"teams"`
	Teams *[]models.Team `json:"teams"`
}
func (lt *StatsTeamAdminListResponse) Render(w http.ResponseWriter, r *http.Request) error {

	return nil
}

func (s *StatsTeamAdminResource) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(render.SetContentType(render.ContentTypeJSON))
	// URLFormat is a middleware that parses the url extension from a request path and stores it
	// on the context as a string under the key `middleware.URLFormatCtxKey`. The middleware will
	// trim the suffix from the routing path and continue routing.
	r.Use(middleware.URLFormat)

	r.Get("/", s.ListView)
	r.Post("/", s.CreateView)

	r.Route("/{id}", func(in chi.Router) {
		in.Use(StatsTeamAdminCtx)
		in.Get("/", s.ShowView)
		in.Put("/", s.EditView)
		in.Delete("/", s.DeleteView)
	})


	return r
}
func StatsTeamAdminCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			team  = &models.Team{}
			err   error
			objID   = chi.URLParam(r, "id")
			tx         = db.GetDBFromRequestContext(r)
		)

		if objID == "" {
			render.Render(w, r, ErrNotFound)
			return
		}

		err = tx.Model(team).Where("id = ?", objID).
			Preload("Leagues").
			Find(team).Error
		if err == gorm.ErrRecordNotFound {
			render.Render(w, r, ErrNotFound)
			return
		}
		if err != nil {
			e := ErrInternalServerError(err)
			render.Render(w, r, e)
			return
		}

		ctx := context.WithValue(r.Context(), StatsTeamAdminResourceContextKey, team)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func GetStatsTeamCtx(r *http.Request) *models.Team {
	obj, _ := r.Context().Value(StatsTeamAdminResourceContextKey).(*models.Team)
	return obj
}


func (s *StatsTeamAdminResource) ListView(w http.ResponseWriter, r *http.Request) {
	var (
		teams = []models.Team{}
		tx    = db.GetDBFromRequestContext(r)
	)

	if err := tx.Table("teams").Order("created_at").
		Preload("Leagues").
		Find(&teams).Error; err != nil {
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}

	var localizedTeams []map[string]interface{}
	var tmpTeams []l10n.L10NModel
	for _, t := range teams {
		teamTmp := l10n.L10NModel(t)
		tmpTeams = append(tmpTeams, teamTmp)
	}

	localizedTeams = l10n.List(&models.Team{}, tmpTeams)
	render.JSON(w, r, localizedTeams)
	//render.Status(r, http.StatusOK)
	//if err := render.Render(w, r, &teams); err != nil {
	//	render.Render(w, r, ErrRender(err))
	//	return
	//}

	return
}
func (s *StatsTeamAdminResource) CreateView(w http.ResponseWriter, r *http.Request) {
	var (
		tx = db.GetDBFromRequestContext(r)
	)
	data := &StatsTeamAdminRequest{}
	// serialize the payload
	if err := render.Bind(r, data); err != nil {
		fmt.Errorf("bind failed %v", err)
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	team := data.Team;

	if err := tx.Create(&team).Error; err != nil {
		fmt.Errorf("create failed %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, l10n.Unpack(team))
}
func (s *StatsTeamAdminResource) ShowView(w http.ResponseWriter, r *http.Request) {
	var (
		obj = GetStatsTeamCtx(r)
	)
	if obj == nil {
		render.Render(w, r, ErrNotFound)
		return
	}

	log.Printf("STeamAdminR.ShowView: Obj: %v", obj)

	//render.Status(r, http.StatusOK)
	//render.JSON(w, r, l10n.Unpack(obj))
	render.Status(r, http.StatusOK)
	if err := render.Render(w, r, &StatsTeamAdminResponse{Team: obj}); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}
func (s *StatsTeamAdminResource) EditView(w http.ResponseWriter, r *http.Request) {
	var (
		obj = GetStatsTeamCtx(r)
	)
	tx := db.GetDBFromRequestContext(r)
	if obj == nil {
		render.Render(w, r, ErrNotFound)
		return
	}

	data := StatsTeamAdminRequest{}
	if err := render.Bind(r, &data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	team := data.Team
	log.Printf("STAR.EV: %#v", team)

	if err := tx.Model(team).Where("id = ?", obj.ID).Update(&team).Error; err != nil {
		log.Printf("STAR.EditView failed: %v", err)
		render.Render(w, r, ErrRender(err))
		return
	}

	log.Printf("STeamAdminR.EditView: Obj: %v", obj)

	render.Status(r, http.StatusOK)
	if err := render.Render(w, r, &StatsTeamAdminResponse{Team: team}); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}
func (s *StatsTeamAdminResource) DeleteView(w http.ResponseWriter, r *http.Request) {
	var (
		obj = GetStatsTeamCtx(r)
	)
	tx := db.GetDBFromRequestContext(r)
	if obj == nil {
		render.Render(w, r, ErrNotFound)
		return
	}
	if err := tx.Model(obj).Where("id = ?", obj.ID).Delete(&obj).Error; err != nil {
		log.Printf("STAR.EditView failed: %v", err)
		render.Render(w, r, ErrRender(err))
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, l10n.Unpack(obj))
}
