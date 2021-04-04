import { APITableCell, APITableHeader } from '@/components/core/apiTable/types';
import type { webhookModelTranslations } from '@/i18n';
import { Pagination } from '@/types/api';
import { defaultFactories } from '@/types/fakes';
import { renderUnixTime } from '@/utils';
import * as Factory from 'factory.ts';
import faker from 'faker';

export class WebhookList extends Pagination {
  webhooks: Webhook[];

  constructor() {
    super();

    this.webhooks = [];
  }
}

export class Webhook {
  id: number;
  externalID: string;
  name: string;
  contentType: string;
  url: string;
  method: string;
  events: string[];
  dataTypes: string[];
  topics: string[];
  createdOn: number;
  lastUpdatedOn?: number;
  archivedOn?: number;
  belongsToAccount: number;

  constructor(
    id: number = 0,
    externalID: string = '',
    name: string = '',
    contentType: string = '',
    url: string = '',
    method: string = '',
    events: string[] = [],
    dataTypes: string[] = [],
    topics: string[] = [],
    createdOn: number = 0,
    lastUpdatedOn?: number,
    archivedOn?: number,
    belongsToAccount: number = 0,
  ) {
    this.id = id;
    this.externalID = externalID;
    this.name = name;
    this.contentType = contentType;
    this.url = url;
    this.method = method;
    this.events = events;
    this.dataTypes = dataTypes;
    this.topics = topics;
    this.createdOn = createdOn;
    this.lastUpdatedOn = lastUpdatedOn;
    this.archivedOn = archivedOn;
    this.belongsToAccount = belongsToAccount;
  }

  static areEqual = function (x: Webhook, y: Webhook): boolean {
    return (
      x.id === y.id &&
      x.name === y.name &&
      x.contentType === y.contentType &&
      x.url === y.url &&
      x.method === y.method &&
      x.events === y.events &&
      x.dataTypes === y.dataTypes &&
      x.topics === y.topics
    );
  };

  // this function should return everything there are no presumed fields
  static headers = (
    translations: Readonly<webhookModelTranslations>,
  ): APITableHeader[] => {
    const columns = translations.columns;
    return [
      { content: columns.id, requiresAdmin: false },
      { content: columns.name, requiresAdmin: false },
      { content: columns.contentType, requiresAdmin: false },
      { content: columns.url, requiresAdmin: false },
      { content: columns.method, requiresAdmin: false },
      { content: columns.events, requiresAdmin: false },
      { content: columns.dataTypes, requiresAdmin: false },
      { content: columns.topics, requiresAdmin: false },
      { content: columns.externalID, requiresAdmin: true },
      { content: columns.createdOn, requiresAdmin: true },
      { content: columns.lastUpdatedOn, requiresAdmin: true },
    ];
  };

  // this function should return everything there are no presumed fields
  static asRow = (x: Webhook): APITableCell[] => {
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
        fieldName: 'contentType',
        content: x.contentType,
      }),
      new APITableCell({ fieldName: 'url', content: x.url }),
      new APITableCell({
        fieldName: 'method',
        content: x.method,
      }),
      new APITableCell({
        fieldName: 'events',
        content: x.events.toString(),
      }),
      new APITableCell({
        fieldName: 'dataTypes',
        content: x.dataTypes.toString(),
      }),
      new APITableCell({
        fieldName: 'topics',
        content: (x.topics || []).toString(),
      }),
      new APITableCell({
        fieldName: 'externalID',
        content: x.externalID,
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
      }),
    ];
  };
}

export const fakeWebhookFactory = Factory.Sync.makeFactory<Webhook>({
  externalID: Factory.Sync.each(() => faker.random.word()),
  name: Factory.Sync.each(() => faker.random.word()),
  contentType: Factory.Sync.each(() => faker.random.word()),
  url: Factory.Sync.each(() => faker.random.word()),
  method: Factory.Sync.each(() => faker.random.word()),
  events: Factory.Sync.each(() => faker.random.words().split(' ')),
  dataTypes: Factory.Sync.each(() => faker.random.words().split(' ')),
  topics: Factory.Sync.each(() => faker.random.words().split(' ')),
  belongsToAccount: Factory.Sync.each(() => faker.random.number()),
  ...defaultFactories,
});

export class WebhookCreationInput {
  name: string;
  contentType: string;
  url: string;
  method: string;
  events: string[];
  dataTypes: string[];
  topics: string[];

  constructor(
    name: string = '',
    contentType: string = '',
    url: string = '',
    method: string = '',
    events: string[] = [],
    dataTypes: string[] = [],
    topics: string[] = [],
  ) {
    this.name = name;
    this.contentType = contentType;
    this.url = url;
    this.method = method;
    this.events = events;
    this.dataTypes = dataTypes;
    this.topics = topics;
  }
}
