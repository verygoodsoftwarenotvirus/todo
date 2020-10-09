<script lang="typescript">
  import { Router, Route } from "svelte-routing";

  // components for this layout
  import PlainNavbar from "../components/Navbars/PlainNavbar.svelte";
  import Sidebar from "../components/Sidebar/Sidebar.svelte";
  import HeaderStats from "../components/Headers/HeaderStats.svelte";
  import FooterAdmin from "../components/Footers/FooterAdmin.svelte";

  // pages for this layout
  import Items from "../views/things/Items.svelte";

  export let location: Location;
  export let admin: string = "";
</script>

<div>
  <Sidebar location={location}/>
  <div class="relative md:ml-64 bg-gray-200">
    <PlainNavbar />
    <HeaderStats />
    <div class="px-4 md:px-10 mx-auto w-full -m-24">
      <Router url="things">
        <Route path="items" component="{Items}" />
      </Router>
      <FooterAdmin />
    </div>
  </div>
</div>
