package gosportmonks

import (
	"fmt"
	"log"
	"net/http"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/gosportmonks/context"
)

const seasonsBasePath = "v2.0/seasons"

// SeasonsService is an interface for fetching soccer seasons from Sportmonks API.
// See: https://www.sportmonks.com/products/soccer/docs/2.0/seasons/25
type SeasonsService interface {
	List(ctx context.Context, opt *ListOptions) ([]Season, *Response, error)
	Get(context.Context, uint, *ListOptions) (*Season, *Response, error)
}

// SeasonServiceOp handles communication with the season related methods of the
// Sportmonks API
type SeasonsServiceOp struct {
	client *Client
}

var _ SeasonsService = &SeasonsServiceOp{}

type Season struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	LeagueID uint   `json:"league_id"`
	// the below work, but they change a lot ... it's better to only reflect this in league
	//IsCurrentSeason bool `json:"is_current_season"`
	//CurrentRoundID uint `json:"current_round_id"`
	//CurrentStageID uint `json:"current_stage_id"`
	StagesData struct {
		Stages []Stage `json:"data"`
	} `json:"stages,omitempty"`
	RoundsData struct {
		Rounds []Round `json:"data"`
	} `json:"rounds,omitempty"`
	FixturesInclude struct {
		Fixtures []Fixture `json:"data"`
	} `json:"fixtures,omitempty"`
}
type Round struct {
	ID       uint   `json:"id"`
	Name     uint   `json:"name"`
	LeagueID uint   `json:"league_id"`
	SeasonID uint   `json:"season_id"`
	Start    string `json:"start"`
	End      string `json:"end"`
}

type SeasonRoot struct {
	Season *Season `json:"data"`
}
type SeasonsRoot struct {
	Seasons []Season `json:"data"`
}

// List all seasons by season - season is a combined key, the has league and date frame
func (s SeasonsServiceOp) List(ctx context.Context, opt *ListOptions) ([]Season, *Response, error) {

	path := fmt.Sprintf("%s%s", seasonsBasePath, "")
	if opt == nil {
		// append default includes unless told otherwise
		opt = &ListOptions{
			Include: "stages,rounds",
		}
	}

	path, err := addOptions(path, opt)
	if err != nil {
		return nil, nil, err
	}

	log.Printf("SSO: path %v", path)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("SSO: req: %v", req.URL.String())

	root := new(SeasonsRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Seasons, resp, err
}

// Get individual season. It requires a non-empty season id
func (s *SeasonsServiceOp) Get(ctx context.Context, seasonID uint, opt *ListOptions) (*Season, *Response, error) {
	if seasonID <= 0 {
		return nil, nil, NewArgError("seasonID", "cannot be/or less than 0")
	}

	path := fmt.Sprintf("%s/%v", seasonsBasePath, seasonID)
	path, err := addOptions(path, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(SeasonRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Season, resp, err
}
