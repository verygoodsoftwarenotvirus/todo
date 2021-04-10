import * as Factory from 'factory.ts';
import faker from 'faker';
import {
  APITableCell,
  APITableHeader,
} from '../components/core/apiTable/types';
import type { accountModelTranslations } from '../i18n';
import { renderUnixTime } from '../utils';
import { Pagination } from './api';
import { defaultFactories } from './fakes';

export class AccountList extends Pagination {
  accounts: Account[];

  constructor() {
    super();

    this.accounts = [];
  }
}

export class Account {
  id: number;
  name: string;
  externalID: string;
  accountSubscriptionPlanID?: number;
  createdOn: number;
  lastUpdatedOn?: number;
  archivedOn?: number;
  belongsToUser: number;

  constructor(
    id: number = 0,
    name: string = '',
    externalID: string = '',
    accountSubscriptionPlanID?: number,
    createdOn: number = 0,
    belongsToUser: number = 0,
  ) {
    this.id = id;
    this.name = name;
    this.externalID = externalID;
    this.accountSubscriptionPlanID = accountSubscriptionPlanID;
    this.createdOn = createdOn;
    this.belongsToUser = belongsToUser;
  }

  static areEqual = function (x: Account, y: Account): boolean {
    return x.id === y.id && x.externalID === y.externalID && x.name === y.name;
  };

  // this function should return everything there are no presumed fields
  static headers = (
    translations: Readonly<accountModelTranslations>,
  ): APITableHeader[] => {
    const columns = translations.columns;
    return [
      { content: columns.id, requiresAdmin: false },
      { content: columns.externalID, requiresAdmin: false },
      { content: columns.name, requiresAdmin: false },
      { content: columns.accountSubscriptionPlanID, requiresAdmin: false },
      { content: columns.createdOn, requiresAdmin: false },
      { content: columns.lastUpdatedOn, requiresAdmin: false },
      { content: columns.belongsToUser, requiresAdmin: true },
    ];
  };

  // this function should return everything there are no presumed fields
  static asRow = (x: Account): APITableCell[] => {
    return [
      new APITableCell({
        isIDCell: true,
        content: x.id.toString(),
      }),
      new APITableCell({
        content: x.externalID.toString(),
      }),
      new APITableCell({
        content: x.name,
      }),
      new APITableCell({
        content: x.accountSubscriptionPlanID?.toString() || '',
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

export class AccountCreationInput {
  name: string;
  details: string;

  constructor(name: string = '', details: string = '') {
    this.name = name;
    this.details = details;
  }
}

export const fakeAccountFactory = Factory.Sync.makeFactory<Account>({
  name: Factory.Sync.each(() => faker.random.word()),
  externalID: Factory.Sync.each(() => faker.random.uuid()),
  accountSubscriptionPlanID: Factory.Sync.each(() => faker.random.number()),
  belongsToUser: Factory.Sync.each(() => faker.random.number()),
  ...defaultFactories,
});
