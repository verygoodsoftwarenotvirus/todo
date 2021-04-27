package frontend

import (
	"bytes"
	"image/png"
	"net/url"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"github.com/tebeka/selenium/firefox"
	"github.com/tebeka/selenium/log"
)

func runTestOnAllSupportedBrowsers(t *testing.T, tp testProvider) {
	firefoxCaps, chromeCaps := selenium.Capabilities{}, selenium.Capabilities{}

	firefoxCaps.AddFirefox(firefox.Capabilities{
		Log: &firefox.Log{Level: firefox.Debug},
	})
	chromeCaps.AddChrome(chrome.Capabilities{})

	capabilities := map[string]selenium.Capabilities{
		"firefox": firefoxCaps,
		"chrome":  chromeCaps,
	}

	for browserName, capability := range capabilities {
		capability["browserName"] = browserName
		capability.AddLogging(log.Capabilities{
			log.Server:      log.Debug,
			log.Browser:     log.Debug,
			log.Client:      log.Debug,
			log.Driver:      log.Debug,
			log.Performance: log.Debug,
			log.Profiler:    log.Debug,
		})

		wd, err := selenium.NewRemote(capability, seleniumHubAddr)
		if err != nil {
			require.NoError(t, err)
		}

		t.Run(browserName, tp(wd))
		assert.NoError(t, wd.Quit())
	}
}

type testProvider func(driver selenium.WebDriver) func(t *testing.T)

func TestLoginPage(T *testing.T) {
	runTestOnAllSupportedBrowsers(T, func(driver selenium.WebDriver) func(t *testing.T) {
		return func(t *testing.T) {
			// Navigate to the login page.
			reqURI := urlToUse + "/auth/login"
			require.NoError(t, driver.Get(reqURI))

			t.Logf("USING url: %q", reqURI)

			time.Sleep(time.Second)

			waitForLoginButtonErr := driver.WaitWithTimeoutAndInterval(func(wd selenium.WebDriver) (bool, error) {
				elem, err := driver.FindElement(selenium.ByID, "loginButton")
				if err != nil {
					return false, err
				}
				return elem.IsDisplayed()
			}, 10*time.Second, time.Second)

			require.NoError(t, waitForLoginButtonErr, "unable to find loginButton: %v", waitForLoginButtonErr)

			// fetch the button.
			loginButton, loginButtonFindErr := driver.FindElement(selenium.ByID, "loginButton")
			require.NoError(t, loginButtonFindErr, "unable to find loginButton: %v", loginButtonFindErr)

			// check that it is visible.
			actual, isDisplayedErr := loginButton.IsDisplayed()
			assert.NoError(t, isDisplayedErr)
			assert.True(t, actual)
		}
	})
}

