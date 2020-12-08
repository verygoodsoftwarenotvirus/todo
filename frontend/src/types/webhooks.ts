import { APITableCell, APITableHeader } from '@/components/APITable/types';
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
  belongsToUser: number;

  constructor() {
    this.id = 0;
    this.name = '';
    this.contentType = '';
    this.url = '';
    this.method = '';
    this.events = [];
    this.dataTypes = [];
    this.topics = [];
    this.createdOn = 0;
    this.belongsToUser = 0;
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
      { content: columns.id, requiresAdminMode: false },
      { content: columns.name, requiresAdminMode: false },
      { content: columns.contentType, requiresAdminMode: false },
      { content: columns.url, requiresAdminMode: false },
      { content: columns.method, requiresAdminMode: false },
      { content: columns.events, requiresAdminMode: false },
      { content: columns.dataTypes, requiresAdminMode: false },
      { content: columns.topics, requiresAdminMode: false },
      { content: columns.createdOn, requiresAdminMode: false },
      { content: columns.lastUpdatedOn, requiresAdminMode: false },
      { content: columns.belongsToUser, requiresAdminMode: true },
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
        content: x.topics.toString(),
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
        fieldName: 'belongsToUser',
        content: x.belongsToUser.toString(),
      }),
    ];
  };
}

export const fakeWebhookFactory = Factory.Sync.makeFactory<Webhook>({
  name: Factory.Sync.each(() => faker.random.word()),
  url: Factory.Sync.each(() => faker.internet.url()),
  method: Factory.Sync.each(() => faker.hacker.verb()),
  contentType: 'application/fake',
  events: ['things', 'and', 'stuff'],
  dataTypes: ['stuff', 'and', 'things'],
  topics: ['blah', 'blarg', 'blorp'],
  belongsToUser: Factory.Sync.each(() => faker.random.number()),
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
