// +build wasm

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/frontend/html"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/frontend/html/components/table"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

func (a *frontendApp) buildItemCreationFunc(nameInput, detailsInput *html.Input) func() {
	return func() {
		input := &models.ItemInput{
			Name:    nameInput.Value(),
			Details: detailsInput.Value(),
		}

		creationBody, _ := json.Marshal(input)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/items/", bytes.NewReader(creationBody))
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			a.logger.Fatal(fmt.Errorf("error executing request: %v", err))
		}

		if res.StatusCode >= http.StatusOK { // http.StatusCreated {
			nameInput.SetValue("")
			detailsInput.SetValue("")
		}
	}
}

func (a *frontendApp) buildItemCreationPage() *html.Div {
	container := html.NewDiv()

	formDiv := html.NewDiv()
	formDiv.SetStyle("margin-top: 3rem; text-align: center;")

	nameP, nameInput := buildFormP("name", "name")
	detailsP, detailsInput := buildFormP("details", "details")

	submit := html.NewInput(html.SubmitInputType)
	submit.SetValue("create")
	submit.OnClick(a.buildItemCreationFunc(nameInput, detailsInput))

	formDiv.AppendChildren(
		nameP,
		detailsP,
		submit,
	)

	container.AppendChild(formDiv)
	return container
}

func up(u uint64) *uint64 {
	return &u
}

func (a *frontendApp) buildItemsPage() *html.Div {
	res, err := http.Get("/api/v1/items")
	if err != nil {
		log.Fatal(err)
	}

	var itemsRes *models.ItemList
	json.NewDecoder(res.Body).Decode(&itemsRes)

	container := html.NewDiv()

	table, err := table.NewTableFromStructs("items", itemsRes.Items)
	if err != nil {
		a.logger.Fatal(err)
	}

	container.AppendChild(table)

	return container
}
