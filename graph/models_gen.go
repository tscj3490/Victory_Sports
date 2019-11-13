// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package graph

import (
	time "time"
)

type Address struct {
	ID           int        `json:"ID"`
	CreatedAt    *time.Time `json:"createdAt"`
	UpdatedAt    *time.Time `json:"updatedAt"`
	DeletedAt    *time.Time `json:"deletedAt"`
	UserID       *int       `json:"userId"`
	ContactName  string     `json:"contactName"`
	Telephone    string     `json:"telephone"`
	City         string     `json:"city"`
	Country      string     `json:"country"`
	AddressLine1 string     `json:"addressLine1"`
	AddressLine2 string     `json:"addressLine2"`
	PostalCode   string     `json:"postalCode"`
	Notes        string     `json:"notes"`
}
type Badge struct {
	CreatedAt *time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
	Name      string     `json:"name"`
	Image     string     `json:"image"`
	Thumbnail string     `json:"thumbnail"`
	Price     float64    `json:"price"`
}
type Brand struct {
	CreatedAt *time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
	Name      string     `json:"name"`
}
type Cards struct {
	YellowCards int `json:"yellowCards"`
	RedCards    int `json:"redCards"`
}
type Category struct {
	CreatedAt *time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
	Name      string     `json:"name"`
	Code      string     `json:"code"`
}
type CollectionM struct {
	Name      string     `json:"name"`
	Code      string     `json:"code"`
	CreatedAt *time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}
type Coverage struct {
	TopscorerGoals   bool `json:"topscorerGoals"`
	TopscorerAssists bool `json:"topscorerAssists"`
	TopscorerCards   bool `json:"topscorerCards"`
}
type Events struct {
	LocalTeamEvents   []*FixtureEventArray `json:"localTeamEvents"`
	VisitorTeamEvents []*FixtureEventArray `json:"visitorTeamEvents"`
}
type EventsInclude struct {
	Events []FixtureEvent `json:"events"`
}
type Fixture struct {
	ID                 int                `json:"id"`
	LeagueID           int                `json:"leagueId"`
	SeasonID           int                `json:"seasonId"`
	StageID            int                `json:"stageId"`
	RoundID            int                `json:"roundId"`
	GroupID            int                `json:"groupId"`
	VenueID            int                `json:"venueId"`
	LocalTeamID        int                `json:"localTeamId"`
	VisitorTeamID      int                `json:"visitorTeamId"`
	Formation          FixtureFormation   `json:"formation"`
	Scores             FixtureScore       `json:"scores"`
	Time               FixtureTime        `json:"time"`
	Coaches            FixtureCoaches     `json:"coaches"`
	Standings          FixtureStandings   `json:"standings"`
	Deleted            bool               `json:"deleted"`
	LocalTeamInclude   LocalTeamInclude   `json:"localTeamInclude"`
	VisitorTeamInclude VisitorTeamInclude `json:"visitorTeamInclude"`
	EventsInclude      EventsInclude      `json:"eventsInclude"`
	LineupInclude      LineupInclude      `json:"lineupInclude"`
}
type FixtureAndEvents struct {
	Fixture *Fixture `json:"fixture"`
	Events  *Events  `json:"events"`
}
type FixtureCoaches struct {
	LocalTeamCoachID   int `json:"localTeamCoachId"`
	VisitorTeamCoachID int `json:"visitorTeamCoachId"`
}
type FixtureEvent struct {
	ID                int    `json:"id"`
	TeamID            string `json:"teamId"`
	Type              string `json:"type"`
	FixtureID         int    `json:"fixtureId"`
	PlayerID          int    `json:"playerId"`
	PlayerName        string `json:"playerName"`
	RelatedPlayerID   int    `json:"relatedPlayerId"`
	RelatedPlayerName string `json:"relatedPlayerName"`
	Minute            int    `json:"minute"`
	ExtraMinute       int    `json:"extraMinute"`
	Reason            string `json:"reason"`
	Injuried          bool   `json:"injuried"`
	Result            string `json:"result"`
}
type FixtureEventArray struct {
	Key   string         `json:"key"`
	Value []FixtureEvent `json:"value"`
}
type FixtureFormation struct {
	LocalTeamFormation   string `json:"localTeamFormation"`
	VisitorTeamFormation string `json:"visitorTeamFormation"`
}
type FixturePlayer struct {
	TeamID            int                `json:"teamId"`
	FixtureID         int                `json:"fixtureId"`
	PlayerID          int                `json:"playerId"`
	PlayerName        int                `json:"playerName"`
	Number            int                `json:"number"`
	Position          string             `json:"position"`
	FormationPosition int                `json:"formationPosition"`
	PosX              int                `json:"posX"`
	PosY              int                `json:"posY"`
	Stats             FixturePlayerStats `json:"stats"`
}
type FixturePlayerStats struct {
	Shorts  Shots   `json:"shorts"`
	Goals   Goals   `json:"goals"`
	Fouls   Fouls   `json:"fouls"`
	Cards   Cards   `json:"cards"`
	Passing Passing `json:"passing"`
	Other   Other   `json:"other"`
}
type FixtureScore struct {
	LocalTeamScore      int    `json:"localTeamScore"`
	VisitorTeamScore    int    `json:"visitorTeamScore"`
	LocalTeamScorePen   int    `json:"localTeamScorePen"`
	VisitorTeamScorePen int    `json:"visitorTeamScorePen"`
	HalfTimeScore       string `json:"halfTimeScore"`
	FullTimeScore       string `json:"fullTimeScore"`
	ExtraTimeScore      string `json:"extraTimeScore"`
}
type FixtureStandings struct {
	LocalTeamPosition   int `json:"localTeamPosition"`
	VisitorTeamPosition int `json:"visitorTeamPosition"`
}
type FixtureStartingAt struct {
	DateTime  string `json:"dateTime"`
	Date      string `json:"date"`
	Time      string `json:"time"`
	Timestamp int    `json:"timestamp"`
	Timezone  string `json:"timezone"`
}
type FixtureTime struct {
	Status     string            `json:"status"`
	StartingAt FixtureStartingAt `json:"startingAt"`
	Minute     int               `json:"minute"`
	Second     int               `json:"second"`
	AddedTime  int               `json:"addedTime"`
	ExtraTime  int               `json:"extraTime"`
	InjuryTime int               `json:"injuryTime"`
}
type FormSelection struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
	Href string `json:"href"`
}
type Fouls struct {
	Drawn    int `json:"drawn"`
	Comitted int `json:"comitted"`
}
type Goals struct {
	Scored   int `json:"scored"`
	Conceded int `json:"conceded"`
}
type LeagueM struct {
	League    League     `json:"league"`
	CreatedAt *time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}
