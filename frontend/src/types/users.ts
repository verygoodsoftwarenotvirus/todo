import * as Factory from 'factory.ts';
import faker from 'faker';
import {
  APITableCell,
  APITableHeader,
} from '../components/core/apiTable/types';
import type { userModelTranslations } from '../i18n';
import { isNumeric, renderUnixTime } from '../utils';
import {DatabaseRecord, Pagination} from './api';
import { defaultFactories } from './fakes';

export class UserList extends Pagination {
  users: User[];

  constructor() {
    super();

    this.users = [];
  }
}

export class User extends DatabaseRecord {
  externalID: string;
  username: string;
  serviceAdminPermissions: number; // TODO: implement permission bitmask
  requiresPasswordChange: boolean;
  passwordLastChangedOn?: number;
  reputation: string;
  reputationExplanation: string;
  adminPermissions: AdminPermissionSummary;

  constructor(
    id: number = 0,
    externalID: string = '',
    username: string = '',
    requiresPasswordChange: boolean = false,
    passwordLastChangedOn: number = 0,
    reputation: string = '',
    reputationExplanation: string = '',
    serviceAdminPermissions: number = 0,
    adminPermissions: AdminPermissionSummary = new AdminPermissionSummary(),
    createdOn: number = 0,
    lastUpdatedOn: number = 0,
    archivedOn?: number,
  ) {
    super(id, createdOn, lastUpdatedOn, archivedOn)
    this.externalID = externalID;
    this.username = username;
    this.serviceAdminPermissions = serviceAdminPermissions;
    this.requiresPasswordChange = requiresPasswordChange;
    this.passwordLastChangedOn = passwordLastChangedOn;
    this.reputation = reputation;
    this.reputationExplanation = reputationExplanation;
    this.adminPermissions = adminPermissions;
  }

  static areEqual = function (x: User, y: User): boolean {
    return (
      x.id === y.id &&
      x.username === y.username &&
      x.serviceAdminPermissions === y.serviceAdminPermissions &&
      x.requiresPasswordChange === y.requiresPasswordChange
    );
  };

  // this function should return everything there are no presumed fields
  static headers = (
    translations: Readonly<userModelTranslations>,
  ): APITableHeader[] => {
    const columns = translations.columns;
    return [
      { content: columns.id, requiresAdmin: false },
      { content: columns.username, requiresAdmin: false },
      { content: columns.reputation, requiresAdmin: false },
      { content: columns.reputationExplanation, requiresAdmin: false },
      { content: columns.serviceAdminPermissions, requiresAdmin: false },
      { content: columns.requiresPasswordChange, requiresAdmin: false },
      { content: columns.passwordLastChangedOn, requiresAdmin: false },
      { content: columns.externalID, requiresAdmin: true },
      { content: columns.createdOn, requiresAdmin: true },
      { content: columns.lastUpdatedOn, requiresAdmin: true },
      { content: columns.archivedOn, requiresAdmin: true },
    ];
  };

  // this function should return everything there are no presumed fields
  static asRow = (x: User): APITableCell[] => {
    return [
      new APITableCell({
        isIDCell: true,
        content: x.id.toString(),
      }),
      new APITableCell({
        content: x.username,
      }),
      new APITableCell({
        content: x.reputation,
      }),
      new APITableCell({
        content: x.reputationExplanation,
      }),
      new APITableCell({
        content: x.serviceAdminPermissions.toString(),
      }),
      new APITableCell({
        content: x.requiresPasswordChange.toString(),
      }),
      new APITableCell({
        content: (x.passwordLastChangedOn || 'never').toString(),
      }),
      new APITableCell({
        content: x.externalID,
        requiresAdmin: true,
      }),
      new APITableCell({
        content: renderUnixTime(x.createdOn),
        requiresAdmin: true,
      }),
      new APITableCell({
        content: renderUnixTime(x.lastUpdatedOn),
        requiresAdmin: true,
      }),
      new APITableCell({
        content: renderUnixTime(x.archivedOn),
        requiresAdmin: true,
      }),
    ];
  };
}

const maximumPossiblePermissions: number = 4294967295;

export const fakeUserFactory = Factory.Sync.makeFactory<User>({
  externalID: Factory.Sync.each(() => faker.datatype.uuid()),
  username: Factory.Sync.each(() => faker.random.word()),
  serviceAdminPermissions: Factory.Sync.each(() =>
    faker.datatype.number(maximumPossiblePermissions),
  ),
  requiresPasswordChange: Factory.Sync.each(() => faker.datatype.boolean()),
  passwordLastChangedOn: Factory.Sync.each(() => faker.datatype.number()),
  reputation: Factory.Sync.each(() => faker.random.word()),
  reputationExplanation: Factory.Sync.each(() => faker.random.words(10)),
  adminPermissions: Factory.Sync.each(() => new AdminPermissionSummary()),
  ...defaultFactories,
});

export class UserRegistrationResponse {
  id: number;
  username: string;
  qrCode: string;
  createdOn: number;

  constructor(
    id: number = 0,
    username: string = '',
    isAdmin: boolean = false,
    qrCode: string = '',
    createdOn: number = 0,
  ) {
    this.id = id;
    this.username = username;
    this.qrCode = qrCode;
    this.createdOn = createdOn;
  }
}

