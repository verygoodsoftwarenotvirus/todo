# todo

archetypal todo process

## dev dependencies

you'll need:

- make
- go >= 1.12
- docker
- docker-compose

the following tools are occasionally required for development:

- [wire](https://github.com/google/wire) for dependency management
- [golangci-lint](https://github.com/golangci/golangci-lint) for linting (see included config file)
- [gocov](https://github.com/axw/gocov) for coverage report generation

assuming you have go installed, you can install these by running `make dev-tools`

## running the server

1. clone this repository
2. run `make dev`
3. [http://localhost](http://localhost)

## working on the frontend

1. run `make dev`
2. in a different terminal, cd into `frontend/v1` and run `npm run autobuild`
3. edit and have fun
