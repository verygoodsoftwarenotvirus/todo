export interface APITableHeader {
  content: string;
  requiresAdmin?: boolean;
}

interface APITableCellParams {
  content: string;
  isIDCell?: boolean;
  isJSON?: boolean;
  requiresAdmin?: boolean;
}

export class APITableCell {
  content: string;
  isIDCell: boolean;
  isJSON: boolean;
  requiresAdmin: boolean;

  constructor({
    content = '',
    isIDCell = false,
    isJSON = false,
    requiresAdmin = false,
  }: APITableCellParams) {
    this.isIDCell = isIDCell;
    this.content = content;
    this.isJSON = isJSON;
    this.requiresAdmin = requiresAdmin;
  }
}
