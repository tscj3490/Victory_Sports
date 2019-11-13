package sportmonks_api

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sort"
	"strconv"
	"time"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/db"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/gosportmonks"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/models"
	"github.com/jinzhu/gorm"
	"github.com/patrickmn/go-cache"
	"golang.org/x/net/context"
)

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

type SportmonksAPI struct {
	Ctx                    context.Context
	RebuildCache           bool
	DontTriggerCacheDBSave bool
}

type FixtureAndEvents struct {
	Fixture *gosportmonks.Fixture
	Events  struct {
		LocalTeamEvents   map[string][]gosportmonks.FixtureEvent
		VisitorTeamEvents map[string][]gosportmonks.FixtureEvent
	}
}

type StatsCalendar struct {
	DateParameter     string
	DateParameterTime time.Time
	Entries           []StatsCalendarEntry
	TodayOrNextIdx    int
	Count             int
}
type StatsCalendarEntry struct {
	DateTime    *time.Time
	NextOrToday bool
	Idx         int
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

// GetLeague by leagueID
func (f *SportmonksAPI) GetLeague(leagueID int) *gosportmonks.League {
	league := &gosportmonks.League{}
	var ctx = f.Ctx

	// league options:
	leagueOpt := &gosportmonks.ListOptions{
		Include: CacheKeyLeaguesAndSeasonsInclude,
	}

	var fnCacheKey = fmt.Sprintf(CacheKeyLeaguesAndSeasons, leagueID)
	log.Printf("FnCacheKey: %v", fnCacheKey)
	if !f.RebuildCache {
		if x, ok := db.InMemoryCache.Get(fnCacheKey); ok {
			cachedObj := x.(gosportmonks.League)
			league = &cachedObj
			return league
		}
	}
	log.Printf("%v - nothing found - lets query ... :) \n", fnCacheKey)

	// DO REQUEST
	league, _, err := client.Leagues.GetWithOptions(ctx, leagueID, leagueOpt)
	if err != nil {
		log.Printf("GetLeague - Leagues for loop failed: %v", err)
		return league
	}

	// SET CACHE
	db.InMemoryCache.Set(fnCacheKey, *league, cache.NoExpiration)

	// by default trigger CacheDBSave
	if !f.DontTriggerCacheDBSave {
		db.CacheDBSave(ctx)
	}

	return league
}

// ListSeasons by competitionID
func (f *SportmonksAPI) ListSeasons(competitionID int) []formSelection {
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

		if leagueWithSeasons == nil {
			log.Printf("query failed: leagues with season doesn't exist.")
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

	//	return leagueWithSeasons.SeasonsInclude.Seasons
	return seasons
}

// ListTeams by season
func (f *SportmonksAPI) ListTeamsBy(seasonID int) []gosportmonks.Team {
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
		if len(teams) <= 0 {
			log.Printf("FnCacheKey: team array doesn't exist")
		}
		log.Printf("FnCacheKey: %v Queried Result Count: %v", fnCacheKey, len(teams))
		// SET CACHE
		db.InMemoryCache.Set(fnCacheKey, teams, cache.NoExpiration)
		db.CacheDBSave(ctx)
	}

	return teams
}

// ListFixturesBySeason
func (f *SportmonksAPI) ListFixturesBySeason(seasonID int) []gosportmonks.Fixture {
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

// ListFixturesBySeasonAndDate
func (f *SportmonksAPI) ListFixturesBySeasonAndDate(seasonID int, filterDate time.Time, teamID int) []gosportmonks.Fixture {
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

// ListFixturesBySeasonDateAndTeam
func (f *SportmonksAPI) ListFixturesBySeasonDateAndTeam(seasonID int, filterDate time.Time) []gosportmonks.Fixture {
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

// GetFixture returns a mixed objects, Fixture: the requested match; Events a map of events
func (f *SportmonksAPI) GetFixture(seasonID int, fixtureID int) FixtureAndEvents {
	fixtureAndEvents := FixtureAndEvents{}
	var fixture *gosportmonks.Fixture
	fixtureEventsLocal := map[string][]gosportmonks.FixtureEvent{}
	fixtureEventsVisitor := map[string][]gosportmonks.FixtureEvent{}

	fixtures := f.ListFixturesBySeason(seasonID)
	if len(fixtures) <= 0 {
		log.Printf("StatsFilter.GetFixture - Error: no fixtures returned from ListFixturesBySeason")
		return fixtureAndEvents
	}

	equalFlag := false
	for _, f := range fixtures {
		if f.ID == uint(fixtureID) {
			fixture = &f
			equalFlag = true
			break
		}
	}

	if !equalFlag {
		log.Printf("StatsFilter.GetFixture - Error: fixture ID doesn't exist")
		return fixtureAndEvents
	}

	if len(fixture.EventsInclude.Events) <= 0 {
		log.Printf("StatsFilter.GetFixture - Error: fixture.EventsInclude.Events doesn't exist")
		return fixtureAndEvents
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

// GetTopscorers by seasonId
func (f *SportmonksAPI) GetTopscorers(seasonID int) []gosportmonks.Topscorer {
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

// GetStandings by seasonID
func (f *SportmonksAPI) GetStandings(seasonID int) []gosportmonks.Standing {
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

// GetStatsCalendar
func (f *SportmonksAPI) GetStatsCalendar(seasonID int, dateParam string, teamIDs []uint) StatsCalendar {

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

// ListCompetitions return the list of leagues - 1) Return the list of leagues so the template can render them
func (f *SportmonksAPI) ListCompetitions() []*gosportmonks.League {
	leagues := []*gosportmonks.League{}

	for _, comp := range competitions {
		league := f.GetLeague(comp.ID)
		leagues = append(leagues, league)
	}

	return leagues
}

// GetStatsCalendarAll by dateParam
func (f *SportmonksAPI) GetStatsCalendarAll(dateParam string) StatsCalendar {

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

// GetTeamsGroupByCollectionCode
func (f *SportmonksAPI) GetTeamsGroupByCollectionCode(tx *gorm.DB) map[string][]models.Team {
	teams := []models.Team{}
	if err := tx.Model(teams).Preload("Collections").Find(&teams).Error; err != nil {
		log.Println("Failed to fetch teams: %v", err)
	}
	menu_clubs := map[string][]models.Team{}
	for _, t := range teams {
		for _, c := range t.Collections {
			if _, ok := menu_clubs[c.Code]; !ok {
				menu_clubs[c.Code] = []models.Team{}
			}
			menu_clubs[c.Code] = append(menu_clubs[c.Code], t)
		}
	}
	return menu_clubs
}

// GetBrands
func (f *SportmonksAPI) GetBrands(tx *gorm.DB) []models.Brand {
	brands := []models.Brand{}
	if err := tx.Model(brands).Find(&brands).Error; err != nil {
		log.Printf("WR.AddShopContext cant find brands. %v", err)
	}
	return brands
}

// GetProductSizes
func (f *SportmonksAPI) GetProductSizes(tx *gorm.DB) []models.ProductSize {
	productSizes := []models.ProductSize{}
	if err := tx.Model(productSizes).Find(&productSizes).Error; err != nil {
		log.Printf("WR.AddShopContext cant find productSizes. %v", err)
	}
	return productSizes
}

// GetProductKits
func (f *SportmonksAPI) GetProductKits() []models.ProductAttrs {
	kitList := []models.ProductAttrs{}
	for _, k := range models.Kits {
		kitList = append(kitList, k)
	}
	return kitList
}

// GetProductKits
func (f *SportmonksAPI) GetProductGenders() []models.ProductAttrs {
	kitList := []models.ProductAttrs{}
	for _, k := range models.Genders {
		kitList = append(kitList, k)
	}
	return kitList
}

// GetCollections
func (f *SportmonksAPI) GetCollections(tx *gorm.DB, shopType string) []models.Collection {
	collection := []models.Collection{}
	if shopType == "" {
		if err := tx.Model(collection).Find(&collection).Error; err != nil {
			fmt.Printf("Failed to fetch collection: %v\n", err)
		}
	} else {
		if err := tx.Model(collection).Where("code = ?", shopType).Find(&collection).Error; err != nil {
			fmt.Printf("Failed to fetch collection: %v\n", err)
		}
	}

	return collection
}

// GetProductDetails
func (f *SportmonksAPI) GetProductDetails(tx *gorm.DB, productId int) []models.Product {
	product := []models.Product{}

	if err := tx.Model(product).First(&product, productId).Error; err != nil {
		fmt.Printf("Failed to fetch product: %v\n", err)
	}

	return product
}

// GetProductVariations
func (f *SportmonksAPI) GetProductVariations(tx *gorm.DB, productId int) []models.ProductVariation {
	variations := []models.ProductVariation{}
	if err := tx.Model(variations).Where("product_id = ? AND available_quantity > 0", productId).
		Preload("Size").
		Preload("Badge").
		Find(&variations).Error; err != nil {
		fmt.Printf("failed to query product variations: %v", err)
	}

	return variations
}

// GetProduct
func (f *SportmonksAPI) GetProducts(tx *gorm.DB, collectionId int, collectionCode string, kitCode string, productSort string) ([]models.Product, error) {
	products := []models.Product{}
	collection := models.Collection{}

	if collectionCode != "any" {
		if err := tx.Model(collection).Where("code = ?", collectionCode).First(&collection).Error; err != nil {
			log.Printf("Failed to fetch collection: %v\n", err)
		}
	}

	query := tx.Model(&products).Select("products.*")

	if collection.Name != "" {
		query = query.
			Joins("inner join product_collections as pc on pc.product_id = id").
			Where("pc.collection_id = ?", collection.ID)
	} else {
		fmt.Println("no collection found ... :(")
	}

	if kitCode != "" {
		query = query.Where("kit_code like ?", kitCode)
	}

	if productSort == "ToHigh" {
		query = query.Order("products.price asc")
	} else {
		query = query.Order("products.price desc")
	}

	err := query.Find(&products).Error

	return products, err

}

// GetProfileHistory
func (f *SportmonksAPI) GetProfileHistory(tx *gorm.DB, userId int) ([]models.Order, error) {
	orders := []models.Order{}

	query := tx.Model(orders).Where("user_id = ?", userId).
		Preload("OrderItems.ProductVariation.Product")

	err := query.Find(&orders).Error

	return orders, err
}

// GetProfileView
func (f *SportmonksAPI) GetProfileView(tx *gorm.DB, userId int) (models.Address, error) {
	address := models.Address{}

	query := tx.Model(address).Where("user_id = ?", userId)

	err := query.First(&address).Error

	return address, err
}

// ShowCartContent
func (f *SportmonksAPI) ShowCartContent(tx *gorm.DB, cartItemIDs []uint) ([]models.ProductVariation, error) {
	selectedProductVariations := []models.ProductVariation{}

	query := tx.Model(models.ProductVariation{}).
		Where("id in (?)", cartItemIDs).
		Preload("Product").
		Preload("Size").
		Preload("Badge")

	err := query.Find(&selectedProductVariations).Error

	return selectedProductVariations, err

}
