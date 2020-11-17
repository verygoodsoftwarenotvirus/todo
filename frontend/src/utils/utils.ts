export function isNumeric(input: string) {
  return !isNaN(parseFloat(input)); // ...and ensure strings of whitespace fail
}
