<script lang="typescript">
  import {onMount} from "svelte";
  import axios, { AxiosResponse, AxiosError } from "axios";
  import {Router, Route, navigate} from "svelte-routing";

  // components for this layout
  import PlainNavbar from "../components/Navbars/PlainNavbar.svelte";
  import Sidebar from "../components/Sidebar/Sidebar.svelte";
  import HeaderStats from "../components/Headers/HeaderStats.svelte";
  import FooterAdmin from "../components/Footers/FooterAdmin.svelte";

  // pages for this layout
  import ItemsTablePage from "../views/things/ItemsTableContainer.svelte";
  import ReadUpdateDeleteItem from "../components/Things/Items/ReadUpdateDeleteItem.svelte";
  import CreateItem from "../components/Things/Items/CreateItem.svelte";

  import {AuthStatus} from "../models";
  import {authStatusStore} from "../stores";

  export let location: Location;
  export let admin: string = "";

  onMount(async () => {
    console.debug("checking status from Things layout");
    axios.get("/users/status", { withCredentials: true })
         .then((res: AxiosResponse<AuthStatus>) => {
           authStatusStore.setAuthStatus(res.data);
         })
         .catch((error: AxiosError) => {
           navigate("/auth/login", { state: {}, replace: true });
         });
  })
</script>

<div>
  <Sidebar location={location}/>
  <div class="relative md:ml-64 bg-gray-200">
    <PlainNavbar />
    <HeaderStats />
    <div class="px-4 md:px-10 mx-auto w-full -m-24">
      <Router url="things">
        <Route path="items" component="{ItemsTablePage}" />
        <Route path="items/:id" let:params>
          <ReadUpdateDeleteItem id="{params.id}" />
        </Route>
        <Route path="items/new" component="{CreateItem}" />
      </Router>
      <FooterAdmin />
    </div>
  </div>
</div>
