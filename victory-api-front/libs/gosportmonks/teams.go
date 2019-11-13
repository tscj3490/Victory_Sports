package gosportmonks

import (
	"fmt"
	"log"
	"net/http"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/gosportmonks/context"
)

const teamsBasePath = "v2.0/teams"

// TeamsService is an interface for fetching soccer teams from Sportmonks API.
// See: https://www.sportmonks.com/products/soccer/docs/2.0/teams/25
type TeamsService interface {
	List(ctx context.Context, seasonID int, opt *ListOptions) ([]Team, *Response, error)
	Get(context.Context, uint) (*Team, *Response, error)
}

// TeamServiceOp handles communication with the team related methods of the
// Sportmonks API
type TeamsServiceOp struct {
	client *Client
}

var _ TeamsService = &TeamsServiceOp{}

type Team struct {
	ID           uint   `json:"id"`
	LegacyID     uint   `json:"legacy_id"`
	CountryID    uint   `json:"country_id"`
	Name         string `json:"name"`
	ShortCode    string `json:"short_code"`
	NationalTeam bool   `json:"national_team"`
	Founded      uint   `json:"founded"`
	LogoPath     string `json:"logo_path"`
	VenueID      uint   `json:"venue_id"`
}

type teamRoot struct {
	Team *Team `json:"team"`
}
type teamsRoot struct {
	Teams []Team `json:"data"`
}

// List all teams by season - season is a combined key, the has league and date frame
func (s TeamsServiceOp) List(ctx context.Context, seasonID int, opt *ListOptions) ([]Team, *Response, error) {
	if seasonID <= 0 {
		return nil, nil, NewArgError("teamID", "cannot be/or less than 0")
	}

	path := fmt.Sprintf("%s/season/%v", teamsBasePath, seasonID)
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

	root := new(teamsRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Teams, resp, err
}

// Get individual team. It requires a non-empty team id
func (s *TeamsServiceOp) Get(ctx context.Context, teamID uint) (*Team, *Response, error) {
	if teamID <= 0 {
		return nil, nil, NewArgError("teamID", "cannot be/or less than 0")
	}

	path := fmt.Sprintf("%s/%v", teamsBasePath, teamID)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(teamRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Team, resp, err
}
