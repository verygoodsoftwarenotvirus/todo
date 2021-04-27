/*
Package server provides a master container struct for any service that
implements a protocol. For now, it's merely an outer layer for the HTTP
implementations of our service, but in the event we wanted the same
application to listen to multiple ports for multiple protocol implementations,
this package is where those services would be started.
*/
package server
