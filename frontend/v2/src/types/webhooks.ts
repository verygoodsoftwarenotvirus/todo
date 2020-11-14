import * as Factory from "factory.ts";
import faker from "faker";

import { Pagination } from "@/types/api";
import { defaultFactories } from "@/types/fakes";
import type { APITableCell, APITableHeader } from "@/components/APITable/types";
import { renderUnixTime } from "@/utils";
import type {webhookModelTranslations} from "@/i18n";

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
    updatedOn?: number;
    archivedOn?: number;
    belongsToUser: number;

    constructor() {
        this.id = 0;
        this.name = "";
        this.contentType = "";
        this.url = "";
        this.method = "";
        this.events = [];
        this.dataTypes = [];
        this.topics = [];
        this.createdOn = 0;
        this.belongsToUser = 0;
    }

    static areEqual = function(
        x: Webhook,
        y: Webhook,
    ): boolean {
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
    }

    // this function should return everything there are no presumed fields
    static headers = (translations: Readonly<webhookModelTranslations>): APITableHeader[] => {
        const columns = translations.columns;
        return [
            {content: columns.id, requiresAdmin: false},
            {content: columns.name, requiresAdmin: false},
            {content: columns.contentType, requiresAdmin: false},
            {content: columns.url, requiresAdmin: false},
            {content: columns.method, requiresAdmin: false},
            {content: columns.events, requiresAdmin: false},
            {content: columns.dataTypes, requiresAdmin: false},
            {content: columns.topics, requiresAdmin: false},
            {content: columns.createdOn, requiresAdmin: false},
            {content: columns.lastUpdatedOn, requiresAdmin: false},
            {content: columns.belongsToUser, requiresAdmin: true},
        ];
    }

    // this function should return everything there are no presumed fields
    static asRow = (x: Webhook): APITableCell[] => {
        return [
            { fieldName: 'id', content: x.id.toString(), requiresAdmin: false },
            { fieldName: 'name', content: x.name, requiresAdmin: false },
            { fieldName: 'contentType', content: x.contentType, requiresAdmin: false },
            { fieldName: 'url', content: x.url, requiresAdmin: false },
            { fieldName: 'method', content: x.method, requiresAdmin: false },
            { fieldName: 'events', content: x.events.toString(), requiresAdmin: false },
            { fieldName: 'dataTypes', content: x.dataTypes.toString(), requiresAdmin: false },
            { fieldName: 'topics', content: x.topics.toString(), requiresAdmin: false },
            { fieldName: 'createdOn', content: renderUnixTime(x.createdOn), requiresAdmin: false },
            { fieldName: 'lastUpdatedOn', content: renderUnixTime(x.updatedOn), requiresAdmin: false },
            { fieldName: 'belongsToUser', content: x.belongsToUser.toString(), requiresAdmin: true },
        ]
    }
}

export const fakeWebhookFactory = Factory.Sync.makeFactory<Webhook> ({
    name: Factory.Sync.each(() =>  faker.random.word()),
    url: Factory.Sync.each( () => faker.internet.url()),
    method: Factory.Sync.each( () => faker.hacker.verb()),
    contentType: "application/fake",
    events: ["things", "and", "stuff"],
    dataTypes: ["stuff", "and", "things"],
    topics: ["blah", "blarg", "blorp"],
    belongsToUser: Factory.Sync.each(() =>  faker.random.number()),
    ...defaultFactories,
});

export class WebhookCreationInput {
    contentType: string;
    url: string;
    method: string;
    events: string[];
    dataTypes: string[];
    topics: string[];

    constructor(
        contentType: string = "",
        url: string = "",
        method: string = "",
        events: string[] = [],
        dataTypes: string[] = [],
        topics: string[] = [],
    ) {
        this.contentType = contentType;
        this.url = url;
        this.method = method;
        this.events = events;
        this.dataTypes = dataTypes;
        this.topics = topics;
    }
}
