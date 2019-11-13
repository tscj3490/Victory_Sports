package filters

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"sort"

	"runtime"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/insights"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/sportmonks_api"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/db"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/gosportmonks"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/models"
	"github.com/flosch/pongo2"
	cache "github.com/patrickmn/go-cache"
	"github.com/robfig/cron"
)

/*
controllers.filters provides data to the template.
The first usecase is for stats, we will be querying the gosportmonks api directly and caching the result.
This is the file that does that, it exposes a map[string]interface{} that can be used to in the template.
*/

const (
	CronSpec                           = "@every 10m"
	CacheKeyLeaguesAndSeasons          = "StatsFilter_LeaguesAndSeasons_%v"
	CacheKeyLeaguesAndSeasonsInclude   = "country,seasons:limit(3|1):order(id|desc)"
	CacheKeyListFixtureBySeason        = "StatsFilter_ListFixtureBySeason_%v"
	CacheKeyListFixtureBySeasonInclude = "fixtures.localTeam,fixtures.visitorTeam,fixtures.events,fixtures.lineup"
	CacheKeyGetSquad                   = "StatsFilter_GetSquadBySeason_%v_Team_%v"
	CacheKeyGetStandingsBySeason       = "StatsFilter_GetStandingsBySeason_%v"
	CacheKeyGetTopscorerBySeason       = "StatsFilter_GetTopscorerBySeason_%v"
	CacheKeyListTeams                  = "StatsFilter_ListTeams_%v"
)

var (
	netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second}
	netClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}
	client = gosportmonks.NewClient(netClient)
)

var (
	Cron *cron.Cron
)

func init() {
	Cron = cron.New()
	log.Printf("cronjobs.init(%v)", CronSpec)
	ctx := context.Background()
	backgroundJob := BackgroundJob{
		Ctx: ctx,
		StatsFilter: StatsFilter{
			Ctx: ctx,
		},
	}
	Cron.AddFunc(CronSpec, func() {
		backgroundJob.UpdateListFixturesBySeason()
	})
	Cron.Start()
}

type formSelection struct {
	Name string
	ID   int
	Href string
}

type competition struct {
	Name string
	ID   int
}

var competitions = []competition{
	{"Worldcup2018", 732},
	{"UAE - UAE League", 959},
	{"UAE - Division 1", 962},
	{"UAE - Arabian Gulf Cup", 965},
	{"KSA - Pro League", 944},
	{"KSA - Division 1", 947},
	{"KSA - Kings Cup", 950},
	{"ESP - La Liga", 564},
	{"ITA - Serie A", 384},
	{"FRA - Ligue 1", 301},
	{"GBR - Premier League", 8},
}

// StatsFilter Holds all StatsFilter function
type StatsFilter struct {
	Ctx          context.Context
	RebuildCache bool
}

// ListCompetitions return the list of leagues - 1) Return the list of leagues so the template can render them
func (f *StatsFilter) ListCompetitions() []*gosportmonks.League {
	leagues := []*gosportmonks.League{}
	var ctx = f.Ctx

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx:                    ctx,
		RebuildCache:           f.RebuildCache,
		DontTriggerCacheDBSave: true,
	}

	for _, comp := range competitions {
		// Don't append Arabic Gulf League, La Liga, or Premier League
		if comp.ID != 959 && comp.ID != 564 && comp.ID != 8 {
			league := smApi.GetLeague(comp.ID)
			leagues = append(leagues, league)
		}
	}

	// Append La Liga, Arabic Gulf League, and Premier League. Premier League will be first in list.
	leagues = append([]*gosportmonks.League{smApi.GetLeague(564)}, leagues...)
	leagues = append([]*gosportmonks.League{smApi.GetLeague(959)}, leagues...)
	leagues = append([]*gosportmonks.League{smApi.GetLeague(8)}, leagues...)

	db.CacheDBSave(ctx)

	return leagues
}

// ListCompetitionsSimple returns the simple list based on the hardcoded values
func (f *StatsFilter) ListCompetitionsSimple() []competition {
	return competitions
}

