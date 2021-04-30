package main

import (
	"fmt"
	"golang.org/x/net/html"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	exampleInput := `<form>
    <img class="mb-4" src="/docs/5.0/assets/brand/bootstrap-logo.svg" alt="" width="72" height="57">
    <h1 class="h3 mb-3 fw-normal">Please sign in</h1>
    <div class="form-floating">
      <input type="email" class="form-control" id="floatingInput" placeholder="name@example.com">
      <label for="floatingInput">Email address</label>
    </div>
    <div class="form-floating">
      <input type="password" class="form-control" id="floatingPassword" placeholder="Password">
      <label for="floatingPassword">Password</label>
    </div>
    <div class="checkbox mb-3">
      <label>
        <input type="checkbox" value="remember-me"> Remember me
      </label>
    </div>
    <button class="w-100 btn btn-lg btn-primary" type="submit">Sign in</button>
    <p class="mt-5 mb-3 text-muted">© 2017–2021</p>
  </form>`

	exampleInputReader := strings.NewReader(exampleInput)

	doc, err := goquery.NewDocumentFromReader(exampleInputReader)
	if err != nil {
		panic(err)
	}

	outputCodes := []string{}

	dazCmd := "New(\n\t\"html\"\n\t"
	doc.Children().Each(func(_ int, s *goquery.Selection) {
		for _, node := range s.Nodes {
			if node.Data != "html" {
				dazCmd = fmt.Sprintf("New(\n\t%q", node.Data)

				if node.Attr != nil {
					dazCmd += ",\n\t Attr{\n\t\t"
					for _, attrVal := range node.Attr {
						dazCmd += fmt.Sprintf("%q: %q,\n\t\t", attrVal.Key, attrVal.Val)
					}
					dazCmd += "},\n\t"
				}

				if node.FirstChild != nil && node.FirstChild.Type == html.TextNode && strings.TrimSpace(node.FirstChild.Data) != "" {
					dazCmd += fmt.Sprintf("%q,\n\t\t", node.FirstChild.Data)
				}
				dazCmd += "),\n"
			}
		}
		dazCmd += "),\n"
		outputCodes = append(outputCodes, dazCmd)
	})

	fmt.Println(len(outputCodes))
	fmt.Println(outputCodes)

	fmt.Println()
}
