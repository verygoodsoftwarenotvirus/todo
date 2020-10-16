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
	// ExampleUsers blah blah blah
	ExampleUsers = []*models.User{
		{
			Username:        defaultUsername,
			HashedPassword:  "$2a$10$JzD3CNBqPmwq.IidQuO7eu3zKdu8vEIi3HkLk8/qRjrzb7eNLKlKG",
			Salt:            []byte("aaaaaaaaaaaaaaaa"),
			TwoFactorSecret: "IFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQI=",
			IsAdmin:         true,
		},
		{
			Username:        exampleUsername1,
			HashedPassword:  "$2a$10$JzD3CNBqPmwq.IidQuO7eu3zKdu8vEIi3HkLk8/qRjrzb7eNLKlKG",
			Salt:            []byte("aaaaaaaaaaaaaaaa"),
			TwoFactorSecret: "IFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQI=",
			IsAdmin:         true,
		},
		{
			Username:        exampleUsername2,
			HashedPassword:  "$2a$10$JzD3CNBqPmwq.IidQuO7eu3zKdu8vEIi3HkLk8/qRjrzb7eNLKlKG",
			Salt:            []byte("aaaaaaaaaaaaaaaa"),
			TwoFactorSecret: "IFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQI=",
			IsAdmin:         true,
		},
		{
			Username:        exampleUsername3,
			HashedPassword:  "$2a$10$JzD3CNBqPmwq.IidQuO7eu3zKdu8vEIi3HkLk8/qRjrzb7eNLKlKG",
			Salt:            []byte("aaaaaaaaaaaaaaaa"),
			TwoFactorSecret: "IFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQKBIFAUCQI=",
			IsAdmin:         true,
		},
	}

	// ExampleItems blah blah blah
	ExampleItems = [][]*models.Item{
		{
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
		{
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
		{
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
		{
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

	// ExampleOAuth2Clients blah blah blah
	ExampleOAuth2Clients = []*models.OAuth2Client{
		{
			Name:            "example client 1",
			ClientID:        "FAKEANDUNIMPORTANT",
			ClientSecret:    "EOFITDNMKRANPTNAAU",
			RedirectURI:     "https://localhost",
			Scopes:          []string{"*"},
			ImplicitAllowed: true,
			BelongsToUser:   1,
		},
		{
			Name:            "example client 2",
			ClientID:        "FAKEANDUNIMPORTANT",
			ClientSecret:    "OTTNANIANFAEKRUPMD",
			RedirectURI:     "https://localhost",
			Scopes:          []string{"*"},
			ImplicitAllowed: true,
			BelongsToUser:   1,
		},
		{
			Name:            "example client 3",
			ClientID:        "FAKEANDUNIMPORTANT",
			ClientSecret:    "ATMNFUDRANEPOKTANI",
			RedirectURI:     "https://localhost",
			Scopes:          []string{"*"},
			ImplicitAllowed: true,
			BelongsToUser:   1,
		},
		{
			Name:            "example client 4",
			ClientID:        "FAKEANDUNIMPORTANT",
			ClientSecret:    "NEIATMOFDRNATNPKAU",
			RedirectURI:     "https://localhost",
			Scopes:          []string{"*"},
			ImplicitAllowed: true,
			BelongsToUser:   1,
		},
		{
			Name:            "example client 5",
			ClientID:        "FAKEANDUNIMPORTANT",
			ClientSecret:    "IKNNDAUOMANEFTATRP",
			RedirectURI:     "https://localhost",
			Scopes:          []string{"*"},
			ImplicitAllowed: true,
			BelongsToUser:   2,
		},
		{
			Name:            "example client 6",
			ClientID:        "FAKEANDUNIMPORTANT",
			ClientSecret:    "EOFITDNMKRANPTNAAU",
			RedirectURI:     "https://localhost",
			Scopes:          []string{"*"},
			ImplicitAllowed: true,
			BelongsToUser:   2,
		},
		{
			Name:            "example client 7",
			ClientID:        "FAKEANDUNIMPORTANT",
			ClientSecret:    "OTTNANIANFAEKRUPMD",
			RedirectURI:     "https://localhost",
			Scopes:          []string{"*"},
			ImplicitAllowed: true,
			BelongsToUser:   2,
		},
		{
			Name:            "example client 8",
			ClientID:        "FAKEANDUNIMPORTANT",
			ClientSecret:    "ATMNFUDRANEPOKTANI",
			RedirectURI:     "https://localhost",
			Scopes:          []string{"*"},
			ImplicitAllowed: true,
			BelongsToUser:   3,
		},
		{
			Name:            "example client 9",
			ClientID:        "FAKEANDUNIMPORTANT",
			ClientSecret:    "NEIATMOFDRNATNPKAU",
			RedirectURI:     "https://localhost",
			Scopes:          []string{"*"},
			ImplicitAllowed: true,
			BelongsToUser:   3,
		},
		{
			Name:            "example client 10",
			ClientID:        "FAKEANDUNIMPORTANT",
			ClientSecret:    "IKNNDAUOMANEFTATRP",
			RedirectURI:     "https://localhost",
			Scopes:          []string{"*"},
			ImplicitAllowed: true,
			BelongsToUser:   3,
		},
	}

	// ExampleWebhooks blah blah blah
	ExampleWebhooks = []*models.Webhook{
		{
			Name:          "example webhook 1",
			ContentType:   "application/json",
			URL:           "https://farts.org",
			Method:        "POST",
			Events:        []string{"*"},
			DataTypes:     []string{"*"},
			Topics:        []string{"*"},
			BelongsToUser: 1,
		},
		{
			Name:          "example webhook 2",
			ContentType:   "application/json",
			URL:           "https://farts.org",
			Method:        "POST",
			Events:        []string{"*"},
			DataTypes:     []string{"*"},
			Topics:        []string{"*"},
			BelongsToUser: 1,
		},
		{
			Name:          "example webhook 3",
			ContentType:   "application/json",
			URL:           "https://farts.org",
			Method:        "POST",
			Events:        []string{"*"},
			DataTypes:     []string{"*"},
			Topics:        []string{"*"},
			BelongsToUser: 1,
		},
		{
			Name:          "example webhook 4",
			ContentType:   "application/json",
			URL:           "https://farts.org",
			Method:        "POST",
			Events:        []string{"*"},
			DataTypes:     []string{"*"},
			Topics:        []string{"*"},
			BelongsToUser: 1,
		},
		{
			Name:          "example webhook 5",
			ContentType:   "application/json",
			URL:           "https://farts.org",
			Method:        "POST",
			Events:        []string{"*"},
			DataTypes:     []string{"*"},
			Topics:        []string{"*"},
			BelongsToUser: 2,
		},
		{
			Name:          "example webhook 6",
			ContentType:   "application/json",
			URL:           "https://farts.org",
			Method:        "POST",
			Events:        []string{"*"},
			DataTypes:     []string{"*"},
			Topics:        []string{"*"},
			BelongsToUser: 2,
		},
		{
			Name:          "example webhook 7",
			ContentType:   "application/json",
			URL:           "https://farts.org",
			Method:        "POST",
			Events:        []string{"*"},
			DataTypes:     []string{"*"},
			Topics:        []string{"*"},
			BelongsToUser: 2,
		},
		{
			Name:          "example webhook 8",
			ContentType:   "application/json",
			URL:           "https://farts.org",
			Method:        "POST",
			Events:        []string{"*"},
			DataTypes:     []string{"*"},
			Topics:        []string{"*"},
			BelongsToUser: 3,
		},
		{
			Name:          "example webhook 9",
			ContentType:   "application/json",
			URL:           "https://farts.org",
			Method:        "POST",
			Events:        []string{"*"},
			DataTypes:     []string{"*"},
			Topics:        []string{"*"},
			BelongsToUser: 3,
		},
		{
			Name:          "example webhook 10",
			ContentType:   "application/json",
			URL:           "https://farts.org",
			Method:        "POST",
			Events:        []string{"*"},
			DataTypes:     []string{"*"},
			Topics:        []string{"*"},
			BelongsToUser: 3,
		},
	}
)