type LineupInclude struct {
	Players []FixturePlayer `json:"players"`
}
type LocalTeamInclude struct {
	LocalTeam Team `json:"localTeam"`
}
type Order struct {
	CreatedAt         *time.Time  `json:"createdAt"`
	UpdatedAt         *time.Time  `json:"updatedAt"`
	DeletedAt         *time.Time  `json:"deletedAt"`
	Reference         string      `json:"reference"`
	UserID            *int        `json:"userId"`
	User              User        `json:"user"`
	Email             string      `json:"email"`
	PaymentAmount     float64     `json:"paymentAmount"`
	AbandonedReason   string      `json:"abandonedReason"`
	DiscountValue     int         `json:"discountValue"`
	TrackingNumber    string      `json:"trackingNumber"`
	ShippedAt         *time.Time  `json:"shippedAt"`
	ReturnedAt        *time.Time  `json:"returnedAt"`
	CancelledAt       *time.Time  `json:"cancelledAt"`
	ShippingAddressID int         `json:"shippingAddressId"`
	PaymentMethod     string      `json:"paymentMethod"`
	ShippingMethod    string      `json:"shippingMethod"`
	Subtotal          float64     `json:"subtotal"`
	VAT               float64     `json:"VAT"`
	ShippingCost      float64     `json:"shippingCost"`
	Total             float64     `json:"total"`
	Notes             string      `json:"notes"`
	OrderItems        []OrderItem `json:"orderItems"`
}
type OrderItem struct {
	CreatedAt                 *time.Time       `json:"createdAt"`
	UpdatedAt                 *time.Time       `json:"updatedAt"`
	DeletedAt                 *time.Time       `json:"deletedAt"`
	OrderID                   int              `json:"orderId"`
	ProductVariationID        int              `json:"productVariationId"`
	ProductVariation          ProductVariation `json:"productVariation"`
	PersistProductDetailsJSON string           `json:"persistProductDetailsJSON"`
	CustomizedName            string           `json:"customizedName"`
	CustomizedNumber          int              `json:"customizedNumber"`
	Quantity                  int              `json:"quantity"`
	Price                     float64          `json:"price"`
	DiscountRate              int              `json:"discountRate"`
}
type Other struct {
	Assists       int `json:"assists"`
	Offsides      int `json:"offsides"`
	Saves         int `json:"saves"`
	PenScored     int `json:"penScored"`
	PenMissed     int `json:"penMissed"`
	PenSaved      int `json:"penSaved"`
	PenCommitted  int `json:"penCommitted"`
	PenWon        int `json:"penWon"`
	HitWoodwork   int `json:"hitWoodwork"`
	Tackles       int `json:"tackles"`
	Blocks        int `json:"blocks"`
	Interceptions int `json:"interceptions"`
	Clearances    int `json:"clearances"`
	MinutesPlayed int `json:"minutesPlayed"`
}
type Passing struct {
	TotalCrosses    int `json:"totalCrosses"`
	CrossesAccuracy int `json:"crossesAccuracy"`
	Passes          int `json:"passes"`
	PassesAccuracy  int `json:"passesAccuracy"`
}
type Player struct {
	PlayerID         int              `json:"playerId"`
	TeamID           int              `json:"teamId"`
	CountryID        int              `json:"countryId"`
	PositionID       int              `json:"positionId"`
	CommonName       string           `json:"commonName"`
	FullName         string           `json:"fullName"`
	FirstName        string           `json:"firstName"`
	LastName         string           `json:"lastName"`
	Nationality      string           `json:"nationality"`
	Birthdate        string           `json:"birthdate"`
	Birthcountry     string           `json:"birthcountry"`
	Birthplace       string           `json:"birthplace"`
	Height           string           `json:"height"`
	Weight           string           `json:"weight"`
	ImagePath        string           `json:"imagePath"`
	TeamInclude      TeamInclude      `json:"teamInclude"`
	StatsInclude     StatsInclude     `json:"statsInclude"`
	TransfersInclude TransfersInclude `json:"transfersInclude"`
}
type PlayerInclude struct {
	Player Player `json:"player"`
}
type PlayerPosition struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
type PlayerSquadStats struct {
	PlayerID          int             `json:"playerId"`
	TeamID            int             `json:"teamId"`
	LeagueID          int             `json:"leagueId"`
	Minutes           int             `json:"minutes"`
	Appearences       int             `json:"appearences"`
	Lineups           int             `json:"lineups"`
	SubstituteIn      int             `json:"substituteIn"`
	SubstituteOut     int             `json:"substituteOut"`
	SubstituteOnBench int             `json:"substituteOnBench"`
	Goals             int             `json:"goals"`
	YellowCards       int             `json:"yellowCards"`
	Yellowred         int             `json:"yellowred"`
	RedCards          int             `json:"redCards"`
	Type              string          `json:"type"`
	PlayerInclude     PlayerInclude   `json:"playerInclude"`
	PositionInclude   PositionInclude `json:"positionInclude"`
}
type PlayerStats struct {
	PlayerID          int    `json:"playerId"`
	TeamID            int    `json:"teamId"`
	LeagueID          int    `json:"leagueId"`
	Minutes           int    `json:"minutes"`
	Appearences       int    `json:"appearences"`
	Lineups           int    `json:"lineups"`
	SubstituteIn      int    `json:"substituteIn"`
	SubstituteOut     int    `json:"substituteOut"`
	SubstituteOnBench int    `json:"substituteOnBench"`
	Goals             int    `json:"goals"`
	YellowCards       int    `json:"yellowCards"`
	Yellowred         int    `json:"yellowred"`
	RedCards          int    `json:"redCards"`
	Type              string `json:"type"`
}
type PlayerTransfers struct {
	PlayerID int    `json:"playerId"`
	ToTeamID int    `json:"toTeamId"`
	SeasonID int    `json:"seasonId"`
	Transfer string `json:"transfer"`
	Type     string `json:"type"`
	Date     string `json:"date"`
	Amount   string `json:"amount"`
}
type PositionInclude struct {
	Position PlayerPosition `json:"position"`
}
type Product struct {
	ID          int                `json:"id"`
	CreatedAt   *time.Time         `json:"createdAt"`
	UpdatedAt   *time.Time         `json:"updatedAt"`
	DeletedAt   *time.Time         `json:"deletedAt"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Price       float64            `json:"price"`
	Gender      string             `json:"gender"`
	Image       string             `json:"image"`
	Thumbnail   string             `json:"thumbnail"`
	Collections []CollectionM      `json:"collections"`
	Sizes       []ProductSize      `json:"sizes"`
	CategoryID  int                `json:"categoryId"`
	Category    *Category          `json:"category"`
	KitCode     string             `json:"kitCode"`
	Variations  []ProductVariation `json:"variations"`
	BrandID     int                `json:"brandId"`
	Brand       *Brand             `json:"brand"`
	LeagueID    int                `json:"leagueId"`
	TeamID      int                `json:"teamID"`
	Team        *TeamM             `json:"team"`
	PlayerID    int                `json:"playerId"`
}
type ProductAttrs struct {
	Code          string `json:"code"`
	Name          string `json:"name"`
	TranslationID string `json:"translationId"`
}
type ProductSize struct {
	CreatedAt *time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
	Name      string     `json:"name"`
}
type ProductVariation struct {
	ID                int         `json:"id"`
	CreatedAt         *time.Time  `json:"createdAt"`
	UpdatedAt         *time.Time  `json:"updatedAt"`
	DeletedAt         *time.Time  `json:"deletedAt"`
	ProductID         int         `json:"productId"`
	Product           Product     `json:"product"`
	BadgeID           *int        `json:"badgeId"`
	Badge             Badge       `json:"badge"`
	SizeID            int         `json:"sizeId"`
	Size              ProductSize `json:"size"`
	SKU               string      `json:"SKU"`
	AvailableQuantity int         `json:"availableQuantity"`
}
type Round struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	LeagueID int    `json:"leagueId"`
	SeasonID int    `json:"seasonId"`
	Start    string `json:"start"`
	End      string `json:"end"`
}
type SeasonsRoot struct {
	Seasons []Season `json:"seasons"`
}
type Shots struct {
	ShotsTotal  int `json:"shotsTotal"`
	ShotsOnGoal int `json:"shotsOnGoal"`
}
type SquadRoot struct {
	Players []PlayerSquadStats `json:"players"`
}
type Stage struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	LeagueID int    `json:"leagueId"`
	SeasonID int    `json:"seasonId"`
}
type Standing struct {
	Position    int           `json:"position"`
	TeamID      int           `json:"teamId"`
	TeamName    string        `json:"teamName"`
	GroupID     int           `json:"groupId"`
	GroupName   string        `json:"groupName"`
	Overall     StandingStats `json:"overall"`
	Home        StandingStats `json:"home"`
	Away        StandingStats `json:"away"`
	Total       Total         `json:"total"`
	Result      string        `json:"result"`
	Points      int           `json:"points"`
	RecentForm  string        `json:"recentForm"`
	Status      string        `json:"status"`
	TeamInclude TeamInclude   `json:"teamInclude"`
}
type StandingStats struct {
	GamesPlayed  int `json:"gamesPlayed"`
	Won          int `json:"won"`
	Draw         int `json:"draw"`
	Lost         int `json:"lost"`
	GoalsScored  int `json:"goalsScored"`
	GoalsAgainst int `json:"goalsAgainst"`
}
type StartingAt struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type StatsCalendar struct {
	DateParameter     string               `json:"dateParameter"`
	DateParameterTime time.Time            `json:"dateParameterTime"`
	Entries           []StatsCalendarEntry `json:"entries"`
	TodayOrNextIdx    int                  `json:"todayOrNextIdx"`
	Count             int                  `json:"count"`
}
type StatsCalendarEntry struct {
	DateTime    *time.Time `json:"dateTime"`
	NextOrToday bool       `json:"nextOrToday"`
	Idx         int        `json:"idx"`
}
type StatsInclude struct {
	Stats PlayerStats `json:"stats"`
}
type Team struct {
	ID           int    `json:"id"`
	LegacyID     int    `json:"legacyId"`
	CountryID    int    `json:"countryId"`
	Name         string `json:"name"`
	ShortCode    string `json:"shortCode"`
	NationalTeam bool   `json:"nationalTeam"`
	Founded      int    `json:"founded"`
	LogoPath     string `json:"logoPath"`
	VenueID      int    `json:"venueId"`
}
type TeamArray struct {
	Key   string  `json:"key"`
	Value []TeamM `json:"value"`
}
type TeamInclude struct {
	Team Team `json:"team"`
}
type TeamM struct {
	CreatedAt   *time.Time    `json:"createdAt"`
	UpdatedAt   *time.Time    `json:"updatedAt"`
	DeletedAt   *time.Time    `json:"deletedAt"`
	Name        string        `json:"name"`
	LeaguesM    []LeagueM     `json:"leaguesM"`
	Collections []CollectionM `json:"collections"`
	BrandID     int           `json:"brandID"`
	Brand       Brand         `json:"brand"`
	Logo        string        `json:"logo"`
}
type Topscorer struct {
	Position      int           `json:"position"`
	PlayerID      int           `json:"playerId"`
	TeamID        int           `json:"teamId"`
	StageID       int           `json:"stageId"`
	Goals         int           `json:"goals"`
	PenaltyGoals  int           `json:"penaltyGoals"`
	PlayerInclude PlayerInclude `json:"playerInclude"`
	TeamInclude   TeamInclude   `json:"teamInclude"`
}
type Total struct {
	Points int `json:"points"`
}
type TransfersInclude struct {
	Transfers PlayerTransfers `json:"transfers"`
}
type User struct {
	ID              int        `json:"ID"`
	CreatedAt       *time.Time `json:"createdAt"`
	UpdatedAt       *time.Time `json:"updatedAt"`
	DeletedAt       *time.Time `json:"deletedAt"`
	Email           string     `json:"email"`
	FirebaseI       string     `json:"firebaseI"`
	UserAccessLevel int        `json:"userAccessLevel"`
	Telephone       string     `json:"telephone"`
}
type VisitorTeamInclude struct {
	VisitorTeam Team `json:"visitorTeam"`
}
type PlayerRoot struct {
	Player Player `json:"player"`
}