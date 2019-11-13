package controllers

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"strings"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/config"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/cart"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/i18n"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/i18n/l10n"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/insights"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/controllers/filters"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/db"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/graph"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/authentic"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/models"
	"github.com/99designs/gqlgen/handler"
	"github.com/alexedwards/scs"
	"github.com/flosch/pongo2"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/googollee/go-socket.io"
	"github.com/qor/i18n/inline_edit"
)

// Homepage and all shared items

// TplContextWebResourceContextKey global html template context
const TplContextWebResourceContextKey = "pongoCtx"

// ShopContextWebResourceContextKey global html template context for shop
const ShopContextWebResourceContextKey = "shopCtx"

// StatsContextWebResourceContextKey global html template context for stats
const StatsContextWebResourceContextKey = "statsCtx"

// KShopType context variable
const KShopType = "shopType"

// KItemID context variable
const KItemID = "itemID"

// KStatsID context variable
const KStatsID = "statsID"

// KStatsType context variable
const KStatsType = "statsType"

// KStatsLeagueID context variable
const KStatsLeagueID = "statsLeagueID"

// KStatsSeasonID context variable
const KStatsSeasonID = "statsSeasonID"

// KStatsTeamID context variable
const KStatsTeamID = "statsTeamID"

// KStatsFixtureID context variable
const KStatsFixtureID = "statsFixtureID"

// WebResource is the main Resource struct that all views start from
type WebResource struct{}

var authResource *authentic.AuthResource

// FilterGetValueByKeyString a template helper function
func FilterGetValueByKeyString(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	m := param.Interface().(map[string]interface{})
	return pongo2.AsValue(m[in.String()]), nil
}

func init() {
	pongo2.RegisterFilter("key_string", FilterGetValueByKeyString)
	pongo2.RegisterFilter("fixture_player_from_int_map", filters.FilterGetFixturePlayerByKeyInt)
	pongo2.RegisterFilter("player_from_int_map", filters.FilterGetPlayerSquadStatsByKeyFixturePlayer)
}

// Routes the router hook
func (web WebResource) Routes() chi.Router {
	r := chi.NewRouter()
	// Add language
	r.Use(i18n.Middleware)

	authResource = authentic.AuthResource{}.NewAuthResource(
		"/user/auth/",
		"user",
		"/",
		scs.NewCookieManager(config.Config.SessionManagerKey))

	r.Use(authResource.SessionContext(false, UserResource{}.GetUser))

	// AddTplContext needs to run after SessionMiddleware
	r.Use(WebResource{}.AddTplContext)

	r.Use(render.SetContentType(render.ContentTypeHTML))

	r.Get("/", web.Homepage)
	r.Handle("/shop/", http.RedirectHandler("/shop/any/", http.StatusSeeOther))
	r.Get("/r/{OrderReference}", func(w http.ResponseWriter, r *http.Request) {
		reference := chi.URLParam(r, "OrderReference")
		newURL := fmt.Sprintf("/cart/checkout/order-received/%v/", reference)
		http.Redirect(w, r, newURL, http.StatusSeeOther)
		return
	})

	r.Handle("/graphiql", handler.Playground("GraphQL playground", "/graphql"))

	graphHandler := handler.GraphQL(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}),
		handler.RecoverFunc(func(ctx context.Context, err interface{}) error {
			// notify bug tracker...
			insights.GenericRecoverer(err)
			return fmt.Errorf("internal server error")
		}))

	r.Handle("/graphql", GraphMiddleware(graphHandler))

	if socketioHandler, err := socketio.NewServer(nil); err == nil {
		livescores := Livescores{
			server: socketioHandler,
			ch:     make(chan string),
		}
		livescoresAlreadyRunning = true
		socketioHandler.On("connection", livescores.OnConnection)
		socketioHandler.On("error", livescores.OnError)
		// hook the socketio server in
		r.Handle("/socket.io/", socketioHandler)

	} else {
		log.Printf("socketio handler failed")
	}

	r.Route(fmt.Sprintf("/shop/{%v}/", KShopType), func(s chi.Router) {
		s.Use(web.AddShopContext)
		s.Get("/", web.Shop)
		s.Get(fmt.Sprintf("/{%v}/", KItemID), web.ShopDetails)
	})

	r.Route("/stats/", func(s chi.Router) {
		s.Use(web.AddStatsContext)
		s.Get("/", web.Stats)
		s.Get(fmt.Sprintf("/league/{%v}/season/{%v}/", KStatsLeagueID, KStatsSeasonID), web.StatsLeague)
		s.Get(fmt.Sprintf("/league/{%v}/season/{%v}/team/{%v}/", KStatsLeagueID, KStatsSeasonID, KStatsTeamID), web.StatsLeague)
		s.Get(fmt.Sprintf("/league/{%v}/season/{%v}/match/{%v}/", KStatsLeagueID, KStatsSeasonID, KStatsFixtureID), web.StatsMatchDetails)
	})

	cart := CartResource{BaseURL: "/cart/"}
	r.Mount(cart.BaseURL, cart.Routes())

	userRes := UserResource{
		BaseURL: "/user/",
	}
	r.Mount(userRes.BaseURL, userRes.Routes())

	r.Handle("/admin", http.RedirectHandler("/admin/", http.StatusSeeOther))
	adminRes := AdminResource{
		BaseURL: "/admin/",
	}
	r.Mount(adminRes.BaseURL, adminRes.Routes())

	r.Handle("/m", http.RedirectHandler("/m/", http.StatusSeeOther))
	mobileRes := MobileResource{
		BaseURL: "/m/",
	}
	r.Mount(mobileRes.BaseURL, mobileRes.Routes())

	r.Get("/about/", web.About)
	r.Get("/customersupport/", web.CustomerSupport)
	r.Post("/emailcapture", web.EmailCapture)

	return r
}

