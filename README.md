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

## running the server

1. clone this repository
2. run `make dev`
3. [http://localhost](http://localhost)

## working on the frontend

1. run `make dev`
2. in a different terminal, cd into `frontend/v2` and run `npm run autobuild`
3. edit and have fun
