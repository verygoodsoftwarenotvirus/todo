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
import AdminFooter from '../components/Footers/AdminFooter.svelte';

// pages for this layout
import Dashboard from '../views/admin/Dashboard.svelte';
import Settings from '../views/admin/Settings.svelte';

import UsersAdmin from '../views/admin/Users.svelte';
import AuditLogEntries from '../views/admin/AuditLogEntries.svelte';

import UserEditor from '../components/Editors/User.svelte';
import { Superstore } from '../stores/superstore';

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
      </Router>
      <AdminFooter />
    </div>
  </div>
</div>
