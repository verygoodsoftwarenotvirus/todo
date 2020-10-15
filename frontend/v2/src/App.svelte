<!-- App.svelte -->
<script lang="typescript">
  import axios, { AxiosResponse, AxiosError } from "axios";
  import {Router, Route, navigate} from "svelte-routing";
  import { onMount } from 'svelte';

  import { AuthStatus } from "./models";
  import { authStatusStore } from "./stores/auth_store";

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

  onMount(() => {
    console.debug("fetching user status from App.svelte")
    axios.get("/users/status", { withCredentials: true })
          .then((response: AxiosResponse<AuthStatus>) => {
            console.debug("setting auth status from App.svelte");
            authStatusStore.setAuthStatus(response.data);
          })
          .catch((error: AxiosError) => {
            console.error(error);
          });
  })
</script>

<Router url="{url}">
  <!-- admin layout -->
  <Route path="admin/*admin" component="{Admin}" />
  <!-- auth layout -->
  <Route path="auth/*auth" component="{Auth}" />
  <!-- users layout -->
  <Route path="user/*user" component="{User}" />
  <!-- things layout -->
  <Route path="things/*things" component="{Things}" />
  <!-- no layout pages -->
  <Route path="/" component="{Index}" />
</Router>
