import * as dayjs from 'dayjs';

const defaultTimeFormat = 'HH:mm MM/DD/YYYY';

export function renderUnixTime(time?: number): string {
  return time ? dayjs.unix(time).format(defaultTimeFormat) : 'never';
}
