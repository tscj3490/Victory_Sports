import {
    ProductGender,
    ordersizes,
    ProductBrands,
    ProductCollections,
    ProductKits,
} from './admin-data';

import {
    mapExtractArrayField, mapExtractEntryField,
} from "./admin-utils";

export function NGAdminOrders (nga, allEntities /* all other entities */) {
    let orders = allEntities.orders;  // outside reference
    let teams = allEntities.teams;

    // MARK: orders
    orders.listView().fields([
        nga.field('ID').isDetailLink(true),
        nga.field('Shipping Address'),
        nga.field('Email'),
        nga.field('Subotal'),
        nga.field('VAT'),
        nga.field('Shipping Costs'),
        nga.field('Total'),
        nga.field('Shipping Method'),
        nga.field('Notes')
    ]);

    orders.showView().fields(orders.listView().fields().concat([]));

    let commonFields = [];

    orders.creationView().fields(commonFields.concat([]));
    orders.editionView().fields(commonFields.concat([]));
}