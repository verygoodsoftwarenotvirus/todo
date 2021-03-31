import { APITableCell, APITableHeader } from '@/components/APITable/types';
import type { itemModelTranslations } from '@/i18n';
import { Pagination } from '@/types/api';
import { defaultFactories } from '@/types/fakes';
import { renderUnixTime } from '@/utils';
import * as Factory from 'factory.ts';
import faker from 'faker';

export class ItemList extends Pagination {
  items: Item[];

  constructor() {
    super();

    this.items = [];
  }
}

export class Item {
  id: number;
  name: string;
  details: string;
  createdOn: number;
  lastUpdatedOn?: number;
  archivedOn?: number;
  belongsToAccount: number;

  constructor(
    id: number = 0,
    name: string = '',
    details: string = '',
    createdOn: number = 0,
    belongsToAccount: number = 0,
  ) {
    this.id = id;
    this.name = name;
    this.details = details;
    this.createdOn = createdOn;
    this.belongsToAccount = belongsToAccount;
  }

  static areEqual = function (x: Item, y: Item): boolean {
    return x.id === y.id && x.name === y.name && x.details === y.details;
  };

  // this function should return everything there are no presumed fields
  static headers = (
    translations: Readonly<itemModelTranslations>,
  ): APITableHeader[] => {
    const columns = translations.columns;
    return [
      { content: columns.id, requiresAdminMode: false },
      { content: columns.name, requiresAdminMode: false },
      { content: columns.details, requiresAdminMode: false },
      { content: columns.createdOn, requiresAdminMode: false },
      { content: columns.lastUpdatedOn, requiresAdminMode: false },
      { content: columns.belongsToAccount, requiresAdminMode: true },
    ];
  };

  // this function should return everything there are no presumed fields
  static asRow = (x: Item): APITableCell[] => {
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
        fieldName: 'details',
        content: x.details,
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

export class ItemCreationInput {
  name: string;
  details: string;

  constructor(name: string = '', details: string = '') {
    this.name = name;
    this.details = details;
  }
}

export const fakeItemFactory = Factory.Sync.makeFactory<Item>({
  name: Factory.Sync.each(() => faker.random.word()),
  details: Factory.Sync.each(() => faker.random.word()),
  belongsToAccount: Factory.Sync.each(() => faker.random.number()),
  ...defaultFactories,
});
