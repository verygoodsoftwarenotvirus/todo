<script lang="typescript">
  import axios, { AxiosResponse } from "axios";
  import { link, navigate } from "svelte-routing";
  import { createPopper } from "@popperjs/core";  // library for creating dropdown menu appear on click

  let dropdownPopoverShow: Boolean = false;

  let btnDropdownRef;
  let popoverDropdownRef;

  function goToSettings() {
    navigate("/admin/settings", { state: {}, replace: true });
  }

  function logout() {
    axios.post("/users/logout", {
      withCredentials: true,
    }).then((response: AxiosResponse) => {
      if (response.status === 200) {
        navigate("/auth/login", { state: {}, replace: true })
      }
    });
  }

  const toggleDropdown = (event) => {
    event.preventDefault();
    if (dropdownPopoverShow) {
      dropdownPopoverShow = false;
    } else {
      dropdownPopoverShow = true;
      createPopper(btnDropdownRef, popoverDropdownRef, {
        placement: "bottom-start",
      });
    }
  };
</script>

<div>
  <a
    class="text-gray-600 block"
    href="#pablo"
    bind:this="{btnDropdownRef}"
    on:click="{toggleDropdown}"
  >
    <div class="items-center flex">
      <span
        class="w-12 h-12 text-sm text-white bg-gray-300 inline-flex items-center justify-center rounded-full"
      >
        <img
          alt="..."
          class="w-full rounded-full align-middle border-none shadow-lg"
          src="https://picsum.photos/seed/todo/256/256"
        />
      </span>
    </div>
  </a>
  <div
    bind:this="{popoverDropdownRef}"
    class="bg-white text-base z-50 float-left py-2 list-none text-left rounded shadow-lg min-w-48 {dropdownPopoverShow ? 'block':'hidden'}"
  >
    <button
      on:click={goToSettings}
      class="text-sm py-2 px-4 font-normal block w-full whitespace-no-wrap bg-transparent text-gray-800"
    >
      <i class="fa fa-cogs"></i>
      Settings
    </button>
    <div class="h-0 my-2 border border-solid border-gray-200" />
    <button
        on:click={logout}
        class="text-sm py-2 px-4 font-normal block w-full whitespace-no-wrap bg-transparent text-red-600"
    >
      <i class="fa fa-sign-out-alt"></i>
      Log Out
    </button>
  </div>
</div>
