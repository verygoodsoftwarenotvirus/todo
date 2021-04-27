import * as Factory from 'factory.ts';
import faker from 'faker';
import {
  APITableCell,
  APITableHeader,
} from '../components/core/apiTable/types';
import type { accountUserMembershipModelTranslations } from '../i18n';
import { Logger } from '../logger';
import { renderUnixTime } from '../utils';
import { DatabaseRecord, Pagination } from './api';
import { defaultFactories } from './fakes';

let logger = new Logger().withDebugValue(
  'source',
  'src/types/account_user_membership.ts',
);

export class AccountUserMembershipList extends Pagination {
  accountUserMemberships: AccountUserMembership[];

  constructor() {
    super();

    this.accountUserMemberships = [];
  }
}

export class AccountUserMembership extends DatabaseRecord {
  belongsToUser: number;
  userAccountPermissions: number;
  belongsToAccount: number;
  defaultAccount: boolean;

  constructor(
    id: number = 0,
    belongsToUser: number = 0,
    userAccountPermissions: number = 0,
    belongsToAccount: number = 0,
    defaultAccount: boolean = false,
    createdOn: number = 0,
    lastUpdatedOn?: number,
    archivedOn?: number,
  ) {
    super(id, createdOn, lastUpdatedOn, archivedOn);
    this.belongsToUser = belongsToUser;
    this.userAccountPermissions = userAccountPermissions;
    this.belongsToAccount = belongsToAccount;
    this.defaultAccount = defaultAccount;
  }

  // this function should return everything there are no presumed fields
  static headers = (
    translations: Readonly<accountUserMembershipModelTranslations>,
  ): APITableHeader[] => {
    const columns = translations.columns;
    return [
      { content: columns.belongsToUser, requiresAdmin: false },
      { content: columns.userAccountPermissions, requiresAdmin: false },
      { content: columns.defaultAccount, requiresAdmin: false },
      { content: columns.createdOn, requiresAdmin: false },
      { content: columns.archivedOn, requiresAdmin: true },
    ];
  };

  // this function should return everything there are no presumed fields
  static asRow = (x: AccountUserMembership): APITableCell[] => {
    return [
      new APITableCell({
        content: x.belongsToUser.toString(),
      }),
      new APITableCell({
        content: x.userAccountPermissions.toString(),
      }),
      new APITableCell({
        content: x.defaultAccount.toString(),
      }),
      new APITableCell({
        content: renderUnixTime(x.createdOn),
      }),
      new APITableCell({
        content: x.belongsToAccount.toString(),
        requiresAdmin: true,
      }),
    ];
  };

  static areEqual = function (
    x: AccountUserMembership,
    y: AccountUserMembership,
  ): boolean {
    return (
      x.id === y.id &&
      x.belongsToUser === y.belongsToUser &&
      x.userAccountPermissions === y.userAccountPermissions &&
      x.belongsToAccount === y.belongsToAccount &&
      x.defaultAccount === y.defaultAccount &&
      x.createdOn === y.createdOn &&
      x.archivedOn === y.archivedOn
    );
  };
}

export class AccountUserMembershipCreationInput {
  name: string;
  details: string;

  constructor(name: string = '', details: string = '') {
    this.name = name;
    this.details = details;
  }
}

export const fakeAccountUserMembershipFactory = Factory.Sync.makeFactory<AccountUserMembership>(
  {
    belongsToUser: Factory.Sync.each(() => faker.datatype.number()),
    userAccountPermissions: Factory.Sync.each(() => faker.datatype.number()),
    belongsToAccount: Factory.Sync.each(() => faker.datatype.number()),
    defaultAccount: Factory.Sync.each(() => faker.datatype.boolean()),
    ...defaultFactories,
  },
);
