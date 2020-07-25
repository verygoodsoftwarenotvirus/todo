# todo

archetypal todo process

## dev dependencies

you'll need:

- make
- go >= 1.14
- docker
- docker-compose

The following tools are prerequisites for development work:

- [wire](https://github.com/google/wire) for dependency management
- [golangci-lint](https://github.com/golangci/golangci-lint) for linting (see included config file)

Assuming you have go installed, you can install these by running `make dev-tools`

## `make`  targets of note

- `lint` - lints the codebase
- `format` - runs `go fmt` on the codebase
- `quicktest` - runs unit tests in almost all packages, with `-failfast` turned on (skips integration/load tests, mock packages, and the `cmd` folder)
- `coverage` - will display the total test coverage in the aforementioned selection of packages
- `integration-tests-<dbprovider>` - runs the integration tests suite against an instance of the server connected to the given database. So, for instance, `integration-tests-postgres` would run the integration test suite against a Postgres database
- `load-tests-<dbprovider>` - similar to the integration tests, runs the load test suite against an instance of the server connected to the given database
- `integration-tests` - runs integration tests for all supported databases
- `lintegration-tests` - runs the integration tests and lint

It's a good idea to run `make quicktest lintegration-tests` before commits. You won't catch every error, but you'll catch the simplest ones that waste CI (and consequently your) time.

## repo organization overview

```
├── artifacts                      // gitignored, where coverage reports and such go
├── client
│   └── v1
│       └── http                   // service client, used in integration tests
├── cmd
│   ├── config_gen                 // helper binary that generates our config files programmatically so they're always correct
│   │   └── v1
│   ├── server                     // the main attraction
│   │   └── v1
│   └── tools              
│       ├── index_initializer      // tool that hydrates a bleve search index folder
│       └── two_factor             // tool for debugging login stuff, displays a given TOTP token on a loop
├── database
│   └── v1
│       ├── client                 // dbclient, wraps all querier calls in tracing and log statements
│       └── queriers               // where all the supported databases actually get queried
│           ├── mariadb
│           ├── postgres
│           └── sqlite
├── environments
│   ├── local
│   │   ├── grafana
│   │   │   └── dashboards
│   │   └── prometheus
│   └── testing
│       ├── compose_files         // docker-compose configurations
│       │   ├── integration_tests
│       │   └── load_tests
│       ├── config_files          // service configuration files
│       ├── dockerfiles           // Dockerfiles for the various test servers/runners
│       ├── grafana
│       │   └── dashboards
│       └── prometheus
├── frontend                       // the lipstick on this pig
│   └── v1
│       ├── node_modules           // the notorious, also gitignored
│       ├── public                 // the only thing in here that isn't built is the home index.html page
│       └── src
│           ├── components         // the place for common frontend elements
│           └── pages
│               └── items
├── internal                       // packages really not meant for use outside this repository, unlike the clients
│   └── v1
│       ├── auth                   // where password encryption/TOTP token verification happens
│       │   └── mock
│       ├── config                 // configuration parsing/client initialization
│       ├── encoding               // helper lib for encoding HTTP responses
│       │   └── mock
│       ├── metrics                // abstraction to provide some stability in the telemetry library space for myself
│       │   └── mock
│       ├── search
│       │   ├── bleve
│       │   └── mock
│       └── tracing                // helper libs for attaching certain IDs to spans, other such things
├── models                         // one models repository to rule them all
│   └── v1
│       ├── fake                   // the one blessed way of creating fake variables in this repo
│       └── mock
├── server                         // notice how there's room for multiple protocols, HTTP is simply the only one present
│   └── v1
│       └── http
├── services                       // I tried to make it so that you could very easily spin any of these up in their own service if your little heart so desired.
│   └── v1
│       ├── auth
│       ├── frontend
│       ├── items
│       ├── oauth2clients
│       ├── users
│       └── webhooks
└── tests
    └── v1
        ├── frontend               // selenium webdriver tests in go
        ├── integration            // using the built in go testing tool, the aforementioned client and server to test everybody working in harmony
        ├── load                   // load tests, run primarily in CI, but able to be run locally as well
        └── testutil               // some helper functions for these more involved tests only
```

## running the server

1. clone this repository
2. run `make dev`
3. [http://localhost](http://localhost)

## working on the frontend

1. run `make dev`
2. in a different terminal, cd into `frontend/v1` and run `npm run autobuild`
3. edit and have fun
