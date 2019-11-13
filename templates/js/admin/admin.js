/*global angular*/

let myApp = angular.module('myApp', ['ng-admin']);

import { uploader } from './uploader-ngadmin';

import {
    UserAccessLevels,
    ProductGender,
    ProductSizes,
    ProductBrands,
    ProductCollections,
    ProductKits,
} from './admin-data';

import { NGAdminBadges } from "./admin-badges";
import { NGAdminProducts } from "./admin-products";
import { NGAdminUsers } from "./admin-users";
import { NGAdminLeagues } from "./admin-leagues";
import { NGAdminTeams } from "./admin-teams";
import { NGAdminProductVariations } from "./admin-variations";
import { NGAdminSportmonks } from "./admin-sportmonks";
import { NGAdminOrders } from "./admin-orders";

myApp.directive('uploader', uploader);

myApp.config(['NgAdminConfigurationProvider', function (nga) {
    // create an admin application
    let admin = nga.application('Victory Admin')
        .debug(true)
        .baseApiUrl('/admin/'); // main API endpoint

    // create a user entity
    // the API endpoint for this entity will be 'http://jsonplaceholder.typicode.com/users/:id
    let users = nga.entity('users')
        .identifier(nga.field('ID'));
    let leagues = nga.entity('leagues')
        .identifier(nga.field('ID'));
    let teams = nga.entity('teams')
        .identifier(nga.field('ID'));
    let products = nga.entity('products')
        .identifier(nga.field('ID'));
    let collections = nga.entity('collections')
        .identifier(nga.field('ID'));
    let variations = nga.entity('variations')
        .identifier(nga.field('ID'));
    let badges = nga.entity('badges')
        .identifier(nga.field('ID'));
    let sportmonksTeams = nga.entity('sportmonksTeams').label('Sportmonks Teams').baseApiUrl('sportmonks/')
        .identifier(nga.field('combinedLeagueSeasonTeamId'));
    let orders = nga.entity('orders')
        .identifier(nga.field('ID'))
    admin
        .addEntity(users)
        .addEntity(products)
        .addEntity(variations)
        .addEntity(badges)
        .addEntity(collections)
        .addEntity(teams)
        .addEntity(leagues)
        .addEntity(sportmonksTeams)
        .addEntity(orders);

    let allEntities = {
        users,
        leagues,
        teams,
        products,
        variations,
        badges,
        collections,
        sportmonksTeams,
        orders
    };
    NGAdminProducts(nga, allEntities);
    NGAdminUsers(nga, allEntities);
    NGAdminLeagues(nga, allEntities);
    NGAdminTeams(nga, allEntities);
    NGAdminBadges(nga, allEntities);
    NGAdminProductVariations(nga, allEntities);
    NGAdminSportmonks(nga, allEntities);
    NGAdminOrders(nga, allEntities);

    // attach the admin application to the DOM and execute it
    nga.configure(admin);
}]);
