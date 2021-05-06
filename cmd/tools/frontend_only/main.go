package main

import (
	"log"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/frontend"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing/chi"
)

func main() {
	logger := zerolog.NewLogger()
	router := chi.NewRouter(logger)

	service := frontend.ProvideService(logger, nil)
	service.SetupRoutes(router)

	if err := http.ListenAndServe(":8888", router.Handler()); err != nil {
		log.Fatalln(err)
	}
}
