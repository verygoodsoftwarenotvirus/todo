<script lang="typescript">
  import axios, {AxiosError, AxiosResponse} from "axios";
  import { onDestroy } from "svelte";
  import { navigate, Router, Route } from "svelte-routing";

  import { authStatusStore } from "../stores";
  import { UserStatus } from "../models";

  // components for this layout
  import AdminNavbar from "../components/Navbars/AdminNavbar.svelte";
  import Sidebar from "../components/Sidebar/Sidebar.svelte";
  import FooterAdmin from "../components/Footers/FooterAdmin.svelte";

  // pages for this layout
  import Dashboard from "../views/admin/Dashboard.svelte";
  import Settings from "../views/admin/Settings.svelte";
  import AdminUsersTable from "../components/Things/Tables/AdminUsersTable.svelte";
  import ReadUpdateDeleteUser from "../components/Things/ReadUpdateDelete/User.svelte";

  export let location: Location;

  let currentAuthStatus = {};
  const unsubscribeFromAuthStatusUpdates = authStatusStore.subscribe((value: UserStatus) => {
    currentAuthStatus = value;
    if (!currentAuthStatus) {
      navigate("/auth/login", { state: {}, replace: true });
    }
  });
  // onDestroy(unsubscribeFromAuthStatusUpdates);

  import { Logger } from "../logger"
  let logger = new Logger();
</script>

<div>
  <Sidebar location={location}/>
  <div class="relative md:ml-64 bg-gray-200">
    <AdminNavbar />
    <div class="px-4 md:px-10 mx-auto w-full -m-24">
      <Router url="admin">
        <Route path="dashboard" component="{Dashboard}" />
        <Route path="settings" component="{Settings}" />
        <Route path="users" component="{AdminUsersTable}" />
        <Route path="users/:id" let:params>
          <ReadUpdateDeleteUser id="{params.id}" />
        </Route>
      </Router>
      <FooterAdmin />
    </div>
  </div>
</div>
