package fakes

import (
	"encoding/base32"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions/bitmask"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
	"github.com/pquerna/otp/totp"
)

// BuildFakeUser builds a faked User.
func BuildFakeUser() *types.User {
	return &types.User{
		ID:         uint64(fake.Uint32()),
		ExternalID: uuid.New().String(),
		Username:   fake.Password(true, true, true, false, false, 32),
		// HashedPassword: "",
		// Salt:           []byte(fakes.Word()),
		TwoFactorSecret:           base32.StdEncoding.EncodeToString([]byte(fake.Password(false, true, true, false, false, 32))),
		TwoFactorSecretVerifiedOn: func(i uint64) *uint64 { return &i }(uint64(uint32(fake.Date().Unix()))),
		IsSiteAdmin:               false,
		AdminPermissions:          bitmask.NewPermissionBitmask(0),
		CreatedOn:                 uint64(uint32(fake.Date().Unix())),
	}
}

// BuildUserCreationResponseFromUser builds a faked UserCreationResponse.
func BuildUserCreationResponseFromUser(user *types.User) *types.UserCreationResponse {
	return &types.UserCreationResponse{
		ID:        user.ID,
		Username:  user.Username,
		IsAdmin:   user.IsSiteAdmin,
		CreatedOn: user.CreatedOn,
	}
}

// BuildFakeUserList builds a faked UserList.
func BuildFakeUserList() *types.UserList {
	var examples []*types.User
	for i := 0; i < exampleQuantity; i++ {
		examples = append(examples, BuildFakeUser())
	}

	return &types.UserList{
		Pagination: types.Pagination{
			Page:          1,
			Limit:         20,
			FilteredCount: exampleQuantity / 2,
			TotalCount:    exampleQuantity,
		},
		Users: examples,
	}
}

// BuildFakeUserCreationInput builds a faked NewUserCreationInput.
func BuildFakeUserCreationInput() *types.NewUserCreationInput {
	exampleUser := BuildFakeUser()

	return &types.NewUserCreationInput{
		Username: exampleUser.Username,
		Password: fake.Password(true, true, true, true, true, 32),
	}
}

// BuildTestUserCreationConfig builds a faked TestUserCreationConfig.
func BuildTestUserCreationConfig() *types.TestUserCreationConfig {
	exampleUser := BuildFakeUserCreationInput()

	return &types.TestUserCreationConfig{
		Username:       exampleUser.Username,
		Password:       exampleUser.Password,
		HashedPassword: "hashed password",
		IsSiteAdmin:    false,
	}
}

// BuildFakeUserCreationInputFromUser builds a faked NewUserCreationInput.
func BuildFakeUserCreationInputFromUser(user *types.User) *types.NewUserCreationInput {
	return &types.NewUserCreationInput{
		Username: user.Username,
		Password: fake.Password(true, true, true, true, true, 32),
	}
}

// BuildFakeUserDataStoreCreationInputFromUser builds a faked UserDataStoreCreationInput.
func BuildFakeUserDataStoreCreationInputFromUser(user *types.User) types.UserDataStoreCreationInput {
	return types.UserDataStoreCreationInput{
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

// BuildFakeTOTPSecretVerificationInput builds a faked TOTPSecretVerificationInput.
func BuildFakeTOTPSecretVerificationInput() *types.TOTPSecretVerificationInput {
	user := BuildFakeUser()

	token, err := totp.GenerateCode(user.TwoFactorSecret, time.Now().UTC())
	if err != nil {
		log.Panicf("error generating TOTP token for fakes user: %v", err)
	}

	return &types.TOTPSecretVerificationInput{
		UserID:    user.ID,
		TOTPToken: token,
	}
}

// BuildFakeTOTPSecretVerificationInputForUser builds a faked TOTPSecretVerificationInput for a given user.
func BuildFakeTOTPSecretVerificationInputForUser(user *types.User) *types.TOTPSecretVerificationInput {
	token, err := totp.GenerateCode(user.TwoFactorSecret, time.Now().UTC())
	if err != nil {
		log.Panicf("error generating TOTP token for fakes user: %v", err)
	}

	return &types.TOTPSecretVerificationInput{
		UserID:    user.ID,
		TOTPToken: token,
	}
}
