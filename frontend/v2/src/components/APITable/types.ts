export interface APITableHeader {
  content: string;
  requiresAdminMode: boolean;
}

export interface APITableCell {
  fieldName: string;
  content: string;
  requiresAdmin: boolean;
}
