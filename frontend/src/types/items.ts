import * as Factory from 'factory.ts';
import faker from 'faker';
import {
  APITableCell,
  APITableHeader,
} from '../components/core/apiTable/types';
import type { itemModelTranslations } from '../i18n';
import { renderUnixTime } from '../utils';
import { DatabaseRecord, Pagination } from './api';
import { defaultFactories } from './fakes';

export class ItemList extends Pagination {
  items: Item[];

  constructor() {
    super();

    this.items = [];
  }
}

export class Item extends DatabaseRecord {
  externalID: string;
  name: string;
  details: string;
  belongsToAccount: number;

  constructor(
    id: number = 0,
    externalID: string = '',
    name: string = '',
    details: string = '',
    createdOn: number = 0,
    lastUpdatedOn?: number,
    archivedOn?: number,
    belongsToAccount: number = 0,
  ) {
    super(id, createdOn, lastUpdatedOn, archivedOn);
    this.name = name;
    this.externalID = externalID;
    this.details = details;
    this.belongsToAccount = belongsToAccount;
  }

  static areEqual = function (x: Item, y: Item): boolean {
    return (
      x.id === y.id &&
      x.externalID === y.externalID &&
      x.name === y.name &&
      x.details === y.details
    );
  };

  // this function should return everything there are no presumed fields
  static headers = (
    translations: Readonly<itemModelTranslations>,
  ): APITableHeader[] => {
    const columns = translations.columns;
    return [
      { content: columns.id, requiresAdmin: false },
      { content: columns.name, requiresAdmin: false },
      { content: columns.details, requiresAdmin: false },
      { content: columns.externalID, requiresAdmin: true },
      { content: columns.createdOn, requiresAdmin: false },
      { content: columns.lastUpdatedOn, requiresAdmin: true },
      { content: columns.archivedOn, requiresAdmin: true },
      { content: columns.belongsToAccount, requiresAdmin: true },
    ];
  };

  // this function should return everything there are no presumed fields
  static asRow = (x: Item): APITableCell[] => {
    return [
      new APITableCell({
        isIDCell: true,
        content: x.id.toString(),
      }),
      new APITableCell({
        content: x.name,
      }),
      new APITableCell({
        content: x.details,
      }),
      new APITableCell({
        content: x.externalID,
        requiresAdmin: true,
      }),
      new APITableCell({
        content: renderUnixTime(x.createdOn),
      }),
      new APITableCell({
        content: renderUnixTime(x.lastUpdatedOn),
        requiresAdmin: true,
      }),
      new APITableCell({
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
  externalID: Factory.Sync.each(() => faker.datatype.uuid()),
  details: Factory.Sync.each(() => faker.random.word()),
  belongsToAccount: Factory.Sync.each(() => faker.datatype.number()),
  ...defaultFactories,
});
