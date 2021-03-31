import { APITableCell, APITableHeader } from '@/components/APITable/types';
import type { userModelTranslations } from '@/i18n';
import { Pagination } from '@/types/api';
import { defaultFactories } from '@/types/fakes';
import { isNumeric, renderUnixTime } from '@/utils';
import * as Factory from 'factory.ts';
import faker from 'faker';

// import { AdminPermissions } from '@/types/permissions/permission_mask';

export class UserList extends Pagination {
  users: User[];

  constructor() {
    super();

    this.users = [];
  }
}

export class User {
  id: number;
  username: string;
  isAdmin: boolean;
  requiresPasswordChange: boolean;
  passwordLastChangedOn?: number;
  reputation: string;
  accountStatusExplanation: string;
  adminPermissions: AdminPermissionSummary;
  createdOn: number;
  lastUpdatedOn: number;
  archivedOn?: number;

  constructor(
    id: number = 0,
    username: string = '',
    isAdmin: boolean = false,
    requiresPasswordChange: boolean = false,
    passwordLastChangedOn: number = 0,
    reputation: string = '',
    accountStatusExplanation: string = '',
    adminPermissions: AdminPermissionSummary = new AdminPermissionSummary(),
    createdOn: number = 0,
    lastUpdatedOn: number = 0,
    archivedOn?: number,
  ) {
    this.id = id;
    this.username = username;
    this.isAdmin = isAdmin;
    this.requiresPasswordChange = requiresPasswordChange;
    this.passwordLastChangedOn = passwordLastChangedOn;
    this.reputation = reputation;
    this.accountStatusExplanation = accountStatusExplanation;
    this.adminPermissions = adminPermissions;
    this.createdOn = createdOn;
    this.lastUpdatedOn = lastUpdatedOn;
    this.archivedOn = archivedOn;
  }

  static areEqual = function (x: User, y: User): boolean {
    return (
      x.id === y.id &&
      x.username === y.username &&
      x.isAdmin === y.isAdmin &&
      x.requiresPasswordChange === y.requiresPasswordChange
    );
  };

  // this function should return everything there are no presumed fields
  static headers = (
    translations: Readonly<userModelTranslations>,
  ): APITableHeader[] => {
    const columns = translations.columns;
    return [
      { content: columns.id, requiresAdminMode: false },
      { content: columns.username, requiresAdminMode: false },
      { content: columns.isAdmin, requiresAdminMode: false },
      { content: columns.requiresPasswordChange, requiresAdminMode: false },
      { content: columns.passwordLastChangedOn, requiresAdminMode: false },
      { content: columns.createdOn, requiresAdminMode: false },
      { content: columns.lastUpdatedOn, requiresAdminMode: false },
      { content: columns.archivedOn, requiresAdminMode: false },
    ];
  };

  // this function should return everything there are no presumed fields
  static asRow = (x: User): APITableCell[] => {
    return [
      new APITableCell({
        fieldName: 'id',
        content: x.id.toString(),
      }),
      new APITableCell({
        fieldName: 'username',
        content: x.username,
      }),
      new APITableCell({
        fieldName: 'isAdmin',
        content: x.isAdmin.toString(),
      }),
      new APITableCell({
        fieldName: 'requiresPasswordChange',
        content: x.requiresPasswordChange.toString(),
      }),
      new APITableCell({
        fieldName: 'passwordLastChangedOn',
        content: (x.passwordLastChangedOn || 'never').toString(),
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
        fieldName: 'archivedOn',
        content: renderUnixTime(x.archivedOn),
        requiresAdmin: true,
      }),
    ];
  };
}

export const fakeUserFactory = Factory.Sync.makeFactory<User>({
  username: Factory.Sync.each(() => faker.random.word()),
  isAdmin: Factory.Sync.each(() => faker.random.boolean()),
  requiresPasswordChange: Factory.Sync.each(() => faker.random.boolean()),
  passwordLastChangedOn: Factory.Sync.each(() => faker.random.number()),
  reputation: Factory.Sync.each(() => faker.random.word()),
  accountStatusExplanation: Factory.Sync.each(() => faker.random.words(10)),
  adminPermissions: Factory.Sync.each(() => new AdminPermissionSummary()),
  ...defaultFactories,
});

