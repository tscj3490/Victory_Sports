package controllers

import (
	"github.com/go-chi/chi"
	"github.com/jinzhu/gorm"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/models"
	"net/http"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/db"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/i18n/l10n"
	"log"
	"github.com/go-chi/render"
	"github.com/go-chi/chi/middleware"
	"context"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/sportmonks_api"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/gosportmonks"
	"fmt"
	"strconv"
	"strings"
)

const SportmonksAdminResourceContextKey = "sportmonksAdminCtx"
type SportmonksAdminResource struct {
	BaseURL string
}

func (s *SportmonksAdminResource) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(render.SetContentType(render.ContentTypeJSON))
	// URLFormat is a middleware that parses the url extension from a request path and stores it
	// on the context as a string under the key `middleware.URLFormatCtxKey`. The middleware will
	// trim the suffix from the routing path and continue routing.
	r.Use(middleware.URLFormat)

	r.Route("/sportmonksTeams", func(in chi.Router) {
		in.Get("/", s.ListTeamsView)
		in.Route("/{leagueSeasonTeamId}", func(t chi.Router) {
			//t.Use(SportmonksAdminCtx)
			t.Get("/", s.ShowTeamView)
			t.Put("/", s.EditTeamView)
			//t.Delete("/", s.DeleteView)
		})
	})
	//r.Post("/", s.CreateView)
	//
	//r.Route("/{id}", func(in chi.Router) {
	//	in.Use(SportmonksAdminCtx)
	//	in.Get("/", s.ShowView)
	//	in.Put("/", s.EditView)
	//	in.Delete("/", s.DeleteView)
	//})

	return r
}

type SportmonksAdminRequest struct {
	*models.StatsTeam
}
func (t *SportmonksAdminRequest) Bind(r *http.Request) error {
	// post-processing after decode
	t.StatsTeam.NameL10N = l10n.SetAll(t.NameL10NMap)

	return nil
}
type SportmonksAdminResponse struct {
	models.StatsTeam
}
func (t *SportmonksAdminResponse) Render(w http.ResponseWriter, r *http.Request) error {
	log.Printf("SAR.Render 1: t.Name: %v, t.Team.Name: %v, t.StatsTeam.Name: %v - mapped: %v", t.Name, t.Team.Name, t.StatsTeam.Name, t.NameL10N)
	log.Printf("SAR.Render 2: %#v", t.StatsTeam)
	if t.NameL10N == "" {
		t.NameL10NMap = map[string]string{
			"en": t.Name,
		}
	} else {
		t.NameL10NMap = l10n.GetAll(t.NameL10N)
	}
	log.Printf("SAR.Render: t.Name: %v, t.Team.Name: %v, t.StatsTeam.Name: %v - mapped: %v", t.Name, t.Team.Name, t.StatsTeam.Name, t.NameL10N)

	return nil
}
type SportmonksAdminListResponse struct {
	statsTeams []models.StatsTeam
	L10NStatsTeams []SportmonksAdminResponse
}
func SportmonksAdminCtx(next http.Handler) http.Handler {
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

		ctx := context.WithValue(r.Context(), SportmonksAdminResourceContextKey, team)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func GetSportmonksCtx(r *http.Request) *models.Team {
	obj, _ := r.Context().Value(SportmonksAdminResourceContextKey).(*models.Team)
	return obj
}

func (s *SportmonksAdminResource) ListTeamsView(w http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
		tx = db.GetDBFromRequestContext(r)
	)

	pageNr, err := strconv.Atoi(r.URL.Query().Get("_page"))
	if err != nil {
		pageNr = 1
	}
	itemsPerPage, err := strconv.Atoi(r.URL.Query().Get("_perPage"))
	if err != nil {
		itemsPerPage = 30
	}
	//sortDirection := r.URL.Query().Get("_sortDir")
	//sortField := r.URL.Query().Get("_sortField")

	start := itemsPerPage * (pageNr-1)
	end := itemsPerPage * pageNr

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	statsTeamsDB := []models.StatsTeam{}
	statsTeamsDBMap := map[uint]models.StatsTeam{}

	if err := tx.Model(statsTeamsDB).Find(&statsTeamsDB).Error; err != nil {
		// something else went wrong
		render.Render(w, r, ErrRender(err))
		return
	}
	for _, t := range statsTeamsDB {
		statsTeamsDBMap[t.ID] = t
	}


	//league_sp = smApi.GetLeague(leagueID)
	leagues := smApi.ListCompetitions()
	teams := []models.StatsTeam{}

	for _, l := range leagues {
		tmpTeams := smApi.ListTeamsBy(int(l.CurrentSeasonID))

		for _, t := range tmpTeams {
			if t.ID == 0 {
				continue
			}
			//t.Name = l10n.SetAll(map[string]string{
			//	"en": t.Name,
			//})

			st := models.StatsTeam{
				Team: t,
				CombinedLeagueSeasonTeamID: fmt.Sprintf("%v-%v-%v", l.ID, l.CurrentSeasonID, t.ID),
				LeagueID: int(l.ID),
				SeasonID: int(l.CurrentSeasonID),
			}
			if stFromDb, ok := statsTeamsDBMap[t.ID]; ok && stFromDb.ID != 0 {
				st.NameL10NMap = l10n.GetAll(stFromDb.NameL10N)
			} else {
				st.NameL10NMap = map[string]string{"en": t.Name}
			}

			teams = append(teams, st)
		}
	}

	teamLength := len(teams)
	if end > teamLength {
		end = teamLength
	}
	var teamSlice []models.StatsTeam
	if end == start {
		teamSlice = teams
	} else {
		teamSlice = teams[start:end]
	}

	log.Printf("start: %v, end: %v, entries: %v", start, end, teamLength)

	w.Header().Set("X-Total-Count", fmt.Sprintf("%v", teamLength))

	render.Status(r, http.StatusOK)
	render.JSON(w, r, teamSlice)
}

