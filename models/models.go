package models

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/config"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/i18n/l10n"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/gosportmonks"
	"github.com/jinzhu/gorm"
)

/*

   displayName: displayName,
   email: email,
   emailVerified: emailVerified,
   phoneNumber: phoneNumber,
   photoURL: photoURL,
   uid: uid,
   accessToken: accessToken,
   providerData: providerData
*/
type FirebaseUserData struct {
	UID           string
	DisplayName   string
	Email         string
	EmailVerified bool
	PhoneNumber   string
	PhotoURL      string

	IdToken    string // we could also grab a new one every time but this is easier
	ProviderID string
}

/*
0-49 unknown users
50-99 known users
100-1000 different levels of admin
*/
const (
	UserAccessLevelAnonymous = 49
	UserAccessLevelRegular   = 99
	UserAccessLevelAdmin     = 1000
	DefaultAnonymousLevel    = 20
	DefaultUserLevel         = 50
	DefaultAdminLevel        = 100
)

// no password or other auth needed we use firebase for auth
type User struct {
	gorm.Model

	Email           string
	FirebaseID      string
	UserAccessLevel int64
	Telephone       string
}

func (u User) GetUser(tx *gorm.DB, id string) (*User, error) {
	if id == "" {
		return nil, nil
	}
	userId, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	user := &User{}
	if err := tx.Model(user).Where("id == ?", userId).Find(user).Error; err != nil {
		fmt.Errorf("User.GetUser failed: %v\n", err)
		return nil, err
	}
	return user, nil
}

type StatsLeague struct {
	gosportmonks.League

	// the league already has an ID field gorm complains about duplicates ...
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}
type StatsTeam struct {
	gosportmonks.Team

	// the league already has an ID field gorm complains about duplicates ...
	CreatedAt                  time.Time         `json:"createdAt"`
	UpdatedAt                  time.Time         `json:"updatedAt"`
	DeletedAt                  *time.Time        `sql:"index" json:"deletedAt"`
	Name                       string            `l10nTarget:"NameL10N"`
	NameL10N                   string            `l10n:"fields:en,ar;" json:"-"`
	NameL10NMap                map[string]string `sql:"-" gorm:"-" json:"name"`
	CombinedLeagueSeasonTeamID string            `json:"combinedLeagueSeasonTeamId"`
	LeagueID                   int               `json:"leagueId"`
	SeasonID                   int               `json:"seasonId"`
}

func (t *StatsTeam) FromGosportmonksTeam(team gosportmonks.Team) {
	t.Team = team
	t.Name = team.Name
}
func (t StatsTeam) L10NFields() []string {
	return l10n.L10NFieldsInner(t)
}
func (t StatsTeam) Map() map[string]interface{} {
	return l10n.L10NMapInner(t)
}

