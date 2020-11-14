import * as Factory from "factory.ts";
import faker from "faker";

import { Pagination } from "@/types/api";
import { defaultFactories } from "@/types/fakes";
import type { APITableCell, APITableHeader } from "@/components/APITable/types";
import { renderUnixTime } from "@/utils";
import type {oauth2ClientModelTranslations} from "@/i18n";

export class OAuth2ClientList extends Pagination {
    clients: OAuth2Client[];

    constructor() {
        super();

        this.clients = [];
    }
}

export class OAuth2Client {
    id: number;
    name: string;
    clientID: string;
    clientSecret: string;
    redirectURI: string;
    scopes: string[];
    implicitAllowed: boolean;
    createdOn: number;
    updatedOn?: number;
    archivedOn?: number;
    belongsToUser: number;

    constructor(
      id: number = 0,
      name: string = '',
      clientID: string = '',
      clientSecret: string = '',
      redirectURI: string = '',
      scopes: string[] = [],
      implicitAllowed: boolean = false,
      createdOn: number = 0,
      belongsToUser: number = 0,
    ) {
        this.id = id;
        this.name = name;
        this.clientID = clientID;
        this.clientSecret = clientSecret;
        this.redirectURI = redirectURI;
        this.scopes = scopes;
        this.implicitAllowed = implicitAllowed;
        this.createdOn = createdOn;
        this.belongsToUser = belongsToUser;
    }

    static areEqual = function(
        x: OAuth2Client,
        y: OAuth2Client,
    ): boolean {
        return (
            x.id === y.id &&
            x.name === y.name &&
            x.clientID === y.clientID &&
            x.clientSecret === y.clientSecret &&
            x.redirectURI === y.redirectURI &&
            x.scopes === y.scopes &&
            x.implicitAllowed === y.implicitAllowed
        );
    }

    // this function should return everything there are no presumed fields
    static headers = (translations: Readonly<oauth2ClientModelTranslations>): APITableHeader[] => {
        const columns = translations.columns;
        return [
            {content: columns.id, requiresAdmin: false},
            {content: columns.name, requiresAdmin: false},
            {content: columns.clientID, requiresAdmin: false},
            {content: columns.clientSecret, requiresAdmin: false},
            {content: columns.redirectURI, requiresAdmin: false},
            {content: columns.scopes, requiresAdmin: false},
            {content: columns.implicitAllowed, requiresAdmin: false},
            {content: columns.createdOn, requiresAdmin: false},
            {content: columns.lastUpdatedOn, requiresAdmin: false},
            {content: columns.belongsToUser, requiresAdmin: true},
        ];
    }

    // this function should return everything there are no presumed fields
    static asRow = (x: OAuth2Client): APITableCell[] => {
        return [
            { fieldName: 'id', content: x.id.toString(), requiresAdmin: false },
            { fieldName: 'name', content: x.name, requiresAdmin: false },
            { fieldName: 'clientID', content: x.clientID, requiresAdmin: false },
            { fieldName: 'clientSecret', content: x.clientSecret, requiresAdmin: false },
            { fieldName: 'redirectURI', content: x.redirectURI, requiresAdmin: false },
            { fieldName: 'scopes', content: x.scopes.toString(), requiresAdmin: false },
            { fieldName: 'implicitAllowed', content: x.implicitAllowed.toString(), requiresAdmin: false },
            { fieldName: 'createdOn', content: renderUnixTime(x.createdOn), requiresAdmin: false },
            { fieldName: 'lastUpdatedOn', content: renderUnixTime(x.updatedOn), requiresAdmin: false },
            { fieldName: 'belongsToUser', content: x.belongsToUser.toString(), requiresAdmin: true },
        ]
    }
}

export const fakeOAuth2ClientFactory = Factory.Sync.makeFactory<OAuth2Client> ({
    name: Factory.Sync.each(() =>  faker.random.word()),
    clientID: Factory.Sync.each( () => faker.random.word() ),
    clientSecret: Factory.Sync.each( () => faker.random.word() ),
    redirectURI: Factory.Sync.each( () => faker.random.word() ),
    scopes: Factory.Sync.each( () => [faker.random.word(), faker.random.word(), faker.random.word()] ),
    implicitAllowed: Factory.Sync.each( () => faker.random.boolean() ),
    belongsToUser: Factory.Sync.each(() =>  faker.random.number()),
    ...defaultFactories,
});
