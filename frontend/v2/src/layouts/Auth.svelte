<script lang="typescript">
  import { onDestroy, onMount} from "svelte";
  import axios from "axios";
  import {Router, Route, navigate} from "svelte-routing";

  // components for this layout
  import AuthNavbar from "../components/Navbars/AuthNavbar.svelte";
  import FooterSmall from "../components/Footers/FooterSmall.svelte";

  // pages for this layout
  import Login from "../views/auth/Login.svelte";
  import Register from "../views/auth/Register.svelte";

  const registerBg2: string = "../assets/img/register_bg_2.png";

  export let location: Location;

  import {Logger} from "../logger";
  let logger = new Logger().withDebugValue("source", "src/layouts/Auth.svelte");

  import { userStatusStore } from "../stores";
  import {UserStatus} from "../models";
  let currentAuthStatus = {};
  const unsubscribeFromUserStatusUpdates = userStatusStore.subscribe((value: UserStatus) => {
    currentAuthStatus = value;
  });
  // onDestroy(unsubscribeFromUserStatusUpdates);
</script>

<div>
  <AuthNavbar />
  <main>
    <section class="relative w-full h-full py-40 min-h-screen">
      <div
        class="absolute top-0 w-full h-full bg-gray-900 bg-no-repeat bg-full"
        style="background-image: url({registerBg2});"
      ></div>
      <Router url="auth">
        <Route path="login" component="{Login}" />
        <Route path="register" component="{Register}" />
      </Router>
      <FooterSmall absolute="true" />
    </section>
  </main>
</div>
