package frontend

import (
	"fmt"
	"net/http"
)

const (
	loginShell = `<html>
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
)

func buildLoginPage(loginRoute LoginRoute) []byte {
	return []byte(fmt.Sprintf(loginShell, string(loginRoute)))
}

// LoginPage serves the login page
func (s *Service) LoginPage(res http.ResponseWriter, req *http.Request) {
	res.Write(s.loginPage)
}