// GetCompetition return a competition
func (f *StatsFilter) GetCompetition(competitionID int) *gosportmonks.League {
	var ctx = f.Ctx
	var league *gosportmonks.League
	// league options:
	leagueOpt := &gosportmonks.ListOptions{
		Include: CacheKeyLeaguesAndSeasonsInclude,
	}

	var fnCacheKey = fmt.Sprintf(CacheKeyLeaguesAndSeasons, competitionID)
	log.Printf("FnCacheKey: %v", fnCacheKey)
	if x, ok := db.InMemoryCache.Get(fnCacheKey); ok {
		l := x.(gosportmonks.League)
		league = &l
		return league
	}
	log.Printf("%v - nothing found - lets query ... :) \n", fnCacheKey)

	// DO REQUEST
	var err error
	league, _, err = client.Leagues.GetWithOptions(ctx, competitionID, leagueOpt)
	if err != nil {
		log.Printf("GetCompetition - League Get failed: %v %v %v", competitionID, leagueOpt, err)
		return league
	}

	// SET CACHE
	db.InMemoryCache.Set(fnCacheKey, *league, cache.NoExpiration)
	db.CacheDBSave(ctx)

	return league
}

// ListSeasons 2) Based on the league, return the current season + the two seasons before
func (f *StatsFilter) ListSeasons(competitionID int) []formSelection {
	log.Printf("filters.StatsFilters.ListSeasons(competitionID: %v)", competitionID)
	seasons := []formSelection{}
	var ctx = f.Ctx

	// league options:
	leagueOpt := &gosportmonks.ListOptions{
		Include: CacheKeyLeaguesAndSeasonsInclude,
	}

	if competitionID == 0 {
		return seasons
	}

	var leagueWithSeasons *gosportmonks.League

	var fnCacheKey = fmt.Sprintf(CacheKeyLeaguesAndSeasons, competitionID)

	log.Printf("FnCacheKey: %v", fnCacheKey)
	// GET FROM CACHE
	if x, ok := db.InMemoryCache.Get(fnCacheKey); ok {
		league := x.(gosportmonks.League)
		leagueWithSeasons = &league
	} else {
		// QUERY
		leagueWithSeasons, _, err := client.Leagues.GetWithOptions(ctx, competitionID, leagueOpt)
		if err != nil {
			log.Printf("query failed: %v", err)
			return seasons
		}

		// SET CACHE
		db.InMemoryCache.Set(fnCacheKey, *leagueWithSeasons, cache.NoExpiration)
		db.CacheDBSave(ctx)
	}

	for _, s := range leagueWithSeasons.SeasonsInclude.Seasons {
		log.Printf("Season: %v", s)
		seasons = append(seasons, formSelection{
			s.Name,
			int(s.ID),
			fmt.Sprintf("/stats/league/%v/season/%v/", competitionID, s.ID),
		})
	}

	return seasons
}

// ListTeams by season
func (f *StatsFilter) ListTeamsBy(seasonID int) []gosportmonks.Team {
	teams := []gosportmonks.Team{}
	var ctx = f.Ctx
	var err error

	// team options:
	teamOpt := &gosportmonks.ListOptions{
		Include: "order(name|asc)",
	}

	var fnCacheKey = fmt.Sprintf(CacheKeyListTeams, seasonID)
	log.Printf("FnCacheKey: %v", fnCacheKey)
	// GET FROM CACHE
	if x, ok := db.InMemoryCache.Get(fnCacheKey); ok {
		teamsCached := x.([]gosportmonks.Team)
		teams = teamsCached
		log.Printf("FnCacheKey: %v Cached Result Count: %v", fnCacheKey, len(teams))
	} else {
		// QUERY

		teams, _, err = client.Teams.List(ctx, seasonID, teamOpt)
		if err != nil {
			log.Printf("query failed: %v", err)
			return teams
		}
		log.Printf("FnCacheKey: %v Queried Result Count: %v", fnCacheKey, len(teams))
		// SET CACHE
		db.InMemoryCache.Set(fnCacheKey, teams, cache.NoExpiration)
		db.CacheDBSave(ctx)
	}

	return teams
}

// ListTeams 3) Based on the League selected we need to pull all the clubs in it
func (f *StatsFilter) ListTeams(competitionID int, seasonID int) []formSelection {
	log.Printf("filters.StatsFilters.ListTeams(competitionID: %v, seasonID: %v)", competitionID, seasonID)

	// the most feasible way to get a list of teams per league is to fetch the
	// current __seasonID__ and query the teams endpoint directly.

	teams := []formSelection{}
	teamsData := []gosportmonks.Team{}

	if competitionID == 0 || seasonID == 0 {
		return teams
	}

	teamsData = f.ListTeamsBy(seasonID)

	for _, t := range teamsData {
		teams = append(teams, formSelection{
			ID:   int(t.ID),
			Name: t.Name,
			Href: fmt.Sprintf("/stats/league/%v/season/%v/team/%v/", competitionID, seasonID, t.ID),
		})
	}

	return teams
}

