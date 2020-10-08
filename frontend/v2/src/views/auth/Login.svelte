<script lang="typescript">
  import axios, {AxiosResponse} from 'axios';
  import { link, navigate } from "svelte-routing";

  import type { AuthStatus, LoginRequest } from "../../models";

  export let location: Location;

    let usernameInput = '';
    let passwordInput = '';
    let totpTokenInput = '';

    function buildLoginRequest(): LoginRequest {
        return {
            username: this.usernameInput,
            password: this.passwordInput,
            totpToken: this.totpTokenInput,
        } as LoginRequest
    }

    async function login() {
        const path = "/users/login"

        console.debug("login called!");

        return axios.post(path, this.buildLoginRequest(), {withCredentials: true})
            .then(() => {
              axios.get("/users/status", {withCredentials: true}).then((statusResponse: AxiosResponse<AuthStatus>) => {
                navigate("/", { state: {}, replace: true });
              });
            })
            .catch((reason: object) => {
                console.error(`something went awry: ${reason}`);
            });
    }

</script>

<div class="container mx-auto px-4 h-full">
  <div class="flex content-center items-center justify-center h-full">
    <div class="w-full lg:w-4/12 px-4">
      <div
        class="relative flex flex-col min-w-0 break-words w-full mb-6 shadow-lg rounded-lg bg-gray-300 border-0"
      >
        <div class="rounded-t mb-0 px-6 py-6"></div>
        <div class="flex-auto px-4 lg:px-10 py-10 pt-0">
          <form>
            <div class="relative w-full mb-3">
              <label
                class="block uppercase text-gray-700 text-xs font-bold mb-2"
                for="grid-username"
              >
                Username
              </label>
              <input
                id="grid-username"
                type="text"
                class="px-3 py-3 placeholder-gray-400 text-gray-700 bg-white rounded text-sm shadow focus:outline-none focus:shadow-outline w-full ease-linear transition-all duration-150"
                placeholder="username"
                bind:value={usernameInput}
              />
            </div>

            <div class="relative w-full mb-3">
              <label
                class="block uppercase text-gray-700 text-xs font-bold mb-2"
                for="grid-password"
              >
                Password
              </label>
              <input
                id="grid-password"
                type="password"
                class="px-3 py-3 placeholder-gray-400 text-gray-700 bg-white rounded text-sm shadow focus:outline-none focus:shadow-outline w-full ease-linear transition-all duration-150"
                placeholder="Password"
                bind:value={passwordInput}
              />
            </div>

            <div class="relative w-full mb-3">
              <label
                class="block uppercase text-gray-700 text-xs font-bold mb-2"
                for="grid-two-factor"
              >
                2FA Token
              </label>
              <input
                id="grid-two-factor"
                type="text"
                class="px-3 py-3 placeholder-gray-400 text-gray-700 bg-white rounded text-sm shadow focus:outline-none focus:shadow-outline w-full ease-linear transition-all duration-150"
                placeholder="123456"
                bind:value={totpTokenInput}
              />
            </div>

            <div class="text-center mt-6">
              <button
                class="bg-gray-900 text-white active:bg-gray-700 text-sm font-bold uppercase px-6 py-3 rounded shadow hover:shadow-lg outline-none focus:outline-none mr-1 mb-1 w-full ease-linear transition-all duration-150"
                type="button"
                on:click={login}
              >
                Sign In
              </button>
            </div>
          </form>
        </div>
      </div>
      <div class="flex flex-wrap mt-6 relative">
        <div class="w-1/2">
          <a href="#pablo" on:click={(e) => e.preventDefault()} class="text-gray-300">
            <small>Forgot password?</small>
          </a>
        </div>
        <div class="w-1/2 text-right">
          <a use:link href="/auth/register" class="text-gray-300">
            <small>Create new account</small>
          </a>
        </div>
      </div>
    </div>
  </div>
</div>
