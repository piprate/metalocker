package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Property holds the schema definition for the Property entity.
type Property struct {
	ent.Schema
}

// Fields of the Property.
func (Property) Fields() []ent.Field {
	return []ent.Field{
		field.String("hash").Unique(),
		field.Int32("level"),
		field.String("encrypted_id"),
		field.String("encrypted_body"),
	}
}

// Edges of the Property.
func (Property) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", Account.Type).Ref("properties").Unique(),
	}
}

func (Property) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("hash").Unique(),
	}
}