// ListFixturesBySeason
func (f *StatsFilter) ListFixturesBySeason(seasonID int) []gosportmonks.Fixture {
	log.Printf("filters.StatsFilters.ListFixturesBySeason(seasonID: %v)", seasonID)

	ctx := f.Ctx
	fixtures := []gosportmonks.Fixture{}
	seasonOpt := &gosportmonks.ListOptions{
		Include: CacheKeyListFixtureBySeasonInclude,
	}

	// CHECK CACHE
	var fnCacheKey = fmt.Sprintf(CacheKeyListFixtureBySeason, seasonID)
	log.Printf("FnCacheKey: %v", fnCacheKey)
	if !f.RebuildCache {
		if x, ok := db.InMemoryCache.Get(fnCacheKey); ok {
			obj := x.([]gosportmonks.Fixture)

			for _, f := range obj {
				fixtures = append(fixtures, f)
			}
			return fixtures
		}
	}

	log.Printf("%v - nothing found - lets query ... :) \n", fnCacheKey)

	seasonData, resp, err := client.Seasons.Get(ctx, uint(seasonID), seasonOpt)
	if err != nil {
		log.Printf("filters.ListFixturesBySeason failed: %v", err)
		return fixtures
	}
	if seasonData == nil {
		log.Printf("filters.ListFixturesBySeason failed: seasonData returned was empty or not serialzable")
		log.Printf("filters.ListFixturesBySeasons: Resp: %v", resp)
		return fixtures
	}

	for _, f := range seasonData.FixturesInclude.Fixtures {
		fixtures = append(fixtures, f)
	}

	// SET CACHE
	db.InMemoryCache.Set(fnCacheKey, fixtures, cache.NoExpiration)
	db.CacheDBSave(ctx)

	return fixtures
}

func (f *StatsFilter) ListFixturesBySeasonAndDate(seasonID int, filterDate time.Time, teamID int) []gosportmonks.Fixture {
	fixtures := []gosportmonks.Fixture{}
	filterByTeam := teamID > 0

	// ensure that it's the day we filter on
	filterDay := time.Date(filterDate.Year(), filterDate.Month(), filterDate.Day(), 0, 0, 0, 0, filterDate.Location())

	for _, f := range f.ListFixturesBySeason(seasonID) {
		tmpTime := f.Time.GetStartTime()
		fixtureDay := time.Date(tmpTime.Year(), tmpTime.Month(), tmpTime.Day(), 0, 0, 0, 0, tmpTime.Location())
		if filterByTeam {
			ok := f.VisitorTeamID == uint(teamID)
			ok2 := f.LocalTeamID == uint(teamID)
			if !ok && !ok2 {
				continue
			}
		}
		if fixtureDay.Equal(filterDay) {
			fixtures = append(fixtures, f)
		}
	}

	return fixtures
}

func (f *StatsFilter) ListFixturesBySeasonDateAndTeam(seasonID int, filterDate time.Time) []gosportmonks.Fixture {
	fixtures := []gosportmonks.Fixture{}

	// ensure that it's the day we filter on
	filterDay := time.Date(filterDate.Year(), filterDate.Month(), filterDate.Day(), 0, 0, 0, 0, filterDate.Location())

	for _, f := range f.ListFixturesBySeason(seasonID) {
		tmpTime := f.Time.GetStartTime()
		fixtureDay := time.Date(tmpTime.Year(), tmpTime.Month(), tmpTime.Day(), 0, 0, 0, 0, tmpTime.Location())
		if fixtureDay.Equal(filterDay) {
			fixtures = append(fixtures, f)
		}
	}

	return fixtures
}

type FixtureAndEvents struct {
	Fixture *gosportmonks.Fixture
	Events  struct {
		LocalTeamEvents   map[string][]gosportmonks.FixtureEvent
		VisitorTeamEvents map[string][]gosportmonks.FixtureEvent
	}
}

