package gosportmonks

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"sort"
	"strconv"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/gosportmonks/context"
)

const fixturesBasePath = "v2.0/fixtures"

// FixturesService is an interface for fetching soccer fixtures from Sportmonks API.
// See: https://www.sportmonks.com/products/soccer/docs/2.0/fixtures/25
type FixturesService interface {
	List(ctx context.Context, from time.Time, to time.Time, opt *ListOptions) ([]Fixture, *Response, error)
	Get(context.Context, uint) (*Fixture, *Response, error)
}

// FixtureServiceOp handles communication with the fixture related methods of the
// Sportmonks API
type FixturesServiceOp struct {
	client *Client
}

var _ FixturesService = &FixturesServiceOp{}

/*
	"id": 8903205,
	"league_id": 959,
	"season_id": 8956,
	"stage_id": 56928,
	"round_id": 141132,
	"group_id": null,
	"aggregate_id": null,
	"venue_id": 13474,
	"referee_id": null,
	"localteam_id": 6844,
	"visitorteam_id": 91,
	"weather_report": {
		"code": "clouds",
		"type": "few clouds",
		"icon": "https://cdn.sportmonks.com/images/weather/02d.png",
		"temperature": {
			"temp": 72.39,
			"unit": "fahrenheit"
		},
		"clouds": "20%",
		"humidity": "43%",
		"wind": {
			"speed": "14.99 m/s",
			"degree": 300
		}
	},
	"commentaries": null,
	"attendance": null,
	"pitch": "Good",
	"winning_odds_calculated": true,
	"formations": {
		"localteam_formation": null,
		"visitorteam_formation": null
	},
	"scores": {
		"localteam_score": 0,
		"visitorteam_score": 2,
		"localteam_pen_score": null,
		"visitorteam_pen_score": null,
		"ht_score": "0-1",
		"ft_score": "0-2",
		"et_score": null
	},
	"time": {
		"status": "FT",
		"starting_at": {
			"date_time": "2018-02-02 13:15:00",
			"date": "2018-02-02",
			"time": "13:15:00",
			"timestamp": 1517577300,
			"timezone": "UTC"
		},
		"minute": 90,
		"second": null,
		"added_time": null,
		"extra_minute": null,
		"injury_time": null
	},
	"coaches": {
		"localteam_coach_id": 1552789,
		"visitorteam_coach_id": 1552219
	},
	"standings": {
		"localteam_position": 10,
		"visitorteam_position": 7
	},
	"deleted": false,
	"events": {
		"data": [
			{
				"id": 42790327,
				"team_id": "6380",
				"type": "yellowcard",
				"fixture_id": 5658815,
				"player_id": 56082,
				"player_name": "A. Ropotan",
				"related_player_id": null,
				"related_player_name": null,
				"minute": 44,
				"extra_minute": null,
				"reason": null,
				"injuried": null,
				"result": null
			},
			{
				"id": 42790624,
				"team_id": "6380",
				"type": "substitution",
				"fixture_id": 5658815,
				"player_id": 308254,
				"player_name": "N. Obaid",
				"related_player_id": 307904,
				"related_player_name": "L. Alnofall",
				"minute": 63,
				"extra_minute": null,
				"reason": null,
				"injuried": null,
				"result": null
			}
		]
	}
*/

/*
Can include: fixtures,results,season,league`
*/
type FixtureCoaches struct {
	LocalTeamCoachID   uint `json:"localteam_coach_id"`
	VisitorTeamCoachID uint `json:"visitorteam_coach_id"`
}
type FixtureStandings struct {
	LocalTeamPosition   uint `json:"localteam_position"`
	VisitorTeamPosition uint `json:"visitorteam_position"`
}
type FixtureTimeStartingAt struct {
	DateTime  string `json:"date_time"`
	Date      string `json:"date"`
	Time      string `json:"time"`
	Timestamp int    `json:"timestamp"`
	Timezone  string `json:"timezone"`
}
type FixtureTime struct {
	Status     string                `json:"status"`
	StartingAt FixtureTimeStartingAt `json:"starting_at"`
	Minute     uint                  `json:"minute"`
	Second     *uint                 `json:"second"`
	AddedTime  *uint                 `json:"added_time"`
	ExtraTime  *uint                 `json:"extra_minute"`
	InjuryTime *uint                 `json:"injury_time"`
}
type FixtureEvent struct {
	ID                uint   `json:"id"`
	TeamID            string `json:"team_id"`
	Type              string `json:"type"`
	FixtureID         uint   `json:"fixture_id"`
	PlayerID          uint   `json:"player_id"`
	PlayerName        string `json:"player_name"`
	RelatedPlayerID   uint   `json:"related_player_id"`
	RelatedPlayerName string `json:"related_player_name"`
	Minute            uint   `json:"minute"`
	ExtraMinute       uint   `json:"extra_minute"`
	Reason            string `json:"reason"`
	Injuried          bool   `json:"injuried"`
	Result            string `json:"result"`
}

