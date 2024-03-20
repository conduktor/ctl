package schema

import (
	"github.com/conduktor/ctl/utils"
	"github.com/pb33f/libopenapi"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
	"golang.org/x/exp/slices"
	"strings"
)

type Schema struct {
	doc *libopenapi.DocumentModel[v3high.Document]
}

func New(schema []byte) (*Schema, error) {
	doc, err := libopenapi.NewDocument(schema)
	if err != nil {
		return nil, err
	}
	v3Model, errors := doc.BuildV3Model()
	if len(errors) > 0 {
		return nil, errors[0]
	}

	return &Schema{
		doc: v3Model,
	}, nil
}

func (s *Schema) GetKind() ([]string, error) {
	result := make([]string, 0)
	for path := s.doc.Model.Paths.PathItems.First(); path != nil; path = path.Next() {
		if path.Value().Get != nil {
			for _, tag := range path.Value().Get.Tags {
				if strings.HasPrefix(tag, "self-serve-") {
					newTag := utils.KebabToUpperCamel(strings.TrimPrefix(tag, "self-serve-"))
					if !slices.Contains(result, newTag) {
						result = append(result, newTag)
					}
				}
			}
		}
	}
	return result, nil
}
