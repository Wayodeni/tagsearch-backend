package service

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/lang/ru"
	"github.com/blevesearch/bleve/v2/mapping"
)

func GetIndexMapping() *mapping.IndexMappingImpl {
	indexMapping := bleve.NewIndexMapping()
	documentMapping := bleve.NewDocumentMapping()

	documentNameFieldMapping := bleve.NewTextFieldMapping()
	documentNameFieldMapping.Analyzer = ru.AnalyzerName
	documentMapping.AddFieldMappingsAt("name", documentNameFieldMapping)

	documentBodyFieldMapping := bleve.NewTextFieldMapping()
	documentNameFieldMapping.Analyzer = ru.AnalyzerName
	documentMapping.AddFieldMappingsAt("body", documentBodyFieldMapping)

	documentTagsFieldMapping := bleve.NewKeywordFieldMapping()
	documentMapping.AddFieldMappingsAt("tags", documentTagsFieldMapping)

	indexMapping.DefaultMapping = documentMapping
	indexMapping.DefaultAnalyzer = ru.AnalyzerName

	return indexMapping
}
