<script lang="typescript">
  import { onDestroy } from "svelte";
  import { navigate, Router, Route } from "svelte-routing";

  // components for this layout
  import AdminNavbar from "../components/Navbars/AdminNavbar.svelte";
  import Sidebar from "../components/Sidebar/Sidebar.svelte";
  import AdminFooter from "../components/Footers/AdminFooter.svelte";

  // pages for this layout
  import Settings from "../views/user/Settings.svelte";

  import { userStatusStore } from "../stores";
  import {User, UserStatus} from "../types";
  import { Logger } from "../logger"

  let logger = new Logger().withDebugValue("source", "src/layouts/User.svelte");

  // let currentUserStatus: UserStatus = new UserStatus();
  // const unsubscribeFromUserStatusUpdates = userStatusStore.subscribe((value: UserStatus) => {
  //   currentUserStatus = value;
  //   if (!currentUserStatus || !currentUserStatus.isAuthenticated) {
  //     logger.debug(`navigating to /auth/login because user is unauthenticated`);
  //     navigate("/auth/login", { state: {}, replace: true });
  //   }
  // });
  // onDestroy(unsubscribeFromUserStatusUpdates);

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
      <AdminFooter />
    </div>
  </div>
</div>
