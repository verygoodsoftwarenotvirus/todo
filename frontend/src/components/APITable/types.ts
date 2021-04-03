export interface APITableHeader {
  content: string;
  requiresAdmin?: boolean;
}

interface APITableCellParams {
  fieldName?: string;
  content?: string;
  isJSON?: boolean;
  requiresAdmin?: boolean;
}

export class APITableCell {
  fieldName: string;
  content: string;
  isJSON: boolean;
  requiresAdmin: boolean;

  constructor({
    fieldName = '',
    content = '',
    isJSON = false,
    requiresAdmin = false,
  }: APITableCellParams) {
    this.fieldName = fieldName;
    this.content = content;
    this.isJSON = isJSON;
    this.requiresAdmin = requiresAdmin;
  }
}
