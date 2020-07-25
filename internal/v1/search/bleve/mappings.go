package bleve

import (
	bleve "github.com/blevesearch/bleve"
	en "github.com/blevesearch/bleve/analysis/lang/en"
	mapping "github.com/blevesearch/bleve/mapping"
)

func buildItemMapping() *mapping.IndexMappingImpl {
	m := mapping.NewIndexMapping()

	englishTextFieldMapping := bleve.NewTextFieldMapping()
	englishTextFieldMapping.Analyzer = en.AnalyzerName

	itemMapping := bleve.NewDocumentMapping()
	itemMapping.AddFieldMappingsAt("name", englishTextFieldMapping)
	itemMapping.AddFieldMappingsAt("details", englishTextFieldMapping)
	itemMapping.AddFieldMappingsAt("belongsToUser", bleve.NewNumericFieldMapping())
	m.AddDocumentMapping("item", itemMapping)

	return m
}
