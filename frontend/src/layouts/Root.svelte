<script lang="typescript">
import { Router, Route } from 'svelte-routing';

import { UserStatus } from '../types';
import { Logger } from '../logger';

let logger = new Logger().withDebugValue('source', 'src/layouts/Root.svelte');

export let location: Location;

// components for this layout
import AdminNavbar from '../components/Navbars/AdminNavbar.svelte';
import Sidebar from '../components/Sidebar/Sidebar.svelte';
import UserEditor from '../components/Users/Editor.svelte';
import WebhookEditor from '../components/Webhooks/Editor.svelte';
import AccountEditor from '../components/Accounts/Editor.svelte';
import AdminFooter from '../components/Footers/AdminFooter.svelte';

// pages for this layout
import ServerSettings from '../views/admin/Settings.svelte';
import UsersAdmin from '../views/admin/Users.svelte';
import AccountsAdmin from '../views/admin/Accounts.svelte';
import Webhooks from '../views/accounts/Webhooks.svelte';
import AuditLogEntries from '../views/admin/AuditLogEntries.svelte';

import { Superstore } from '../stores';

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
