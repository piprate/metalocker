package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/piprate/metalocker/model/account"
)

// Account holds the schema definition for the Account entity.
type Account struct {
	ent.Schema
}

// Fields of the Account.
func (Account) Fields() []ent.Field {
	return []ent.Field{
		field.String("did").Unique(),
		field.String("state"),
		field.String("email").Optional(),
		field.String("parent_account").Optional(),
		field.JSON("body", &account.Account{}),
	}
}

// Edges of the Account.
func (Account) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("recovery_codes", RecoveryCode.Type).StorageKey(edge.Column("account")),
		edge.To("access_keys", AccessKey.Type).StorageKey(edge.Column("account")),
		edge.To("identities", Identity.Type).StorageKey(edge.Column("account")),
		edge.To("lockers", Locker.Type).StorageKey(edge.Column("account")),
		edge.To("properties", Property.Type).StorageKey(edge.Column("account")),
	}
}

func (Account) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("did"),
		index.Fields("state"),
		index.Fields("email"),
		index.Fields("parent_account"),
	}
}
