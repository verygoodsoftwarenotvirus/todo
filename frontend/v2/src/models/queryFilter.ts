const enum SortBy {
    ASCENDING = "asc",
    DESCENDING = "desc",
}

export class QueryFilter {
    page: number;
    createdBefore: number;
    createdAfter: number;
    updatedBefore: number;
    updatedAfter: number;
    includeArchived: boolean;
    sortBy: SortBy;

    constructor() {
        this.page = 0;
        this.createdBefore = 0;
        this.createdAfter = 0;
        this.updatedBefore = 0;
        this.updatedAfter = 0;
        this.includeArchived = false;
        this.sortBy = SortBy.ASCENDING;
    }
}
