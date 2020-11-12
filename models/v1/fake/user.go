package fakemodels

import (
	"encoding/base32"
	"fmt"
	"log"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	fake "github.com/brianvoe/gofakeit/v5"
	"github.com/pquerna/otp/totp"
)

// BuildFakeUser builds a faked User.
func BuildFakeUser() *models.User {
	return &models.User{
		ID:       uint64(fake.Uint32()),
		Username: fake.Password(true, true, true, false, false, 32),
		// HashedPassword: "",
		// Salt:           []byte(fake.Word()),
		TwoFactorSecret:           base32.StdEncoding.EncodeToString([]byte(fake.Password(false, true, true, false, false, 32))),
		TwoFactorSecretVerifiedOn: func(i uint64) *uint64 { return &i }(uint64(uint32(fake.Date().Unix()))),
		IsAdmin:                   false,
		CreatedOn:                 uint64(uint32(fake.Date().Unix())),
	}
}

// BuildDatabaseCreationResponse builds a faked UserCreationResponse.
func BuildDatabaseCreationResponse(user *models.User) *models.UserCreationResponse {
	return &models.UserCreationResponse{
		ID:                    user.ID,
		Username:              user.Username,
		PasswordLastChangedOn: user.PasswordLastChangedOn,
		IsAdmin:               user.IsAdmin,
		CreatedOn:             user.CreatedOn,
		LastUpdatedOn:         user.LastUpdatedOn,
		ArchivedOn:            user.ArchivedOn,
	}
}

// BuildFakeUserList builds a faked UserList.
func BuildFakeUserList() *models.UserList {
	exampleUser1 := BuildFakeUser()
	exampleUser2 := BuildFakeUser()
	exampleUser3 := BuildFakeUser()

	return &models.UserList{
		Pagination: models.Pagination{
			Page:  1,
			Limit: 20,
		},
		Users: []models.User{
			*exampleUser1,
			*exampleUser2,
			*exampleUser3,
		},
	}
}

// BuildFakeUserCreationInput builds a faked UserCreationInput.
func BuildFakeUserCreationInput() *models.UserCreationInput {
	exampleUser := BuildFakeUser()

	return &models.UserCreationInput{
		Username: exampleUser.Username,
		Password: fake.Password(true, true, true, true, true, 32),
	}
}

// BuildFakeUserCreationInputFromUser builds a faked UserCreationInput.
func BuildFakeUserCreationInputFromUser(user *models.User) *models.UserCreationInput {
	return &models.UserCreationInput{
		Username: user.Username,
		Password: fake.Password(true, true, true, true, true, 32),
	}
}

// BuildFakeUserDatabaseCreationInputFromUser builds a faked UserDatabaseCreationInput.
func BuildFakeUserDatabaseCreationInputFromUser(user *models.User) models.UserDatabaseCreationInput {
	return models.UserDatabaseCreationInput{
		Username:        user.Username,
		HashedPassword:  user.HashedPassword,
		TwoFactorSecret: user.TwoFactorSecret,
	}
}

// BuildFakeUserLoginInputFromUser builds a faked UserLoginInput.
func BuildFakeUserLoginInputFromUser(user *models.User) *models.UserLoginInput {
	return &models.UserLoginInput{
		Username:  user.Username,
		Password:  fake.Password(true, true, true, true, true, 32),
		TOTPToken: fmt.Sprintf("0%s", fake.Zip()),
	}
}

// BuildFakePasswordUpdateInput builds a faked PasswordUpdateInput.
func BuildFakePasswordUpdateInput() *models.PasswordUpdateInput {
	return &models.PasswordUpdateInput{
		NewPassword:     fake.Password(true, true, true, true, true, 32),
		CurrentPassword: fake.Password(true, true, true, true, true, 32),
		TOTPToken:       fmt.Sprintf("0%s", fake.Zip()),
	}
}

// BuildFakeTOTPSecretRefreshInput builds a faked TOTPSecretRefreshInput.
func BuildFakeTOTPSecretRefreshInput() *models.TOTPSecretRefreshInput {
	return &models.TOTPSecretRefreshInput{
		CurrentPassword: fake.Password(true, true, true, true, true, 32),
		TOTPToken:       fmt.Sprintf("0%s", fake.Zip()),
	}
}

// BuildFakeTOTPSecretValidationInputForUser builds a faked TOTPSecretVerificationInput for a given user.
func BuildFakeTOTPSecretValidationInputForUser(user *models.User) *models.TOTPSecretVerificationInput {
	token, err := totp.GenerateCode(user.TwoFactorSecret, time.Now().UTC())
	if err != nil {
		log.Panicf("error generating TOTP token for fake user: %v", err)
	}

	return &models.TOTPSecretVerificationInput{
		UserID:    user.ID,
		TOTPToken: token,
	}
}
