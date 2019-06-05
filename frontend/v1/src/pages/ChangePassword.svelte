<script>
  import { Link, navigate } from "svelte-routing";

  let currentPassword = "";
  let newPassword = "";
  let newPasswordRepeat = "";
  let totpToken = "";
  let canSubmit = false;

  function evaluateSubmission() {
    canSubmit =
      newPassword.length > 0 &&
      currentPassword.length > 0 &&
      totpToken.length > 0 &&
      newPassword !== currentPassword &&
      newPasswordRepeat === newPassword;
  }

  function submitChangeRequest() {
    fetch("/users/password/new", {
      method: "PUT",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        new_password: newPassword,
        current_password: currentPassword,
        totp_token: totpToken
      })
    }).then(function(response) {
      if (response.status != 202) {
        console.error("something has gone awry");
      } else {
        navigate("/login", { replace: true });
      }
    });
  }
</script>

<form
  id="loginForm"
  on:submit|preventDefault={submitChangeRequest}
  style="margin-top: 7.5%; text-align: center;">
  <p>
    current password:
    <input
      bind:value={currentPassword}
      on:keyup={evaluateSubmission}
      type="password"
      name="username" />
  </p>
  <p>
    new password:
    <input
      bind:value={newPassword}
      on:keyup={evaluateSubmission}
      type="password"
      name="password" />
  </p>
  <p>
    once more so you're sure:
    <input
      bind:value={newPasswordRepeat}
      on:keyup={evaluateSubmission}
      type="password" />
  </p>
  <p>
    2FA code:
    <input bind:value={totpToken} on:keyup={evaluateSubmission} type="text" />
  </p>
  <input type="submit" value="change password" disabled={!canSubmit} />
</form>
