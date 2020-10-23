<script lang="typescript">
  import axios, {AxiosError, AxiosResponse} from 'axios';
  import { link, navigate } from "svelte-routing";

  import type { UserStatus, LoginRequest } from "../../models";

  export let location: Location;

    let usernameInput: string = '';
    let passwordInput: string = '';
    let totpTokenInput: string = '';

    let canLogin: boolean = false;
    let loginError: string = '';

    import { Logger } from "../../logger";
    let logger = new Logger().withDebugValue("source", "src/views/auth/Login.svelte");

    function buildLoginRequest(): LoginRequest {
        return {
            username: usernameInput,
            password: passwordInput,
            totpToken: totpTokenInput,
        };
    }

    function evaluateInputs(): void {
      canLogin = usernameInput !== '' && passwordInput !== '' && totpTokenInput.length > 0 && totpTokenInput.length < 7;
    }

    let dumpingGround: string = '';

    import { userStatusStore } from "../../stores"
    import { V1APIClient } from "../../requests";

    async function login() {
        const path = "/users/login"

        logger.debug("login called!");
        dumpingGround = "login called!";

        evaluateInputs();
        if (!canLogin) {
          dumpingGround = "error thrown!";
          throw new Error("invalid input!");
        }

        dumpingGround = "login can proceed!";

        return V1APIClient.login(buildLoginRequest()).then(() => {
              dumpingGround = "login response received!";

              V1APIClient.checkAuthStatusRequest().then((statusResponse: AxiosResponse<UserStatus>) => {
                dumpingGround = "status response received!";

                userStatusStore.setAuthStatus(statusResponse.data);

                if (statusResponse.data.isAdmin) {
                  dumpingGround = "navigating to admin dashboard!";
                  logger.debug(`navigating to /admin/dashboard because user is an authenticated admin`);
                  navigate("/admin/dashboard", { state: {}, replace: true });
                  location.reload();
                } else {
                  dumpingGround = "navigating to homepage!";
                  logger.debug(`navigating to homepage because user is a plain user`);
                  navigate("/", { state: {}, replace: true });
                }
              });
            })
            .catch((reason: AxiosError) => {
              dumpingGround = "login promise catch called! "+ JSON.stringify(reason)
              if (reason.response) {
                if (reason.response.status === 401) {
                  loginError = 'invalid credentials: please try again'
                } else {
                  loginError = reason.response.toString();
                  logger.error(JSON.stringify(reason.response));
                }
              }
            });
    }

</script>

<div class="container mx-auto px-4 h-full">
  <div class="flex content-center items-center justify-center h-full">
    <div class="w-full lg:w-4/12 px-4">
      <div class="relative flex flex-col min-w-0 break-words w-full mb-6 shadow-lg rounded-lg bg-gray-300 border-0">
        <div class="rounded-t mb-0 px-6 py-6"></div>
        <div class="flex-auto px-4 lg:px-10 py-10 pt-0">
          <form on:submit|preventDefault="{login}">
            <div class="relative w-full mb-3">
              <label
                class="block uppercase text-gray-700 text-xs font-bold mb-2"
                for="usernameInput"
              >
                Username
              </label>
              <input
                id="usernameInput"
                type="text"
                class="px-3 py-3 placeholder-gray-400 text-gray-700 bg-white rounded text-sm shadow focus:outline-none focus:shadow-outline w-full ease-linear transition-all duration-150"
                placeholder="username"
                on:keyup={evaluateInputs}
                on:blur={evaluateInputs}
                bind:value={usernameInput}
              />
            </div>

            <div class="relative w-full mb-3">
              <label
                class="block uppercase text-gray-700 text-xs font-bold mb-2"
                for="passwordInput"
              >
                Password
              </label>
              <input
                id="passwordInput"
                type="password"
                class="px-3 py-3 placeholder-gray-400 text-gray-700 bg-white rounded text-sm shadow focus:outline-none focus:shadow-outline w-full ease-linear transition-all duration-150"
                placeholder="password1"
                on:keyup={evaluateInputs}
                on:blur={evaluateInputs}
                bind:value={passwordInput}
              />
            </div>

            <div class="relative w-full mb-3">
              <label
                class="block uppercase text-gray-700 text-xs font-bold mb-2"
                for="totpTokenInput"
              >
                2FA Token
              </label>
              <input
                id="totpTokenInput"
                type="text"
                class="px-3 py-3 placeholder-gray-400 text-gray-700 bg-white rounded text-sm shadow focus:outline-none focus:shadow-outline w-full ease-linear transition-all duration-150"
                placeholder="123456"
                on:keyup={evaluateInputs}
                on:blur={evaluateInputs}
                bind:value={totpTokenInput}
              />
            </div>

            {#if loginError !== ''}
            <p class="text-red-600">{loginError}</p>
            {/if}

            <p class="text-orange-500">{ dumpingGround }</p>

            <div class="text-center mt-6">
              <button
                on:click={login}
                type="submit"
                id="loginButton"
                class="bg-gray-900 text-white active:bg-gray-700 text-sm font-bold uppercase px-6 py-3 rounded shadow hover:shadow-lg outline-none focus:outline-none mr-1 mb-1 w-full ease-linear transition-all duration-150"
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
