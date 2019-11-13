package gosportmonks

import (
	"fmt"
	"log"
	"net/http"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/gosportmonks/context"
)

const playersBasePath = "v2.0/players"
const squadBasePath = "v2.0/squad"

// PlayersService is an interface for fetching soccer players from Sportmonks API.
// See: https://www.sportmonks.com/products/soccer/docs/2.0/players/25
type PlayersService interface {
	Get(ctx context.Context, playerID uint, opt *ListOptions) (*Player, *Response, error)
	List(ctx context.Context, seasonID int, teamID int, opt *ListOptions) ([]PlayerSquadStats, *Response, error)
}

// PlayerServiceOp handles communication with the player related methods of the
// Sportmonks API
type PlayersServiceOp struct {
	client *Client
}

var _ PlayersService = &PlayersServiceOp{}

/*

id: 120771
name: Group Player
type: null // not sure what type looks like
league_id: 1085
season_id: 11697

*/

/*
Can include: fixtures,results,season,league`
*/

// this might need to be migrate some place else ...
type Player struct {
	PlayerID     uint   `json:"player_id"`
	TeamID       uint   `json:"team_id"`
	CountryID    uint   `json:"country_id"`
	PositionID   uint   `json:"position_id"`
	CommonName   string `json:"common_name"`
	Fullname     string `json:"fullname"`
	Firstname    string `json:"firstname"`
	Lastname     string `json:"lastname"`
	Nationality  string `json:"nationality"`
	Birthdate    string `json:"birthdate"`
	Birthcountry string `json:"birthcountry"`
	Birthplace   string `json:"birthplace"`
	Height       string `json:"height"`
	Weight       string `json:"weight"`
	ImagePath    string `json:"image_path"`
	TeamInclude  struct {
		Team Team `json:"data"`
	} `json:"team"`
	StatsInclude struct {
		Stats PlayerStats `json:"data"`
	} `json:"stats"`
	TransfersInclude struct {
		Transfers PlayerTransfers `json:"data"`
	} `json:"transfers"`
}
type PlayerStats struct {
	PlayerID          uint   `json:"player_id"`
	TeamID            uint   `json:"team_id"`
	LeagueID          uint   `json:"league_id"`
	Minutes           uint   `json:"minutes"`
	Appearences       uint   `json:"appearences"`
	Lineups           uint   `json:"lineups"`
	SubstituteIn      uint   `json:"substitute_in"`
	SubstituteOut     uint   `json:"substitute_out"`
	SubstituteOnBench uint   `json:"substitute_on_bench"`
	Goals             uint   `json:"goals"`
	YellowCards       uint   `json:"yellowcards"`
	Yellowred         uint   `json:"yellowred"`
	RedCards          uint   `json:"redcards"`
	Type              string `json:"type"`
}
type PlayerTransfers struct {
	PlayerID uint   `json:"player_id"`
	ToTeamID uint   `json:"to_team_id"`
	SeasonID uint   `json:"season_id"`
	Transfer string `json:"transfer"`
	Type     string `json:"type"`
	Date     string `json:"date"`
	Amount   string `json:"amount"`
}

type playerRoot struct {
	Player *Player `json:"data"`
}

type PlayerSquadStats struct {
	PlayerID          uint   `json:"player_id"`
	TeamID            uint   `json:"team_id"`
	LeagueID          uint   `json:"league_id"`
	Minutes           uint   `json:"minutes"`
	Appearences       uint   `json:"appearences"`
	Lineups           uint   `json:"lineups"`
	SubstituteIn      uint   `json:"substitute_in"`
	SubstituteOut     uint   `json:"substitute_out"`
	SubstituteOnBench uint   `json:"substitute_on_bench"`
	Goals             uint   `json:"goals"`
	YellowCards       uint   `json:"yellowcards"`
	Yellowred         uint   `json:"yellowred"`
	RedCards          uint   `json:"redcards"`
	Type              string `json:"type"`
	PlayerInclude  struct {
		Player Player `json:"data"`
	} `json:"player"`
	PositionInclude struct {
		Position PlayerPosition `json:"data"`
	} `json:"position"`
}
type PlayerPosition struct {
	ID uint `json:"id"`
	Name string `json:"name"`
}
type squadRoot struct {
	Players []PlayerSquadStats `json:"data"`
}

// Get individual player. It requires a non-empty player id
func (s *PlayersServiceOp) Get(ctx context.Context, playerID uint, opt *ListOptions) (*Player, *Response, error) {
	if playerID <= 0 {
		return nil, nil, NewArgError("playerID", "cannot be/or less than 0")
	}

	path := fmt.Sprintf("%s/%v", playersBasePath, playerID)
	path, err := addOptions(path, opt)
	if err != nil {
		return nil, nil, err
	}

	log.Printf("TSO: path %v", path)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(playerRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Player, resp, err
}

// List all players, and their statistics by teams and season
func (s PlayersServiceOp) List(ctx context.Context, seasonID int, teamID int, opt *ListOptions) ([]PlayerSquadStats, *Response, error) {
	if seasonID <= 0 {
		return nil, nil, NewArgError("seasonID", "cannot be/or less than 0")
	}
	if teamID <= 0 {
		return nil, nil, NewArgError("teamID", "cannot be/or less than 0")
	}

	path := fmt.Sprintf("%s/season/%v/team/%v/", squadBasePath, seasonID, teamID)
	if opt == nil {
		// append default includes unless told otherwise
		opt = &ListOptions{
			Include: "player,position",
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

	root := new(squadRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Players, resp, err
}
