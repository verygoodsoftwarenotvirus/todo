// QueryFilter keys
import { isNumeric, parseBool } from './misc';

const queryFilterKeyPage = 'page',
  queryFilterKeyCreatedBefore = 'createdBefore',
  queryFilterKeyCreatedAfter = 'createdAfter',
  queryFilterKeyUpdatedBefore = 'updatedBefore',
  queryFilterKeyUpdatedAfter = 'updatedAfter',
  queryFilterKeyIncludeArchived = 'includeArchived',
  queryFilterKeySortBy = 'sortBy';

const validQueryFilterKeys: string[] = [
  queryFilterKeyPage,
  queryFilterKeyCreatedBefore,
  queryFilterKeyCreatedAfter,
  queryFilterKeyUpdatedBefore,
  queryFilterKeyUpdatedAfter,
  queryFilterKeyIncludeArchived,
  queryFilterKeySortBy,
];

export function inheritQueryFilterSearchParams(
  pageURLParams: URLSearchParams,
): URLSearchParams {
  const outboundURLParams: URLSearchParams = new URLSearchParams();

  validQueryFilterKeys.forEach((key: string) => {
    const x = (pageURLParams.get(key) || '').trim();

    if (x) {
      if (
        key === queryFilterKeyIncludeArchived &&
        parseBool(x.toLowerCase()) !== null
      ) {
        outboundURLParams.set(key, x);
      } else if (
        key === queryFilterKeySortBy &&
        ['asc', 'desc'].includes(x.toLowerCase())
      ) {
        outboundURLParams.set(key, x);
      } else if (isNumeric(x)) {
        outboundURLParams.set(key, x);
      }
    }
  });

  return outboundURLParams;
}
