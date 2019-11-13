export function NGAdminBadges (nga, allEntities /* all other entities */) {
    let badges = allEntities.badges;

    let commonFields = [

        nga.field('Name').isDetailLink(true)
            .validation({ required: true }),
        nga.field('Price','float')
            .validation({ required: true }),
        // nga.field('Thumbnail','template')
        //     .template('<img src="{{ entry.values.Thumbnail }}" class="product-thumbnail" alt="{{ entry.values.Thumbnail }}"/>'),
        // nga.field('Image', 'file')
        //     .uploadInformation({
        //         'url': '/admin/products/image-upload',
        //         'apifilename': 'image_name' })
        //     .validation({ required: true }),
    ];
    badges.listView().fields(commonFields.concat([

        nga.field('Thumbnail','template')
            .template('<img src="{{ entry.values.Thumbnail }}" class="product-thumbnail" alt="{{ entry.values.Thumbnail }}"/>'),
    ]));
    badges.showView().fields(commonFields.concat([

        nga.field('Thumbnail','template')
            .template('<img src="{{ entry.values.Thumbnail }}" class="product-thumbnail" alt="{{ entry.values.Thumbnail }}"/>'),
        nga.field('Image', 'file')
            .uploadInformation({
                'url': '/admin/products/image-upload',
                'apifilename': 'image_name' })
            .validation({ required: true }),
    ]));
    badges.creationView().fields(commonFields.concat([

        nga.field('Image', 'file')
            .uploadInformation({
                'url': '/admin/products/image-upload',
                'apifilename': 'image_name' })
            .validation({ required: true }),
    ]));
    badges.editionView().fields(commonFields.concat([

        nga.field('Thumbnail','template')
            .template('<img src="{{ entry.values.Thumbnail }}" class="product-thumbnail" alt="{{ entry.values.Thumbnail }}"/>'),
        nga.field('Image', 'file')
            .uploadInformation({
                'url': '/admin/products/image-upload',
                'apifilename': 'image_name' })
            .validation({ required: true }),
    ]));
    badges.deletionView();
}
