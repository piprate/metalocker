package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Identity holds the schema definition for the Identity entity.
type Identity struct {
	ent.Schema
}

// Fields of the Identity.
func (Identity) Fields() []ent.Field {
	return []ent.Field{
		field.String("hash").Unique(),
		field.Int32("level"),
		field.String("encrypted_id"),
		field.String("encrypted_body"),
	}
}

// Edges of the Identity.
func (Identity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", Account.Type).Ref("identities").Unique(),
	}
}

func (Identity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("hash").Unique(),
	}
}
