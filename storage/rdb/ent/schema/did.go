package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/piprate/metalocker/model"
)

// DID holds the schema definition for the DID entity.
type DID struct {
	ent.Schema
}

// Annotations of the DID.
func (DID) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "did_documents"},
	}
}

// Fields of the DID.
func (DID) Fields() []ent.Field {
	return []ent.Field{
		field.String("did").Unique(),
		field.JSON("body", &model.DIDDocument{}),
	}
}

// Edges of the DID.
func (DID) Edges() []ent.Edge {
	return nil
}

func (DID) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("did").Unique(),
	}
}
