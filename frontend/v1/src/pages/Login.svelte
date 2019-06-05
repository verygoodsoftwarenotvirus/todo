<script>
  import { Link, navigate } from "svelte-routing";

  let username = "";
  let password = "";
  let totp_token = "";
  let canSubmit = false;

  function evaluateSubmission() {
    canSubmit =
      password.length > 0 && username.length > 0 && totp_token.length > 0;
  }

  function handleLogin() {
    fetch("/users/login", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        username,
        password,
        totp_token
      })
    }).then(function(response) {
      if (response.status != 204) {
        console.error("something has gone awry");
      } else {
        console.log("login request was good");

        window.location.replace("/");
      }
    });
  }
</script>

<form
  id="loginForm"
  on:submit|preventDefault={handleLogin}
  style="margin-top: 7.5%; text-align: center;">
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
    2FA code:
    <input bind:value={totp_token} on:keyup={evaluateSubmission} type="text" />
  </p>
  <input id="loginButton" type="submit" value="login" disabled={!canSubmit} />
  <Link to="/register">register instead</Link>
</form>
