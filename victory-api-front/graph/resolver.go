//go:generate gorunpkg github.com/99designs/gqlgen

package graph

import (
	context "context"
	"fmt"
	"time"

	"sort"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/sportmonks_api"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/db"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/gosportmonks"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/models"
	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
)

type SportmonksAPI struct {
	Ctx                    context.Context
	RebuildCache           bool
	DontTriggerCacheDBSave bool
}

type SeasonRoot struct {
	Season *Season `json:"data"`
}

type Resolver struct{}

func (r *Resolver) League() LeagueResolver {
	return &leagueResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}
func (r *Resolver) Season() SeasonResolver {
	return &seasonResolver{r}
}

type leagueResolver struct{ *Resolver }

func (r *leagueResolver) Coverage(ctx context.Context, obj *League) (Coverage, error) {
	panic("not implemented")
}
func (r *leagueResolver) SeasonsInclude(ctx context.Context, obj *League, orderBy string, limit int) (*SeasonsRoot, error) {
	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	seasons := []Season{}
	seasons_sp := smApi.ListSeasons(obj.ID)

	/*
		val := t.(Season)
		v := reflect.ValueOf(val)
		fieldVal := reflect.Indirect(v).FieldByName(orderBy)
		switch fieldVal.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			ret = strconv.FormatInt(fieldVal.Int(), 10)
		case reflect.String:
			ret = fieldVal.String()
		}
	*/
	copier.Copy(&seasons, &seasons_sp)

	if orderBy == "Id" {
		sort.Slice(seasons, func(i, j int) bool {
			return seasons[i].ID > seasons[j].ID
		})
	}
	if orderBy == "-Id" {
		sort.Slice(seasons, func(i, j int) bool {
			return seasons[i].ID < seasons[j].ID
		})
	}
	if orderBy == "Name" {
		sort.Slice(seasons, func(i, j int) bool {
			return seasons[i].Name > seasons[j].Name
		})
	}
	if orderBy == "LeagueId" {
		sort.Slice(seasons, func(i, j int) bool {
			return seasons[i].LeagueID > seasons[j].LeagueID
		})
	}
	seasonRoot := &SeasonsRoot{
		Seasons: seasons,
	}

	maxLen := len(seasons)

	if limit > maxLen {
		limit = maxLen - 1
	}
	if maxLen == 0 {
		return seasonRoot, nil
	}
	seasonRoot.Seasons = seasons[:limit]

	return seasonRoot, nil
}

type seasonResolver struct{ *Resolver }

func (r *seasonResolver) StageData(ctx context.Context, obj *Season) ([]Stage, error) {
	panic("not implemented")
}
func (r *seasonResolver) RoundsData(ctx context.Context, obj *Season) ([]Round, error) {
	panic("not implemented")
}
func (r *seasonResolver) FixturesInclude(ctx context.Context, obj *Season, orderBy string, limit int) ([]Fixture, error) {

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}
	fixtures := []Fixture{}

	fixtures_sp := smApi.ListFixturesBySeason(obj.ID)

	copier.Copy(&fixtures, &fixtures_sp)

	if orderBy == "Id" {
		sort.Slice(fixtures, func(i, j int) bool {
			return fixtures[i].ID > fixtures[j].ID
		})
	}
	if orderBy == "-Id" {
		sort.Slice(fixtures, func(i, j int) bool {
			return fixtures[i].ID < fixtures[j].ID
		})
	}
	if orderBy == "StartingAt" {
		sort.Slice(fixtures, func(i, j int) bool {
			return fixtures[i].Time.StartingAt.Timestamp > fixtures[j].Time.StartingAt.Timestamp
		})
	}
	if orderBy == "-StartingAt" {
		sort.Slice(fixtures, func(i, j int) bool {
			return fixtures[i].Time.StartingAt.Timestamp < fixtures[j].Time.StartingAt.Timestamp
		})
	}

	maxLen := len(fixtures)
	if limit > maxLen {
		limit = maxLen - 1
	}
	if maxLen == 0 {
		return fixtures, nil
	}
	fixtures = fixtures[:limit]

	return fixtures, nil
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) GetLeague(ctx context.Context, leagueId int) (*League, error) {
	fmt.Printf("Called GetLeague:ID = %d\n", leagueId)
	league_sp := &gosportmonks.League{}
	league := &League{}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	league_sp = smApi.GetLeague(leagueId)

	league.ID = int(league_sp.ID)
	league.Name = league_sp.Name
	league.CountryID = int(league_sp.CountryID)
	league.LegacyID = int(league_sp.LegacyID)
	league.CurrentSeasonID = int(league_sp.CurrentSeasonID)
	league.CurrentRoundID = int(league_sp.CurrentRoundID)
	league.CurrentStageID = int(league_sp.CurrentStageID)
	league.LiveStandings = league_sp.LiveStandings
	league.IsCup = league_sp.IsCup

	return league, nil
}
func (r *queryResolver) ListSeasons(ctx context.Context, competitionId int) ([]FormSelection, error) {

	seasons := []FormSelection{}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	seasons_sp := smApi.ListSeasons(competitionId)

	copier.Copy(&seasons, &seasons_sp)

	return seasons, nil
}
func (r *queryResolver) ListTeamsBy(ctx context.Context, seasonId int) ([]Team, error) {
	teams := []Team{}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	teams_sp := smApi.ListTeamsBy(seasonId)
	copier.Copy(&teams, &teams_sp)

	return teams, nil
}

