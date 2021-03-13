package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing/chi"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/images"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/storage"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging/zerolog"
)

var uploadCount = 0

func main() {
	ctx := context.TODO()
	logger := zerolog.NewLogger()
	const viewPath = "/view/"

	storageConfig := &storage.Config{
		Provider:          "filesystem",
		BucketName:        "avatars",
		UploadFilenameKey: "avatar",
		AzureConfig:       nil,
		GCSConfig:         nil,
		S3Config:          nil,
		FilesystemConfig: &storage.FilesystemConfig{
			RootDirectory: "./artifacts/avatars",
		},
	}

	x, err := storage.NewUploadManager(
		ctx,
		logger,
		storageConfig,
		chi.NewRouteParamManager(),
	)
	if err != nil {
		log.Fatal(err)
	}

	um := uploads.ProvideUploadManager(x)
	up := images.NewImageUploadProcessor(logger)

	http.HandleFunc("/upload", func(res http.ResponseWriter, req *http.Request) {
		uploadCount++
		fileName := fmt.Sprintf("%d.png", uploadCount)

		img, parseErr := up.Process(req.Context(), req, "avatar")
		if parseErr != nil {
			log.Fatalf("error parsing upload: %v", parseErr)
		}

		if saveErr := um.SaveFile(ctx, fileName, img.Data); saveErr != nil {
			log.Fatalf("error saving file: %v", saveErr)
		}

		// return that we have successfully uploaded our file!
		fmt.Fprintf(res, "Successfully Uploaded File\n")
	})

	http.HandleFunc(viewPath, um.ServeFiles)

	if serveErr := http.ListenAndServe(":8080", http.DefaultServeMux); serveErr != nil {
		panic(serveErr)
	}
}
