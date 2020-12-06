<script lang="typescript">
import { Router, Route } from 'svelte-routing';

// components for this layout
import AuthNavbar from '../components/Navbars/AuthNavbar.svelte';
import SmallFooter from '../components/Footers/SmallFooter.svelte';

// pages for this layout
import Login from '../views/auth/Login.svelte';
import Register from '../views/auth/Register.svelte';

const registerBg2: string = '../assets/img/register_bg_2.png';

export let location: Location;

import { Logger } from '../logger';

let logger = new Logger().withDebugValue('source', 'src/layouts/Auth.svelte');
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
      <SmallFooter absolute="true" />
    </section>
  </main>
</div>