type FixturePlayerStats struct {
	Shots struct {
		ShotsTotal  uint `json:"shots_total"`
		ShotsOnGoal uint `json:"shots_on_goal"`
	} `json:"shots"`
	Goals struct {
		Scored   uint `json:"scored"`
		Conceded uint `json:"conceded"`
	} `json:"goals"`
	Fouls struct {
		Drawn     uint `json:"drawn"`
		Committed uint `json:"comitted"`
	} `json:"fouls"`
	Cards struct {
		YellowCards uint `json:"yellowcards"`
		RedCards    uint `json:"redcards"`
	} `json:"cards"`
	Passing struct {
		TotalCrosses    uint `json:"total_crosses"`
		CrossesAccuracy uint `json:"crosses_accuracy"`
		Passes          uint `json:"passes"`
		PassesAccuracy  uint `json:"passes_accuracy"`
	} `json:"passing"`
	Other struct {
		Assists       uint `json:"assists"`
		Offsides      uint `json:"offsides"`
		Saves         uint `json:"saves"`
		PenScored     uint `json:"pen_scored"`
		PenMissed     uint `json:"pen_missed"`
		PenSaved      uint `json:"pen_saved"`
		PenCommitted  uint `json:"pen_committed"`
		PenWon        uint `json:"pen_won"`
		HitWoodwork   uint `json:"hit_woodwork"`
		Tackles       uint `json:"tackles"`
		Blocks        uint `json:"blocks"`
		Interceptions uint `json:"interceptions"`
		Clearances    uint `json:"clearances"`
		MinutesPlayed uint `json:"minutes_played"`
	} `json:"other"`
}
type FixturePlayer struct {
	TeamID            uint               `json:"team_id"`
	FixtureID         uint               `json:"fixture_id"`
	PlayerID          uint               `json:"player_id"`
	PlayerName        string             `json:"player_name"`
	Number            uint               `json:"number"`
	Position          string             `json:"position"`
	FormationPosition uint               `json:"formation_position"`
	PosX              uint               `json:"posx"`
	PosY              uint               `json:"posy"`
	Stats             FixturePlayerStats `json:"stats"`
}

/*
	"date_time": "2018-02-02 13:15:00",
	"date": "2018-02-02",
	"time": "13:15:00",
	"timestamp": 1517577300,
	"timezone": "UTC"
*/
var (
	LocUTC   *time.Location
	LocDubai *time.Location
)

func init() {
	locUTC, err := time.LoadLocation("UTC")
	locDubai, err := time.LoadLocation("Asia/Dubai")
	if err != nil {
		panic(fmt.Sprintf("Fixtures.go not able to time.LoadLocation: %v", err))
	}
	LocUTC = locUTC
	LocDubai = locDubai
}

func (f FixtureTime) GetStartTime() *time.Time {
	startingAt := f.StartingAt.DateTime
	t, err := time.ParseInLocation("2006-01-02 15:04:05", startingAt, LocUTC) // parse as UTC
	if err != nil {
		log.Printf("FixtureTime.GetStartTime err: %v", err)
		return nil
	}
	t = t.In(LocDubai)
	return &t
}

type FixtureScore struct {
	LocalTeamScore      uint   `json:"localteam_score"`
	VisitorTeamScore    uint   `json:"visitorteam_score"`
	LocalTeamScorePen   uint   `json:"localteam_pen_score"`
	VisitorTeamScorePen uint   `json:"visitorteam_pen_score"`
	HalfTimeScore       string `json:"ht_score"`
	FullTimeScore       string `json:"ft_score"`
	ExtraTimeScore      string `json:"et_score"`
}
type FixtureFormation struct {
	LocalTeamFormation   string `json:"localteam_formation"`
	VisitorTeamFormation string `json:"visitorteam_formation"`
}

