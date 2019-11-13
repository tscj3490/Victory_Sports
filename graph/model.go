package graph

type League struct {
	ID              int    `json:"id"`
	LegacyID        int    `json:"legacyId"`
	CountryID       int    `json:"countryId"`
	Name            string `json:"name"`
	IsCup           bool   `json:"isCup"`
	CurrentSeasonID int    `json:"currentSeasonId"`
	CurrentRoundID  int    `json:"currentRoundId"`
	CurrentStageID  int    `json:"currentStageId"`
	LiveStandings   bool   `json:"liveStandings"`
	CoverageID      int
	SeasonsRootID   int
}

type Season struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	LeagueID          int    `json:"leagueId"`
	FixturesIncludeID int
}
