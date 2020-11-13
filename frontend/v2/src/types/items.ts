import * as Factory from "factory.ts";
import faker from "faker";

import { Pagination } from "@/types/api";
import { defaultFactories } from "@/types/fakes";
import type { APITableCell, APITableHeader } from "@/components/APITable/types";
import { renderUnixTime } from "@/utils";
import type {itemModelTranslations} from "@/i18n";

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
    updatedOn?: number;
    archivedOn?: number;
    belongsToUser: number;

    constructor() {
        this.id = 0;
        this.name = "";
        this.details = "";
        this.createdOn = 0;
        this.belongsToUser = 0;
    }

    static areEqual = function(
        i1: Item,
        i2: Item,
    ): boolean {
        return (
            i1.id === i2.id &&
            i1.name === i2.name &&
            i1.details === i2.details
        );
    }

    // this function should return everything there are no presumed fields
    static headers = (translations: Readonly<itemModelTranslations>): APITableHeader[] => {
        const columns = translations.columns;
        return [
            {content: columns.id, requiresAdmin: false},
            {content: columns.name, requiresAdmin: false},
            {content: columns.details, requiresAdmin: false},
            {content: columns.createdOn, requiresAdmin: false},
            {content: columns.lastUpdatedOn, requiresAdmin: false},
            {content: columns.belongsToUser, requiresAdmin: true},
        ];
    }

    // this function should return everything there are no presumed fields
    static asRow = (x: Item): APITableCell[] => {
        return [
            { fieldName: 'id', content: x.id.toString(), requiresAdmin: false },
            { fieldName: 'name', content: x.name, requiresAdmin: false },
            { fieldName: 'details', content: x.details, requiresAdmin: false },
            { fieldName: 'createdOn', content: renderUnixTime(x.createdOn), requiresAdmin: false },
            { fieldName: 'lastUpdatedOn', content: renderUnixTime(x.updatedOn), requiresAdmin: false },
            { fieldName: 'belongsToUser', content: x.belongsToUser.toString(), requiresAdmin: true },
        ]
    }
}

export const fakeItemFactory = Factory.Sync.makeFactory<Item> ({
    name: Factory.Sync.each(() =>  faker.random.word()),
    details: Factory.Sync.each(() =>  faker.random.word()),
    belongsToUser: Factory.Sync.each(() =>  faker.random.number()),
    ...defaultFactories,
});

export class ItemCreationInput {
    name: string;
    details: string;

    constructor() {
        this.name = "";
        this.details = "";
    }
}
