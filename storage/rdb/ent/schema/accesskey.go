package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/piprate/metalocker/model"
)

// AccessKey holds the schema definition for the AccessKey entity.
type AccessKey struct {
	ent.Schema
}

// Fields of the AccessKey.
func (AccessKey) Fields() []ent.Field {
	return []ent.Field{
		field.String("did").Unique(),
		field.JSON("body", &model.AccessKey{}),
	}
}

// Edges of the AccessKey.
func (AccessKey) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", Account.Type).Ref("access_keys").Unique(),
	}
}

func (AccessKey) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("did").Unique(),
	}
}
