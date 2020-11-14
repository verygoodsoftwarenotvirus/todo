export function parseBool(input: string): boolean {
  switch (input.trim().toLowerCase()) {
    case "1":
    case "t":
    case "true":
      return true;
    case "0":
    case "f":
    case "false":
      return false;
    default:
      return false;
  }
}
