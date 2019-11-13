package gosportmonks

import (
	"fmt"
	"log"
	"net/http"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/gosportmonks/context"
)

const livescoresBasePath = "v2.0/livescores"
const livescoresNowBasePath = "v2.0/livescores/now"

// LivescoresService is an interface for fetching soccer fixtures from Sportmonks API.
// See: https://www.sportmonks.com/products/soccer/docs/2.0/livescores/25
type LivescoresService interface {
	ListToday(ctx context.Context, opt *ListOptions) ([]LivescoreFixture, *Response, error)
	ListNow(ctx context.Context, opt *ListOptions) ([]LivescoreFixture, *Response, error)
}

// LivescoresServiceOp handles communication with the team related methods of the Sportmonks API
type LivescoresServiceOp struct {
	client *Client
}

// WeatherReport The live variation with different fields
type WeatherReport struct {
	Code        string `json:"code,omitempty"`
	Type        string `json:"type,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Temperature struct {
		Temp float32 `json:"temp,omitempty"`
		Unit string  `json:"unit,omitempty"`
	} `json:"temperature,omitempty"`
	Clouds   string `json:"clouds"`
	Humidity string `json:"humidity"`
	Wind     struct {
		Speed  string  `json:"speed,omitempty"`
		Degree float32 `json:"degree,omitempty"`
	} `json:"wind,omitempty"`
}

// Colors The live socket variation with different fields
type Colors struct {
	LocalTeam struct {
		Color     string `json:"color,omitempty"`
		KitColors string `json:"kit_colors,omitempty"`
	} `json:"localteam,omitempty"`
	VisitorTeam struct {
		Color     string `json:"color,omitempty"`
		KitColors string `json:"kit_colors,omitempty"`
	} `json:"visitorteam,omitempty"`
}

// LivescoresFixtureEvent The live variation with additional fields
type LivescoresFixtureEvent struct {
	FixtureEvent
}

// LivescoreFixture The live socket endpoint variation of Fixtures with additional fields etc.
type LivescoreFixture struct {
	Fixture
	LivescoresFixtureInclude struct {
		Events []LivescoresFixtureEvent `json:"data,omitempty"`
	} `json:"events,omitempty"`
	AggregateID           uint          `json:"aggregate_id,omitempty"`
	RefereeIDID           uint          `json:"referee_id,omitempty"`
	WeatherReport         WeatherReport `json:"weather_report,omitempty"`
	Commentaries          bool          `json:"commentaries,omitempty"`
	Attendance            uint          `json:"attendance,omitempty"`
	Pitch                 string        `json:"pitch,omitempty"`
	WinningOddsCalculated bool          `json:"winning_odds_calculated,omitempty"`
	Leg                   string        `json:"leg,omitempty"`
	Colors                Colors        `json:"colors,omitempty"`
}

var _ LivescoresService = &LivescoresServiceOp{}

type livescoresRoot struct {
	LivescoreFixtures []LivescoreFixture `json:"data"`
}

// ListToday List all fixtures by season - season is a combined key, the has league and date frame
func (s LivescoresServiceOp) ListToday(ctx context.Context, opt *ListOptions) ([]LivescoreFixture, *Response, error) {

	path := fmt.Sprintf("%s/", livescoresBasePath)
	path, err := addOptions(path, opt)
	if err != nil {
		return nil, nil, err
	}

	log.Printf("TSO: path %v", path)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("TSO: req: %v", req.URL.String())

	root := new(livescoresRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.LivescoreFixtures, resp, err
}

// ListNow List all fixtures by season - season is a combined key, the has league and date frame
func (s LivescoresServiceOp) ListNow(ctx context.Context, opt *ListOptions) ([]LivescoreFixture, *Response, error) {

	path := fmt.Sprintf("%s/", livescoresNowBasePath)
	path, err := addOptions(path, opt)
	if err != nil {
		return nil, nil, err
	}

	log.Printf("TSO: path %v", path)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("TSO: req: %v", req.URL.String())

	root := new(livescoresRoot)

	resp, err := s.client.Do(ctx, req, root)

	if err != nil {
		return nil, resp, err
	}

	return root.LivescoreFixtures, resp, err
}
