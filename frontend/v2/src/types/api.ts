export class Pagination {
    page: number;
    limit: number;
    totalCount: number;

    constructor() {
        this.page = 0;
        this.limit = 0;
        this.totalCount = 0;
    }
}

export class ErrorResponse {
    message: string;
    code: number;

    constructor() {
        this.message = '';
        this.code = 0;
    }
}
