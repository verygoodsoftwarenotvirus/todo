<script lang="typescript">
  import { navigate } from 'svelte-routing';
  import { onDestroy, onMount } from 'svelte';
  import { AxiosError, AxiosResponse } from 'axios';

  import { UserSiteSettings, User } from '../../../types';
  import { Logger } from '../../../logger';
  import { V1APIClient } from '../../../requests';
  import { translations } from '../../../i18n';
  import { sessionSettingsStore } from '../../../stores';
  import AuditLogTable from '../../AuditLogTable/AuditLogTable.svelte';

  export let id: number = 0;

  // local state
  let userRetrievalError: string = '';
  let originalUser: User = new User();
  let user: User = new User();
  let needsToBeSaved: boolean = false;

  function evaluateChanges() {
    needsToBeSaved = !User.areEqual(user, originalUser);
  }

  let logger = new Logger().withDebugValue(
    'source',
    'src/components/Types/User/Editor.svelte',
  );

  // set up translations
  let currentSessionSettings = new UserSiteSettings();
  let translationsToUse = translations.messagesFor(
    currentSessionSettings.language,
  ).pages.registration;
  const unsubscribeFromSettingsUpdates = sessionSettingsStore.subscribe(
    (value: UserSiteSettings) => {
      currentSessionSettings = value;
      translationsToUse = translations.messagesFor(
        currentSessionSettings.language,
      ).pages.registration;
    },
  );
  onDestroy(unsubscribeFromSettingsUpdates);

  function saveUser(): void {
    logger.debug(`saveUser called`);

    if (id === 0) {
      throw new Error('id cannot be zero!');
    } else if (!needsToBeSaved) {
      throw new Error('no changes to save!');
    }

    V1APIClient.saveUser(user)
      .then((response: AxiosResponse<User>) => {
        user = { ...response.data };
        originalUser = { ...response.data };
        needsToBeSaved = false;
      })
      .catch((error: AxiosError) => {
        if (error.response) {
          if (error.response.data) {
            userRetrievalError = error.response.data;
          }
        }
      });
  }

  function fetchUser(): void {
    logger.debug(`fetchUser called`);

    if (id === 0) {
      throw new Error('id cannot be zero!');
    }

    V1APIClient.fetchUser(id)
      .then((response: AxiosResponse<User>) => {
        user = { ...response.data };
        originalUser = { ...response.data };
      })
      .catch((error: AxiosError) => {
        if (error.response) {
          if (error.response.data) {
            userRetrievalError = error.response.data;
          }
        }
      });
  }

  function deleteUser(): void {
    logger.debug(`fetchUser called`);

    if (id === 0) {
      throw new Error('id cannot be zero!');
    }

    V1APIClient.deleteUser(id)
      .then((response: AxiosResponse<User>) => {
        user = { ...response.data };
        originalUser = { ...response.data };
      })
      .catch((error: AxiosError) => {
        if (error.response) {
          if (error.response.data) {
            userRetrievalError = error.response.data;
          }
        }
      });
  }

  onMount(fetchUser);
</script>

<div
  class="relative flex flex-col min-w-0 break-words bg-white w-full mb-6 shadow-lg rounded">
  <div class="rounded-t mb-0 px-4 py-3 bg-transparent justify-between ">
    <div class="flex flex-wrap items-center">
      <div class="relative w-full max-w-full flex-grow flex-1">
        {#if originalUser.id !== 0}
          <h2 class="text-gray-800 text-xl font-semibold">
            #{originalUser.id}:
            {originalUser.username}
          </h2>
        {/if}
      </div>
      <div class="flex w-full max-w-full flex-grow justify-end flex-1">
        <button
          class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded {needsToBeSaved ? '' : 'opacity-50 cursor-not-allowed'}"
          on:click={saveUser}><i class="fa fa-save" />
          {translationsToUse.actions.save}</button>
      </div>
    </div>
  </div>
  <div>
    <div class="flex flex-wrap -mx-3 mb-6">
      <div class="w-full md:w-1/2 px-3 mb-6 md:mb-0">
        <label
          class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
          for="grid-first-name">
          {translationsToUse.labels.name}
        </label>
        <input
          class="appearance-none block w-full text-gray-700 border border-red-500 rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white"
          id="grid-first-name"
          type="text"
          placeholder={translationsToUse.inputPlaceholders.name}
          on:keyup={evaluateChanges}
          bind:value={user.username} />
        <!--  <p class="text-red-500 text-xs italic">Please fill out this field.</p>-->
      </div>
      <div class="flex w-full mr-3 mt-4 max-w-full flex-grow justify-end flex-1">
        <button class="bg-red-500 hover:bg-red-700 text-white font-bold py-2 px-4 rounded" on:click={deleteUser}><i class="fa fa-trash-alt"></i> {translationsToUse.actions.delete}</button>
      </div>
    </div>
  </div>

  {#if currentUserStatus.isAdmin}
    <AuditLogTable entries={auditLogEntries} />
  {/if}
</div>
