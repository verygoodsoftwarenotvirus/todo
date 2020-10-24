<script lang="typescript">
  import axios, {AxiosError, AxiosResponse} from "axios";
  import { onDestroy } from "svelte";
  import { navigate, Router, Route } from "svelte-routing";

  import { userStatusStore } from "../stores";
  import { UserStatus } from "../models";

  // components for this layout
  import AdminNavbar from "../components/Navbars/AdminNavbar.svelte";
  import Sidebar from "../components/Sidebar/Sidebar.svelte";
  import AdminFooter from "../components/Footers/AdminFooter.svelte";

  // pages for this layout
  import Dashboard from "../views/admin/Dashboard.svelte";
  import Settings from "../views/admin/Settings.svelte";
  import ReadUpdateDeleteUser from "../components/Things/ReadUpdateDelete/User.svelte";

  import { Logger } from "../logger";
  let logger = new Logger().withDebugValue("source", "src/layouts/Admin.svelte");

  export let location: Location;

  let currentAuthStatus = {};
  const unsubscribeFromUserStatusUpdates = userStatusStore.subscribe((value: UserStatus) => {
    currentAuthStatus = value;
    // if (!currentAuthStatus || !currentAuthStatus.isAuthenticated || !currentAuthStatus.isAdmin) {
    //   logger.debug(`navigating to /auth/login because the user is not authenticated`);
    //   navigate("/auth/login", { state: {}, replace: true });
    // }
  });
  onDestroy(unsubscribeFromUserStatusUpdates);
</script>

<div>
  <Sidebar location={location}/>
  <div class="relative md:ml-64 bg-gray-200">
    <AdminNavbar />
    <div class="px-4 md:px-10 mx-auto w-full -m-24">
      <Router url="admin">
        <Route path="dashboard" component="{Dashboard}" />
        <Route path="settings" component="{Settings}" />
<!--        <Route path="users" component="{AdminUsersTable}" />-->
        <Route path="users/:id" let:params>
          <ReadUpdateDeleteUser id="{params.id}" />
        </Route>
      </Router>
      <AdminFooter />
    </div>
  </div>
</div>
