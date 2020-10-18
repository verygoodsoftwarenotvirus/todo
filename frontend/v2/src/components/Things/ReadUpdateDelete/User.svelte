<script lang="typescript">
  import { navigate } from "svelte-routing";
  import { onMount } from "svelte";
  import axios, {AxiosError, AxiosResponse} from "axios";

  import { User } from "../../../models";
  import { Logger } from "../../../logger";

  export let id: number = 0;

  // local state
  let userRetrievalError: string = '';
  let originalUser: User = new User();
  let user: User = new User();
  let needsToBeSaved: boolean = false;

  function evaluateChanges() {
    needsToBeSaved = !User.areEqual(user, originalUser);
  }

  onMount(fetchUser);

  let logger = new Logger();

  function fetchUser(): void {
    logger.debug(`vies/things/User.fetchUser called`);

    if (id === 0) {
      throw new Error("id cannot be zero!");
    }

    const path: string = `/api/v1/users/${id}`;

    axios.get(path, { withCredentials: true })
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

  function saveUser(): void {
    logger.debug(`vies/things/User.saveUser called`);

    if (id === 0) {
      throw new Error("id cannot be zero!");
    } else if (!needsToBeSaved) {
      throw new Error("no changes to save!");
    }

    const path: string = `/api/v1/users/${id}`;

    axios.put(path, user, { withCredentials: true })
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

  function deleteUser(): void {
    logger.debug(`vies/things/User.deleteUser called`);

    if (id === 0) {
      throw new Error("id cannot be zero!");
    }

    const path: string = `/api/v1/users/${id}`;

    axios.delete(path, { withCredentials: true })
            .then((response: AxiosResponse<User>) => {
              if (response.status === 204) {
                navigate("/things/users", { state: {}, replace: true });
              }
            })
            .catch((error: AxiosError) => {
              if (error.response) {
                if (error.response.data) {
                  userRetrievalError = error.response.data;
                }
              }
            });
  }
</script>

<div class="relative flex flex-col min-w-0 break-words bg-white w-full mb-6 shadow-lg rounded">
  <div class="rounded-t mb-0 px-4 py-3 bg-transparent justify-between ">
    <div class="flex flex-wrap items-center">
      <div class="relative w-full max-w-full flex-grow flex-1">
        {#if originalUser.id !== 0}
          <h2 class="text-gray-800 text-xl font-semibold">
            #{originalUser.id}: { originalUser.username }
          </h2>
        {/if}
      </div>
      <div class="flex w-full max-w-full flex-grow justify-end flex-1">
        <button class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded {needsToBeSaved ? '' : 'opacity-50 cursor-not-allowed'}" on:click={saveUser}><i class="fa fa-save"></i> Save</button>
      </div>
    </div>
  </div>
  <div>
    <div class="flex flex-wrap -mx-3 mb-6">
      <div class="w-full md:w-1/2 px-3 mb-6 md:mb-0">
        <label class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2" for="grid-first-name">
          Name
        </label>
        <input class="appearance-none block w-full text-gray-700 border border-red-500 rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white" id="grid-first-name" type="text" on:keyup={evaluateChanges} bind:value={user.username}>
        <!--  <p class="text-red-500 text-xs italic">Please fill out this field.</p>-->
      </div>
<!--      <div class="w-full md:w-1/2 px-3">-->
<!--        <label class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2" for="grid-last-name">-->
<!--          Details-->
<!--        </label>-->
<!--        <input class="appearance-none block w-full text-gray-700 border border-gray-200 rounded py-3 px-4 leading-tight focus:outline-none focus:bg-white focus:border-gray-500" id="grid-last-name" type="text" on:keyup={evaluateChanges} bind:value={user.details}>-->
<!--      </div>-->
      <div class="flex w-full mr-3 mt-4 max-w-full flex-grow justify-end flex-1">
        <button class="bg-red-500 hover:bg-red-700 text-white font-bold py-2 px-4 rounded" on:click={deleteUser}><i class="fa fa-trash-alt"></i> Delete</button>
      </div>
    </div>
  </div>
</div>
