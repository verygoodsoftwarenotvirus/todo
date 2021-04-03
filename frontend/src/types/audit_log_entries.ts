import { APITableCell, APITableHeader } from '@/components/APITable/types';
import type { auditLogEntryTableTranslations } from '@/i18n';
import { Pagination } from '@/types/api';
import { defaultFactories } from '@/types/fakes';
import { renderUnixTime } from '@/utils';
import * as Factory from 'factory.ts';
import faker from 'faker';

export class AuditLogEntryList extends Pagination {
  entries: AuditLogEntry[];

  constructor() {
    super();

    this.entries = [];
  }
}

export class AuditLogEntry {
  id: number;
  eventType: string;
  context: object;
  createdOn: number;

  constructor(
    id: number = 0,
    eventType: string = '',
    context: object = {},
    createdOn: number = 0,
  ) {
    this.id = id;
    this.eventType = eventType;
    this.context = context;
    this.createdOn = createdOn;
  }

  static areEqual = function (x: AuditLogEntry, y: AuditLogEntry): boolean {
    return (
      x.id === y.id && x.eventType === y.eventType && x.context === y.context
    );
  };

  // this function should return everything there are no presumed fields
  static headers = (
    translations: Readonly<auditLogEntryTableTranslations>,
  ): APITableHeader[] => {
    const columns = translations.columns;
    return [
      { content: columns.id, requiresAdmin: false },
      { content: columns.eventType, requiresAdmin: false },
      { content: columns.context, requiresAdmin: false },
      { content: columns.createdOn, requiresAdmin: false },
    ];
  };

  // this function should return everything there are no presumed fields
  static asRow = (x: AuditLogEntry): APITableCell[] => {
    return [
      new APITableCell({
        fieldName: 'id',
        content: x.id.toString(),
      }),
      new APITableCell({
        fieldName: 'eventType',
        content: x.eventType,
      }),
      new APITableCell({
        fieldName: 'context',
        content: JSON.stringify(x.context),
        isJSON: true,
      }),
      new APITableCell({
        fieldName: 'createdOn',
        content: renderUnixTime(x.createdOn),
      }),
    ];
  };
}

export const fakeAuditLogEntryFactory = Factory.Sync.makeFactory<AuditLogEntry>(
  {
    eventType: Factory.Sync.each(() => faker.random.word()),
    context: Factory.Sync.each(() => {
      return { performed_by: faker.random.number() };
    }),
    ...defaultFactories,
  },
);