export class UserRegistrationResponse {
  id: number;
  username: string;
  isAdmin: boolean;
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
    this.isAdmin = isAdmin;
    this.qrCode = qrCode;
    this.createdOn = createdOn;
  }
}

// QR Code for `otpauth://totp/todo:username?secret=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=&issuer=todo`
const fakeTwoFactorQRCode = `data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAIAAAACACAAAAADmVT4XAAADKUlEQVR4nOxa0Y7kIAwLp/3/X85p1SsksUN5OREkLM2q06HdYFLjZOZHZS/+bP7/N4AbwA3gBnADuAHcAG4A+wP4Ieeae6f9vbrP1Yz7PR4vMX/jvQAlGbCRjtn62UmYs3ZOWr9S6b0CijIwZi5m/uLm7Y9nsPcClGUA8TvT1tda+9n4DHi2PnEEA839tev/Auf/lRcdZRnA1Ys6MEa1ngvjNb+XQVEG4gpqon9cF+O+MMV2BtrCo9q6zmvY5zwLY7R8rfzAdgZyP6Amv6Ur4TNfddogJgfEXOudRDvID4yn26+61z4L9AY4Ltk1SzLwzlf+7X8NOHnhtd9z5r1RujNuZ2CmA17JrBKg3mVX43FAUQZQydXooYAXzDwRjgOUZMBHzOodvILnxjT/HxRmgOmh/VzIGlvwHAFsZyCrDZvRQ+uLeR3wIrqFBS3czkCuAxLcDauNpX8eNcDOuB3nBwRWzTLSiP/nOsh6JIDtDMxdsfXD0SsK1YHsdqfpALrZ2PFZ8QMScugcHWDeT1xOr/oBi6o5kDEwq/2YOvC62PcPKLYzkFdGD1DR/fp6jbQvy95pfoApm/UHCvtbrBTF1cm1/UDuiN5+wHvG74Vf/YFYVUhWH21nIHNELxRmERVekrpAAi+nOSLp/QH7Lh7ZHIjPPtNHgpIMcB1Q6IMJ3ePw3ax+qsmAn58Ety+kT5D7geaqCoLtDHA/4JU+07MvHWAjAUUZeBBdIVaMX98MKWgCoCQDvAeiJKf5Ph+Zknz+BRiY6YDVd+x5SNIXkD4i5kjNHJj1B0a8GryPuMqB9wXGKIV7GJRlQEMtKJAHjfIiocf2VUPvZ2CuA/7Zj2fm34Vw9wgoyYDXb+YCY0av9AdO6pDAGNB/cUfZd0krmVKTgbi+47z3O9/fFyj0lQDbGVj5LVkzvWENcxn9gagTDT6v+RSs/paM60KD8b4ndLIfiPDRK3n2hfTG3mvfqwiOYABrpJEBti+YVVTx2KEsA1jlx/0wu4ZVjpX9wNpvyVpQhFlfYOh/c4pQdS9Y8QP/FdsZuAHcAG4AN4AbwA3gBnAD2B7A3wAAAP//lX36HfFGOAwAAAAASUVORK5CYII=`;

export const fakeUserRegistrationResponseFactory = Factory.Sync.makeFactory<UserRegistrationResponse>(
  {
    username: Factory.Sync.each(() => faker.random.word()),
    isAdmin: Factory.Sync.each(() => faker.random.boolean()),
    qrCode: fakeTwoFactorQRCode,
    ...defaultFactories,
  },
);

export class AdminPermissionSummary {
  canCycleCookieSecrets: boolean;

  constructor(canCycleCookieSecrets: boolean = false) {
    this.canCycleCookieSecrets = canCycleCookieSecrets;
  }
}

export class UserStatus {
  isAuthenticated: boolean;
  isAdmin: boolean;
  adminPermissions?: AdminPermissionSummary;

  constructor(
    isAuthenticated: boolean = false,
    isAdmin: boolean = false,
    adminPermissions?: AdminPermissionSummary,
  ) {
    this.isAuthenticated = isAuthenticated;
    this.isAdmin = isAdmin;
    this.adminPermissions = adminPermissions;
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
