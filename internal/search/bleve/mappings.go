package bleve

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/lang/en"
	"github.com/blevesearch/bleve/v2/mapping"
)

func buildItemMapping() *mapping.IndexMappingImpl {
	m := mapping.NewIndexMapping()

	englishTextFieldMapping := bleve.NewTextFieldMapping()
	englishTextFieldMapping.Analyzer = en.AnalyzerName

	itemMapping := bleve.NewDocumentMapping()
	itemMapping.AddFieldMappingsAt("name", englishTextFieldMapping)
	itemMapping.AddFieldMappingsAt("details", englishTextFieldMapping)
	itemMapping.AddFieldMappingsAt("belongsToAccount", bleve.NewNumericFieldMapping())
	m.AddDocumentMapping("item", itemMapping)

	return m
}
