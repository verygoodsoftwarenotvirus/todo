<script lang="typescript">
import { Route, Router } from 'svelte-routing';

// components for this layout
import AdminNavbar from '../components/navbars/adminNavbar.svelte';
import Sidebar from '../components/sidebar/dashboardSidebar.svelte';
import AdminFooter from '../components/footers/adminFooter.svelte';
import WebhookEditor from '../components/editors/webhook.svelte';
import WebhookCreator from '../components/creators/webhook.svelte';

// pages for this layout
import Webhooks from '../components/tables/webhooks.svelte';
import AccountSettings from '../views/settings/accountSettings.svelte';

import { Logger } from '../logger';

let logger = new Logger().withDebugValue('source', 'src/layouts/account.svelte');

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