func TestRegistrationFlow(T *testing.T) {
	const (
		loginButtonID           = "loginButton"
		registrationButtonID    = "registrationButton"
		usernameInputID         = "usernameInput"
		passwordInputID         = "passwordInput"
		passwordRepeatInputID   = "passwordRepeatInput"
		twoFactorSecretQRCodeID = "twoFactorSecretQRCode"
		totpTokenSubmitButtonID = "totpTokenSubmitButton"
		totpTokenInputID        = "totpTokenInput"
	)

	runTestOnAllSupportedBrowsers(T, func(driver selenium.WebDriver) func(t *testing.T) {
		return func(t *testing.T) {
			// Navigate to the registration page.
			reqURI := urlToUse + "/auth/register"
			require.NoError(t, driver.Get(reqURI))

			time.Sleep(time.Second)

			require.NoError(t, driver.WaitWithTimeoutAndInterval(func(wd selenium.WebDriver) (bool, error) {
				elem, err := wd.FindElement(selenium.ByID, registrationButtonID)
				if err != nil {
					return false, err
				}
				return elem.IsDisplayed()
			}, 10*time.Second, time.Second))

			user := fakes.BuildFakeUserCreationInput()

			// fetch the username field and fill it
			usernameField, usernameFieldFindErr := driver.FindElement(selenium.ByID, usernameInputID)
			require.NoError(t, usernameFieldFindErr, "unexpected error finding registration page username input field: %v", usernameFieldFindErr)
			usernameEntryErr := usernameField.SendKeys(user.Username)
			require.NoError(t, usernameEntryErr, "unexpectedError filling out generated username %q: %v", user.Username, usernameEntryErr)

			t.Logf("found registration page username field and filled it with: %q", user.Username)

			// fetch the passwords field and fill it
			passwordField, passwordFieldFindErr := driver.FindElement(selenium.ByID, passwordInputID)
			require.NoError(t, passwordFieldFindErr, "unexpected error finding registration page passwords input field: %v", passwordFieldFindErr)
			registrationPagePasswordFieldInputErr := passwordField.SendKeys(user.Password)
			require.NoError(t, registrationPagePasswordFieldInputErr, "unexpected error sending input to registration page passwords input field: %v", registrationPagePasswordFieldInputErr)

			t.Logf("found registration page passwords field and filled it with: %q", user.Password)

			// fetch the passwords confirm field and fill it
			passwordRepeatField, passwordRepeatFieldFindErr := driver.FindElement(selenium.ByID, passwordRepeatInputID)
			require.NoError(t, passwordRepeatFieldFindErr, "unexpected error finding passwords repeat input field: %v", passwordRepeatFieldFindErr)
			passwordRepeatFieldInputErr := passwordRepeatField.SendKeys(user.Password)
			require.NoError(t, passwordRepeatFieldInputErr, "unexpected error sending input to blah: %v", passwordRepeatFieldInputErr)

			t.Logf("found registration page passwords repeat field and filled it with: %q", user.Password)

			// fetch the button.
			registerButton, registerButtonFindErr := driver.FindElement(selenium.ByID, registrationButtonID)
			require.NoError(t, registerButtonFindErr, "unexpected error finding registration button: %v", registerButtonFindErr)
			require.NoError(t, registerButton.Click())

			t.Logf("clicked registration button")

			time.Sleep(time.Second)

			require.NoError(t, driver.WaitWithTimeoutAndInterval(func(wd selenium.WebDriver) (bool, error) {
				qrCode, err := wd.FindElement(selenium.ByID, twoFactorSecretQRCodeID)
				if err != nil {
					return false, err
				}
				return qrCode.IsDisplayed()
			}, 10*time.Second, time.Second))

			// check that it is visible.
			qrCode, twoFactorQRCodeFindErr := driver.FindElement(selenium.ByID, twoFactorSecretQRCodeID)
			assert.NoError(t, twoFactorQRCodeFindErr, "unexpected error finding two factor QR code: %v", twoFactorQRCodeFindErr)

			t.Logf("found QR code")

			qrCodeIsDisplayed, qrCodeIsDisplayedErr := qrCode.IsDisplayed()
			require.NoError(t, qrCodeIsDisplayedErr)
			require.True(t, qrCodeIsDisplayed)

			qrScreenshotBytes, qrCodeScreenshotErr := qrCode.Screenshot(false)
			require.NoError(t, qrCodeScreenshotErr)

			t.Logf("took screenshot of QR code: %q", user.Username)

			img, err := png.Decode(bytes.NewReader(qrScreenshotBytes))
			require.NoError(t, err)

			// prepare BinaryBitmap
			bmp, bitmapErr := gozxing.NewBinaryBitmapFromImage(img)
			require.NoError(t, bitmapErr)

			// decode image
			qrReader := qrcode.NewQRCodeReader()
			result, qrCodeDecodeErr := qrReader.Decode(bmp, nil)
			require.NoError(t, qrCodeDecodeErr)

			u, secretParseErr := url.Parse(result.String())
			require.NoError(t, secretParseErr)
			twoFactorSecret := u.Query().Get("secret")
			require.NotEmpty(t, twoFactorSecret)

			code, firstCodeGenerationErr := totp.GenerateCode(twoFactorSecret, time.Now().UTC())
			require.NoError(t, firstCodeGenerationErr)
			require.NotEmpty(t, code)

			t.Logf("generated TOTP verification code: %q", code)

			// fetch the totp confirmation field and fill it
			totpTokenInputField, totpInputFieldFindErr := driver.FindElement(selenium.ByID, totpTokenInputID)
			require.NoError(t, totpInputFieldFindErr, "unexpected error finding TOTP token input field: %v", totpInputFieldFindErr)
			totpTokenConfirmationInputFieldInputErr := totpTokenInputField.SendKeys(code)
			require.NoError(t, totpTokenConfirmationInputFieldInputErr, "unexpected error sending input to TOTP token confirmation : %v", totpTokenConfirmationInputFieldInputErr)

			t.Logf("found registration page TOTP token secret validation field and filled it with: %q", code)

			// fetch the button.
			totpTokenSubmitButton, err := driver.FindElement(selenium.ByID, totpTokenSubmitButtonID)
			require.NoError(t, err)
			require.NoError(t, totpTokenSubmitButton.Click())

			t.Logf("clicked TOTP token secret validation button")

			time.Sleep(3 * time.Second)

			expectedURL := urlToUse + "/auth/login"

			actualURL, err := driver.CurrentURL()
			require.NoError(t, err)
			assert.Equal(t, expectedURL, actualURL, "expected %q to equal %q", actualURL, expectedURL)

			t.Logf("navigated to login page")

			// fetch the username field and fill it
			usernameField, usernameFieldFindErr = driver.FindElement(selenium.ByID, usernameInputID)
			require.NoError(t, usernameFieldFindErr, "unexpected error finding login page username input field: %v", usernameFieldFindErr)
			loginPageUsernameFieldInputErr := usernameField.SendKeys(user.Username)
			require.NoError(t, loginPageUsernameFieldInputErr, "unexpected error sending input to blah: %v", loginPageUsernameFieldInputErr)

			t.Logf("found login page username field and filled it with: %q", user.Username)

			// fetch the passwords field and fill it
			passwordField, passwordFieldFindErr = driver.FindElement(selenium.ByID, passwordInputID)
			require.NoError(t, passwordFieldFindErr, "unexpected error finding login page passwords input field: %v", passwordFieldFindErr)
			loginPagePasswordFieldInputErr := passwordField.SendKeys(user.Password)
			require.NoError(t, loginPagePasswordFieldInputErr, "unexpected error sending input to blah: %v", loginPagePasswordFieldInputErr)

			t.Logf("found login page passwords field and filled it with: %q", user.Password)

			code, secondCodeGenerationErr := totp.GenerateCode(twoFactorSecret, time.Now().UTC())
			require.NoError(t, secondCodeGenerationErr)
			require.NotEmpty(t, code)

			// fetch the TOTP code field and fill it
			totpTokenInputField, totpTokenInputFieldErr := driver.FindElement(selenium.ByID, totpTokenInputID)
			require.NoError(t, totpTokenInputFieldErr)
			loginPageTOTPTokenInputFieldInputErr := totpTokenInputField.SendKeys(code)
			require.NoError(t, loginPageTOTPTokenInputFieldInputErr, "unexpected error sending input to blah: %v", loginPageTOTPTokenInputFieldInputErr)

			t.Logf("found login page TOTP input field and filled it with: %q", code)

			// fetch the button.
			loginButton, loginButtonFindErr := driver.FindElement(selenium.ByID, loginButtonID)
			require.NoError(t, loginButtonFindErr, "unexpected error finding login button: %v", loginButtonFindErr)
			require.NoError(t, loginButton.Click())

			t.Logf("clicked login page login button")

			time.Sleep(5 * time.Second)

			expectedURL = urlToUse + "/"
			actualURL, err = driver.CurrentURL()
			require.NoError(t, err)
			assert.Equal(t, expectedURL, actualURL, "expected final url %q to equal %q", actualURL, expectedURL)
		}
	})
}
