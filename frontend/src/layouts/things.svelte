<script lang="typescript">
import { Router, Route } from 'svelte-routing';

// components for this layout
import Sidebar from '../components/sidebar/dashboardSidebar.svelte';
import AdminFooter from '../components/footers/adminFooter.svelte';
import AdminNavbar from '../components/navbars/adminNavbar.svelte';

// custom components for this layout
import ItemsList from '../components/tables/items.svelte';
import ItemEditor from '../components/editors/item.svelte';
import ItemCreator from '../components/creators/item.svelte';

// pages for this layout

import { UserStatus } from '../types';
import { Logger } from '../logger';
import { Superstore } from '../stores';

export let location: Location;

let logger = new Logger().withDebugValue('source', 'src/layouts/things.svelte');

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
      <Router url="things">
        <!--     ITEMS     -->
        <Route path="items" component="{ItemsList}" />
        <Route path="items/:id" let:params>
          <ItemEditor itemID="{params.id}" />
        </Route>
        <Route path="items/new" component="{ItemCreator}" />
      </Router>
      <AdminFooter />
    </div>
  </div>
</div>
