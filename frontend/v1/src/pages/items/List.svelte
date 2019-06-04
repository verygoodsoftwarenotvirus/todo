<script>
  import { Link, navigate } from "svelte-routing";

  import Table from "./components/Table.svelte";

  const columns = [
    {
      title: "ID",
      key: "id"
    },
    {
      title: "Name",
      key: "name"
    },
    {
      title: "Details",
      key: "details"
    },
    {
      title: "Created On",
      key: "created_on"
    },
    {
      title: "Updated On",
      key: "updated_on"
    }
  ];
  let items = [];

  function deleteItem(row) {
    if (confirm("are you sure you want to delete this item?")) {
      fetch(`/api/v1/items/${row.id}`, {
        method: "DELETE"
      }).then(response => {
        if (response.status != 204) {
          console.error("something has gone awry");
        }
        items = items.filter(item => {
          return item.id != row.id;
        });
      });
    }
  }

  function goToItem(row) {
    navigate(`/items/${row.id}`, { replace: true });
  }

  fetch("/api/v1/items/")
    .then(response => response.json())
    .then(data => {
      items = data["items"];
    });
</script>

<!-- Items.svelte -->

<Table
  {columns}
  tableStyle={'margin: 0px auto;'}
  rows={items}
  rowClickFunc={goToItem}
  rowDeleteFunc={deleteItem} />
