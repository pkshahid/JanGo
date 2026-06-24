package contenttypes

import (
	"fmt"
	"reflect"
)

// GenericForeignKey represents a generic relation to any content type.
// Equivalent to Django's GenericForeignKey.
type GenericForeignKey struct {
	ContentTypeID int
	ObjectID      interface{}
}

// NewGenericForeignKey creates a new generic foreign key reference.
func NewGenericForeignKey(contentTypeID int, objectID interface{}) *GenericForeignKey {
	return &GenericForeignKey{
		ContentTypeID: contentTypeID,
		ObjectID:      objectID,
	}
}

// Resolve returns the ContentType associated with this generic FK.
func (gfk *GenericForeignKey) Resolve() (*ContentType, error) {
	return Get(gfk.ContentTypeID)
}

// GenericRelation represents the reverse side of a generic foreign key.
// Equivalent to Django's GenericRelation.
type GenericRelation struct {
	ContentType    *ContentType
	RelatedModel   reflect.Type
	ContentTypeField string
	ObjectIDField    string
}

// NewGenericRelation creates a generic relation configuration.
func NewGenericRelation(relatedModel reflect.Type, ctField, objIDField string) *GenericRelation {
	if ctField == "" {
		ctField = "ContentTypeID"
	}
	if objIDField == "" {
		objIDField = "ObjectID"
	}
	return &GenericRelation{
		RelatedModel:     relatedModel,
		ContentTypeField: ctField,
		ObjectIDField:    objIDField,
	}
}

// GenericInlineModelAdmin represents inline admin for generic relations.
type GenericInlineModelAdmin struct {
	Model          reflect.Type
	ContentTypeField string
	ObjectIDField    string
	Extra          int
}

// TaggedItem is a common pattern: tagging using generic relations.
type TaggedItem struct {
	ID            int
	Tag           string
	ContentTypeID int
	ObjectID      int
}

// GetContentType returns the ContentType for a tagged item.
func (ti *TaggedItem) GetContentType() (*ContentType, error) {
	return Get(ti.ContentTypeID)
}

// PolymorphicQuery provides type-aware queries across content types.
type PolymorphicQuery struct {
	ContentTypes []*ContentType
}

// NewPolymorphicQuery creates a query that can span multiple content types.
func NewPolymorphicQuery(types ...*ContentType) *PolymorphicQuery {
	return &PolymorphicQuery{ContentTypes: types}
}

// FilterByType returns content types matching the given model name.
func (pq *PolymorphicQuery) FilterByType(model string) []*ContentType {
	var result []*ContentType
	for _, ct := range pq.ContentTypes {
		if ct.Model == model {
			result = append(result, ct)
		}
	}
	return result
}

// FilterByApp returns content types matching the given app label.
func (pq *PolymorphicQuery) FilterByApp(appLabel string) []*ContentType {
	var result []*ContentType
	for _, ct := range pq.ContentTypes {
		if ct.AppLabel == appLabel {
			result = append(result, ct)
		}
	}
	return result
}

// GetObjectForType creates a new instance of the model for a given content type.
func GetObjectForType(ct *ContentType) (interface{}, error) {
	if ct.GoType == nil {
		return nil, fmt.Errorf("contenttypes: content type %s has no Go type", ct)
	}
	return reflect.New(ct.GoType).Interface(), nil
}
