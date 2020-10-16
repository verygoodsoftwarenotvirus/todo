import * as Factory from "factory.ts";
import faker from "faker";

import { Pagination } from "@/models/api";
import {defaultFactories} from "@/models/fakes";

export class UserList extends Pagination {
    items: User[];

    constructor() {
        super();

        this.items = [];
    }
}

export class User {
    id: number;
    username: string;
    isAdmin: boolean;
    requiresPasswordChange: boolean;
    passwordLastChangedOn: number;
    createdOn: number;
    lastUpdatedOn: number;
    archivedOn?: number;

    constructor() {
        this.id = 0;
        this.username = '';
        this.isAdmin = false;
        this.requiresPasswordChange = false;
        this.passwordLastChangedOn = 0;
        this.createdOn = 0;
        this.lastUpdatedOn = 0;
        this.archivedOn = 0;
    }

    static areEqual = function(
        u1: User,
        u2: User,
    ): boolean {
        return (
            u1.id === u2.id &&
            u1.username === u2.username &&
            u1.isAdmin === u2.isAdmin &&
            u1.requiresPasswordChange === u2.requiresPasswordChange
        );
    }
}

export const fakeUserFactory = Factory.Sync.makeFactory<User> ({
    username: Factory.Sync.each(() =>  faker.random.word()),
    isAdmin: Factory.Sync.each(() =>  faker.random.boolean()),
    requiresPasswordChange: Factory.Sync.each(() =>  faker.random.boolean()),
    passwordLastChangedOn: Factory.Sync.each(() =>  faker.random.number()),
    ...defaultFactories,
});

export class UserRegistrationResponse {
    id: number;
    username: string;
    isAdmin: boolean;
    qrCode: string;
    createdOn: number;
    lastUpdatedOn: number;
    archivedOn: number;
    passwordLastChangedOn: number;

    constructor() {
        this.id = 0;
        this.username = '';
        this.isAdmin = false;
        this.qrCode = '';
        this.createdOn = 0;
        this.lastUpdatedOn = 0;
        this.archivedOn = 0;
        this.passwordLastChangedOn = 0;
    }
}