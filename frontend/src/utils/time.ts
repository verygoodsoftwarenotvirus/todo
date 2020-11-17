import * as dayjs from 'dayjs';

export function renderUnixTime(time?: number): string {
  if (time && time !== 0) {
    return dayjs.unix(time).format('HH:mm MM/DD/YYYY');
  }
  return 'never';
}
