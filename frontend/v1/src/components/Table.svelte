<script>
  export let columns = [];
  export let rows = [];
  export let tableStyle = "";
  export let rowClickFunc = () => {};
  export let rowDeleteFunc = () => {};

  let sortOrder = 1;
  let sortKey = "";
  let sortBy = r => "";
  let filterSettings = {};
  let columnByKey = {};

  columns.forEach(col => {
    columnByKey[col.key] = col;
  });

  $: c_rows = rows
    .filter(r =>
      Object.keys(filterSettings).every(f => {
        return (
          filterSettings[f] === undefined ||
          filterSettings[f] === columnByKey[f].filterValue(r)
        );
      })
    )
    .map(r => {
      return { ...r, $sortOn: sortBy(r) };
    })
    .sort((a, b) => {
      if (a.$sortOn > b.$sortOn) return sortOrder;
      else if (a.$sortOn < b.$sortOn) return -sortOrder;
      return 0;
    });

  const handleSort = col => {
    if (!col.unsortable) {
      if (sortKey === col.key) {
        sortOrder = sortOrder === 1 ? -1 : 1;
      } else {
        sortOrder = 1;
        sortKey = col.key;
        sortBy = r => r[sortKey];
      }
    }
  };
</script>

<style>
  .isSortable {
    cursor: pointer;
  }
</style>

<!-- heavily borrowed from/inspired by https://github.com/dasDaniel/svelte-table/blob/402a9eb3803ae2367f19651bddcb26ff46d29601/src/SvelteTable.svelte -->

{#if columns.length === 0}
  <h4>no data available :(</h4>
{:else}
  <table style={tableStyle}>
    <tr>
      {#each columns as col}
        <th
          on:click={() => handleSort(col)}
          class={col.sortable ? 'isSortable' : ''}>
           {col.title}
          {#if sortKey === col.key}{sortOrder === 1 ? '▲' : '▼'}{/if}
        </th>
      {/each}
      <th>
        <!--🗑 -->
      </th>
    </tr>
    {#each c_rows as row}
      <tr on:click={() => rowClickFunc(row)} style={row._style || ''}>
        {#each columns as col}
          <td>
            {@html col.renderValue ? col.renderValue(row) : row[col.key]}
          </td>
        {/each}
        <td>
          <button on:click={() => rowDeleteFunc(row)}>🗑️</button>
        </td>
      </tr>
    {/each}
  </table>
{/if}
