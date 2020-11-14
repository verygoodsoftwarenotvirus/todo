package fakes

import (
	"encoding/base32"
	"fmt"
	"log"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions/bitmask"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
	"github.com/pquerna/otp/totp"
)

// BuildFakeUser builds a faked User.
func BuildFakeUser() *types.User {
	return &types.User{
		ID:       uint64(fake.Uint32()),
		Username: fake.Password(true, true, true, false, false, 32),
		// HashedPassword: "",
		// Salt:           []byte(fakes.Word()),
		TwoFactorSecret:           base32.StdEncoding.EncodeToString([]byte(fake.Password(false, true, true, false, false, 32))),
		TwoFactorSecretVerifiedOn: func(i uint64) *uint64 { return &i }(uint64(uint32(fake.Date().Unix()))),
		IsAdmin:                   false,
		AdminPermissions:          bitmask.NewPermissionBitmask(0),
		CreatedOn:                 uint64(uint32(fake.Date().Unix())),
	}
}

// BuildDatabaseCreationResponse builds a faked UserCreationResponse.
func BuildDatabaseCreationResponse(user *types.User) *types.UserCreationResponse {
	return &types.UserCreationResponse{
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
func BuildFakeUserList() *types.UserList {
	exampleUser1 := BuildFakeUser()
	exampleUser2 := BuildFakeUser()
	exampleUser3 := BuildFakeUser()

	return &types.UserList{
		Pagination: types.Pagination{
			Page:  1,
			Limit: 20,
		},
		Users: []types.User{
			*exampleUser1,
			*exampleUser2,
			*exampleUser3,
		},
	}
}

// BuildFakeUserCreationInput builds a faked UserCreationInput.
func BuildFakeUserCreationInput() *types.UserCreationInput {
	exampleUser := BuildFakeUser()

	return &types.UserCreationInput{
		Username: exampleUser.Username,
		Password: fake.Password(true, true, true, true, true, 32),
	}
}

// BuildFakeUserCreationInputFromUser builds a faked UserCreationInput.
func BuildFakeUserCreationInputFromUser(user *types.User) *types.UserCreationInput {
	return &types.UserCreationInput{
		Username: user.Username,
		Password: fake.Password(true, true, true, true, true, 32),
	}
}

// BuildFakeUserDatabaseCreationInputFromUser builds a faked UserDatabaseCreationInput.
func BuildFakeUserDatabaseCreationInputFromUser(user *types.User) types.UserDatabaseCreationInput {
	return types.UserDatabaseCreationInput{
		Username:        user.Username,
		HashedPassword:  user.HashedPassword,
		TwoFactorSecret: user.TwoFactorSecret,
	}
}

// BuildFakeUserLoginInputFromUser builds a faked UserLoginInput.
func BuildFakeUserLoginInputFromUser(user *types.User) *types.UserLoginInput {
	return &types.UserLoginInput{
		Username:  user.Username,
		Password:  fake.Password(true, true, true, true, true, 32),
		TOTPToken: fmt.Sprintf("0%s", fake.Zip()),
	}
}

// BuildFakePasswordUpdateInput builds a faked PasswordUpdateInput.
func BuildFakePasswordUpdateInput() *types.PasswordUpdateInput {
	return &types.PasswordUpdateInput{
		NewPassword:     fake.Password(true, true, true, true, true, 32),
		CurrentPassword: fake.Password(true, true, true, true, true, 32),
		TOTPToken:       fmt.Sprintf("0%s", fake.Zip()),
	}
}

// BuildFakeTOTPSecretRefreshInput builds a faked TOTPSecretRefreshInput.
func BuildFakeTOTPSecretRefreshInput() *types.TOTPSecretRefreshInput {
	return &types.TOTPSecretRefreshInput{
		CurrentPassword: fake.Password(true, true, true, true, true, 32),
		TOTPToken:       fmt.Sprintf("0%s", fake.Zip()),
	}
}

// BuildFakeTOTPSecretValidationInputForUser builds a faked TOTPSecretVerificationInput for a given user.
func BuildFakeTOTPSecretValidationInputForUser(user *types.User) *types.TOTPSecretVerificationInput {
	token, err := totp.GenerateCode(user.TwoFactorSecret, time.Now().UTC())
	if err != nil {
		log.Panicf("error generating TOTP token for fakes user: %v", err)
	}

	return &types.TOTPSecretVerificationInput{
		UserID:    user.ID,
		TOTPToken: token,
	}
}