// GetFixture returns a mixed objects, Fixture: the requested match; Events a map of events
func (f *StatsFilter) GetFixture(seasonID int, fixtureID int) FixtureAndEvents {
	fixtureAndEvents := FixtureAndEvents{}
	var fixture *gosportmonks.Fixture
	fixtureEventsLocal := map[string][]gosportmonks.FixtureEvent{}
	fixtureEventsVisitor := map[string][]gosportmonks.FixtureEvent{}

	fixtures := f.ListFixturesBySeason(seasonID)
	if len(fixtures) <= 0 {
		log.Printf("StatsFilter.GetFixture - Error: no fixtures returned from ListFixturesBySeason")
		return fixtureAndEvents
	}

	for _, f := range fixtures {
		if f.ID == uint(fixtureID) {
			fixture = &f
			break
		}
	}

	for _, e := range fixture.EventsInclude.Events {
		teamID, err := strconv.Atoi(e.TeamID)
		if err != nil {
			log.Printf("StatsFilter.GetFixture - Events loop failed: %v", err)
			continue
		}
		fixtureEvents := fixtureEventsVisitor
		if teamID == int(fixture.LocalTeamID) {
			fixtureEvents = fixtureEventsLocal
		}
		eventType := e.Type
		tmpEventsType := fixtureEvents[eventType]
		log.Printf("StatsFilter.GetFixture.Event TYpe: %v", eventType)

		tmpEventsType = append(tmpEventsType, e)

		fixtureEvents[eventType] = tmpEventsType
	}

	fixtureAndEvents.Fixture = fixture
	fixtureAndEvents.Events.LocalTeamEvents = fixtureEventsLocal
	fixtureAndEvents.Events.VisitorTeamEvents = fixtureEventsVisitor

	log.Printf("StatsFilter.GetFixture - Events LocalTeam: %v", fixtureEventsLocal)

	return fixtureAndEvents
}

func (f *StatsFilter) GetTopscorers(seasonID int) []gosportmonks.Topscorer {
	topscorers := []gosportmonks.Topscorer{}

	ctx := f.Ctx

	// CHECK CACHE
	var fnCacheKey = fmt.Sprintf(CacheKeyGetTopscorerBySeason, seasonID)
	log.Printf("FnCacheKey: %v", fnCacheKey)
	if x, ok := db.InMemoryCache.Get(fnCacheKey); ok {
		obj := x.([]gosportmonks.Topscorer)

		for _, f := range obj {
			topscorers = append(topscorers, f)
		}
		return topscorers
	}

	log.Printf("%v - nothing found - lets query ... :) \n", fnCacheKey)

	topscorersData, resp, err := client.Topscorers.List(ctx, seasonID, nil)
	if err != nil {
		log.Printf("filters.GetTopscorerBySeason failed: %v", err)
		return topscorers
	}
	if topscorersData == nil {
		log.Printf("filters.GetTopscorerBySeason failed: seasonData returned was empty or not serialzable")
		log.Printf("filters.GetTopscorerBySeasons: Resp: %v", resp)
		return topscorers
	}

	topscorers = topscorersData

	// SET CACHE
	db.InMemoryCache.Set(fnCacheKey, topscorers, cache.NoExpiration)
	db.CacheDBSave(ctx)

	return topscorers
}

func (f *StatsFilter) GetStandings(seasonID int) []gosportmonks.Standing {
	standings := []gosportmonks.Standing{}
	ctx := f.Ctx

	// CHECK CACHE
	var fnCacheKey = fmt.Sprintf(CacheKeyGetStandingsBySeason, seasonID)
	log.Printf("FnCacheKey: %v", fnCacheKey)
	if x, ok := db.InMemoryCache.Get(fnCacheKey); ok {
		obj := x.([]gosportmonks.Standing)

		for _, f := range obj {
			standings = append(standings, f)
		}
		return standings
	}

	log.Printf("%v - nothing found - lets query ... :) \n", fnCacheKey)

	standingsData, resp, err := client.Standings.List(ctx, uint(seasonID), nil)
	if err != nil {
		log.Printf("filters.GetTopscorerBySeason failed: %v", err)
		return standings
	}
	if standingsData == nil {
		log.Printf("filters.GetTopscorerBySeason failed: seasonData returned was empty or not serialzable")
		log.Printf("filters.GetTopscorerBySeasons: Resp: %v", resp)
		return standings
	}

	standings = standingsData

	// SET CACHE
	db.InMemoryCache.Set(fnCacheKey, standings, cache.NoExpiration)
	db.CacheDBSave(ctx)

	return standings
}

