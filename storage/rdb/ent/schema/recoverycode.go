package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RecoveryCode holds the schema definition for the RecoveryCode entity.
type RecoveryCode struct {
	ent.Schema
}

// Fields of the RecoveryCode.
func (RecoveryCode) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").Unique(),
		field.Time("expires_at").Optional().Nillable(),
	}
}

// Edges of the RecoveryCode.
func (RecoveryCode) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", Account.Type).Ref("recovery_codes").Unique(),
	}
}

func (RecoveryCode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
	}
}
