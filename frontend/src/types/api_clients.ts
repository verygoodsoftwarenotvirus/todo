import * as Factory from 'factory.ts';
import faker from 'faker';
import {
  APITableCell,
  APITableHeader,
} from '../components/core/apiTable/types';
import type { apiClientModelTranslations } from '../i18n';
import { Pagination } from '../types/api';
import { defaultFactories } from '../types/fakes';
import { renderUnixTime } from '../utils';

export class APIClientList extends Pagination {
  clients: APIClient[];

  constructor() {
    super();

    this.clients = [];
  }
}

export class APIClient {
  id: number;
  externalID: string;
  name: string;
  clientID: string;
  clientSecret: string;
  createdOn: number;
  lastUpdatedOn?: number;
  archivedOn?: number;
  belongsToUser: number;

  constructor(
    id: number = 0,
    externalID: string = '',
    name: string = '',
    clientID: string = '',
    clientSecret: string = '',
    createdOn: number = 0,
    belongsToUser: number = 0,
  ) {
    this.id = id;
    this.name = name;
    this.externalID = externalID;
    this.clientID = clientID;
    this.clientSecret = clientSecret;
    this.createdOn = createdOn;
    this.belongsToUser = belongsToUser;
  }

  static areEqual = function (x: APIClient, y: APIClient): boolean {
    return (
      x.id === y.id &&
      x.externalID === y.externalID &&
      x.name === y.name &&
      x.clientID === y.clientID &&
      x.clientSecret === y.clientSecret
    );
  };

  // this function should return everything there are no presumed fields
  static headers = (
    translations: Readonly<apiClientModelTranslations>,
  ): APITableHeader[] => {
    const columns = translations.columns;
    return [
      { content: columns.id, requiresAdmin: false },
      { content: columns.externalID, requiresAdmin: false },
      { content: columns.name, requiresAdmin: false },
      { content: columns.clientID, requiresAdmin: false },
      { content: columns.createdOn, requiresAdmin: false },
      { content: columns.lastUpdatedOn, requiresAdmin: false },
      { content: columns.belongsToUser, requiresAdmin: true },
    ];
  };

  // this function should return everything there are no presumed fields
  static asRow = (x: APIClient): APITableCell[] => {
    return [
      new APITableCell({
        isIDCell: true,
        content: x.id.toString(),
      }),
      new APITableCell({
        content: x.externalID,
      }),
      new APITableCell({
        content: x.name,
      }),
      new APITableCell({
        content: x.clientID,
      }),
      new APITableCell({
        content: renderUnixTime(x.createdOn),
      }),
      new APITableCell({
        content: renderUnixTime(x.lastUpdatedOn),
      }),
      new APITableCell({
        content: x.belongsToUser.toString(),
        requiresAdmin: true,
      }),
    ];
  };
}

export class APIClientCreationInput {
  name: string;
  username: string;
  password: string;
  totpToken: string;

  constructor(
    name: string = '',
    username: string = '',
    password: string = '',
    totpToken: string = '',
  ) {
    this.name = name;
    this.username = username;
    this.password = password;
    this.totpToken = totpToken;
  }

  complete(): boolean {
    return (
      this.name !== '' &&
      this.username !== '' &&
      this.password !== '' &&
      this.totpToken !== ''
    );
  }
}

export const fakeAPIClientFactory = Factory.Sync.makeFactory<APIClient>({
  name: Factory.Sync.each(() => faker.random.word()),
  externalID: Factory.Sync.each(() => faker.datatype.uuid()),
  clientID: Factory.Sync.each(() => faker.datatype.uuid()),
  clientSecret: Factory.Sync.each(() => faker.datatype.uuid()),
  belongsToUser: Factory.Sync.each(() => faker.datatype.number()),
  ...defaultFactories,
});
