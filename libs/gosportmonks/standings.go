package gosportmonks

import (
	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/gosportmonks/context"
	"log"
	"net/http"
	"fmt"
)

const standingsBasePath = "v2.0/standings"

// StandingsService is an interface for fetching soccer standings from Sportmonks API.
// See: https://www.sportmonks.com/products/soccer/docs/2.0/standings/25
type StandingsService interface {
	List(ctx context.Context, seasonID uint,opt *ListOptions) ([]Standing, *Response, error)
}

// StandingServiceOp handles communication with the standing related methods of the
// Sportmonks API
type StandingsServiceOp struct {
	client *Client
}

var _ StandingsService = &StandingsServiceOp{}

type StandingsInSeason struct {
	SeasonID       uint   `json:"season_id"`
	SeasonName     string `json:"name"`
	LeagueID uint   `json:"league_id"`
	StandingsInclude struct {
		Standings []Standing `json:"data"`
	} `json:"standings"`
}

type StandingStats struct {
	GamesPlayed uint `json:"games_played"`
	Won uint `json:"won"`
	Draw uint `json:"draw"`
	Lost uint `json:"lost"`
	GoalsScored uint `json:"goals_scored"`
	GoalsAgainst uint `json:"goals_against"`
}
type Standing struct {
	Position uint `json:"position"`
	TeamID uint `json:"team_id"`
	TeamName string `json:"team_name"`
	GroupID uint `json:"group_id"`
	GroupName string `json:"group_name"`
	Overall StandingStats
	Home StandingStats
	Away StandingStats
	Total struct {
		GoalDifference interface{} `json:"goal_difference"`
		Points uint `json:"points"`
	} `json:"total"`
	Result string `json:"result"`
	Points uint `json:"points"`
	RecentForm string `json:"recent_form"`
	Status string `json:"status"`
	TeamInclude struct {
		Team Team `json:"data"`
	} `json:"team"`
}

type standingsRoot struct {
	Standings []StandingsInSeason `json:"data"`
}

// List all standings by season - season is a combined key, the has league and date frame
func (s StandingsServiceOp) List(ctx context.Context, seasonID uint, opt *ListOptions) ([]Standing, *Response, error) {
	if seasonID <= 0 {
		return nil, nil, NewArgError("standingID", "cannot be/or less than 0")
	}

	path := fmt.Sprintf("%s/season/%v", standingsBasePath, seasonID)
	if opt == nil {
		// append default includes unless told otherwise
		opt = &ListOptions{
			Include: "standings.team",
		}
	}
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

	root := new(standingsRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	// TODO: support group stages
	standings := []Standing{}
	for _, s := range root.Standings {
		standings = s.StandingsInclude.Standings
		break
	}

	return standings, resp, err
}

