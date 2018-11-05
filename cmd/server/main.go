package main

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo"
)

const (
	secure   = false
	certFile = "certs/server/cert.pem"
	keyFile  = "certs/server/key.pem"
)

func main() {
	srv := todo.NewDebug(certFile, keyFile)
	srv.Serve()
}