type StatsSeason struct {
	gosportmonks.Season

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}
type StatsStage struct {
	gosportmonks.Stage

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}
type StatsRound struct {
	gosportmonks.Round

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

type League struct {
	ID            uint   `json:"ID"`
	Name          string `l10n:"fields:en,ar;"`
	StatsLeagueID int    `json:"StatsLeagueId"`

	// the league already has an ID field gorm complains about duplicates ...
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

func (t League) L10NFields() []string {
	return l10n.L10NFieldsInner(t)
}
func (t League) Map() map[string]interface{} {
	return l10n.L10NMapInner(t)
}

// collections are used to group products and groups of products
/*
For example: La Liga is a League in stats, but in the shop it's a collection.
*/
const (
	KCollectionWhatsNew         = "whatsnew"
	KCollectionFeaturedProducts = "featuredproducts"
	KCollectionBestSellers      = "bestsellers"
)

type Collection struct {
	gorm.Model

	Name string
	Code string
}

/*
Teams are in multiple Leagues and multiple Collections
*/
type Team struct {
	gorm.Model

	Name                   string       `l10n:"fields:en,ar;"`
	Leagues                []League     `gorm:"many2many:team_leagues"` // teams can be in multiple leagues/competitions
	Collections            []Collection `gorm:"many2many:team_collections;"`
	BrandID                uint
	Brand                  Brand
	Logo                   string
	StatsTeamID            int
	StatsTeamIDCombinedKey string `json:"StatsTeamIDCombinedKey"`
	StatsBuyButtonState    int
}

const (
	StatsBuyButtonButtonStateEnabled  = 2
	StatsBuyButtonButtonStateDisabled = 3
)

func (t *Team) GetName(locale string) string {
	return l10n.Get(t.Name, locale)
}
func (t *Team) SetName(locale string, value string) {
	l10n.Set(t.Name, locale, value)
}
func (t *Team) GetShopFilterURL() string {
	if t.ID == 0 || t.StatsBuyButtonState == StatsBuyButtonButtonStateDisabled {
		return ""
	}
	resp := fmt.Sprintf("/shop/any/?team_select=%v", t.ID)
	return resp
}
func (t Team) L10NFields() []string {
	return l10n.L10NFieldsInner(t)
}
func (t Team) Map() map[string]interface{} {
	return l10n.L10NMapInner(t)
}

func (t Team) GetTeamsGroupByCollectionCode(tx *gorm.DB) map[string][]Team {
	teams := []Team{}
	if err := tx.Model(teams).Preload("Collections").Find(&teams).Error; err != nil {
		log.Println("Failed to fetch teams: %v", err)
	}
	menu_clubs := map[string][]Team{}
	for _, t := range teams {
		for _, c := range t.Collections {
			if _, ok := menu_clubs[c.Code]; !ok {
				menu_clubs[c.Code] = []Team{}
			}
			menu_clubs[c.Code] = append(menu_clubs[c.Code], t)
		}
	}
	return menu_clubs
}

type Player struct {
	gorm.Model

	Name        string
	TeamID      int
	Collections []Collection `l10n:"sync" gorm:"many2many:player_collections;"`
}
type ProductAttrs struct {
	Code          string
	Name          string
	TranslationID string
}

func (k ProductAttrs) Get(code string) *ProductAttrs {
	for _, k := range AllProductAttrs {
		if k.Code == code {
			return &k
		}
	}
	return nil
}

var Kits = []ProductAttrs{
	{"homekit", "Home Kit", "sidebar.options.homekit"},
	{"awaykit", "Away Kit", "sidebar.options.awaykit"},
	{"thirdkit", "Third Kit", "sidebar.options.thirdkit"},
	{"goalkeeperkit", "Goalkeeper Kit", "sidebar.options.goalkeeperkit"},
	{"accessories", "Accessories", "sidebar.options.accessories"},
}

const FilterKeyKit = "filter_kit"

var Genders = []ProductAttrs{
	{"male", "Male", "sidebar.options.male"},
	{"female", "Female", "sidebar.options.female"},
	{"unisex", "Unisex", "sidebar.options.unisex"},
	{"youth", "Youth", "sidebar.options.youth"},
}

const FilterKeyGender = "filter_gender"

const FilterTeams = "team_select"

const FilterBrands = "brand_select"

var AllProductFilterKeys = map[string]bool{
	FilterKeyGender:      true,
	FilterKeyKit:         true,
	FilterKeyProductSize: true,
}
var AllProductAttrs = []ProductAttrs{}
var QueryAttributeLookupMap = map[string]func(*gorm.DB, []string) *gorm.DB{}

func init() {
	// concat all product attrs together
	AllProductAttrs = append(AllProductAttrs, Kits...)
	AllProductAttrs = append(AllProductAttrs, Genders...)

	// build a map of query params
	QueryAttributeLookupMap[FilterKeyKit] = func(query *gorm.DB, attrs []string) *gorm.DB {
		if len(attrs) == 0 {
			return query
		}
		return query.Where("kit_code in (?)", attrs)
	}
	QueryAttributeLookupMap[FilterKeyGender] = func(query *gorm.DB, attrs []string) *gorm.DB {
		if len(attrs) == 0 {
			return query
		}
		return query.Where("gender in (?)", attrs)
	}
	QueryAttributeLookupMap[FilterKeyProductSize] = func(query *gorm.DB, attrs []string) *gorm.DB {
		if len(attrs) == 0 {
			return query
		}
		return query.Joins("INNER JOIN product_productsizes as ps on ps.product_id = products.id and ps.product_size_id in (?)", attrs)
		//return query.Where("gender in (?)", attrs)
	}

	QueryAttributeLookupMap[FilterTeams] = func(query *gorm.DB, attrs []string) *gorm.DB {
		if len(attrs) == 0 {
			return query
		}
		return query.Joins("INNER JOIN teams as t on t.id = products.team_id and t.id in (?)", attrs)
	}

	QueryAttributeLookupMap[FilterBrands] = func(query *gorm.DB, attrs []string) *gorm.DB {
		if len(attrs) == 0 {
			return query
		}
		return query.Joins("INNER JOIN brands as b on b.id = products.brand_id and b.name in (?)", attrs)
	}
}

type Product struct {
	gorm.Model

	Name        string
	Description string
	Price       float64 // is supposed to show the lowest price of all product variations
	Gender      string  // "youth", "unisex", "male", "female"

	Image       string // for now we only use one product image - makes things easier
	Image2 		string
	Image3 		string
	Image4 		string
	Thumbnail   string
	Collections []Collection `l10n:"sync" gorm:"many2many:product_collections;"`
	// all sizes in which this product is available
	Sizes      []ProductSize `gorm:"many2many:product_productsizes;"`
	CategoryID uint          //
	Category   *Category     // options are: t-shirt,pants,shoes,merchandise, - adidas had: apparel, shoes, accessories
	KitCode    string

	Variations []ProductVariation
	/*
	  - name: Jersey
	  - name: T-Shirt
	  - name: Shirt
	  - name: Jacket
	  - name: Pants
	  - name: Sweatshirt
	  - name: Shorts
	*/
	BrandID uint // BrandReference
	Brand   *Brand

	LeagueID int  // possible league reference
	TeamID   uint // possible team reference
	Team     *Team
	PlayerID int // possible player reference
}

func (p *Product) GetRelativeURL() string {
	prodUrl := fmt.Sprintf("/shop/any/%v/", p.ID)
	return prodUrl
}

// ProductVariation holds all different sub types of products
// that are based of the same base product
// such as:
// - size
// - badge
type ProductVariation struct {
	gorm.Model
	ProductID uint
	Product   Product

	BadgeID     *uint
	Badge       Badge
	SizeID      uint
	Size        ProductSize
	CustomPrint bool

	SKU               string
	AvailableQuantity uint
}

const FilterKeyProductSize = "filter_size"

type ProductSize struct {
	gorm.Model

	Name string // XXS, ... XXL
}
type Category struct {
	gorm.Model
	Name string
	Code string // lowercase version of Category Name
}
type Brand struct {
	gorm.Model

	Name string
}

// Badges
type Badge struct {
	gorm.Model

	Name      string
	Image     string
	Thumbnail string
	Price     float64
}

type PageContent struct {
	gorm.Model

	Page 					   string			 `json:"page"`
	Identifier 				   string			 `json:"identifier"`
	Text                       string            `l10nTarget:"TextL10N"`
	TextL10N                   string            `l10n:"fields:en,ar;" json:"-"`
	Link 					   string            `json:"link"`
}
func (t PageContent) L10NFields() []string {
	return l10n.L10NFieldsInner(t)
}
func (t PageContent) Map() map[string]interface{} {
	return l10n.L10NMapInner(t)
}

// GetProduct
func (p Product) GetProducts(tx *gorm.DB, id int, collectionId int, collectionCode string, kitCode string, productSort string) ([]Product, error) {
	products := []Product{}
	collection := Collection{}

	if collectionCode != "any" {
		if err := tx.Model(collection).Where("code = ?", collectionCode).First(&collection).Error; err != nil {
			log.Printf("Failed to fetch collection: %v\n", err)
		}
	}

	patterns := []string{}
	values := []interface{}{}
	middle := []interface{}{}
	pattern1 := "%v = ?"
	pattern2 := "%v in ?"
	comma := " AND "
	query := tx.Model(&products).Select("products.*").Joins("inner join product_collections as pc on pc.product_id = id")

	if collection.Name != "" {
		patterns = append(patterns, pattern1)
		middle = append(middle, "pc.collection_id")
		values = append(values, collection.ID)
	}

	if kitCode != "" {
		patterns = append(patterns, pattern2)
		middle = append(middle, "products.kit_code")
		values = append(values, kitCode)
	}

	whereFormat := strings.Join(patterns, comma)
	whereStr := fmt.Sprintf(whereFormat, middle...)

	if collection.Name != "" || kitCode != "" {
		query = query.Where(whereStr, values...)
	}

	if productSort == "ToHigh" {
		query = query.Order("products.price asc")
	} else {
		query = query.Order("products.price desc")
	}

	err := query.Find(&products).Error

	return products, err

}

// GetBrands
func (b Brand) GetBrands(tx *gorm.DB) []Brand {
	brands := []Brand{}
	if err := tx.Model(brands).Find(&brands).Error; err != nil {
		log.Printf("WR.AddShopContext cant find brands. %v", err)
	}
	return brands
}

// GetProductSizes
func (ps ProductSize) GetProductSizes(tx *gorm.DB) []ProductSize {
	productSizes := []ProductSize{}
	if err := tx.Model(productSizes).Find(&productSizes).Error; err != nil {
		log.Printf("WR.AddShopContext cant find productSizes. %v", err)
	}
	return productSizes
}

// GetProductKits
func (k ProductAttrs) GetProductKits() []ProductAttrs {
	kitList := []ProductAttrs{}
	for _, k := range Kits {
		kitList = append(kitList, k)
	}
	return kitList
}

// GetProductGenders
func (k ProductAttrs) GetProductGenders() []ProductAttrs {
	kitList := []ProductAttrs{}
	for _, k := range Genders {
		kitList = append(kitList, k)
	}
	return kitList
}

// GetCollections
func (c Collection) GetCollections(tx *gorm.DB, shopType string) []Collection {
	collection := []Collection{}
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
func (c Product) GetProductDetails(tx *gorm.DB, productId int) []Product {
	product := []Product{}

	if err := tx.Model(product).First(&product, productId).Error; err != nil {
		fmt.Printf("Failed to fetch product: %v\n", err)
	}

	return product
}

// ProductVariations Static Methods
// GetProductVariations
func (pv ProductVariation) GetProductVariations(tx *gorm.DB, productId int) []ProductVariation {
	variations := []ProductVariation{}
	if err := tx.Model(variations).Where("product_id = ? AND available_quantity > 0", productId).
		Preload("Size").
		Preload("Badge").
		Find(&variations).Error; err != nil {
		fmt.Errorf("failed to query product variations: %v", err)
	}

	return variations
}

// ProductVariations Instance Methods
func (pv *ProductVariation) GetPrice() float64 {
	var price float64 = 0

	product := pv.Product
	price = product.Price

	// address special variation pricing for Badges
	// this will have to be moved once we have more price variations
	if pv.BadgeID != nil {
		// alright, we've got a badge

		badgePrice := config.DefaultBadgePrice // 10 AED
		badge := pv.Badge
		badgePrice = badge.Price
		price += badgePrice
	}

	return price
}
