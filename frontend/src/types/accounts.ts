import * as Factory from 'factory.ts';
import faker from 'faker';
import {
  APITableCell,
  APITableHeader,
} from '../components/core/apiTable/types';
import type { accountModelTranslations } from '../i18n';
import { renderUnixTime } from '../utils';
import {DatabaseRecord, Pagination} from './api';
import { defaultFactories } from './fakes';
import {
  AccountUserMembership,
  fakeAccountUserMembershipFactory,
} from "../types/account_user_membership";
import {create} from "string-format";

export class AccountList extends Pagination {
  accounts: Account[];

  constructor() {
    super();

    this.accounts = [];
  }
}

export class Account extends DatabaseRecord {
  name: string;
  externalID: string;
  accountSubscriptionPlanID?: number;
  defaultNewMemberPermissions: number;
  belongsToUser: number;
  members: AccountUserMembership[];

  constructor(
    id: number = 0,
    name: string = '',
    externalID: string = '',
    accountSubscriptionPlanID?: number,
    defaultNewMemberPermissions: number = 0,
    members: AccountUserMembership[] = [],
    createdOn: number = 0,
    lastUpdatedOn?: number,
    archivedOn?: number,
    belongsToUser: number = 0,
  ) {
    super(id, createdOn, lastUpdatedOn, archivedOn);
    this.name = name;
    this.externalID = externalID;
    this.accountSubscriptionPlanID = accountSubscriptionPlanID;
    this.belongsToUser = belongsToUser;
    this.defaultNewMemberPermissions = defaultNewMemberPermissions;
    this.members = members;
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
  externalID: Factory.Sync.each(() => faker.datatype.uuid()),
  accountSubscriptionPlanID: Factory.Sync.each(() => faker.datatype.number()),
  belongsToUser: Factory.Sync.each(() => faker.datatype.number()),
  defaultNewMemberPermissions: Number.MAX_SAFE_INTEGER,
  members: fakeAccountUserMembershipFactory.buildList(10),
  ...defaultFactories,
});
