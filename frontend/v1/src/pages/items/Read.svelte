<script>
  import { Link } from "svelte-routing";

  let item = {};

  const itemID = window.location.pathname.replace("/items/", "");

  fetch(`/api/v1/items/${ itemID }`)
    .then(response => response.json())
    .then(data => {
      item = data;
    });
</script>

<!-- Items.svelte -->

<div style="text-align: center; margin: 0px auto;">
  <p>
    name:
    {@html item.name}
  </p>
  <p>
    details:
    {@html item.details}
  </p>

  <Link to="/items">see all</Link>

</div>