type StatsCalendar struct {
	DateParameter     string
	DateParameterTime time.Time
	Entries           []StatsCalendarEntry
	TodayOrNextIdx    int
	Count             int
}
type StatsCalendarEntry struct {
	DateTime        *time.Time
	NextOrToday     bool
	Idx             int
	HasMatchesToday bool
}

func (c StatsCalendarEntry) ToString() string {
	today := time.Now().Format("Mon, Jan 2")
	dateString := c.DateTime.Format("Mon, Jan 2")
	if today == dateString {
		return "Today"
	}
	return dateString
}
func (c StatsCalendarEntry) GetHrefParameter() string {
	return c.DateTime.Format("2006-01-02")
}

func (f *StatsFilter) GetStatsCalendar(seasonID int, dateParam string, teamIDs []uint) StatsCalendar {

	fixtures := f.ListFixturesBySeason(seasonID)
	statsCalendarEntries := []StatsCalendarEntry{}
	gotTeamIDsToFilter := len(teamIDs) > 0
	teamIDsLookupMap := map[uint]bool{}
	for _, tID := range teamIDs {
		teamIDsLookupMap[tID] = true
	}

	sort.Slice(fixtures, func(i, j int) bool {
		return fixtures[i].Time.GetStartTime().Before(*fixtures[j].Time.GetStartTime())
	})
	today := time.Now()
	var tmpTime = ""
	switchedToDatesAfterToday := false
	nextOrTodayIndex := 0
	idx := 0
	for _, c := range fixtures {
		newTime := c.Time.GetStartTime()
		if tmpTime == c.Time.GetStartTime().Format("Mon, Jan 2") {
			continue
		}
		if gotTeamIDsToFilter {
			_, ok := teamIDsLookupMap[c.VisitorTeamID]
			_, ok2 := teamIDsLookupMap[c.LocalTeamID]
			if !ok && !ok2 {
				continue
			}
		}
		tmpTime = newTime.Format("Mon, Jan 2")
		entry := StatsCalendarEntry{DateTime: newTime, Idx: idx}
		if !switchedToDatesAfterToday {
			found := false
			if today.Format("Mon, Jan 2") == newTime.Format("Mon, Jan 2") {
				// found today
				switchedToDatesAfterToday = true
				entry.NextOrToday = true
				found = true
			} else {
				// not today
				if newTime.After(today) {
					// but a day after today
					switchedToDatesAfterToday = true
					entry.NextOrToday = true
					found = true
				}
			}
			if found {
				// remember where NextOrToday is located
				nextOrTodayIndex = idx
				if dateParam == "" {
					dateParam = entry.GetHrefParameter()
				}
			}
		}

		statsCalendarEntries = append(statsCalendarEntries, entry)
		idx += 1
	}

	// in case we still haven't gound a dateParam ... the matches are all
	// in the past ... ok, lets take a random one
	if dateParam == "" {
		dateParam = statsCalendarEntries[idx-1].GetHrefParameter()
		nextOrTodayIndex = idx - 1
	}
	if nextOrTodayIndex == 0 {
		// we are still in the past
		// lets find the idx
		for i, c := range statsCalendarEntries {
			if c.GetHrefParameter() == dateParam {
				nextOrTodayIndex = i
			}
		}
	}
	log.Printf("StatsFilters.GetCalendar - DateParam: %v nextOrTodayIdx: %v", dateParam, nextOrTodayIndex)

	dateParamTime, err := time.ParseInLocation("2006-01-02", dateParam, gosportmonks.LocDubai)
	if err != nil {
		dateParamTime = time.Now()
	}

	return StatsCalendar{
		DateParameterTime: dateParamTime,
		DateParameter:     dateParam,
		Entries:           statsCalendarEntries,
		TodayOrNextIdx:    nextOrTodayIndex,
		Count:             len(statsCalendarEntries),
	}
}

