import {parseBool} from "@/utils";

const enum SortBy {
    ASCENDING = "asc",
    DESCENDING = "desc",
}

export function NewQueryFilterFromCurrentPage() {
    const pageURLParams: URLSearchParams = new URLSearchParams(window.location.search);


}

const queryFilterKeyPage: string = "page";
const queryFilterKeyCreatedBefore: string = "createdBefore";
const queryFilterKeyCreatedAfter: string = "createdAfter";
const queryFilterKeyUpdatedBefore: string = "updatedBefore";
const queryFilterKeyUpdatedAfter: string = "updatedAfter";
const queryFilterKeyIncludeArchived: string = "includeArchived";
const queryFilterKeySortBy: string = "sortBy";

export class QueryFilter {
    page: number;
    limit: number;
    createdBefore: number;
    createdAfter: number;
    updatedBefore: number;
    updatedAfter: number;
    includeArchived: boolean;
    sortBy: SortBy;

    constructor() {
        this.page = 0;
        this.limit = 20;
        this.createdBefore = 0;
        this.createdAfter = 0;
        this.updatedBefore = 0;
        this.updatedAfter = 0;
        this.includeArchived = false;
        this.sortBy = SortBy.ASCENDING;
    }

    static fromURLSearchParams(input?: URLSearchParams): QueryFilter {
        const out = new QueryFilter();

        if (!input) {
            input = new URLSearchParams(window.location.search);
        }

        out.page = parseInt(input.get(queryFilterKeyPage) || '1');
        out.createdBefore = parseInt(input.get(queryFilterKeyCreatedBefore) || '0');
        out.createdAfter = parseInt(input.get(queryFilterKeyCreatedAfter) || '0');
        out.updatedBefore = parseInt(input.get(queryFilterKeyUpdatedBefore) || '0');
        out.updatedAfter = parseInt(input.get(queryFilterKeyUpdatedAfter) || '0');
        out.includeArchived = parseBool(input.get(queryFilterKeyIncludeArchived) || 'false');
        out.sortBy = input.get(queryFilterKeySortBy) as SortBy || SortBy.ASCENDING;

        return out
    }

    toURLSearchParams(adminMode: boolean = false): URLSearchParams {
        const out = new URLSearchParams();

        if (this.page !== 0) {
            out.set(queryFilterKeyPage, this.page.toString());
        }
        if (this.createdBefore !== 0) {
            out.set(queryFilterKeyCreatedBefore, this.createdBefore.toString());
        }
        if (this.createdAfter !== 0) {
            out.set(queryFilterKeyCreatedAfter, this.createdAfter.toString());
        }
        if (this.updatedBefore !== 0) {
            out.set(queryFilterKeyUpdatedBefore, this.updatedBefore.toString());
        }
        if (this.updatedAfter !== 0) {
            out.set(queryFilterKeyUpdatedAfter, this.updatedAfter.toString());
        }

        out.set(queryFilterKeyIncludeArchived, this.includeArchived.toString());
        out.set(queryFilterKeySortBy, this.sortBy);

        if (adminMode) {
            out.set("admin", "true");
        }

        return out
    }
}
