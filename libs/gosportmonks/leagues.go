package gosportmonks

import (
	"fmt"
	"log"
	"net/http"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/gosportmonks/context"
)

const leaguesBasePath = "v2.0/leagues"

// LeaguesService is an interface for fetching soccer leagues from Sportmonks API.
// See: https://www.sportmonks.com/products/soccer/docs/2.0/leagues/16
type LeaguesService interface {
	List(context.Context, *ListOptions) ([]League, *Response, error)
	Get(context context.Context, competitionID int) (*League, *Response, error)
	GetWithOptions(context context.Context, competitionID int, opt *ListOptions) (*League, *Response, error)
}

// LeaguesServiceOp handles communication with the league related methods of the
type LeaguesServiceOp struct {
	client *Client
}

var _ LeaguesService = &LeaguesServiceOp{}

/*
   "id": 271,
   "legacy_id": 43,
   "country_id": 320,
   "name": "Superliga",
   "is_cup": false,
   "current_season_id": 6361,
   "current_round_id": 144937,
   "current_stage_id": 48049,
   "live_standings": true,
   "coverage": {
     "topscorer_goals": true,
     "topscorer_assists": true,
     "topscorer_cards": true
   }
*/

// League represents a Sportmonks League
type League struct {
	ID              uint   `json:"id"`
	LegacyID        uint   `json:"legacy_id"`
	CountryID       uint   `json:"country_id"`
	Name            string `json:"name"`
	IsCup           bool   `json:"is_cup"`
	CurrentSeasonID uint   `json:"current_season_id"`
	CurrentRoundID  uint   `json:"current_round_id"`
	CurrentStageID  uint   `json:"current_stage_id"`
	LiveStandings   bool   `json:"live_standings"`
	Coverage        struct {
		TopscorerGoals   bool `json:"topscorer_goals"`
		TopscorerAssists bool `json:"topscorer_assists"`
		TopscorerCards   bool `json:"topscorer_cards"`
	} `json:"coverage"`
	SeasonsInclude *SeasonsRoot `json:"seasons"`
	CountryInclude *CountryRoot `json:"country"`
}

/*
id: 320
name: Denmark
extra: {
	continent: Europe
	sub_region: Northern Europe
	world_region: EMEA
	fifa: DEN
	iso: DNK
	longitude: 9.555907249450684
	latitude: 56.10176086425781
	flag: <svg xmlns="http://www.w3.org/2000/svg" width="370" height="280"><path fill="#c60c30" d="M0 0h370v280h-370z"/><path fill="#fff" d="M120 0h40v280h-40zM0 120h370v40h-370z"/></svg>
}
*/

// CountryRoot the packaged country object
type CountryRoot struct {
	Country Country `json:"data"`
}

// Country the Embedded Country Data for each league
type Country struct {
	ID    uint         `json:"id"`
	Name  string       `json:"name"`
	Extra CountryExtra `json:"extra"`
}

// CountryExtra the Embedded detail in each country object
type CountryExtra struct {
	Continent   string `json:"continent"`
	SubRegion   string `json:"sub_region"`
	WorldRegion string `json:"world_region"`
	Fifa        string `json:"fifa"`
	ISO         string `json:"ISO"`
	Longitude   string `json:"longitude"`
	Latitude    string `json:"latitude"`
	Flag        string `json:"flag"`
}

/*
{
  "data": [
    {
LEAGUE ...
    },
LEAGUE ...
    }
  ],
  "meta": {
    "subscription": {
      "started_at": {
        "date": "2017-03-20 21:02:56.000000",
        "timezone_type": 3,
        "timezone": "UTC"
      },
      "trial_ends_at": null,
      "ends$at": null
    },
    "plan": {
      "name": "Free Plan",
      "price": "0.00",
      "request_limit": "3,1"
    },
    "sports": [
      {
        "id": 1,
        "name": "Soccer",
        "current": true
      }
    ],
    "pagination": {
      "total": 2,
      "count": 2,
      "per_page": 10,
      "current_page": 1,
      "total_pages": 1,
      "links": []
    }
  }
}
*/

// leagueRoot represents a response from the Sportmonk API
type leagueRoot struct {
	League *League `json:"data"`
}

type leaguesRoot struct {
	Leagues []League `json:"data"`
}

func (l League) String() string {
	return Stringify(l)
}

// List all leagues
func (s LeaguesServiceOp) List(ctx context.Context, opt *ListOptions) ([]League, *Response, error) {
	path := leaguesBasePath
	path, err := addOptions(path, opt)
	if err != nil {
		return nil, nil, err
	}

	log.Printf("LSO: path %v", path)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("LSO: req: %v", req.URL.String())

	root := new(leaguesRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Leagues, resp, err
}

// Get individual league. It requires a non-empty league name
func (s *LeaguesServiceOp) Get(ctx context.Context, competitionID int) (*League, *Response, error) {
	if competitionID < 1 {
		return nil, nil, NewArgError("competitionID", "cannot be undefined or 0")
	}

	path := fmt.Sprintf("%s/%v", leaguesBasePath, competitionID)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(leagueRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.League, resp, err
}

// GetWithOptions - Retrieve the league and it's last 3 seasons, newest first
func (s *LeaguesServiceOp) GetWithOptions(ctx context.Context, competitionID int, opt *ListOptions) (*League, *Response, error) {
	if competitionID < 1 {
		return nil, nil, NewArgError("competitionID", "cannot be undefined or 0")
	}

	path := fmt.Sprintf("%s/%v", leaguesBasePath, competitionID)
	path, err := addOptions(path, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(leagueRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.League, resp, err
}
