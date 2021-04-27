/*
Package querier provides an abstraction around database querybuilding. The primary
purpose of this abstraction is to contain all the necessary logging and tracing
steps in a single place, so that the actual package that is responsible for
executing queries and loading their return values into structs isn't burdened
with inconsistent logging
*/
package querier
