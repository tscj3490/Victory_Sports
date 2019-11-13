import {
    UserAccessLevels,
    ProductGender,
    ProductSizes,
    ProductBrands,
    ProductCollections,
    ProductKits,
} from './admin-data';

export function NGAdminUsers (nga, allEntities /* all other entities */) {
    let users = allEntities.users;

    // set the fields of the user entity list view
    users.listView().fields([
        nga.field('ID'),
        nga.field('Email').isDetailLink(true),
        nga.field('UserAccessLevel', 'choice').choices([
            { value: 10, label: 'Anonymous' },
            { value: 50, label: 'User' },
            { value: 100, label: 'Admin' }
        ]),
        nga.field('Name'),
        nga.field('FirebaseId'),
        nga.field('CreatedAt', 'datetime'),
        nga.field('UpdatedAt', 'datetime'),
        nga.field('DeletedAt', 'datetime')
    ]).filters([
        nga.field('q')
            .label('User Search')
            .pinned(true)
    ]);
    users.creationView().fields([
        nga.field('Email', 'email').validation({ required: true }),
        // nga.field('Name'),
        // nga.field('Email', 'Email'),
        nga.field('UserAccessLevel','choice')
            .choices(UserAccessLevels)
            .validation({ required: true })
        // nga.field('address.street').label('Street'),
        // nga.field('address.city').label('City'),
        // nga.field('address.zipcode').label('Zipcode'),
        // nga.field('phone'),
        // nga.field('website')
    ]);
    users.editionView().fields(users.creationView().fields());
    users.deletionView();
    // add the user entity to the admin application
}
