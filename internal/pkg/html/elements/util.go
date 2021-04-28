package elements

import "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/html"

func htmlArrayToInterfaces(in []html.HTML) []interface{} {
	out := []interface{}{}

	for _, x := range in {
		out = append(out, x)
	}

	return out
}
