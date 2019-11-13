package gosportmonks

import (
	"fmt"
	"log"
	"net/http"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/gosportmonks/context"
)

const stagesBasePath = "v2.0/stages"

// StagesService is an interface for fetching soccer stages from Sportmonks API.
// See: https://www.sportmonks.com/products/soccer/docs/2.0/stages/25
type StagesService interface {
	List(ctx context.Context, seasonID uint, opt *ListOptions) ([]Stage, *Response, error)
	Get(ctx context.Context, stageID uint) (*Stage, *Response, error)
}

// StageServiceOp handles communication with the stage related methods of the
// Sportmonks API
type StagesServiceOp struct {
	client *Client
}

var _ StagesService = &StagesServiceOp{}

/*

id: 120771
name: Group Stage
type: null // not sure what type looks like
league_id: 1085
season_id: 11697

*/

/*
Can include: fixtures,results,season,league`
*/

type Stage struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	LeagueID uint   `json:"league_id"`
	SeasonID uint   `json:"season_id"`
}

type stageRoot struct {
	Stage *Stage `json:"data"`
}
type stagesRoot struct {
	Stages []Stage `json:"data"`
}

// List all stages by season - season is a combined key, the has league and date frame
func (s StagesServiceOp) List(ctx context.Context, seasonID uint, opt *ListOptions) ([]Stage, *Response, error) {
	if seasonID <= 0 {
		return nil, nil, NewArgError("stageID", "cannot be/or less than 0")
	}

	path := fmt.Sprintf("%s/season/%v", stagesBasePath, seasonID)
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

	root := new(stagesRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Stages, resp, err
}

// Get individual stage. It requires a non-empty stage id
func (s *StagesServiceOp) Get(ctx context.Context, stageID uint) (*Stage, *Response, error) {
	if stageID <= 0 {
		return nil, nil, NewArgError("stageID", "cannot be/or less than 0")
	}

	path := fmt.Sprintf("%s/%v", stagesBasePath, stageID)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(stageRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Stage, resp, err
}
