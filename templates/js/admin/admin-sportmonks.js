import {
    UserAccessLevels,
    ProductGender,
    ProductSizes,
    ProductBrands,
    ProductCollections,
    ProductKits,
} from './admin-data';

export function NGAdminSportmonks (nga, allEntities /* all other entities */) {
    let teams = allEntities.sportmonksTeams;
    // teams
    /*
            "id": 65,
        "legacy_id": 376,
        "country_id": 462,
        "name": "Southampton",
        "short_code": "SOU",
        "national_team": false,
        "founded": 1885,
        "logo_path": "https://cdn.sportmonks.com/images/soccer/teams/1/65.png",
        "venue_id": 167
     */
    var commonFields = [
        nga.field('name.en').isDetailLink(true),
        nga.field('name.ar'),
        nga.field('id'),
        nga.field('Logo', 'template')
            .label('Logo')
            .template(`<img
                src="{{ entry.values.logo_path }}"
                class="club-logo"
                alt="{{ entry.values.logo_path }}"/>`),
        nga.field('short_code'),
        nga.field('national_team'),
        nga.field('combinedLeagueSeasonTeamId')
            .label('Combined Key'),
    ];
    teams.listView().fields(commonFields);
    teams.editionView().fields(commonFields.concat([
        nga.field('founded'),
        nga.field('leagueId'),
        nga.field('seasonId')
    ]));
}
