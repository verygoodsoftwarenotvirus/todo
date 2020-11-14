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

    constructor(
      id: number = 0,
      name: string = '',
      details: string = '',
      createdOn: number = 0,
      belongsToUser: number = 0,
    ) {
        this.id = id;
        this.name = name;
        this.details = details;
        this.createdOn = createdOn;
        this.belongsToUser = belongsToUser;
    }

    static areEqual = function(
        x: Item,
        y: Item,
    ): boolean {
        return (
            x.id === y.id &&
            x.name === y.name &&
            x.details === y.details
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
