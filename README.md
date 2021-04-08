# todo

archetypal todo process

## dev dependencies

The following tools are prerequisites for development work:

- [mage](https://www.magefile.org)
    - If you don't have `mage` installed, and you do have `go` installed, you can run `go run mage.go ensureMage` to install it.
    - If you don't have `go` installed, I can't help you.
- [go](https://golang.org/) 1.16+
- [node.js](https://nodejs.org/) and [pnpm](https://pnpm.js.org/)
- [docker](https://docs.docker.com/get-docker/) and [docker-compose](https://docs.docker.com/compose/install/)
- [wire](https://github.com/google/wire) for dependency management

## `mage`  targets of note

Assuming you have go installed, you can install prerequisite tools by running `mage dev-tools`

- `dev` - run the backend
- `dev_frontend` - watch and build the frontend
- `lint` - lints the codebase
- `format` - formats the codebase  
- `coverage` - will display the total test coverage in the aforementioned selection of packages
- `quicktest` - runs unit tests in almost all packages, with `-failfast` turned on (skips integration/load tests, mock packages, and the `cmd` folder)
- `integrationTests <provider>` -  runs the integration tests suite against an instance of the server connected to the given database. So, for instance, `integrationTests postgres` would run the integration test suite against a Postgres database
- `integrationTests` - runs the integration tests for all valid providers
- `lintegrationTests` - runs the linter and the integration tests
- `loadTests <provider>` - similar to the integration tests, runs the load test suite against an instance of the server connected to the given database
- `browserDrivenTests` - selenium webdriver tests against the frontend
- `frontendOnly` - watch and build the frontend in a mode that substitutes API calls for development purposes
- `scaffoldUsers <count>` - initialize data in a running instance of the backend server 

It's a good idea to run `mage quicktest lintegrationTests` before commits. You won't catch every error, but you'll catch the simplest ones that waste CI (and consequently your) time.

## running the server

1. clone this repository
2. run `mage run`
3. [http://localhost:8888/](http://localhost:8888/)

## working on the frontend

2. run `mage run`
2. in a different terminal, run `mage frontendAutoBuild`
3. edit and have fun