func (r *queryResolver) ListFixturesBySeason(ctx context.Context, seasonId int) ([]Fixture, error) {
	fixtures := []Fixture{}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	fixtures_sp := smApi.ListFixturesBySeason(seasonId)
	copier.Copy(&fixtures, &fixtures_sp)

	return fixtures, nil
}

func (r *queryResolver) ListFixturesBySeasonAndDate(ctx context.Context, seasonId int, filterDate time.Time, teamId int) ([]Fixture, error) {
	fixtures := []Fixture{}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	fixtures_sp := smApi.ListFixturesBySeasonAndDate(seasonId, filterDate, teamId)
	copier.Copy(&fixtures, &fixtures_sp)

	return fixtures, nil
}

func (r *queryResolver) ListFixturesBySeasonDateAndTeam(ctx context.Context, seasonId int, filterDate time.Time) ([]Fixture, error) {
	fixtures := []Fixture{}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	fixtures_sp := smApi.ListFixturesBySeasonDateAndTeam(seasonId, filterDate)
	copier.Copy(&fixtures, &fixtures_sp)

	return fixtures, nil
}

func (r *queryResolver) GetTopscorers(ctx context.Context, seasonId int) ([]Topscorer, error) {
	topscorers := []Topscorer{}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	topscorers_sp := smApi.GetTopscorers(seasonId)
	copier.Copy(&topscorers, &topscorers_sp)

	return topscorers, nil
}

func (r *queryResolver) GetStandings(ctx context.Context, seasonId int) ([]Standing, error) {
	standings := []Standing{}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	standings_sp := smApi.GetStandings(seasonId)
	copier.Copy(&standings, &standings_sp)

	return standings, nil
}

func (r *queryResolver) GetStatsCalendar(ctx context.Context, seasonId int, dateParam string, teamIds []int) (StatsCalendar, error) {
	statsCalendar := StatsCalendar{}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	teamIds_uint := []uint{}
	for _, teamId := range teamIds {
		teamId_uint := uint(teamId)
		teamIds_uint = append(teamIds_uint, teamId_uint)
	}
	statsCalendar_sp := smApi.GetStatsCalendar(seasonId, dateParam, teamIds_uint)
	copier.Copy(&statsCalendar, &statsCalendar_sp)

	return statsCalendar, nil
}

func (r *queryResolver) GetStatsCalendarAll(ctx context.Context, dateParam string) (StatsCalendar, error) {
	statsCalendar := StatsCalendar{}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	statsCalendar_sp := smApi.GetStatsCalendarAll(dateParam)
	copier.Copy(&statsCalendar, &statsCalendar_sp)

	return statsCalendar, nil
}

func (r *queryResolver) GetFixture(ctx context.Context, seasonId int, fixtureID int) (FixtureAndEvents, error) {
	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	fixtureAndEvents_sp := smApi.GetFixture(seasonId, fixtureID)

	var localTeamEvents []*FixtureEventArray
	var visitorTeamEvents []*FixtureEventArray
	for key, val := range fixtureAndEvents_sp.Events.LocalTeamEvents {
		value := []FixtureEvent{}
		copier.Copy(&value, &val)
		localTeamEvents = append(localTeamEvents, &FixtureEventArray{
			Key:   key,
			Value: value,
		})
	}
	for key, val := range fixtureAndEvents_sp.Events.VisitorTeamEvents {
		value := []FixtureEvent{}
		copier.Copy(&value, &val)
		visitorTeamEvents = append(visitorTeamEvents, &FixtureEventArray{
			Key:   key,
			Value: value,
		})
	}
	fixture := &Fixture{}
	copier.Copy(fixture, &fixtureAndEvents_sp.Fixture)
	fixtureAndEvents := FixtureAndEvents{
		Fixture: fixture,
		Events: &Events{
			LocalTeamEvents:   localTeamEvents,
			VisitorTeamEvents: visitorTeamEvents,
		},
	}

	return fixtureAndEvents, nil
}

