import * as Factory from "factory.ts";
import faker from "faker";

import { Pagination } from "@/types/api";
import { defaultFactories } from "@/types/fakes";
import type { APITableCell, APITableHeader } from "@/components/APITable/types";
import { renderUnixTime } from "@/utils";
import type {auditLogEntryTableTranslations} from "@/i18n";

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

    constructor() {
        this.id = 0;
        this.eventType = "";
        this.context = {};
        this.createdOn = 0;
    }

    static areEqual = function(
        e1: AuditLogEntry,
        e2: AuditLogEntry,
    ): boolean {
        return (
            e1.id === e2.id &&
            e1.eventType === e2.eventType &&
            e1.context === e2.context
        );
    }

    // this function should return everything there are no presumed fields
    static headers = (translations: Readonly<auditLogEntryTableTranslations>): APITableHeader[] => {
        const columns = translations.columns;
        return [
            {content: columns.id, requiresAdmin: false},
            {content: columns.eventType, requiresAdmin: false},
            {content: columns.context, requiresAdmin: false},
            {content: columns.createdOn, requiresAdmin: false},
        ];
    }

    // this function should return everything there are no presumed fields
    static asRow = (x: AuditLogEntry): APITableCell[] => {
        return [
            { fieldName: 'id', content: x.id.toString(), requiresAdmin: false },
            { fieldName: 'eventType', content: x.eventType, requiresAdmin: false },
            { fieldName: 'context', content: JSON.stringify(x.context), requiresAdmin: false },
            { fieldName: 'createdOn', content: renderUnixTime(x.createdOn), requiresAdmin: false },
        ]
    }
}

export const fakeAuditLogEntryFactory = Factory.Sync.makeFactory<AuditLogEntry> ({
    eventType: Factory.Sync.each(() =>  faker.random.word()),
    context: Factory.Sync.each(() =>  faker.random.word()),
    ...defaultFactories,
});
