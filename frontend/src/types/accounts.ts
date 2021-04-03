import { APITableCell, APITableHeader } from '@/components/APITable/types';
import type { accountModelTranslations } from '@/i18n';
import { Pagination } from '@/types/api';
import { defaultFactories } from '@/types/fakes';
import { renderUnixTime } from '@/utils';
import * as Factory from 'factory.ts';
import faker from 'faker';

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
  belongsToAccount: number;

  constructor(
    id: number = 0,
    name: string = '',
    externalID: string = '',
    accountSubscriptionPlanID?: number,
    createdOn: number = 0,
    belongsToAccount: number = 0,
  ) {
    this.id = id;
    this.name = name;
    this.externalID = externalID;
    this.accountSubscriptionPlanID = accountSubscriptionPlanID;
    this.createdOn = createdOn;
    this.belongsToAccount = belongsToAccount;
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
      { content: columns.belongsToAccount, requiresAdmin: true },
    ];
  };

  // this function should return everything there are no presumed fields
  static asRow = (x: Account): APITableCell[] => {
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
        fieldName: 'accountSubscriptionPlanID',
        content: x.accountSubscriptionPlanID?.toString(),
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
        content: x.belongsToAccount.toString(),
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
  belongsToAccount: Factory.Sync.each(() => faker.random.number()),
  ...defaultFactories,
});
