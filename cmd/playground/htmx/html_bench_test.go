package main

import (
	"bytes"
	"fmt"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"html/template"
	"testing"
)

const benchTestExample = `<div id="content" class="">
    <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom"><h1 class="h2">Item #1627715355</h1></div>
    <div class="col-md-8 order-md-1">
        <form class="needs-validation" novalidate="">
            <div class="mb3">
                <label for="name">Name</label>
                <div class="input-group">
                    <input class="form-control" id="name" placeholder="Name" required="" value="" />
                    <div class="invalid-feedback" style="width: 100%;">Name is required.</div>
                </div>
            </div>
            <div class="mb-3">
				<label for="details">Details</label>
				<input class="form-control" id="details" placeholder="Details" value="" />
			</div>
            <hr class="mb-4" />
            <button class="btn btn-primary btn-lg btn-block" type="submit">Save</button>
        </form>
    </div>
</div>`

func benchTestBuildAsConstString(_ *types.Item) string {
	return benchTestExample
}

func BenchmarkBuildItemViewerAsString(b *testing.B) {
	item := fakes.BuildFakeItem()
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		benchTestBuildAsConstString(item)
	}
}

var benchTestExampleTemplate = template.Must(template.New("").Parse(`<div id="content" class="">
    <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom"><h1 class="h2">Item #1627715355</h1></div>
    <div class="col-md-8 order-md-1">
        <form class="needs-validation" novalidate="">
            <div class="mb3">
                <label for="name">Name</label>
                <div class="input-group">
                    <input class="form-control" id="name" placeholder="Name" required="" value="{{ .Name }}" />
                    <div class="invalid-feedback" style="width: 100%;">Name is required.</div>
                </div>
            </div>
            <div class="mb-3">
				<label for="details">Details</label>
				<input class="form-control" id="details" placeholder="Details" value="{{ .Details }}" />
			</div>
            <hr class="mb-4" />
            <button class="btn btn-primary btn-lg btn-block" type="submit">Save</button>
        </form>
    </div>
</div>`))

func testBuildItemViewerAsTemplate(x *types.Item) string {
	var b bytes.Buffer
	if err := benchTestExampleTemplate.Execute(&b, x); err != nil {
		panic(err)
	}
	return b.String()
}

func BenchmarkBuildItemViewerAsTemplate(b *testing.B) {
	item := fakes.BuildFakeItem()
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		testBuildItemViewerAsTemplate(item)
	}
}

const benchTestFormatTemplate = `<div id="content" class="">
    <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom"><h1 class="h2">Item #1627715355</h1></div>
    <div class="col-md-8 order-md-1">
        <form class="needs-validation" novalidate="">
            <div class="mb3">
                <label for="name">Name</label>
                <div class="input-group">
                    <input class="form-control" id="name" placeholder="Name" required="" value="%s" />
                    <div class="invalid-feedback" style="width: 100%%;">Name is required.</div>
                </div>
            </div>
            <div class="mb-3">
				<label for="details">Details</label>
				<input class="form-control" id="details" placeholder="Details" value="%s" />
			</div>
            <hr class="mb-4" />
            <button class="btn btn-primary btn-lg btn-block" type="submit">Save</button>
        </form>
    </div>
</div>`

func benchTestBuildAsSprintf(x *types.Item) string {
	return fmt.Sprintf(benchTestFormatTemplate, x.Name, x.Details)
}

func BenchmarkBuildItemViewerAsSprintf(b *testing.B) {
	item := fakes.BuildFakeItem()
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		benchTestBuildAsSprintf(item)
	}
}
