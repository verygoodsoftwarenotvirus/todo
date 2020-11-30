<script lang="typescript">
import { Route } from 'svelte-routing';

// components for this layout
import AdminNavbar from '../components/Navbars/AdminNavbar.svelte';
import Sidebar from '../components/Sidebar/Sidebar.svelte';
import AdminFooter from '../components/Footers/AdminFooter.svelte';

// pages for this layout
import WebhookEditor from '../components/Editors/Webhook.svelte';
import UserSettings from '../views/user/Settings.svelte';

import { Logger } from '../logger';

let logger = new Logger().withDebugValue('source', 'src/layouts/User.svelte');

export let location: Location;
</script>

<div>
  <Sidebar location="{location}" />
  <div class="relative md:ml-64 bg-gray-200">
    <AdminNavbar />
    <div class="px-4 md:px-10 mx-auto w-full -m-24">
      <Route path="settings" component="{UserSettings}" />
      <AdminFooter />
    </div>
  </div>
</div>
