<script lang="typescript">
  import { onDestroy } from "svelte";
  import { Router, Route, navigate } from "svelte-routing";

  // components for this layout
  import Sidebar from "../components/Sidebar/Sidebar.svelte";
  import FooterAdmin from "../components/Footers/FooterAdmin.svelte";
  import AdminNavbar from "../components/Navbars/AdminNavbar.svelte";

  // custom components for this layout
  import ReadUpdateDeleteItem from "../components/Things/ReadUpdateDelete/Item.svelte";
  import CreateItem from "../components/Things/Creation/CreateItem.svelte";

  // pages for this layout
  import Items from "../views/things/Items.svelte";

  import {authStatusStore} from "../stores";
  import {UserStatus} from "../models";
  import {Logger} from "../logger";

  export let location: Location;

  let logger = new Logger().withDebugValue("source", "src/layouts/Things.svelte");

  let currentAuthStatus = {};
  const unsubscribeFromAuthStatusUpdates = authStatusStore.subscribe((value: UserStatus) => {
    currentAuthStatus = value;
    // if (!currentAuthStatus || !currentAuthStatus.isAuthenticated) {
    //   logger.debug(`navigating to /auth/login because user is unauthenticated`);
    //   navigate("/auth/login", {state: {}, replace: true});
    // }
  });
  onDestroy(unsubscribeFromAuthStatusUpdates);
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
