package exampledata

import "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

/*

	for _, x := range []string{"username", "username1", "username2", "username3"} {
		query, args, err := p.sqlBuilder.
			Insert(usersTableName).
			Columns(
				usersTableUsernameColumn,
				usersTableHashedPasswordColumn,
				usersTableSaltColumn,
				usersTableTwoFactorColumn,
				usersTableIsAdminColumn,
				usersTableTwoFactorVerifiedOnColumn,
			).
			Values(
				x,
				"$2a$10$JzD3CNBqPmwq.IidQuO7eu3zKdu8vEIi3HkLk8/qRjrzb7eNLKlKG",
				[]byte("aaaaaaaaaaaaaaaa"),
				// `otpauth://totp/todo:username?secret=IFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQI=&issuer=todo`
				"IFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQI=",
				true,
				squirrel.Expr(currentUnixTimeQuery),
			).
			ToSql()
		p.logQueryBuildingError(err)

		if _, dbErr := p.db.ExecContext(ctx, query, args...); dbErr != nil {
			return dbErr
		}
	}

	for _, x := range []uint64{1, 2, 3, 4} {
		i := fakemodels.BuildFakeItem()
		i.BelongsToUser = x
		for y := 0; y < 5; y++ {
			z := fakemodels.BuildFakeItem()
			z.Name = fmt.Sprintf("%s #%d", i.Name, y)
			z.BelongsToUser = x
			query, args := p.buildCreateItemQuery(z)
			if _, dbErr := p.db.ExecContext(ctx, query, args...); dbErr != nil {
				return dbErr
			}
		}
	}

*/

const (
	defaultUsername  = "username"
	exampleUsername1 = defaultUsername + "1"
	exampleUsername2 = defaultUsername + "2"
	exampleUsername3 = defaultUsername + "3"
)

var (
	defaultExampleUser = &models.User{
		Username:        defaultUsername,
		HashedPassword:  "$2a$10$JzD3CNBqPmwq.IidQuO7eu3zKdu8vEIi3HkLk8/qRjrzb7eNLKlKG",
		Salt:            []byte("aaaaaaaaaaaaaaaa"),
		TwoFactorSecret: "IFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQI=",
		IsAdmin:         true,
	}

	exampleUser1 = &models.User{
		Username:        exampleUsername1,
		HashedPassword:  "$2a$10$JzD3CNBqPmwq.IidQuO7eu3zKdu8vEIi3HkLk8/qRjrzb7eNLKlKG",
		Salt:            []byte("aaaaaaaaaaaaaaaa"),
		TwoFactorSecret: "IFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQI=",
		IsAdmin:         true,
	}

	exampleUser2 = &models.User{
		Username:        exampleUsername2,
		HashedPassword:  "$2a$10$JzD3CNBqPmwq.IidQuO7eu3zKdu8vEIi3HkLk8/qRjrzb7eNLKlKG",
		Salt:            []byte("aaaaaaaaaaaaaaaa"),
		TwoFactorSecret: "IFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQI=",
		IsAdmin:         true,
	}

	exampleUser3 = &models.User{
		Username:        exampleUsername3,
		HashedPassword:  "$2a$10$JzD3CNBqPmwq.IidQuO7eu3zKdu8vEIi3HkLk8/qRjrzb7eNLKlKG",
		Salt:            []byte("aaaaaaaaaaaaaaaa"),
		TwoFactorSecret: "IFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQI=",
		IsAdmin:         true,
	}

	// ExampleUsers blah blah blah
	ExampleUsers = []*models.User{
		defaultExampleUser,
		exampleUser1,
		exampleUser2,
		exampleUser3,
	}

	// ExampleItemMap blah blah blah
	ExampleItemMap = map[string][]*models.Item{
		defaultUsername: {
			{
				Name:          "Apple #1",
				BelongsToUser: 1,
			},
			{
				Name:          "Apple Pie #2",
				BelongsToUser: 1,
			},
			{
				Name:          "Apple Fritter #3",
				BelongsToUser: 1,
			},
			{
				Name:          "Apple Butter #4",
				BelongsToUser: 1,
			},
			{
				Name:          "Apple Jack #5",
				BelongsToUser: 1,
			},
		},
		exampleUsername1: {
			{
				Name:          "Application #1",
				BelongsToUser: 2,
			},
			{
				Name:          "Application #2",
				BelongsToUser: 2,
			},
			{
				Name:          "Application #3",
				BelongsToUser: 2,
			},
			{
				Name:          "Application #4",
				BelongsToUser: 2,
			},
			{
				Name:          "Application #5",
				BelongsToUser: 2,
			},
		},
		exampleUsername2: {
			{
				Name:          "Appliance #1",
				BelongsToUser: 3,
			},
			{
				Name:          "Appliance #2",
				BelongsToUser: 3,
			},
			{
				Name:          "Appliance #3",
				BelongsToUser: 3,
			},
			{
				Name:          "Appliance #4",
				BelongsToUser: 3,
			},
			{
				Name:          "Appliance #5",
				BelongsToUser: 3,
			},
		},
		exampleUsername3: {
			{
				Name:          "Apple #1",
				BelongsToUser: 4,
			},
			{
				Name:          "Apple Pie #2",
				BelongsToUser: 4,
			},
			{
				Name:          "Apple Fritter #3",
				BelongsToUser: 4,
			},
			{
				Name:          "Apple Butter #4",
				BelongsToUser: 4,
			},
			{
				Name:          "Apple Jack #5",
				BelongsToUser: 4,
			},
		},
	}
)
