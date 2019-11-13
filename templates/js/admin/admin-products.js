import {
    ProductGender,
    ProductSizes,
    ProductBrands,
    ProductCollections,
    ProductKits,
} from './admin-data';

import {
    mapExtractArrayField, mapExtractEntryField,
} from "./admin-utils";

export function NGAdminProducts (nga, allEntities /* all other entities */) {
    let products = allEntities.products;  // outside reference
    let teams = allEntities.teams;

    // MARK: Products
    products.listView().fields([
        nga.field('ID'),
        nga.field('Name').isDetailLink(true),
        nga.field('Description'),
        nga.field('Price'),
        nga.field('Gender'),
        nga.field('TeamID', 'reference')
            .targetEntity(teams)
            .targetField(nga.field('Name.en'))
    ]);

    products.showView().fields(products.listView().fields().concat([
        // nga.field('Image'),
        nga.field('Thumbnail','template')
            .template('<img src="{{ entry.values.Thumbnail }}" class="product-thumbnail" alt="{{ entry.values.Thumbnail }}"/>'),
        nga.field('KitCode'),
        nga.field('Category.Name').label("Category"),
        nga.field('Brand.Name'),
        nga.field('Team.Name.en'),
        nga.field('Team.Name.ar'),
        nga.field('Collections', 'embedded_list') // Define a 1-N relationship with the (embedded) comment entity
            .targetFields([ // which comment fields to display in the datagrid / form
                nga.field('Name'),
                nga.field('Code'),
                nga.field('CreatedAt', 'date')
            ])
    ]));
    let commonFields = [
        nga.field('Name').isDetailLink(true)
            .validation({ required: true }),
        nga.field('Description')
            .validation({ required: true }),
        nga.field('Gender', 'choice')
            .choices(ProductGender)
            .validation({ required: true }),
        nga.field('Brand','choice')
            .map(mapExtractEntryField("Brand.Name"))
            .choices(ProductBrands)
            .validation({ required: true }),
        nga.field('KitCode','choice').label('Kit')
            .choices(ProductKits)
            .validation({ required: true }),
        nga.field('Collections','choices')
            .map(mapExtractArrayField("Code"))
            .choices(ProductCollections),
        nga.field('TeamID', 'reference')
            .targetEntity(teams)
            .targetField(nga.field('Name.en')),
        // nga.field('upload', 'template')
        //     .label('Upload image')
        //     .template(`<uploader prefix="'/admin/products/image-upload'" suffix="''"/>`),
        nga.field('Thumbnail', 'file')
            .uploadInformation({
                'url': '/admin/products/image-upload',
                'apifilename': 'image_name' })
            .validation({ required: true }),
        nga.field('Image', 'file')
            .uploadInformation({
                'url': '/admin/products/image-upload',
                'apifilename': 'image_name' })
            .validation({ required: true }),
        nga.field('Image2', 'file')
            .uploadInformation({
                'url': '/admin/products/image-upload',
                'apifilename': 'image_name' })
            .validation({ required: false }),
        nga.field('Image3', 'file')
            .uploadInformation({
                'url': '/admin/products/image-upload',
                'apifilename': 'image_name' })
            .validation({ required: false }),
        nga.field('Image4', 'file')
            .uploadInformation({
                'url': '/admin/products/image-upload',
                'apifilename': 'image_name' })
            .validation({ required: false }),
        nga.field('Thumbnail Preview','template')
            .template('<img src="{{ entry.values.Thumbnail }}" class="product-thumbnail" alt="{{ entry.values.Thumbnail }}"/>'),
        nga.field('Image Preview','template')
            .template('<img src="{{ entry.values.Image }}" class="product-thumbnail" alt="{{ entry.values.Image }}"/>'),
        nga.field('Image2 Preview','template')
            .template('<img src="{{ entry.values.Image2 }}" class="product-thumbnail" alt="{{ entry.values.Image2 }}"/>'),
        nga.field('Image3 Preview','template')
            .template('<img src="{{ entry.values.Image3 }}" class="product-thumbnail" alt="{{ entry.values.Image3 }}"/>'),
        nga.field('Image4 Preview','template')
            .template('<img src="{{ entry.values.Image4 }}" class="product-thumbnail" alt="{{ entry.values.Image4 }}"/>'),
    ];

    products.creationView().fields(commonFields.concat([
        nga.field('AvailableQuantity','number')
            .validation({ required: true }),
        nga.field('Price','float')
            .validation({ required: true }),
        nga.field('Sizes','choices')
            .map(mapExtractArrayField("Name"))
            .choices(ProductSizes)
            .validation({ required: true }),
    ]));
    products.editionView().fields(commonFields.concat([

    ]));
}