func splitCombinedKey(combinedKey string) (leagueID int, seasonID int, teamID int) {
	idParts := strings.Split(combinedKey, "-")
	if len(idParts) != 3 {
		return
	}
	leagueID, err := strconv.Atoi(idParts[0])
	if err != nil {
		return
	}
	seasonID, err = strconv.Atoi(idParts[1])
	if err != nil {
		return
	}
	teamID, err = strconv.Atoi(idParts[2])
	if err != nil {
		return
	}
	return
}

func (s SportmonksAdminResource) ShowTeamView(w http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
		leagueSeasonTeamId = chi.URLParam(r, "leagueSeasonTeamId")
		tx = db.GetDBFromRequestContext(r)
	)

	leagueID, seasonID, teamID := splitCombinedKey(leagueSeasonTeamId)
	if leagueID == 0 || seasonID == 0 || teamID == 0 {
		render.Render(w, r, ErrNotFound)
		return
	}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}
	var team gosportmonks.Team
	for _, t := range smApi.ListTeamsBy(seasonID) {
		if t.ID == uint(teamID) {
			team = t
			break
		}
	}

	if team.ID == 0 {
		render.Render(w, r, ErrNotFound)
		return
	}
	statsTeam := models.StatsTeam{}

	// TODO: fetch from DB as well
	if err := tx.Model(statsTeam).Where("combined_league_season_team_id = ?", leagueSeasonTeamId).Find(&statsTeam).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// we've not started changing the item in our database
			statsTeam = models.StatsTeam{
				Team: team,
				CombinedLeagueSeasonTeamID: fmt.Sprintf("%v-%v-%v", leagueID, seasonID, teamID),
				LeagueID: leagueID,
				SeasonID: seasonID,
			}
		} else {
			// something else went wrong
			render.Render(w, r, ErrRender(err))
			return
		}
	}

	render.Status(r, http.StatusOK)
	if err := render.Render(w, r, &SportmonksAdminResponse{StatsTeam: statsTeam}); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

func (s SportmonksAdminResource) EditTeamView(w http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
		tx = db.GetDBFromRequestContext(r)
		leagueSeasonTeamId = chi.URLParam(r, "leagueSeasonTeamId")
	)

	leagueID, seasonID, teamID := splitCombinedKey(leagueSeasonTeamId)
	if leagueID == 0 || seasonID == 0 || teamID == 0 {
		render.Render(w, r, ErrNotFound)
		return
	}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}
	var team gosportmonks.Team
	for _, t := range smApi.ListTeamsBy(seasonID) {
		if t.ID == uint(teamID) {
			team = t
			break
		}
	}

	if team.ID == 0 {
		render.Render(w, r, ErrNotFound)
		return
	}
	// do the processing

	data := &SportmonksAdminRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	newTeam := data.StatsTeam
	if err := tx.Save(newTeam).Error; err != nil {
		log.Printf("SAR.EditTeamView failed: %v", err)
		render.Render(w, r, ErrRender(err))
		return
	}

	// return result
	statsTeam := models.StatsTeam{}

	// TODO: fetch from DB as well
	if err := tx.Model(statsTeam).Where("combined_league_season_team_id = ?", leagueSeasonTeamId).Find(&statsTeam).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// we've not started changing the item in our database
			statsTeam = models.StatsTeam{
				Team: team,
				CombinedLeagueSeasonTeamID: fmt.Sprintf("%v-%v-%v", leagueID, seasonID, teamID),
				LeagueID: leagueID,
				SeasonID: seasonID,
			}
		} else {
			// something else went wrong
			render.Render(w, r, ErrRender(err))
			return
		}
	}

	render.Status(r, http.StatusOK)
	if err := render.Render(w, r, &SportmonksAdminResponse{StatsTeam: statsTeam}); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}
