<script lang="typescript">
import { Route, Router } from 'svelte-routing';

// components for this layout
import AdminNavbar from '../components/navbars/adminNavbar.svelte';
import Sidebar from '../components/sidebar/dashboardSidebar.svelte';
import AdminFooter from '../components/footers/adminFooter.svelte';

// pages for this layout
import APIClients from '../components/tables/apiClients.svelte';
import APIClientCreator from '../components/creators/apiClient.svelte';
import APIClientEditor from '../components/editors/apiClient.svelte';
import UserSettings from '../views/settings/userSettings.svelte';

import { Logger } from '../logger';

let logger = new Logger().withDebugValue('source', 'src/layouts/account.svelte');

export let location: Location;
</script>

<div>
  <Sidebar location="{location}" />
  <div class="relative md:ml-64 bg-gray-200">
    <AdminNavbar />
    <div class="px-4 md:px-10 mx-auto w-full -m-24">
      <Router url="user">
        <Route path="api_clients" component="{APIClients}" />
        <Route path="api_clients/new" component="{APIClientCreator}" />
        <Route path="api_clients/:id" let:params>
          <APIClientEditor apiClientID="{params.id}" />
        </Route>
        <Route path="settings" component="{UserSettings}" />
      </Router>
      <AdminFooter />
    </div>
  </div>
</div>
