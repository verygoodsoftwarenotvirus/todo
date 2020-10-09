import faker from 'faker';
import * as Factory from "factory.ts";

export const idFactory = Factory.Sync.each(() => faker.random.number());

// yes I've seen this article: https://swizec.com/blog/a-day-is-not-606024-seconds-long/swizec/6755
const oneDayInMillis = 86400000;

export const createdOnFactory = Factory.Sync.each(() => {
  const yesterday = new Date(Date.now() - oneDayInMillis);
  const twelveHoursAgo = new Date(Date.now() - (oneDayInMillis/2));
  return faker.date.between(
    yesterday,
    twelveHoursAgo,
  ).getTime() / 1000;
});

export const lastUpdatedOnFactory = Factory.Sync.each(() => {
  const twelveHoursAgo = new Date(Date.now() - (oneDayInMillis/2));
  const now = new Date();
  return faker.date.between(
    twelveHoursAgo,
    now,
  ).getTime() / 1000;
});

export const archivedOnFactory = Factory.Sync.each(() => undefined);

export const defaultFactories = {
  id: idFactory,
  createdOn: createdOnFactory,
  lastUpdatedOn: lastUpdatedOnFactory,
  archivedOn: archivedOnFactory,
};
