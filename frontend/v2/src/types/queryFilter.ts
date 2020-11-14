import { parseBool } from "@/utils";

const enum SortBy {
  ASCENDING = "asc",
  DESCENDING = "desc",
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

  constructor(
    page: number = 0,
    limit: number = 0,
    createdBefore: number = 0,
    createdAfter: number = 0,
    updatedBefore: number = 0,
    updatedAfter: number = 0,
    includeArchived: boolean = false,
    sortBy: SortBy = SortBy.ASCENDING
  ) {
    this.page = page;
    this.limit = limit;
    this.createdBefore = createdBefore;
    this.createdAfter = createdAfter;
    this.updatedBefore = updatedBefore;
    this.updatedAfter = updatedAfter;
    this.includeArchived = includeArchived;
    this.sortBy = sortBy;
  }

  static fromURLSearchParams(input?: URLSearchParams): QueryFilter {
    const out = new QueryFilter();

    if (!input) {
      input = new URLSearchParams(window.location.search);
    }

    out.page = parseInt(input.get(queryFilterKeyPage) || "1");
    out.createdBefore = parseInt(input.get(queryFilterKeyCreatedBefore) || "0");
    out.createdAfter = parseInt(input.get(queryFilterKeyCreatedAfter) || "0");
    out.updatedBefore = parseInt(input.get(queryFilterKeyUpdatedBefore) || "0");
    out.updatedAfter = parseInt(input.get(queryFilterKeyUpdatedAfter) || "0");
    out.includeArchived = parseBool(
      input.get(queryFilterKeyIncludeArchived) || "false"
    );
    out.sortBy =
      (input.get(queryFilterKeySortBy) as SortBy) || SortBy.ASCENDING;

    return out;
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

    return out;
  }
}
