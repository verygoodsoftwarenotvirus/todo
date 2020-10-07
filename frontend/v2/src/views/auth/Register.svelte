<script lang="typescript">
  import { link } from "svelte-routing";

  import { RegistrationPage } from "../../pages";

  const page = new RegistrationPage();

  export let location: Location;
</script>

<div class="container mx-auto px-4 h-full">
  <div class="flex content-center items-center justify-center h-full">
    <div class="w-full lg:w-6/12 px-4">
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
                placeholder="Username"
                on:keyup={page.evaluateInputs()}
                bind:value={page.usernameInput}
              />
            </div>

            <div class="relative w-full mb-3">
              <label
                class="block uppercase text-gray-700 text-xs font-bold mb-2"
                for="grid-password-1"
              >
                Password
              </label>
              <input
                id="grid-password-1"
                type="password"
                class="px-3 py-3 placeholder-gray-400 text-gray-700 bg-white rounded text-sm shadow focus:outline-none focus:shadow-outline w-full ease-linear transition-all duration-150"
                placeholder="Password"
                on:keyup={page.evaluateInputs()}
                bind:value={page.passwordInput}
              />
            </div>

            <div class="relative w-full mb-3">
              <label
                class="block uppercase text-gray-700 text-xs font-bold mb-2"
                for="grid-password-2"
              >
                Confirm Password
              </label>
              <input
                id="grid-password-2"
                type="password"
                class="px-3 py-3 placeholder-gray-400 text-gray-700 bg-white rounded text-sm shadow focus:outline-none focus:shadow-outline w-full ease-linear transition-all duration-150"
                placeholder="Password Again"
                on:keyup={page.evaluateInputs()}
                bind:value={page.passwordRepeatInput}
              />
            </div>

            <!--
            <div>
              <label class="inline-flex items-center cursor-pointer">
                <input
                  id="customCheckLogin"
                  type="checkbox"
                  class="form-checkbox text-gray-800 ml-1 w-5 h-5 ease-linear transition-all duration-150"
                />
                <span class="ml-2 text-sm font-semibold text-gray-700">
                  I agree with the
                  <a href="#pablo" on:click={(e) => e.preventDefault()} class="text-red-500">
                    Privacy Policy
                  </a>
                </span>
              </label>
            </div>
            -->

            <div class="text-center mt-6">
              <button
                class="bg-gray-900 text-white active:bg-gray-700 text-sm font-bold uppercase px-6 py-3 rounded shadow hover:shadow-lg outline-none focus:outline-none mr-1 mb-1 w-full ease-linear transition-all duration-150"
                type="button"
                on:click={page.register()}
              >
                Create Account
              </button>
            </div>
          </form>
        </div>
      </div>
      <div class="flex flex-wrap mt-6 relative">
        <div class="w-1/2"></div>
        <div class="w-1/2 text-right">
          <a use:link href="/auth/login" class="text-gray-300">
            <small>Login instead?</small>
          </a>
        </div>
      </div>
    </div>
  </div>
</div>
