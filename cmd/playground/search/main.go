package main

import (
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search/bleve"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/zerolog"
)

func main() {
	ctx := context.Background()

	im, err := bleve.NewBleveIndexManager("whatever", models.ItemsSearchIndexName, zerolog.NewLogger())
	if err != nil {
		panic(err)
	}

	var items []models.Item
	terms := []string{
		"App",
		"Apple",
		"Apples",
		"Applesauce",
		"Appalachia",
		"Apollonia",
		"Apple Pie",
		"Apple Tree",
		"App Manager",
		"Application",
	}

	for i, t := range terms {
		items = append(items, models.Item{
			ID:   uint64(i),
			Name: t,
		})
	}

	for _, i := range items {
		if err := im.Index(ctx, i.ID, &i); err != nil {
			panic(err)
		}
	}

	results, err := im.SearchForAdmin(ctx, "Ap")
	if err != nil {
		panic(err)
	}

	fmt.Println(results)
}
