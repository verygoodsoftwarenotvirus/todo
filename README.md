# todo

archetypal todo process

## dev dependencies

The following tools are prerequisites for development work:

- [make](https://www.gnu.org/software/make/)
- [go](https://golang.org/)
- [node.js](https://nodejs.org/)
- [pnpm](https://pnpm.js.org/)
- [docker](https://docs.docker.com/get-docker/) or [podman](https://podman.io/)
- [docker-compose](https://docs.docker.com/compose/install/)
- [wire](https://github.com/google/wire) for dependency management

Assuming you have go installed, you can install these by running `make dev-tools`

## `make`  targets of note

- `dev` - run the backend
- `dev_frontend` - watch and build the frontend
- `lint` - lints the codebase
- `format` - runs `go fmt -s -w` on the codebase
- `format_frontend` - runs prettier on the frontend  
- `coverage` - will display the total test coverage in the aforementioned selection of packages
- `quicktest` - runs unit tests in almost all packages, with `-failfast` turned on (skips integration/load tests, mock packages, and the `cmd` folder)
- `integration_tests` - runs integration tests for all supported databases
- `lintegration_tests` - runs the linter, followed by the integration tests
- `integration_tests_<provider>` - runs the integration tests suite against an instance of the server connected to the given database. So, for instance, `integration-tests-postgres` would run the integration test suite against a Postgres database
- `load_tests_<provider>` - similar to the integration tests, runs the load test suite against an instance of the server connected to the given database
- `frontend_tests` - selenium webdriver tests against the frontend
- `frontend_only` - watch and build the frontend in a mode that substitutes API calls for development purposes
- `load_data` - initialize data in a running instance of the backend server 

It's a good idea to run `make quicktest lintegration-tests` before commits. You won't catch every error, but you'll catch the simplest ones that waste CI (and consequently your) time.

## running the server

1. clone this repository
2. run `make dev`
3. [http://localhost:8888/](http://localhost:8888/)

## working on the frontend

1. run `make dev`
2. in a different terminal, cd into `frontend/` and run `pnpm run autobuild`
3. edit and have fun
