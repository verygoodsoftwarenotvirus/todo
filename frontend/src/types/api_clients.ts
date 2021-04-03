import { APITableCell, APITableHeader } from '@/components/APITable/types';
import type { apiClientModelTranslations } from '@/i18n';
import { Pagination } from '@/types/api';
import { defaultFactories } from '@/types/fakes';
import { renderUnixTime } from '@/utils';
import * as Factory from 'factory.ts';
import faker from 'faker';

export class APIClientList extends Pagination {
  apiClients: APIClient[];

  constructor() {
    super();

    this.apiClients = [];
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
  belongsToAccount: number;

  constructor(
    id: number = 0,
    externalID: string = '',
    name: string = '',
    clientID: string = '',
    clientSecret: string = '',
    createdOn: number = 0,
    belongsToAccount: number = 0,
  ) {
    this.id = id;
    this.name = name;
    this.externalID = externalID;
    this.clientID = clientID;
    this.clientSecret = clientSecret;
    this.createdOn = createdOn;
    this.belongsToAccount = belongsToAccount;
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
      { content: columns.details, requiresAdmin: false },
      { content: columns.createdOn, requiresAdmin: false },
      { content: columns.lastUpdatedOn, requiresAdmin: false },
      { content: columns.belongsToAccount, requiresAdmin: true },
    ];
  };

  // this function should return everything there are no presumed fields
  static asRow = (x: APIClient): APITableCell[] => {
    return [
      new APITableCell({
        fieldName: 'id',
        content: x.id.toLocaleString(),
      }),
      new APITableCell({
        fieldName: 'externalID',
        content: x.externalID,
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
        fieldName: 'createdOn',
        content: renderUnixTime(x.createdOn),
      }),
      new APITableCell({
        fieldName: 'lastUpdatedOn',
        content: renderUnixTime(x.lastUpdatedOn),
      }),
      new APITableCell({
        fieldName: 'belongsToAccount',
        content: x.belongsToAccount.toLocaleString(),
        requiresAdmin: true,
      }),
    ];
  };
}

export class APIClientCreationInput {
  name: string;
  details: string;

  constructor(name: string = '', details: string = '') {
    this.name = name;
    this.details = details;
  }
}

export const fakeAPIClientFactory = Factory.Sync.makeFactory<APIClient>({
  name: Factory.Sync.each(() => faker.random.word()),
  externalID: Factory.Sync.each(() => faker.random.uuid()),
  clientID: Factory.Sync.each(() => faker.random.uuid()),
  clientSecret: Factory.Sync.each(() => faker.random.uuid()),
  belongsToAccount: Factory.Sync.each(() => faker.random.number()),
  ...defaultFactories,
});
