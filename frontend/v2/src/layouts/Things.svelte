<script lang="typescript">
import { Router, Route } from 'svelte-routing';

// components for this layout
import Sidebar from '../components/Sidebar/Sidebar.svelte';
import AdminFooter from '../components/Footers/AdminFooter.svelte';
import AdminNavbar from '../components/Navbars/AdminNavbar.svelte';

// custom components for this layout
import Items from '../views/things/Items.svelte';
import ItemEditorComponent from '../components/Types/Items/Editor.svelte';
import ItemCreatorComponent from '../components/Types/Items/Creator.svelte';

// pages for this layout

import { userStatusStore } from '../stores';
import { UserStatus } from '../types';
import { Logger } from '../logger';

export let location: Location;

let logger = new Logger().withDebugValue('source', 'src/layouts/Things.svelte');

let currentAuthStatus = {};
const unsubscribeFromUserStatusUpdates = userStatusStore.subscribe(
  (value: UserStatus) => {
    currentAuthStatus = value;
  },
);
</script>

<div>
  <Sidebar location="{location}" />
  <div class="relative md:ml-64 bg-gray-200">
    <AdminNavbar />
    <div class="px-4 md:px-10 mx-auto w-full -m-24">
      <Router url="things">
        <!--     ITEMS     -->
        <Route path="items" component="{Items}" />
        <Route path="items/:id" let:params>
          <ItemEditorComponent id="{params.id}" />
        </Route>
        <Route path="items/new" component="{ItemCreatorComponent}" />
      </Router>
      <AdminFooter />
    </div>
  </div>
</div>
