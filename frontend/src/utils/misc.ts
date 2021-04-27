export function isNumeric(input: string) {
  return !isNaN(parseFloat(input)); // ...and ensure strings of whitespace fail
}

export function parseBool(input: string): boolean | null {
  switch (input.trim().toLowerCase()) {
    case '1':
    case 't':
    case 'true':
      return true;
    case '0':
    case 'f':
    case 'false':
      return false;
    default:
      return null;
  }
}