type Fixture struct {
	ID               uint             `json:"id"`
	LeagueID         uint             `json:"league_id"`
	SeasonID         uint             `json:"season_id"`
	StageID          uint             `json:"stage_id"`
	RoundID          uint             `json:"round_id"`
	GroupID          uint             `json:"group_id"`
	VenueID          uint             `json:"venue_id"`
	LocalTeamID      uint             `json:"localteam_id"`
	VisitorTeamID    uint             `json:"visitorteam_id"`
	Formation        FixtureFormation `json:"formation"`
	Scores           FixtureScore     `json:"scores"`
	Time             FixtureTime      `json:"time"`
	Coaches          FixtureCoaches   `json:"coaches"`
	Standings        FixtureStandings `json:"standings"`
	Deleted          bool             `json:"deleted"`
	LocalTeamInclude struct {
		LocalTeam Team `json:"data,omitempty"`
	} `json:"localTeam,omitempty"`
	VisitorTeamInclude struct {
		VisitorTeam Team `json:"data,omitempty"`
	} `json:"visitorTeam,omitempty"`
	EventsInclude struct {
		Events []FixtureEvent `json:"data,omitempty"`
	} `json:"events,omitempty"`
	LineupInclude struct {
		Players []FixturePlayer `json:"data,omitempty"`
	} `json:"lineup"`
}

type fixtureRoot struct {
	Fixture *Fixture `json:"data"`
}
type fixturesRoot struct {
	Fixtures []Fixture `json:"data"`
}

func (f Fixture) GetEventsFilteredBy(teamID int) []FixtureEvent {
	events := []FixtureEvent{}

	for _, e := range f.EventsInclude.Events {
		eventTeamID, err := strconv.Atoi(e.TeamID)
		if err != nil {
			continue
		}
		if eventTeamID != teamID {
			continue
		}
		events = append(events, e)
	}

	// sort the events by minute
	sort.Slice(events, func(i, j int) bool {
		return events[i].Minute < events[j].Minute
	})

	return events
}

// GetLocalTeamLineup
func (f Fixture) GetLineupFormationPositionFilteredBy(teamID uint) map[int]FixturePlayer {
	lineup := map[int]FixturePlayer{}

	for _, player := range f.LineupInclude.Players {
		if player.TeamID == teamID {
			lineup[int(player.FormationPosition)] = player
		}
	}

	return lineup
}

// List all fixtures by season - season is a combined key, the has league and date frame
func (s FixturesServiceOp) List(ctx context.Context, from time.Time, to time.Time, opt *ListOptions) ([]Fixture, *Response, error) {
	if from.IsZero() {
		return nil, nil, NewArgError("from time.Time", "cannot be/or less than 0")
	}
	if to.IsZero() {
		return nil, nil, NewArgError("from time.Time", "cannot be/or less than 0")
	}

	path := fmt.Sprintf("%s/between/%v/%v", fixturesBasePath, from.Format("2006-01-02"), to.Format("2006-01-02"))
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

	root := new(fixturesRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Fixtures, resp, err
}

// ListByTeam - all fixtures by team -
func (s FixturesServiceOp) ListByTeam(ctx context.Context,
	from time.Time, to time.Time, teamID uint, opt *ListOptions) ([]Fixture, *Response, error) {
	if from.IsZero() {
		return nil, nil, NewArgError("from time.Time", "cannot be/or less than 0")
	}
	if to.IsZero() {
		return nil, nil, NewArgError("from time.Time", "cannot be/or less than 0")
	}
	if teamID <= 0 {
		return nil, nil, NewArgError("teamID", "cannot be/or less than 0")
	}

	path := fmt.Sprintf("%s/between/%v/%v/%v", fixturesBasePath, from.Format("2006-01-02"), to.Format("2006-01-02"), teamID)
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

	root := new(fixturesRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Fixtures, resp, err
}

// Get individual fixture. It requires a non-empty fixture id
func (s *FixturesServiceOp) Get(ctx context.Context, fixtureID uint) (*Fixture, *Response, error) {
	if fixtureID <= 0 {
		return nil, nil, NewArgError("fixtureID", "cannot be/or less than 0")
	}

	path := fmt.Sprintf("%s/%v", fixturesBasePath, fixtureID)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(fixtureRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Fixture, resp, err
}
