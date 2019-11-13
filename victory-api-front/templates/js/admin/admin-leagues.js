import {
    UserAccessLevels,
    ProductGender,
    ProductSizes,
    ProductBrands,
    ProductCollections,
    ProductKits,
} from './admin-data';

/*
        "country_id": 99474,
        "name": "World Cup",
        "is_cup": true,
        "current_season_id": 892,
        "current_round_id": 0,
        "current_stage_id": 1731,
 */

export function NGAdminLeagues (nga, allEntities /* all other entities */) {
    let leagues = allEntities.leagues;

    let commonFields = [
        nga.field('Name.en').isDetailLink(true),
        nga.field('Name.ar'),
        nga.field('StatsLeagueID'),
    ];

    leagues.listView().fields(commonFields.concat([
        nga.field('ID'),
    ]));
    leagues.creationView().fields(commonFields.concat([
    ]));
    leagues.editionView().fields(commonFields.concat([
        nga.field('ID'),
    ]));
    // leagues.deletionView();
}
