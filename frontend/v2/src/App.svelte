<!-- App.svelte -->
<script lang="typescript">
  import axios, { AxiosError } from "axios";
  import { Router, Route } from "svelte-routing";
  import { onMount } from 'svelte';

  import { AuthStatus } from "./models";
  import { authStatus } from "./stores/auth_store";

  // Admin Layout
  import Admin from "./layouts/Admin.svelte";
  // Auth Layout
  import Auth from "./layouts/Auth.svelte";

  // No Layout Pages
  import Index from "./views/Index.svelte";

  export let url: string = "";

  onMount(async () => {
    try {
      const res = await axios.get("/users/status")
      const data = res.data as AuthStatus;
      authStatus.setAuthStatus(data);
    } catch (e: any) {
      const as = new AuthStatus();
      authStatus.setAuthStatus(as);
    }
  })
</script>

<Router url="{url}">
  <!-- admin layout -->
  <Route path="admin/*admin" component="{Admin}" />
  <!-- auth layout -->
  <Route path="auth/*auth" component="{Auth}" />
  <!-- no layout pages -->
  <Route path="/" component="{Index}" />
</Router>
