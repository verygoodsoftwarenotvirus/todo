<script>
  import { Link, navigate } from "svelte-routing";

  let username = "";
  let password = "";
  let passwordCopy = "";
  let twoFactorQRCode = "";

  // state vars
  let showingSecret = false;
  let canSubmit = false;

  function evaluateSubmission() {
    canSubmit =
      password.length > 0 && username.length > 0 && passwordCopy == password;
  }

  function moseyOn() {
    navigate("/login", { replace: true });
  }

  function handleRegistration() {
    fetch("/users/", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        username,
        password
      })
    })
      .then(response => {
        if (response.status == 201) {
          return response.json();
        } else {
          console.error("something has gone awry: ");
          console.log(response);
        }
      })
      .then(data => {
        twoFactorQRCode = data["qr_code"];
        showingSecret = true;
      });
  }
</script>

<div style="margin-top: 7.5%; text-align: center;">

  {#if !showingSecret}
    <form id="registrationForm" on:submit|preventDefault={handleRegistration}>

      <p>
        username:
        <input
          bind:value={username}
          on:keyup={evaluateSubmission}
          type="text"
          name="username" />
      </p>

      <p>
        password:
        <input
          bind:value={password}
          on:keyup={evaluateSubmission}
          type="password"
          name="password" />
      </p>

      <p>
        once more so you're certain:
        <input
          bind:value={passwordCopy}
          on:keyup={evaluateSubmission}
          type="password" />
      </p>

      <input type="submit" value="register" disabled={!canSubmit} />
      <Link to="/login">log in instead</Link>

    </form>
  {:else}
    <img
      style="width: 20%;"
      src={twoFactorQRCode}
      alt="two factor authentication secret encoded as a QR code" />
    <p>
      You should save the secret this QR code contains, you'll be required to
      generate a token from it on every login.
    </p>
    <button on:click={moseyOn}>I've saved it, I promise</button>
  {/if}

</div>
