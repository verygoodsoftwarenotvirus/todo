<script lang="typescript">
import { Router, Route } from 'svelte-routing';

import { UserStatus } from '../types';
import { Logger } from '../logger';


// components for this layout
import AdminNavbar from '../components/navbars/adminNavbar.svelte';
import Sidebar from '../components/sidebar/dashboardSidebar.svelte';
import UserEditor from '../components/editors/user.svelte';
import WebhookEditor from '../components/editors/apiClient.svelte';
import AccountEditor from '../components/editors/account.svelte';
import AdminFooter from '../components/footers/adminFooter.svelte';

// pages for this layout
import ServerSettings from '../views/settings/adminSettings.svelte';
import UsersAdmin from '../components/tables/users.svelte';
import AccountsAdmin from '../components/tables/accounts.svelte';
import Webhooks from '../components/tables/webhooks.svelte';
import AuditLogEntries from '../components/tables/auditLogEntries.svelte';

import { Superstore } from '../stores';

export let location: Location;

let logger = new Logger().withDebugValue('source', 'src/layouts/dashboard.svelte');
let currentAuthStatus: UserStatus = new UserStatus();

let superstore = new Superstore({
  userStatusStoreUpdateFunc: (value: UserStatus) => {
    currentAuthStatus = value;
  },
});
</script>

<div>
  <Sidebar location="{location}" />
  <div class="relative md:ml-64 bg-gray-200">
    <AdminNavbar />
    <div class="px-4 md:px-10 mx-auto w-full -m-24">
      <!-- no layout pages -->
      <Router url="admin">
        <Route path="settings" component="{ServerSettings}" />
        <Route path="audit_log" component="{AuditLogEntries}" />
        <Route path="accounts" component="{AccountsAdmin}" />
        <Route path="accounts/:id" let:params>
          <AccountEditor accountID="{params.id}" />
        </Route>
        <Route path="users" component="{UsersAdmin}" />
        <Route path="users/:id" let:params>
          <UserEditor userID="{params.id}" />
        </Route>
        <Route path="webhooks" component="{Webhooks}" />
        <Route path="webhooks/:id" let:params>
          <WebhookEditor webhookID="{params.id}" />
        </Route>
      </Router>
      <AdminFooter />
    </div>
  </div>
</div>