func (f *StatsFilter) GetStatsCalendarAll(dateParam string) StatsCalendar {

	fixtures := []gosportmonks.Fixture{}

	competitions := f.ListCompetitions()
	for _, c := range competitions {
		if c == nil {
			continue
		}
		seasonID := int(c.CurrentSeasonID)
		if seasonID == 0 {
			continue
		}
		fixtures = append(fixtures, f.ListFixturesBySeason(seasonID)...)
	}

	statsCalendarEntries := []StatsCalendarEntry{}

	sort.Slice(fixtures, func(i, j int) bool {
		return fixtures[i].Time.GetStartTime().Before(*fixtures[j].Time.GetStartTime())
	})
	today := time.Now()
	var tmpTime = ""
	switchedToDatesAfterToday := false
	nextOrTodayIndex := 0
	idx := 0
	for _, c := range fixtures {
		newTime := c.Time.GetStartTime()
		if tmpTime == c.Time.GetStartTime().Format("Mon, Jan 2") {
			continue
		}
		tmpTime = newTime.Format("Mon, Jan 2")
		entry := StatsCalendarEntry{DateTime: newTime, Idx: idx}
		if !switchedToDatesAfterToday {
			if today.Format("Mon, Jan 2") == newTime.Format("Mon, Jan 2") {
				// found today
				switchedToDatesAfterToday = true
				entry.NextOrToday = true
				entry.HasMatchesToday = true

				nextOrTodayIndex = idx
				if dateParam == "" {
					dateParam = entry.GetHrefParameter()
				}
			} else {
				// not today
				if newTime.After(today) {

					// but a day after today
					switchedToDatesAfterToday = true
					entry.DateTime = &today
					entry.NextOrToday = true
					// we slot in the today entry and move the idx one forward
					nextOrTodayIndex = idx

					statsCalendarEntries = append(statsCalendarEntries, entry)
					idx += 1

					if dateParam == "" {
						dateParam = entry.GetHrefParameter()
					}

					entry = StatsCalendarEntry{DateTime: newTime, Idx: idx}
				}
			}
		}

		statsCalendarEntries = append(statsCalendarEntries, entry)
		idx += 1
	}

	// in case we still haven't gound a dateParam ... the matches are all
	// in the past ... ok, lets take a random one
	if dateParam == "" {
		dateParam = statsCalendarEntries[idx-1].GetHrefParameter()
		nextOrTodayIndex = idx - 1
	}
	if nextOrTodayIndex == 0 {
		// we are still in the past
		// lets find the idx
		for i, c := range statsCalendarEntries {
			if c.GetHrefParameter() == dateParam {
				nextOrTodayIndex = i
			}
		}
	}
	log.Printf("StatsFilters.GetCalendarAll - DateParam: %v nextOrTodayIdx: %v", dateParam, nextOrTodayIndex)

	dateParamTime, err := time.ParseInLocation("2006-01-02", dateParam, gosportmonks.LocDubai)
	if err != nil {
		dateParamTime = time.Now()
	}

	return StatsCalendar{
		DateParameterTime: dateParamTime,
		DateParameter:     dateParam,
		Entries:           statsCalendarEntries,
		TodayOrNextIdx:    nextOrTodayIndex,
		Count:             len(statsCalendarEntries),
	}
}

// GetSquad - return the squad in that season
func (f *StatsFilter) GetSquad(seasonID int, teamID int) []gosportmonks.PlayerSquadStats {
	ctx := f.Ctx

	squad := []gosportmonks.PlayerSquadStats{}

	// CHECK CACHE
	var fnCacheKey = fmt.Sprintf(CacheKeyGetSquad, seasonID, teamID)
	log.Printf("FnCacheKey: %v", fnCacheKey)
	if x, ok := db.InMemoryCache.Get(fnCacheKey); ok {
		obj := x.([]gosportmonks.PlayerSquadStats)

		for _, f := range obj {
			squad = append(squad, f)
		}
		return squad
	}

	log.Printf("%v - nothing found - lets query ... :) \n", fnCacheKey)

	squadData, _, err := client.Players.List(ctx, seasonID, teamID, nil)
	if err != nil {
		log.Printf("filters.GetSquad failed: %v", err)
		return squad
	}
	squad = squadData

	// SET CACHE
	db.InMemoryCache.Set(fnCacheKey, squad, cache.NoExpiration)
	db.CacheDBSave(ctx)

	return squad
}

// GetSquadStatsMap
func (f *StatsFilter) GetSquadIntMap(seasonID int, teamID uint) map[int]gosportmonks.PlayerSquadStats {
	squad := f.GetSquad(seasonID, int(teamID))
	squadMap := map[int]gosportmonks.PlayerSquadStats{}

	for _, s := range squad {
		squadMap[int(s.PlayerID)] = s
	}

	return squadMap
}

