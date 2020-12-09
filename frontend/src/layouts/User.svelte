<script lang="typescript">
import { Route, Router } from 'svelte-routing';

// components for this layout
import AdminNavbar from '../components/Navbars/AdminNavbar.svelte';
import Sidebar from '../components/Sidebar/Sidebar.svelte';
import AdminFooter from '../components/Footers/AdminFooter.svelte';
import WebhookEditor from '../components/Webhooks/Editor.svelte';
import WebhookCreator from '../components/Webhooks/Creator.svelte';
import OAuth2ClientEditor from '../components/OAuth2Clients/Editor.svelte';

// pages for this layout
import Webhooks from '../views/user/Webhooks.svelte';
import UserSettings from '../views/user/Settings.svelte';
import OAuth2Clients from '../views/user/OAuth2Clients.svelte';

import { Logger } from '../logger';

let logger = new Logger().withDebugValue('source', 'src/layouts/User.svelte');

export let location: Location;
</script>

<div>
  <Sidebar location="{location}" />
  <div class="relative md:ml-64 bg-gray-200">
    <AdminNavbar />
    <div class="px-4 md:px-10 mx-auto w-full -m-24">
      <Router url="user">
        <Route path="oauth2_clients" component="{OAuth2Clients}" />
        <Route path="oauth2_clients/:id" let:params>
          <OAuth2ClientEditor oauth2ClientID="{params.id}" />
        </Route>
        <Route path="webhooks" component="{Webhooks}" />
        <Route path="webhooks/new" component="{WebhookCreator}" />
        <Route path="webhooks/nu" component="{WebhookEditor}" />
        <Route path="webhooks/:id" let:params>
          <WebhookEditor webhookID="{params.id}" />
        </Route>
        <Route path="settings" component="{UserSettings}" />
      </Router>
      <AdminFooter />
    </div>
  </div>
</div>
