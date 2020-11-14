export interface APITableHeader {
  content: string;
  requiresAdmin: boolean;
}

export interface APITableCell {
  fieldName: string;
  content: string;
  requiresAdmin: boolean;
}
