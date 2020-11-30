import * as Factory from 'factory.ts';
import faker from 'faker';

import { Pagination } from '@/types/api';
import { defaultFactories } from '@/types/fakes';
import { APITableCell, APITableHeader } from '@/components/APITable/types';
import { renderUnixTime } from '@/utils';
import type { oauth2ClientModelTranslations } from '@/i18n';

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
  lastUpdatedOn?: number;
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

  static areEqual = function (x: OAuth2Client, y: OAuth2Client): boolean {
    return (
      x.id === y.id &&
      x.name === y.name &&
      x.clientID === y.clientID &&
      x.clientSecret === y.clientSecret &&
      x.redirectURI === y.redirectURI &&
      x.scopes === y.scopes &&
      x.implicitAllowed === y.implicitAllowed
    );
  };

  // this function should return everything there are no presumed fields
  static headers = (
    translations: Readonly<oauth2ClientModelTranslations>,
  ): APITableHeader[] => {
    const columns = translations.columns;
    return [
      { content: columns.id, requiresAdminMode: false },
      { content: columns.name, requiresAdminMode: false },
      { content: columns.clientID, requiresAdminMode: false },
      { content: columns.clientSecret, requiresAdminMode: false },
      { content: columns.redirectURI, requiresAdminMode: false },
      { content: columns.scopes, requiresAdminMode: false },
      { content: columns.implicitAllowed, requiresAdminMode: false },
      { content: columns.createdOn, requiresAdminMode: false },
      { content: columns.lastUpdatedOn, requiresAdminMode: false },
      { content: columns.belongsToUser, requiresAdminMode: true },
    ];
  };

  // this function should return everything there are no presumed fields
  static asRow = (x: OAuth2Client): APITableCell[] => {
    return [
      new APITableCell({
        fieldName: 'id',
        content: x.id.toString(),
      }),
      new APITableCell({
        fieldName: 'name',
        content: x.name,
      }),
      new APITableCell({
        fieldName: 'clientID',
        content: x.clientID,
      }),
      new APITableCell({
        fieldName: 'clientSecret',
        content: x.clientSecret,
      }),
      new APITableCell({
        fieldName: 'redirectURI',
        content: x.redirectURI,
      }),
      new APITableCell({
        fieldName: 'scopes',
        content: x.scopes.toString(),
      }),
      new APITableCell({
        fieldName: 'implicitAllowed',
        content: x.implicitAllowed.toString(),
      }),
      new APITableCell({
        fieldName: 'createdOn',
        content: renderUnixTime(x.createdOn),
      }),
      new APITableCell({
        fieldName: 'lastUpdatedOn',
        content: renderUnixTime(x.lastUpdatedOn),
      }),
      new APITableCell({
        fieldName: 'belongsToUser',
        content: x.belongsToUser.toString(),
        requiresAdmin: true,
      }),
    ];
  };
}

export class OAuth2ClientCreationInput {
  name: string;
  redirectURI: string;
  scopes: string;

  constructor(
    name: string = '',
    redirectURI: string = '',
    scopes: string = '',
  ) {
    this.name = name;
    this.redirectURI = redirectURI;
    this.scopes = scopes;
  }
}

export const fakeOAuth2ClientFactory = Factory.Sync.makeFactory<OAuth2Client>({
  name: Factory.Sync.each(() => faker.random.word()),
  clientID: Factory.Sync.each(() => faker.random.word()),
  clientSecret: Factory.Sync.each(() => faker.random.word()),
  redirectURI: Factory.Sync.each(() => faker.random.word()),
  scopes: Factory.Sync.each(() => [
    faker.random.word(),
    faker.random.word(),
    faker.random.word(),
  ]),
  implicitAllowed: Factory.Sync.each(() => faker.random.boolean()),
  belongsToUser: Factory.Sync.each(() => faker.random.number()),
  ...defaultFactories,
});
