package frontend

import (
	"fmt"
	"net/http"
)

const loginShell = `<html>
	<head>
		<title>Login</title>
	</head>
	<body>
		<form id="loginForm" action="%s" method="POST" style="margin-top: 15%%; text-align: center;">
			<p>username: <input type="text" name="username"></p>
			<p>password: <input type="password" name="password"></p>
			<p>2FA code: <input type="text" name="2fa_code"></p>
			<button>login</button>
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
						// Success!
						var twoFactorQRCode = JSON.parse(this.response).qr_code || '';
						if (twoFactorQRCode.length !== 0) {
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
		<div id="registrationForm" onsubmit="createUser()" style="margin-top: 15%%; text-align: center;">
			<p>username: <input id="username" type="text" name="username"></p>
			<p>password: <input id="password" type="password" name="password"></p>
			<button onclick="createUser()">register</button>
		</div>
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
