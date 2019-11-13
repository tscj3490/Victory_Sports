import {
    ProductGender,
    ProductSizes,
    ProductBrands,
    ProductCollections,
    ProductKits,
} from './admin-data';

import {
    mapExtractArrayField, mapExtractEntryField,
    mapExtractField
} from "./admin-utils";

export function NGAdminProductVariations (nga, allEntities /* all other entities */) {
    let variations = allEntities.variations;
    let products = allEntities.products;
    let badges = allEntities.badges;

    let commonFields = [
        nga.field('ProductID', 'reference')
            .targetEntity(products)
            .targetField(nga.field('Name')),
        nga.field('BadgeID', 'reference')
            .targetEntity(badges)
            .targetField(nga.field('Name')),
        nga.field('Size','choice')
            .map(mapExtractEntryField("Size.Name"))
            .choices(ProductSizes)
            .validation({ required: true }),
        nga.field('SKU'),
        nga.field('AvailableQuantity','number'),
    ];

    variations.listView().fields([
        nga.field('SKU')
            .isDetailLink(true),
        nga.field('Price','float')
            .validation({ required: true }),
        ]);
    variations.showView().fields(commonFields.concat([

    ]));
    variations.creationView().fields(commonFields.concat([

    ]));
    variations.editionView().fields(commonFields.concat([

    ]));
    variations.deletionView();

}
