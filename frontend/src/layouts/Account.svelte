<script lang="typescript">
import { Route, Router } from 'svelte-routing';

// components for this layout
import AdminNavbar from '../components/Navbars/AdminNavbar.svelte';
import Sidebar from '../components/Sidebar/Sidebar.svelte';
import AdminFooter from '../components/Footers/AdminFooter.svelte';
import WebhookEditor from '../components/Webhooks/Editor.svelte';
import WebhookCreator from '../components/Webhooks/Creator.svelte';

// pages for this layout
import Webhooks from '../views/accounts/Webhooks.svelte';
import AccountSettings from '../views/accounts/Settings.svelte';

import { Logger } from '../logger';

let logger = new Logger().withDebugValue('source', 'src/layouts/Account.svelte');

export let location: Location;
</script>

<div>
  <Sidebar location="{location}" />
  <div class="relative md:ml-64 bg-gray-200">
    <AdminNavbar />
    <div class="px-4 md:px-10 mx-auto w-full -m-24">
      <Router url="account">
        <Route path="webhooks" component="{Webhooks}" />
        <Route path="webhooks/new" component="{WebhookCreator}" />
        <Route path="webhooks/nu" component="{WebhookEditor}" />
        <Route path="webhooks/:id" let:params>
          <WebhookEditor webhookID="{params.id}" />
        </Route>
        <Route path="settings" component="{AccountSettings}" />
      </Router>
      <AdminFooter />
    </div>
  </div>
</div>
