<!-- App.svelte -->
<script lang="typescript">
import { Router, Route, navigate } from 'svelte-routing';

import { Logger } from '@/logger';

// Admin Layout
import Admin from './layouts/Root.svelte';
// Auth Layout
import Auth from './layouts/Auth.svelte';
// User Layout
import User from './layouts/User.svelte';
// Account Layout
import Account from './layouts/Account.svelte';
// Things Layout
import Things from './layouts/Things.svelte';

// No Layout Pages
import Dashboard from './views/Dashboard.svelte';
import { UserSiteSettings, UserStatus } from '@/types';
import { Superstore } from '@/stores';
import { frontendRoutes } from '@/constants';

export let url: string = '';

let logger = new Logger().withDebugValue('source', 'src/App.svelte');

let currentAuthStatus = new UserStatus();
let currentSessionSettings = new UserSiteSettings();
let translationsToUse = currentSessionSettings.getTranslations().components
  .sidebars.primary;

let superstore = new Superstore({
  userStatusStoreUpdateFunc: (value: UserStatus) => {
    currentAuthStatus = value;
    if (!currentAuthStatus.isAuthenticated && !Superstore.frontendOnlyMode) {
      if (window.location.pathname !== '/') {
        logger.debug('sending unauthenticated user back to login page')
        navigate(frontendRoutes.LOGIN)
      }
    }
  },
  sessionSettingsStoreUpdateFunc: (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = currentSessionSettings.getTranslations().components
      .sidebars.primary;
  },
});
</script>

<Router url="{url}">
  <!-- admin layout -->
  <Route path="admin/*admin" component="{Admin}" />
  <!-- auth layout -->
  <Route path="auth/*auth" component="{Auth}" />
  <!-- account layout -->
  <Route path="account/*" component="{Account}" />
  <!-- user layout -->
  <Route path="user/*" component="{User}" />
  <!-- things layout -->
  <Route path="things/*things" component="{Things}" />
  <!-- no layout pages -->
  <Route path="/" component="{Dashboard}" />
</Router>
