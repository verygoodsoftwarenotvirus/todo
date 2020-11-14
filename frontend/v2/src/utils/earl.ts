// QueryFilter keys
const queryFilterKeyPage = "page";
const queryFilterKeyCreatedBefore = "createdBefore";
const queryFilterKeyCreatedAfter = "createdAfter";
const queryFilterKeyUpdatedBefore = "updatedBefore";
const queryFilterKeyUpdatedAfter = "updatedAfter";
const queryFilterKeyIncludeArchived = "includeArchived";
const queryFilterKeySortBy = "sortBy";

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
  pageURLParams: URLSearchParams
): URLSearchParams {
  // const pageURLParams: URLSearchParams = new URLSearchParams(window.location.search);
  const outboundURLParams: URLSearchParams = new URLSearchParams();

  validQueryFilterKeys.forEach((key: string, _: number) => {
    const x = pageURLParams.get(key);

    if (x) {
      if (key === queryFilterKeyIncludeArchived) {
        const val = (x as string).toLowerCase().trim();
        if (val === "true" || val === "false") {
          outboundURLParams.set(key, x);
        }
      } else if (key === queryFilterKeySortBy) {
        const val = (x as string).toLowerCase().trim();
        if (val === "asc" || val === "desc") {
          outboundURLParams.set(key, x);
        }
      } else {
        // assumed numeric here
        outboundURLParams.set(key, x);
      }
    }
  });

  return outboundURLParams;
}
