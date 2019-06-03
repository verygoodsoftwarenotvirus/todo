# todo

archetypal todo process

## dev dependencies

you'll need:

- make
- go >= 1.12
- [wire](https://github.com/google/wire) for dependency management
- docker
- docker-compose
- [golangci-lint](https://github.com/golangci/golangci-lint) for linting (see included config file)
- [gocov](https://github.com/axw/gocov) for coverage report generation

## running the server

1. clone this repository
2. run `make dev`
3. [http://localhost](http://localhost)

## working on the frontend

1. run `make dev`
2. in a different terminal window, run `npm run autobuild`
