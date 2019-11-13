import {
    UserAccessLevels,
    ProductGender,
    ProductSizes,
    ProductBrands,
    ProductCollections,
    ProductKits,
    StatsBuyButtonStates,
} from './admin-data';

export function NGAdminTeams (nga, allEntities /* all other entities */) {
    let teams = allEntities.teams;
    let leagues = allEntities.leagues;
    let sportmonksTeams = allEntities.sportmonksTeams;
    // teams
    teams.listView().fields([
        nga.field('ID'),
        nga.field('Logo', 'template')
            .label('Logo')
            .template(`<img
                src="{{ entry.values.Logo }}"
                class="club-logo"
                alt="{{ entry.values.Logo }}"/>`),
        nga.field('Name.en').isDetailLink(true),
        nga.field('Name.ar'),
        nga.field('StatsBuyButtonState','choice')
            .choices(StatsBuyButtonStates)
            .label('Buy Button'),
        // nga.field('Leagues','reference_many')
        //     .targetEntity(leagues)
        //     .targetField(nga.field('Name.en')),
    ]);
    teams.creationView().fields([
        nga.field('Name.en'),
        nga.field('Name.ar'),
        nga.field('EnableStatsBuyButton','boolean').label('Show Buy Button'),
        nga.field('StatsTeamIDCombinedKey', 'reference').label('StatsTeam')
            .targetEntity(sportmonksTeams)
            .targetField(nga.field('name')),
        nga.field('Leagues','reference_many')
            .targetEntity(leagues)
            .targetField(nga.field('Name.en'))
            // .remoteComplete(true, {
            //     refreshDelay: 200,
            //     searchQuery: function(search) { return { q: search }; }
            // }),
        // nga.field('LeagueID', 'number')
        // nga.field('LeagueID', 'reference')
        //     .label('League')
        //     .targetEntity(leagues)
        //     .targetField('ID')
        //     .sortField('Name')
        //     .sortDir('ASC')
        //     .validation({ required: true })
    ]);
    teams.editionView().fields(teams.listView().fields().concat([
        nga.field('StatsTeamIDCombinedKey', 'reference').label('StatsTeam')
            .targetEntity(sportmonksTeams)
            .targetField(nga.field('name')),
        nga.field('Leagues','reference_many')
            .targetEntity(leagues)
            .targetField(nga.field('Name.en')),
        nga.field('CreatedAt', 'datetime'),
        nga.field('UpdatedAt', 'datetime'),
        nga.field('DeletedAt', 'datetime')
    ]));
    // teams.deletionView();
}
