<!-- App.svelte -->
<script lang="typescript">
  import axios, { AxiosResponse, AxiosError } from "axios";
  import {Router, Route, navigate} from "svelte-routing";
  import { onMount } from 'svelte';

  import { UserStatus } from "./models";
  import { authStatusStore } from "./stores/auth_store";

  import { Logger, LogLevel } from "./logger";

  // Admin Layout
  import Admin from "./layouts/Admin.svelte";
  // Auth Layout
  import Auth from "./layouts/Auth.svelte";
  // User Layout
  import User from "./layouts/User.svelte";
  // Things Layout
  import Things from "./layouts/Things.svelte";

  // No Layout Pages
  import Index from "./views/Index.svelte";

  export let url: string = "";

  let logger = new Logger().withDebugValue("source", "src/App.svelte");

  // onMount(() => {
  //   axios.get("/auth/status", { withCredentials: true })
  //         .then((response: AxiosResponse<AuthStatus>) => {
  //           logger.debug("setting auth status from App.svelte");
  //           authStatusStore.setAuthStatus(response.data);
  //         })
  //         .catch((error: AxiosError) => {
  //           logger.error(error.toString());
  //         });
  // })
</script>

<Router url="{url}">
  <!-- admin layout -->
  <Route path="admin/*admin" component="{Admin}" />
  <!-- auth layout -->
  <Route path="auth/*auth" component="{Auth}" />
  <!-- user layout -->
  <Route path="user/*user" component="{User}" />
  <!-- things layout -->
  <Route path="things/*things" component="{Things}" />
  <!-- no layout pages -->
  <Route path="/" component="{Index}" />
</Router>
