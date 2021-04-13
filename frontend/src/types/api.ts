export class DatabaseRecord {
  id: number;
  createdOn: number;
  
  constructor(
    id: number = 0,
    createdOn: number = 0,
  ) {
    this.id = id;
    this.createdOn = createdOn;
  }
}

export class Pagination {
  page: number;
  limit: number;
  totalCount: number;

  constructor(page: number = 0, limit: number = 0, totalCount: number = 0) {
    this.page = page;
    this.limit = limit;
    this.totalCount = totalCount;
  }
}

export class ErrorResponse {
  message: string;
  code: number;

  constructor(message: string = '', code: number = 0) {
    this.message = message;
    this.code = code;
  }
}
