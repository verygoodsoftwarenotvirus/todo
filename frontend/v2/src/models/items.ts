import * as Factory from "factory.ts";
import faker from "faker";

import {defaultFactories} from "@/models/fakes";

export class ItemList {
    page: number;
    limit: number;
    totalCount: number;
    items: Item[];

    constructor() {
        this.page = 0;
        this.limit = 0;
        this.totalCount = 0;
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
