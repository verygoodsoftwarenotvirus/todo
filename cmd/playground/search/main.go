package main

import (
	"context"
	"fmt"

	zerolog "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging/zerolog"
	bleve2 "gitlab.com/verygoodsoftwarenotvirus/todo/internal/search/bleve"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

func main() {
	ctx := context.Background()

	im, err := bleve2.NewBleveIndexManager("whatever", types.ItemsSearchIndexName, zerolog.NewLogger())
	if err != nil {
		panic(err)
	}

	var items []types.Item
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
		items = append(items, types.Item{
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
