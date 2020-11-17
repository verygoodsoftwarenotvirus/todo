import * as Factory from 'factory.ts';
import faker from 'faker';

import { Pagination } from '@/types/api';
import { defaultFactories } from '@/types/fakes';
import { isNumeric, renderUnixTime } from '@/utils';
import type { userModelTranslations } from '@/i18n';
import { APITableCell, APITableHeader } from '@/components/APITable/types';
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
  createdOn: number;
  lastUpdatedOn: number;
  archivedOn?: number;

  constructor(
    id: number = 0,
    username: string = '',
    isAdmin: boolean = false,
    requiresPasswordChange: boolean = false,
    passwordLastChangedOn: number = 0,
    createdOn: number = 0,
    lastUpdatedOn: number = 0,
    archivedOn?: number,
  ) {
    this.id = id;
    this.username = username;
    this.isAdmin = isAdmin;
    this.requiresPasswordChange = requiresPasswordChange;
    this.passwordLastChangedOn = passwordLastChangedOn;
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
  ...defaultFactories,
});

export class UserRegistrationResponse {
  id: number;
  username: string;
  isAdmin: boolean;
  qrCode: string;
  createdOn: number;
  lastUpdatedOn: number;
  archivedOn: number;
  passwordLastChangedOn: number;

  constructor(
    id: number = 0,
    username: string = '',
    isAdmin: boolean = false,
    qrCode: string = '',
    createdOn: number = 0,
    lastUpdatedOn: number = 0,
    archivedOn: number = 0,
    passwordLastChangedOn: number = 0,
  ) {
    this.id = id;
    this.username = username;
    this.isAdmin = isAdmin;
    this.qrCode = qrCode;
    this.createdOn = createdOn;
    this.lastUpdatedOn = lastUpdatedOn;
    this.archivedOn = archivedOn;
    this.passwordLastChangedOn = passwordLastChangedOn;
  }
}

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
