<script lang="typescript">
  import axios, { AxiosError, AxiosResponse } from 'axios';
  import { onDestroy } from 'svelte';
  import { navigate, Router, Route } from 'svelte-routing';

  import { userStatusStore } from '../stores';
  import { UserStatus } from '../types';

  // components for this layout
  import AdminNavbar from '../components/Navbars/AdminNavbar.svelte';
  import Sidebar from '../components/Sidebar/Sidebar.svelte';
  import AdminFooter from '../components/Footers/AdminFooter.svelte';

  // pages for this layout
  import UsersAdmin from '../views/admin/Users.svelte';
  import OAuth2ClientsAdmin from '../views/admin/OAuth2Clients.svelte';
  import WebhooksAdmin from '../views/admin/Webhooks.svelte';
  import Dashboard from '../views/admin/Dashboard.svelte';
  import Settings from '../views/admin/Settings.svelte';
  import UserEditor from '../components/Types/Users/Editor.svelte';

  import { Logger } from '../logger';
  let logger = new Logger().withDebugValue(
    'source',
    'src/layouts/Admin.svelte',
  );

  export let location: Location;

  let currentAuthStatus = {};
  const unsubscribeFromUserStatusUpdates = userStatusStore.subscribe(
    (value: UserStatus) => {
      currentAuthStatus = value;
      // if (!currentAuthStatus || !currentAuthStatus.isAuthenticated || !currentAuthStatus.isAdmin) {
      //   logger.debug(`navigating to /auth/login because the user is not authenticated`);
      //   navigate("/auth/login", { state: {}, replace: true });
      // }
    },
  );
  onDestroy(unsubscribeFromUserStatusUpdates);
</script>

<div>
  <Sidebar {location} />
  <div class="relative md:ml-64 bg-gray-200">
    <AdminNavbar />
    <div class="px-4 md:px-10 mx-auto w-full -m-24">
      <Router url="admin">
        <Route path="dashboard" component={Dashboard} />
        <Route path="settings" component={Settings} />
        <Route path="oauth2_clients" component={OAuth2ClientsAdmin} />
        <Route path="webhooks" component={WebhooksAdmin} />
        <Route path="users" component={UsersAdmin} />
        <Route path="users/:id" let:params>
          <UserEditor id={params.id} />
        </Route>
      </Router>
      <AdminFooter />
    </div>
  </div>
</div>
