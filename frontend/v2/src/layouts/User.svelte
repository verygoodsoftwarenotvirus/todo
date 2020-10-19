<script lang="typescript">
  import { onDestroy } from "svelte";
  import { navigate, Router, Route } from "svelte-routing";

  // components for this layout
  import AdminNavbar from "../components/Navbars/AdminNavbar.svelte";
  import Sidebar from "../components/Sidebar/Sidebar.svelte";
  import FooterAdmin from "../components/Footers/FooterAdmin.svelte";

  // pages for this layout
  import Settings from "../views/user/Settings.svelte";

  import { authStatusStore } from "../stores";
  import { UserStatus } from "../models";
  import { Logger } from "../logger"

  let logger = new Logger().withDebugValue("source", "src/layouts/User.svelte");

  let currentAuthStatus = {};
  const unsubscribeFromAuthStatusUpdates = authStatusStore.subscribe((value: UserStatus) => {
    currentAuthStatus = value;
    if (!currentAuthStatus || !currentAuthStatus.isAuthenticated) {
      logger.debug(`navigating to /auth/login because user is unauthenticated`);
      navigate("/auth/login", { state: {}, replace: true });
    }
  });
  // onDestroy(unsubscribeFromAuthStatusUpdates);

  export let location: Location;
</script>

<div>
  <Sidebar location={location}/>
  <div class="relative md:ml-64 bg-gray-200">
    <AdminNavbar />
    <div class="px-4 md:px-10 mx-auto w-full -m-24">
      <Router url="admin">
        <Route path="settings" component="{Settings}" />
      </Router>
      <FooterAdmin />
    </div>
  </div>
</div>
