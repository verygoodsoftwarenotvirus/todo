<script lang="typescript">
import { onDestroy } from 'svelte';
import { Router, Route } from 'svelte-routing';

import { UserStatus } from '../types';
import { Logger } from '../logger';

let logger = new Logger().withDebugValue('source', 'src/layouts/Admin.svelte');

export let location: Location;

// components for this layout
import AdminNavbar from '../components/Navbars/AdminNavbar.svelte';
import Sidebar from '../components/Sidebar/Sidebar.svelte';
import UserEditor from '../components/Users/Editor.svelte';
import AdminFooter from '../components/Footers/AdminFooter.svelte';


// pages for this layout
import Dashboard from '../views/admin/Dashboard.svelte';
import Settings from '../views/admin/Settings.svelte';
import UsersAdmin from '../views/admin/Users.svelte';
import Webhooks from '../views/admin/Webhooks.svelte';
import OAuth2Clients from '../views/admin/OAuth2Clients.svelte';
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
      <Router url="admin">
        <Route path="dashboard" component="{Dashboard}" />
        <Route path="settings" component="{Settings}" />
        <Route path="audit_log" component="{AuditLogEntries}" />
        <Route path="users" component="{UsersAdmin}" />
        <Route path="users/:id" let:params>
          <UserEditor userID="{params.id}" />
        </Route>
        <Route path="oauth2_clients" component="{OAuth2Clients}" />
        <Route path="oauth2_clients/:id" let:params>
          <OAuth2ClientEditor oauth2ClientID="{params.id}" />
        </Route>
        <Route path="webhooks" component="{Webhooks}" />
        <Route path="webhooks/new" component="{WebhookCreator}" />
        <Route path="webhooks/:id" let:params>
          <WebhookEditor webhookID="{params.id}" />
        </Route>
        <Route path="settings" component="{UserSettings}" />
      </Router>
      <AdminFooter />
    </div>
  </div>
</div>