func (r *queryResolver) GetTeamsGroupByCollectionCode(ctx context.Context) ([]*TeamArray, error) {
	var tx *gorm.DB = db.GetDBFromContext(ctx)

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	teamsGroup_sp := smApi.GetTeamsGroupByCollectionCode(tx)

	var teamArray []*TeamArray
	for key, val := range teamsGroup_sp {
		value := []TeamM{}
		copier.Copy(&value, &val)
		teamArray = append(teamArray, &TeamArray{
			Key:   key,
			Value: value,
		})
	}

	return teamArray, nil
}

func (r *queryResolver) GetBrands(ctx context.Context) ([]Brand, error) {
	var tx *gorm.DB = db.GetDBFromContext(ctx)

	brands := []Brand{}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	brands_sp := smApi.GetBrands(tx)
	copier.Copy(&brands, &brands_sp)

	return brands, nil
}

func (r *queryResolver) GetProductSizes(ctx context.Context) ([]ProductSize, error) {
	var tx *gorm.DB = db.GetDBFromContext(ctx)

	productSizes := []ProductSize{}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	productSizes_sp := smApi.GetProductSizes(tx)
	copier.Copy(&productSizes, &productSizes_sp)

	return productSizes, nil
}

func (r *queryResolver) GetProductKits(ctx context.Context) ([]ProductAttrs, error) {
	productAttrs := []ProductAttrs{}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	productAttrs_sp := smApi.GetProductKits()
	copier.Copy(&productAttrs, &productAttrs_sp)

	return productAttrs, nil
}

func (r *queryResolver) GetProductGenders(ctx context.Context) ([]ProductAttrs, error) {
	productAttrs := []ProductAttrs{}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	productAttrs_sp := smApi.GetProductGenders()
	copier.Copy(&productAttrs, &productAttrs_sp)

	return productAttrs, nil
}

func (r *queryResolver) GetCollections(ctx context.Context, shopType string) ([]CollectionM, error) {
	var tx *gorm.DB = db.GetDBFromContext(ctx)

	collections := []CollectionM{}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	collections_sp := smApi.GetCollections(tx, shopType)
	copier.Copy(&collections, &collections_sp)

	return collections, nil
}

func (r *queryResolver) GetProductDetails(ctx context.Context, productId int) ([]Product, error) {
	var tx *gorm.DB = db.GetDBFromContext(ctx)

	products := []Product{}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	products_sp := smApi.GetProductDetails(tx, productId)
	copier.Copy(&products, &products_sp)

	return products, nil
}

func (r *queryResolver) GetProductVariations(ctx context.Context, productId int) ([]ProductVariation, error) {
	var tx *gorm.DB = db.GetDBFromContext(ctx)

	variations := []ProductVariation{}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	variations_sp := smApi.GetProductVariations(tx, productId)
	copier.Copy(&variations, &variations_sp)

	return variations, nil
}

func (r *queryResolver) GetProducts(ctx context.Context, collectionId int, collectionCode string, kitCode string, productSort string) ([]Product, error) {
	var tx *gorm.DB = db.GetDBFromContext(ctx)

	products := []Product{}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	products_sp, _ := smApi.GetProducts(tx, collectionId, collectionCode, kitCode, productSort)

	copier.Copy(&products, &products_sp)

	return products, nil
}

func (r *queryResolver) GetProfileHistory(ctx context.Context) ([]Order, error) {
	user, _ := ctx.Value("user").(*models.User)
	var tx *gorm.DB = db.GetDBFromContext(ctx)

	orders := []Order{}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	orders_sp, _ := smApi.GetProfileHistory(tx, int(user.ID))

	copier.Copy(&orders, &orders_sp)

	return orders, nil
}

func (r *queryResolver) GetProfileView(ctx context.Context) (Address, error) {
	user, _ := ctx.Value("user").(*models.User)
	var tx *gorm.DB = db.GetDBFromContext(ctx)

	address := Address{}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}

	address_sp, _ := smApi.GetProfileView(tx, int(user.ID))

	copier.Copy(&address, &address_sp)

	return address, nil
}

func (r *queryResolver) ShowCartContent(ctx context.Context) ([]ProductVariation, error) {
	CartItemIDs, _ := ctx.Value("Graph_Cart").([]uint)

	selectedProductVariations := []ProductVariation{}

	smApi := &sportmonks_api.SportmonksAPI{
		Ctx: ctx,
		DontTriggerCacheDBSave: true,
	}
	var tx *gorm.DB = db.GetDBFromContext(ctx)

	selectedProductVariations_sp, _ := smApi.ShowCartContent(tx, CartItemIDs)

	copier.Copy(&selectedProductVariations, &selectedProductVariations_sp)

	return selectedProductVariations, nil
}
