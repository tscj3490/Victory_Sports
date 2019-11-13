package gosportmonks

import (
	"fmt"
	"log"
	"net/http"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/gosportmonks/context"
)

const topscorersBasePath = "v2.0/topscorers"

// TopscorersService is an interface for fetching soccer topscorers from Sportmonks API.
// See: https://www.sportmonks.com/products/soccer/docs/2.0/topscorers/25
type TopscorersService interface {
	List(ctx context.Context, seasonID int, opt *ListOptions) ([]Topscorer, *Response, error)
}

// TopscorerServiceOp handles communication with the topscorer related methods of the
// Sportmonks API
type TopscorersServiceOp struct {
	client *Client
}

var _ TopscorersService = &TopscorersServiceOp{}

type TopscorersInSeason struct {
	SeasonID   uint   `json:"id"`
	SeasonName string `json:"name"`
	LeagueID   uint   `json:"league_id"`
	// the below work, but they change a lot ... it's better to only reflect this in league
	TopscorersInclude struct {
		Topscorers []Topscorer `json:"data"`
	} `json:"goalscorers,omitempty"`
}
type Topscorer struct {
	Position      uint `json:"position"`
	PlayerID      uint `json:"player_id"`
	TeamID        uint `json:"team_id"`
	StageID       uint `json:"stage_id"`
	Goals         uint `json:"goals"`
	PenaltyGoals  uint `json:"penalty_goals"`
	PlayerInclude struct {
		Player Player `json:"data"`
	} `json:"player"`
	TeamInclude struct {
		Team Team `json:"data"`
	} `json:"team"`
}

type TopscorersRoot struct {
	Topscorers TopscorersInSeason `json:"data"`
}

// List all topscorers by topscorer - topscorer is a combined key, the has league and date frame
func (s TopscorersServiceOp) List(ctx context.Context, seasonID int, opt *ListOptions) ([]Topscorer, *Response, error) {
	if seasonID <= 0 {
		return nil, nil, NewArgError("seasonID", "cannot be/or less than 0")
	}

	path := fmt.Sprintf("%s/season/%v", topscorersBasePath, seasonID)
	if opt == nil {
		// append default includes unless told otherwise
		opt = &ListOptions{
			Include: "goalscorers.player,goalscorers.team",
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

	root := new(TopscorersRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Topscorers.TopscorersInclude.Topscorers, resp, err
}
