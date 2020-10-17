<script lang="typescript">
  import {onDestroy, onMount} from "svelte";
  import axios, {AxiosResponse, AxiosError} from "axios";
  import {Router, Route, navigate} from "svelte-routing";

  // components for this layout
  import Sidebar from "../components/Sidebar/Sidebar.svelte";
  import FooterAdmin from "../components/Footers/FooterAdmin.svelte";
  import AdminNavbar from "../components/Navbars/AdminNavbar.svelte";

  // custom components for this layout
  import ReadUpdateDeleteItem from "../components/Things/RUD/ReadUpdateDeleteItem.svelte";
  import CreateItem from "../components/Things/Creation/CreateItem.svelte";

  // pages for this layout
  import Items from "../views/things/Items.svelte";

  export let location: Location;

  import {Logger} from "../logger";

  let logger = new Logger();

  import {authStatusStore} from "../stores";
  import {UserStatus} from "../models";

  let currentAuthStatus = {};
  const unsubscribeFromAuthStatusUpdates = authStatusStore.subscribe((value: UserStatus) => {
    currentAuthStatus = value;
    if (!currentAuthStatus) {
      navigate("/auth/login", {state: {}, replace: true});
    }
  });
  // onDestroy(unsubscribeFromAuthStatusUpdates);
</script>

<div>
  <Sidebar location={location}/>
  <div class="relative md:ml-64 bg-gray-200">
    <AdminNavbar />
    <div class="px-4 md:px-10 mx-auto w-full -m-24">
      <Router url="things">
        <Route path="items" component="{Items}" />
        <Route path="items/:id" let:params>
          <ReadUpdateDeleteItem id="{params.id}" />
        </Route>
        <Route path="items/new" component="{CreateItem}" />
      </Router>
      <FooterAdmin />
    </div>
  </div>
</div>