func (f *StatsFilter) GetTeamBuyURL(statsTeamID uint) string {
	resp := ""
	if statsTeamID <= 0 {
		return resp
	}

	shopTeam := models.Team{}

	ctx := f.Ctx
	tx := db.GetDBFromContext(ctx)

	if err := tx.Model(shopTeam).Where("stats_team_id = ?", statsTeamID).First(&shopTeam).Error; err != nil {
		return resp
	}
	resp = shopTeam.GetShopFilterURL()

	return resp
}

func (f *StatsFilter) ConvertGosportmonksTeamToStatsTeam(team gosportmonks.Team) models.StatsTeam {

	ctx := f.Ctx
	tx := db.GetDBFromContext(ctx)
	resp := models.StatsTeam{}

	if err := tx.Model(resp).Where("id = ?", team.ID).First(&resp).Error; err != nil {
		resp.FromGosportmonksTeam(team)
		return resp
	}

	return resp
}

func FilterGetFixturePlayerByKeyInt(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	m := param.Interface().(map[int]gosportmonks.FixturePlayer)
	return pongo2.AsValue(m[in.Integer()]), nil
}
func FilterGetPlayerSquadStatsByKeyInt(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	m := param.Interface().(map[int]gosportmonks.PlayerSquadStats)
	return pongo2.AsValue(m[in.Integer()]), nil
}
func FilterGetPlayerSquadStatsByKeyFixturePlayer(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	fp := in.Interface().(gosportmonks.FixturePlayer)
	m := param.Interface().(map[int]gosportmonks.PlayerSquadStats)
	return pongo2.AsValue(m[int(fp.PlayerID)]), nil
}

// TODO: move this out into its own folder
type BackgroundJob struct {
	StatsFilter StatsFilter
	Ctx         context.Context
}

func (j *BackgroundJob) UpdateListFixturesBySeason() {

	defer func() {
		if rvr := recover(); rvr != nil {

			insights.GenericRecoverer(rvr)

			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("cron: panic running job: %v\n%s", rvr, buf)
		}
	}()

	log.Printf("filters.BackgroundJob.UpdateListFixturesBySeason")
	// find out which seasons need to be refresh

	CacheKeyRefreshAllCompetitions := "background_refreshAllCompetitions"
	var fnCacheKey = CacheKeyRefreshAllCompetitions
	log.Printf("filters.BJ.UpdateListFixturesBySeason FnCacheKey: %v", fnCacheKey)
	if _, ok := db.InMemoryCache.Get(fnCacheKey); ok {
		// ok, we found an cache item, dont query
		log.Printf("filters.BJ.UpdateListFixturesBySeason found the cache item no refresh needed")
		return
	}
	log.Printf("filters.BJ.UpdateListFixturesBySeason cache item not found - refreshing")
	// SET CACHE
	db.InMemoryCache.Set(fnCacheKey, true, 1*time.Hour)
	db.CacheDBSave(j.Ctx)

	j.StatsFilter.RebuildCache = true

	comps := j.StatsFilter.ListCompetitions()
	seasons := []uint{}
	for _, c := range comps {
		seasons = append(seasons, c.CurrentSeasonID)
	}

	// trigger the individual API query
	//j.StatsFilter.RebuildCache = true
	fixtures := map[uint][]gosportmonks.Fixture{}
	for i, s := range seasons {
		fixtures[s] = j.StatsFilter.ListFixturesBySeason(int(s))
		log.Printf("filters.BJ.UpdateListFixturesBySeason sleeping for 10 sec on item: %v of %v", i, len(seasons))
		time.Sleep(10 * time.Second)
	}
	//j.StatsFilter.RebuildCache = false
	// find which game is soon
	//for _, s := range seasons {
	//	for _, f := range fixtures[s] {
	//		inTwoHours := time.Now().Add(120 * time.Minute)
	//		twoHoursAgo := time.Now().Add((-1)*120 * time.Minute)
	//		gameStart := f.Time.GetStartTime()
	//		if gameStart.After(twoHoursAgo) && gameStart.Before(inTwoHours) {
	//			log.Printf("filters.BJ.UpdateListFixturesBySeason match soon: %v %v",
	//				f.LocalTeamInclude.LocalTeam,
	//				f.VisitorTeamInclude.VisitorTeam)
	//		}
	//	}
	//}
}