// QR Code for `otpauth://totp/todo:username?secret=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=&issuer=todo`
const fakeTwoFactorQRCode = `data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAIAAAACACAAAAADmVT4XAAADKUlEQVR4nOxa0Y7kIAwLp/3/X85p1SsksUN5OREkLM2q06HdYFLjZOZHZS/+bP7/N4AbwA3gBnADuAHcAG4A+wP4Ieeae6f9vbrP1Yz7PR4vMX/jvQAlGbCRjtn62UmYs3ZOWr9S6b0CijIwZi5m/uLm7Y9nsPcClGUA8TvT1tda+9n4DHi2PnEEA839tev/Auf/lRcdZRnA1Ys6MEa1ngvjNb+XQVEG4gpqon9cF+O+MMV2BtrCo9q6zmvY5zwLY7R8rfzAdgZyP6Amv6Ur4TNfddogJgfEXOudRDvID4yn26+61z4L9AY4Ltk1SzLwzlf+7X8NOHnhtd9z5r1RujNuZ2CmA17JrBKg3mVX43FAUQZQydXooYAXzDwRjgOUZMBHzOodvILnxjT/HxRmgOmh/VzIGlvwHAFsZyCrDZvRQ+uLeR3wIrqFBS3czkCuAxLcDauNpX8eNcDOuB3nBwRWzTLSiP/nOsh6JIDtDMxdsfXD0SsK1YHsdqfpALrZ2PFZ8QMScugcHWDeT1xOr/oBi6o5kDEwq/2YOvC62PcPKLYzkFdGD1DR/fp6jbQvy95pfoApm/UHCvtbrBTF1cm1/UDuiN5+wHvG74Vf/YFYVUhWH21nIHNELxRmERVekrpAAi+nOSLp/QH7Lh7ZHIjPPtNHgpIMcB1Q6IMJ3ePw3ax+qsmAn58Ety+kT5D7geaqCoLtDHA/4JU+07MvHWAjAUUZeBBdIVaMX98MKWgCoCQDvAeiJKf5Ph+Zknz+BRiY6YDVd+x5SNIXkD4i5kjNHJj1B0a8GryPuMqB9wXGKIV7GJRlQEMtKJAHjfIiocf2VUPvZ2CuA/7Zj2fm34Vw9wgoyYDXb+YCY0av9AdO6pDAGNB/cUfZd0krmVKTgbi+47z3O9/fFyj0lQDbGVj5LVkzvWENcxn9gagTDT6v+RSs/paM60KD8b4ndLIfiPDRK3n2hfTG3mvfqwiOYABrpJEBti+YVVTx2KEsA1jlx/0wu4ZVjpX9wNpvyVpQhFlfYOh/c4pQdS9Y8QP/FdsZuAHcAG4AN4AbwA3gBnAD2B7A3wAAAP//lX36HfFGOAwAAAAASUVORK5CYII=`;

export const fakeUserRegistrationResponseFactory = Factory.Sync.makeFactory<UserRegistrationResponse>(
  {
    username: Factory.Sync.each(() => faker.random.word()),
    qrCode: fakeTwoFactorQRCode,
    ...defaultFactories,
  },
);

export class UserPermissionSummary {
  canManageWebhooks: boolean;
  canManageAPIClients: boolean;

  constructor(
    canManageWebhooks: boolean = false,
    canManageAPIClients: boolean = false,
  ) {
    this.canManageWebhooks = canManageWebhooks;
    this.canManageAPIClients = canManageAPIClients;
  }
}

export class AdminPermissionSummary {
  canCycleCookieSecret: boolean;
  canBanUsers: boolean;
  canTerminateAccounts: boolean;

  constructor(
    canCycleCookieSecret = false,
    canBanUsers = false,
    canTerminateAccounts = false,
  ) {
    this.canCycleCookieSecret = canCycleCookieSecret;
    this.canBanUsers = canBanUsers;
    this.canTerminateAccounts = canTerminateAccounts;
  }
}

export class UserStatus {
  isAuthenticated: boolean;
  userReputation: string;
  reputationExplanation: string;
  activeAccount: number;
  accountPermissions?: Map<string, UserPermissionSummary>;
  adminPermissions?: AdminPermissionSummary;

  constructor(
    userReputation: string = '',
    reputationExplanation: string = '',
    isAuthenticated: boolean = false,
    activeAccount: number = 0,
    accountPermissions?: Map<string, UserPermissionSummary>,
    adminPermissions?: AdminPermissionSummary,
  ) {
    this.userReputation = userReputation;
    this.reputationExplanation = reputationExplanation;
    this.isAuthenticated = isAuthenticated;
    this.activeAccount = activeAccount;
    this.accountPermissions = accountPermissions;
    this.adminPermissions = adminPermissions;
  }

  public isAdmin(): boolean {
    return !!this.adminPermissions;
  }
}

export class UserPasswordUpdateRequest {
  newPassword: string;
  currentPassword: string;
  totpToken: string;

  constructor(
    newPassword: string = '',
    currentPassword: string = '',
    totpToken: string = '',
  ) {
    this.newPassword = newPassword;
    this.currentPassword = currentPassword;
    this.totpToken = totpToken;
  }

  goodToGo(): boolean {
    return (
      this.newPassword !== '' &&
      this.currentPassword != '' &&
      this.currentPassword !== this.newPassword &&
      this.totpToken.length === 6 &&
      isNumeric(this.totpToken)
    );
  }
}

export class UserTwoFactorSecretUpdateRequest {
  currentPassword: string;
  totpToken: string;

  constructor(currentPassword: string = '', totpToken: string = '') {
    this.currentPassword = currentPassword;
    this.totpToken = totpToken;
  }

  goodToGo(): boolean {
    return (
      this.currentPassword != '' &&
      this.totpToken.length === 6 &&
      isNumeric(this.totpToken)
    );
  }
}