const (
	locEN = "en-US"
	locAR = "ar-AE"
)

// GraphMiddleware the graphql middleware
func GraphMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := sessionManager.Load(r)
		shoppingCart, _ := cart.GetCart(w, session)
		cartItemIDs := shoppingCart.GetItemsIDS()
		fmt.Printf("cartItemIDs = %+v\n", cartItemIDs)

		ctx := context.WithValue(r.Context(), "Graph_Cart", cartItemIDs)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// MARK: WebResource Middleware
func (wr WebResource) AddTplContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Print(" \n")
		fmt.Print(" \n")
		fmt.Print(" \n")
		fmt.Printf("------------------- %v %v -------------------\n", r.Method, r.RequestURI)
		var (
			user, _ = authResource.GetUserSession(r).(*models.User)
			ctx     = r.Context()
		)

		fmt.Printf("WebR.AddTplContext, User: [%v]\n", user)

		locale := i18n.GetLocaleContext(r)
		t_en := inline_edit.InlineEdit(i18n.I18n, locEN, false)
		t_ar := inline_edit.InlineEdit(i18n.I18n, locAR, false)

		path := r.URL.Path
		if locale == "ar-AE" {
			path = strings.Replace(path, "/ar/", "/", 1)
		}

		tplContext := pongo2.Context{
			"path":       path,
			"query":      r.URL.Query(),
			"version":    config.ENVVersion(),
			"user":       user,
			"menu_links": config.MenuLinks,
			// add locale and local functions
			"locale": locale,
			"href": func(url string) string {
				newURL := url
				if locale == "ar-AE" {
					newURL = fmt.Sprintf("/ar%v", url)
				}
				return newURL
			},
			"d": func(args ...interface{}) *pongo2.Value {
				log.Printf("debug: %v", args)

				value := args[0].(string)
				return pongo2.AsSafeValue(value)
			},
			// takes localizable database entries
			// and returns the only the one for the current locale
			"l": func(args ...interface{}) *pongo2.Value {
				localized, ok := args[0].(string)
				if !ok {
					return pongo2.AsSafeValue("")
				}
				tmpLocale := locale
				if len(args) > 1 {
					tmpLocale, ok = args[1].(string)
					if !ok {
						tmpLocale = locale
					}
				}
				value := l10n.Get(localized, tmpLocale)
				return pongo2.AsSafeValue(value)
			},
			"l10n": func(args ...interface{}) *pongo2.Value {
				localizedModel, ok := args[0].(models.StatsTeam)
				fieldName, ok := args[1].(string)
				if !ok {
					return pongo2.AsSafeValue("")
				}
				value := l10n.GetByFieldName(localizedModel, fieldName, locale)
				return pongo2.AsSafeValue(value)
			},
			"t": func(args ...interface{}) *pongo2.Value {

				t := t_en
				var value template.HTML
				if locale == locAR {
					t = t_ar
				}

				key, ok := args[0].(string)
				key = strings.Replace(key, " ", "", -1)
				if !ok {
					return pongo2.AsSafeValue("")
				}

				value = t(key, args...)
				if args[0].(string) == "pages.checkoutGateway.description" {
					log.Printf("WR.AddTplContex Locale: %v: %v %#v", locale, value, args)
					log.Printf("WR.Value: %v", value)
				}

				return pongo2.AsSafeValue(value)
			},
			"stats_filters": &filters.StatsFilter{
				Ctx: ctx,
			},
		}
		ctx = context.WithValue(ctx, TplContextWebResourceContextKey, tplContext)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func (wr WebResource) GetTplContext(r *http.Request) pongo2.Context {
	// recipe := r.Context().Value("recipe").(*models.Recipe)
	return r.Context().Value(TplContextWebResourceContextKey).(pongo2.Context)
}
func (wr WebResource) AddShopContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// add shop type custom context
		shopType := chi.URLParam(r, KShopType)
		ctx := context.WithValue(r.Context(), ShopContextWebResourceContextKey, shopType)

		// add menu look up stuff the tplContext
		tplContext := wr.GetTplContext(r)
		tx := db.GetDBFromRequestContext(r)
		teams := []models.Team{}
		//collections := []models.Collection{}
		if err := tx.Model(teams).Preload("Collections").Find(&teams).Error; err != nil {
			log.Printf("Failed to fetch teams: %v", err)
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
		sizes := []models.ProductSize{}
		if err := tx.Model(sizes).Find(&sizes).Error; err != nil {
			log.Printf("Failed to fetch sizes: %v", err)
		}
		brands := []models.Brand{}
		if err := tx.Find(&brands).Error; err != nil {
			log.Printf("WR.AddShopContext cant find brands. %v", err)
		}

		tplContext = tplContext.Update(pongo2.Context{
			"menu_clubs":      menu_clubs,
			"shop_type":       shopType,
			"brands":          brands,
			"product_sizes":   sizes,
			"product_kits":    models.Kits,
			"product_genders": models.Genders,
		})
		ctx = context.WithValue(ctx, TplContextWebResourceContextKey, tplContext)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func (wr WebResource) GetShopContext(r *http.Request) string {
	// recipe := r.Context().Value("recipe").(*models.Recipe)
	return r.Context().Value(ShopContextWebResourceContextKey).(string)
}
func (wr WebResource) AddStatsContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statsType := chi.URLParam(r, KStatsType)
		ctx := context.WithValue(r.Context(), StatsContextWebResourceContextKey, statsType)

		var (
			statsLeagueID = chi.URLParam(r, KStatsLeagueID)
			statsSeasonID = chi.URLParam(r, KStatsSeasonID)
			statsTeamID   = chi.URLParam(r, KStatsTeamID)
		)
		seasonID, err := strconv.Atoi(statsSeasonID)
		if err != nil {
			seasonID = 0
		}
		teamID, err := strconv.Atoi(statsTeamID)
		if err != nil {
			teamID = 0
		}

		// stats filter context
		tplContext := wr.GetTplContext(r)

		tplContext = tplContext.Update(pongo2.Context{
			KStatsLeagueID: statsLeagueID,
			KStatsSeasonID: seasonID,
			KStatsTeamID:   teamID,
		})

		ctx = context.WithValue(ctx, TplContextWebResourceContextKey, tplContext)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func (wr WebResource) GetStatsContext(r *http.Request) string {
	// recipe := r.Context().Value("recipe").(*models.Recipe)
	return r.Context().Value(StatsContextWebResourceContextKey).(string)
}

// MARK: Pages
func (web WebResource) Homepage(w http.ResponseWriter, r *http.Request) {
	var (
		tpl        = pongo2.Must(pongo2.FromFile("templates/html/homepage.html"))
		tplContext = web.GetTplContext(r)
		tx         = db.GetDBFromRequestContext(r)
	)
	// hold data
	productsBestsellers := []models.Product{}
	productWhatsNew := models.Product{}
	productFeatured := models.Product{}
	collectionBestSellers := models.Collection{}
	collectionWhatsNew := models.Collection{}
	collectionFeaturedProducts := models.Collection{}
	teams := []models.Team{}
	// build queries - for productsBestsellers
	if err := tx.Model(collectionBestSellers).Where("code = ?", models.KCollectionBestSellers).First(&collectionBestSellers).Error; err != nil {
		fmt.Printf("Failed to fetch collectionBestSellers: %v\n", err)
	}
	query := tx.Model(&productsBestsellers).Select("distinct products.*")
	query = query.
		Joins("inner join product_collections as pc on pc.product_id = id").
		Where("pc.collection_id = ?", collectionBestSellers.ID)
	// final - trigger query
	query = query.Find(&productsBestsellers)

	// queries for teams
	if err := tx.Model(teams).Preload("Collections").Find(&teams).Error; err != nil {
		log.Printf("Failed to fetch teams: %v", err)
	}
	type MatchStat struct {
		T1 models.Team
		T2 models.Team
	}
	match_stats := []MatchStat{}
	counter := 0
	for _, t := range teams {
		if counter >= 12 {
			break
		}
		for _, c := range t.Collections {
			if c.Code != "uaeleague" {
				continue
			}
			if counter < 6 {
				m := MatchStat{T1: t}
				match_stats = append(match_stats, m)
			} else {
				match_stats[counter-6].T2 = t
			}
			counter += 1
			break
		}
	}
	// query for whats new
	if err := tx.Model(collectionWhatsNew).Where("code = ?", models.KCollectionWhatsNew).First(&collectionWhatsNew).Error; err != nil {
		fmt.Printf("Failed to fetch collectionBestSellers: %v\n", err)
	}
	queryWN := tx.Model(&productWhatsNew).Select("distinct products.*")
	queryWN = queryWN.
		Joins("inner join product_collections as pc on pc.product_id = id").
		Where("pc.collection_id = ?", collectionBestSellers.ID)
	// final - trigger query
	queryWN = queryWN.Order("created_at desc").First(&productWhatsNew)
	// query for latest featured product

	if err := tx.Model(collectionFeaturedProducts).Where("code = ?", models.KCollectionFeaturedProducts).First(&collectionFeaturedProducts).Error; err != nil {
		fmt.Printf("Failed to fetch collectionBestSellers: %v\n", err)
	}
	queryFP := tx.Model(&productFeatured).Select("distinct products.*")
	queryFP = queryFP.
		Joins("inner join product_collections as pc on pc.product_id = id").
		Where("pc.collection_id = ?", collectionFeaturedProducts.ID)
	// final - trigger query
	queryFP = queryFP.Order("created_at asc").First(&productFeatured)

	tplContext = tplContext.Update(pongo2.Context{
		"bestseller_products": productsBestsellers,
		"match_stats":         match_stats,
		"whatsnew_product":    productWhatsNew,
		"featured_product":    productFeatured,
	})

	if err := tpl.ExecuteWriter(tplContext, w); err != nil {
		fmt.Printf("Tpl.ExecuteWriter Failed %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
}

func (web WebResource) Shop(w http.ResponseWriter, r *http.Request) {
	var (
		tpl        = pongo2.Must(pongo2.FromFile("templates/html/shop.html"))
		tplContext = web.GetTplContext(r)
		tx         = db.GetDBFromRequestContext(r)
		shopType   = web.GetShopContext(r)
	)

	products := []models.Product{}
	collection := models.Collection{}
	// 1) apply collection filter
	if shopType != "any" {
		if err := tx.Model(collection).Where("code = ?", shopType).First(&collection).Error; err != nil {
			fmt.Printf("Failed to fetch collection: %v\n", err)
		}
	}
	query := tx.Model(&products).Select("distinct products.*")

	if collection.Name != "" {
		query = query.
			Joins("inner join product_collections as pc on pc.product_id = id").
			Where("pc.collection_id = ?", collection.ID)
	} else {
		fmt.Println("no collection found ... :(")
	}

	// 3) apply kit and team filters defined in the model definition
	sortBy := []string{}
	for key, queryParams := range r.URL.Query() {
		applyQueryFnc, ok := models.QueryAttributeLookupMap[key]
		if !ok {
			// check to see if we can grab the sorting ...
			if key == "sort_by" {
				sortBy = queryParams
			}
			continue
		}
		query = applyQueryFnc(query, queryParams)
	}
	// sortby
	if len(sortBy) == 1 {
		sB := sortBy[0]
		if sB == "ToHigh" {
			query = query.Order("products.price asc")
		} else {
			query = query.Order("products.price desc")
		}
	}

	// final - trigger query
	query = query.Find(&products)

	if err := query.Error; err != nil {
		fmt.Printf("Failed to fetch products: %v\n", err)
	}
	//if err := tx.Model(products).Preload("Collections").Preload("Sizes").Find(&products).Error; err != nil {
	//	fmt.Println("Failed to fetch products: %v", err)
	//}
	// implementing filtering later on :)

	tplContext = tplContext.Update(pongo2.Context{
		"products": products,
	})

	if err := tpl.ExecuteWriter(tplContext, w); err != nil {
		fmt.Printf("Tpl.ExecuteWriter Failed %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
}

func (web WebResource) ShopDetails(w http.ResponseWriter, r *http.Request) {
	var (
		tpl        = pongo2.Must(pongo2.FromFile("templates/html/shop-details.html"))
		tplContext = web.GetTplContext(r)
		tx         = db.GetDBFromRequestContext(r)
		itemID     = chi.URLParam(r, KItemID)
	)

	product := models.Product{}
	if err := tx.Model(product).Preload("Collections").Where("id == ?", itemID).First(&product).Error; err != nil {
		fmt.Printf("Failed to fetch product: %v", err)
		render.Render(w, r, ErrNotFound)
		return
	}

	products := []models.Product{}
	collectionIDs := []uint{}
	for _, c := range product.Collections {
		collectionIDs = append(collectionIDs, c.ID)
	}
	rand.Seed(time.Now().Unix())
	randCollectionID := collectionIDs[rand.Intn(len(collectionIDs))]

	query := tx.Model(&products)
	if len(collectionIDs) > 0 {
		query = query.
			Joins("inner join product_collections as pc on pc.product_id = id").
			Where("pc.collection_id = ?", randCollectionID).
			Limit(4).
			Find(&products)
	}
	if err := query.Error; err != nil {
		log.Printf("failed to query related products: %v", err)
	}

	variations := []models.ProductVariation{}
	badges := map[uint]models.Badge{}
	sizes := map[uint]models.ProductSize{}
	if err := tx.Model(variations).Where("product_id = ? AND available_quantity > 0", product.ID).
		Preload("Size").
		Preload("Badge").
		Find(&variations).Error; err != nil {
		log.Printf("failed to query product variations: %v", err)
	}

	for _, pv := range variations {
		if pv.BadgeID != nil {
			_, ok := badges[*pv.BadgeID]
			if !ok {
				badges[*pv.BadgeID] = pv.Badge
			}
		}
		_, ok := sizes[pv.SizeID]
		if !ok {
			sizes[pv.SizeID] = pv.Size
		}
	}

	tplContext = tplContext.Update(pongo2.Context{
		KItemID:            itemID,
		"product":          product,
		"related_products": products,
		"badges":           badges,
		"productSizes":     sizes,
	})

	if err := tpl.ExecuteWriter(tplContext, w); err != nil {
		fmt.Printf("Tpl.ExecuteWriter Failed %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
}

func ReturnFormat(r *http.Request) string {
	format := "html"

	if f := r.URL.Query().Get("format"); f == "json" {
		// not just blindly accept anything
		format = "json"
	}

	return format
}

func (web WebResource) Stats(w http.ResponseWriter, r *http.Request) {
	var (
		tpl        = pongo2.Must(pongo2.FromFile("templates/html/stats.html"))
		tplContext = web.GetTplContext(r)
	)

	if ReturnFormat(r) == "json" {
		tpl = pongo2.Must(pongo2.FromFile("templates/json/stats.json"))
	}

	dateParam := r.URL.Query().Get("date")

	//return c.DateTime.Format("2006-01-02")
	some := tplContext["stats_filters"].(*filters.StatsFilter)
	statsCalendar := some.GetStatsCalendarAll(dateParam)

	idx := statsCalendar.TodayOrNextIdx
	calEntry := statsCalendar.Entries[idx]

	dateTime, err := time.Parse("2006-01-02", dateParam)
	if err != nil {
		dateTime = *calEntry.DateTime
	}

	tplContext = tplContext.Update(pongo2.Context{
		"stats_calendar_today_or_next_date": dateTime,
		"now":                          time.Now(),
		"nothing":                      "nothing",
		"stats_calendar":               statsCalendar,
		"stats_calendar_next_or_today": calEntry,
	})

	if err := tpl.ExecuteWriter(tplContext, w); err != nil {
		log.Printf("Tpl.ExecuteWriter Failed %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
}

func (web WebResource) StatsMatchDetails(w http.ResponseWriter, r *http.Request) {
	var (
		tpl            = pongo2.Must(pongo2.FromFile("templates/html/stats-details.html"))
		tplContext     = web.GetTplContext(r)
		statsFixtureID = chi.URLParam(r, KStatsFixtureID)
	)

	if ReturnFormat(r) == "json" {
		tpl = pongo2.Must(pongo2.FromFile("templates/json/stats-details.json"))
	}
	var (
		statsLeagueID = chi.URLParam(r, KStatsLeagueID)
		statsSeasonID = chi.URLParam(r, KStatsSeasonID)
	)
	seasonID, err := strconv.Atoi(statsSeasonID)
	if err != nil {
		seasonID = 0
	}

	tplContext = tplContext.Update(pongo2.Context{
		KStatsLeagueID:  statsLeagueID,
		KStatsSeasonID:  seasonID,
		KStatsFixtureID: statsFixtureID,
	})

	if err := tpl.ExecuteWriter(tplContext, w); err != nil {
		log.Printf("Tpl.ExecuteWriter Failed %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
}

func (web WebResource) StatsLeague(w http.ResponseWriter, r *http.Request) {
	var (
		tpl        = pongo2.Must(pongo2.FromFile("templates/html/stats-league.html"))
		tplContext = web.GetTplContext(r)
	)

	if ReturnFormat(r) == "json" {
		tpl = pongo2.Must(pongo2.FromFile("templates/json/stats-league.json"))
	}

	var (
		statsLeagueID = chi.URLParam(r, KStatsLeagueID)
		statsSeasonID = chi.URLParam(r, KStatsSeasonID)
		statsTeamID   = chi.URLParam(r, KStatsTeamID)
	)
	seasonID, err := strconv.Atoi(statsSeasonID)
	if err != nil {
		seasonID = 0
	}
	teamID, err := strconv.Atoi(statsTeamID)
	if err != nil {
		teamID = 0
	}

	teamIDs := []uint{}
	if teamID > 0 {
		teamIDs = append(teamIDs, uint(teamID))
	}

	dateParam := r.URL.Query().Get("date")
	some := tplContext["stats_filters"].(*filters.StatsFilter)
	statsCalendar := some.GetStatsCalendar(seasonID, dateParam, teamIDs)
	idx := statsCalendar.TodayOrNextIdx
	calEntry := statsCalendar.Entries[idx]

	tplContext = tplContext.Update(pongo2.Context{
		"stats_calendar":                    statsCalendar,
		"stats_calendar_today_or_next_date": calEntry.DateTime,
		KStatsLeagueID:                      statsLeagueID,
		KStatsSeasonID:                      seasonID,
		KStatsTeamID:                        teamID,
	})

	if err := tpl.ExecuteWriter(tplContext, w); err != nil {
		fmt.Printf("Tpl.ExecuteWriter Failed %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
}

// OTHER PAGES
func (web WebResource) About(w http.ResponseWriter, r *http.Request) {
	var (
		tpl        = pongo2.Must(pongo2.FromFile("templates/html/about.html"))
		tplContext = web.GetTplContext(r)
	)

	if err := tpl.ExecuteWriter(tplContext, w); err != nil {
		fmt.Printf("Tpl.ExecuteWriter Failed %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
}

func (web WebResource) CustomerSupport(w http.ResponseWriter, r *http.Request) {
	var (
		tpl        = pongo2.Must(pongo2.FromFile("templates/html/customersupport.html"))
		tplContext = web.GetTplContext(r)
	)

	if err := tpl.ExecuteWriter(tplContext, w); err != nil {
		fmt.Printf("Tpl.ExecuteWriter Failed %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
}

func (web WebResource) EmailCapture(w http.ResponseWriter, r *http.Request) {
	resp := struct {
		Done string
	}{
		Done: "Done",
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}
