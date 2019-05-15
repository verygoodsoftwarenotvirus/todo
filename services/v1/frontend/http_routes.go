package frontend

import (
	"fmt"
	"net/http"
)

// Routes returns a map of route to handlerfunc for the parent router to set
// this keeps routing logic in the frontend service and not in the server itself.
func (s *Service) Routes() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"/":         s.HomePage,
		"/login":    s.LoginPage,
		"/register": s.RegistrationPage,
	}
}

const loginShell = `<html>
	<head>
		<title>Login</title>
		<script>
			function login() {
				var request = new XMLHttpRequest();
				request.open('POST', '%s', true);

				request.onload = function() {
					if (this.status >= 200 && this.status < 400) {
						window.location.replace("/");
					}
				}

				request.onerror = () => {
					// There was a connection error of some sort
					console.error('something has gone awry!');
				};

				request.setRequestHeader('Content-Type', 'application/json');
				const body = JSON.stringify({
					username: document.getElementById('username').value,
					password: document.getElementById('password').value,
					totp_token: document.getElementById('totp_token').value,
				})
				request.send(body);

				return false;
			}
		</script>
	</head>
	<body>
		<form id="loginForm" action="#" onsubmit="return login(this);" style="margin-top: 15%%; text-align: center;">
			<p>username: <input id="username" type="text" name="username"></p>
			<p>password: <input id="password" type="password" name="password"></p>
			<p>2FA code: <input id="totp_token" type="text" name="totp_token"></p>
			<input type="submit" value="login"> <a href="/register">register instead</a>
		</form>
	</body>
</html>`

func buildLoginPage(loginRoute LoginRoute) []byte {
	return []byte(fmt.Sprintf(loginShell, string(loginRoute)))
}

// LoginPage serves the login page
func (s *Service) LoginPage(res http.ResponseWriter, req *http.Request) {
	res.Write(s.loginPage)
}

const registerShell = `<html>
	<head>
		<title>Register</title>
		<script>
			function removeElement(id) {
				var elem = document.getElementById(id);
				return elem.parentNode.removeChild(elem);
			}

			function createUser() {
				var request = new XMLHttpRequest();
				request.open('POST', '%s', true);

				request.onload = function() {
					if (this.status >= 200 && this.status < 400) {
						var res = JSON.parse(this.response);
						var twoFactorQRCode = res.qr_code || '';
						if (twoFactorQRCode.length !== 0) {
							console.debug(res.two_factor_secret)

							// gather our container
							var containerDiv = document.getElementById("qrCodeContainer");

							// build our image
							var img = document.createElement("img");
							img.setAttribute('src', twoFactorQRCode);
							img.setAttribute('height', '40%%');

							// build our disclaimer
							var disclaimer = document.createElement("p");
							disclaimer.textContent = "This is your 2FA secret as a QR code. Save it so you can log in.";

							// build our acceptance button
							var button = document.createElement("button");
							button.textContent = "I've saved it, I promise!"
							button.onclick = function() {
								window.location.replace("/login");
							}

							// swap the divs
							containerDiv.appendChild(img);
							containerDiv.appendChild(disclaimer);
							containerDiv.appendChild(button);
							removeElement("registrationForm");
						}
					} else {
						// We reached our target server, but it returned an error
						console.error('something has gone awry!');
					}
				};

				request.onerror = () => {
					// There was a connection error of some sort
					console.error('something has gone awry!');
				};

				request.setRequestHeader('Content-Type', 'application/json');
				const body = JSON.stringify({
					username: document.getElementById('username').value,
					password: document.getElementById('password').value,
				})
				request.send(body);

				return false;
			}
		</script>
	</head>
	<body>
		<form id="registrationForm" action="#" onsubmit="return createUser(this);" style="margin-top: 15%%; text-align: center;">
			<p>username: <input id="username" type="text" name="username"></p>
			<p>password: <input id="password" type="password" name="password"></p>
			<input type="submit" value="register"> <a href="/login">log in instead</a>
		</form>
		<div id="qrCodeContainer" style="text-align: center;">
		</div>
	</body>
</html>`

func buildRegisterPage(registerRoute RegistrationRoute) []byte {
	return []byte(fmt.Sprintf(registerShell, string(registerRoute)))
}

// RegistrationPage serves the registration page
func (s *Service) RegistrationPage(res http.ResponseWriter, req *http.Request) {
	res.Write(s.registrationPage)
}

const homePage = `<html>
	<head>
		<title>Home</title>
	</head>
	<body>
		<ul>
			<li><a href="/">Here</a></li>
			<li><a href="/items">Items</a></li>
			<li><a href="/webhooks">Webhooks</a></li>
			<li><a href="/oauth2_clients">OAuth2 Clients</a></li>
			<li><a href="/login">Login</a></li>
			<li><a href="/register">Register</a></li>
		</ul>
	</body>
</html>`

// HomePage serves our homepage
func (s *Service) HomePage(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte(homePage))
}
