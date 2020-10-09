<script lang="typescript">
  // core components
  import TableDropdown from "../TableDropdowns/TableDropdown.svelte";

  import { renderUnixTime } from "../../utils"
  import { Item, fakeItemFactory } from "../../models/"

  let items: Item[] = fakeItemFactory.buildList(10);

  console.dir(items);

</script>

<div
  class="relative flex flex-col min-w-0 break-words w-full mb-6 shadow-lg rounded bg-white"
>
  <div class="rounded-t mb-0 px-4 py-3 border-0">
    <div class="flex flex-wrap items-center">
      <div class="relative w-full px-4 max-w-full flex-grow flex-1">
        <h3
          class="font-semibold text-lg text-gray-800"
        >
          Items
        </h3>
      </div>
    </div>
  </div>
  <div class="block w-full overflow-x-auto">
    <!-- Projects table -->
    <table class="items-center w-full bg-transparent border-collapse">
      <thead>
        <tr>
          <th
            class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200"
          >
            ID
          </th>
          <th
            class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200"
          >
            Name
          </th>
          <th
            class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200"
          >
            Details
          </th>
          <th
            class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200"
          >
            Created On
          </th>
          <th
            class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200"
          >
            Last Updated On
          </th>
          <th
            class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200"
          ></th>
        </tr>
      </thead>
      <tbody>
      {#each items as item}
        <tr>
          <th
            class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4 text-left flex items-center"
          >
            <span class="ml-3 font-bold btext-gray-700">{item.id}</span>
          </th>
          <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4">
            {item.name}
          </td>
          <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4">
            {item.details}
          </td>
          <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4">
            {renderUnixTime(item.createdOn)}
          </td>
          <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4">
            {renderUnixTime(item.lastUpdatedOn)}
          </td>
          <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4 text-right">
            <TableDropdown />
          </td>
        </tr>
      {/each}

        <!--
        <tr>
          <th
            class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4 text-left flex items-center"
          >
            <img
              src="https://picsum.photos/seed/todo/128/128"
              class="h-12 w-12 bg-white rounded-full border"
              alt="..."
            />
            <span
              class="ml-3 font-bold btext-gray-700"
            >
              React Material Dashboard
            </span>
          </th>
          <td
            class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4"
          >
            $4,400 USD
          </td>
          <td
            class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4"
          >
            <i class="fas fa-circle text-teal-500 mr-2"></i> on schedule
          </td>
          <td
            class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4"
          >
            <div class="flex">
              <img
                src="https://picsum.photos/seed/todo/800/800"
                alt="..."
                class="w-10 h-10 rounded-full border-2 border-gray-100 shadow"
              />
              <img
                src="https://picsum.photos/seed/todo/800/800"
                alt="..."
                class="w-10 h-10 rounded-full border-2 border-gray-100 shadow -ml-4"
              />
              <img
                src="https://picsum.photos/seed/todo/800/800"
                alt="..."
                class="w-10 h-10 rounded-full border-2 border-gray-100 shadow -ml-4"
              />
              <img
                src="https://picsum.photos/seed/todo/800/800"
                alt="..."
                class="w-10 h-10 rounded-full border-2 border-gray-100 shadow -ml-4"
              />
            </div>
          </td>
          <td
            class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4"
          >
            <div class="flex items-center">
              <span class="mr-2">90%</span>
              <div class="relative w-full">
                <div
                  class="overflow-hidden h-2 text-xs flex rounded bg-teal-200"
                >
                  <div
                    style="width: 90%;"
                    class="shadow-none flex flex-col text-center whitespace-nowrap text-white justify-center bg-teal-500"
                  ></div>
                </div>
              </div>
            </div>
          </td>
          <td
            class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4 text-right"
          >
            <TableDropdown />
          </td>
        </tr>
        -->

      </tbody>
    </table>
  </div>
</div>
