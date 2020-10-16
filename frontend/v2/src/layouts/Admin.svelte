<script lang="typescript">
  import axios, {AxiosError, AxiosResponse} from "axios";
  import { onDestroy, onMount } from "svelte";
  import { navigate, Router, Route } from "svelte-routing";

  // components for this layout
  import AdminNavbar from "../components/Navbars/AdminNavbar.svelte";
  import Sidebar from "../components/Sidebar/Sidebar.svelte";
  import HeaderStats from "../components/Headers/HeaderStats.svelte";
  import FooterAdmin from "../components/Footers/FooterAdmin.svelte";

  // pages for this layout
  import Dashboard from "../views/admin/Dashboard.svelte";
  import Settings from "../views/admin/Settings.svelte";
  import Tables from "../views/admin/Tables.svelte";

  export let location: Location;

  import { authStatusStore } from "../stores";
  import { AuthStatus } from "../models";

  let currentAuthStatus = {};
  const unsubscribeFromAuthStatusUpdates = authStatusStore.subscribe((value: AuthStatus) => {
    currentAuthStatus = value;
  });
  // onDestroy(unsubscribeFromAuthStatusUpdates);

  import { Logger } from "../logger"
  let logger = new Logger();

  onMount(() => {
    if (!currentAuthStatus.isAuthenticated) {
      logger.debug("I would fuck you off back to the login page");
    } else {
      logger.debug("Admin layout onMount called");
    }
  })

  // onMount(async () => {
  //   logger.debug("checking status from Admin layout");
  //   const res = await axios.get("/users/status", { withCredentials: true });
  //   const as: AuthStatus = res.data;
  //   authStatusStore.setAuthStatus(as);
  //
  //   if (!as.isAdmin) {
  //     navigate("/", { state: {}, replace: true });
  //   }
  // })
</script>

<div>
  <Sidebar location={location}/>
  <div class="relative md:ml-64 bg-gray-200">
    <AdminNavbar />
    <HeaderStats />
    <div class="px-4 md:px-10 mx-auto w-full -m-24">
      <Router url="admin">
        <Route path="dashboard" component="{Dashboard}" />
        <Route path="settings" component="{Settings}" />
        <Route path="tables" component="{Tables}" />
      </Router>
      <FooterAdmin />
    </div>
  </div>
</div>
