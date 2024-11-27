package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Locker holds the schema definition for the Locker entity.
type Locker struct {
	ent.Schema
}

// Fields of the Locker.
func (Locker) Fields() []ent.Field {
	return []ent.Field{
		field.String("hash").Unique(),
		field.Int32("level"),
		field.String("encrypted_id"),
		field.String("encrypted_body"),
	}
}

// Edges of the Locker.
func (Locker) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", Account.Type).Ref("lockers").Unique(),
	}
}

func (Locker) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("hash").Unique(),
	}
}
