// Copyright 2024 Piprate Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package locker

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/piprate/metalocker/storage/rdb/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id int) predicate.Locker {
	return predicate.Locker(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id int) predicate.Locker {
	return predicate.Locker(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id int) predicate.Locker {
	return predicate.Locker(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...int) predicate.Locker {
	return predicate.Locker(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...int) predicate.Locker {
	return predicate.Locker(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id int) predicate.Locker {
	return predicate.Locker(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id int) predicate.Locker {
	return predicate.Locker(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id int) predicate.Locker {
	return predicate.Locker(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id int) predicate.Locker {
	return predicate.Locker(sql.FieldLTE(FieldID, id))
}

// Hash applies equality check predicate on the "hash" field. It's identical to HashEQ.
func Hash(v string) predicate.Locker {
	return predicate.Locker(sql.FieldEQ(FieldHash, v))
}

// Level applies equality check predicate on the "level" field. It's identical to LevelEQ.
func Level(v int32) predicate.Locker {
	return predicate.Locker(sql.FieldEQ(FieldLevel, v))
}

// EncryptedID applies equality check predicate on the "encrypted_id" field. It's identical to EncryptedIDEQ.
func EncryptedID(v string) predicate.Locker {
	return predicate.Locker(sql.FieldEQ(FieldEncryptedID, v))
}

// EncryptedBody applies equality check predicate on the "encrypted_body" field. It's identical to EncryptedBodyEQ.
func EncryptedBody(v string) predicate.Locker {
	return predicate.Locker(sql.FieldEQ(FieldEncryptedBody, v))
}

// HashEQ applies the EQ predicate on the "hash" field.
func HashEQ(v string) predicate.Locker {
	return predicate.Locker(sql.FieldEQ(FieldHash, v))
}

// HashNEQ applies the NEQ predicate on the "hash" field.
func HashNEQ(v string) predicate.Locker {
	return predicate.Locker(sql.FieldNEQ(FieldHash, v))
}

// HashIn applies the In predicate on the "hash" field.
func HashIn(vs ...string) predicate.Locker {
	return predicate.Locker(sql.FieldIn(FieldHash, vs...))
}

// HashNotIn applies the NotIn predicate on the "hash" field.
func HashNotIn(vs ...string) predicate.Locker {
	return predicate.Locker(sql.FieldNotIn(FieldHash, vs...))
}

// HashGT applies the GT predicate on the "hash" field.
func HashGT(v string) predicate.Locker {
	return predicate.Locker(sql.FieldGT(FieldHash, v))
}

// HashGTE applies the GTE predicate on the "hash" field.
func HashGTE(v string) predicate.Locker {
	return predicate.Locker(sql.FieldGTE(FieldHash, v))
}

// HashLT applies the LT predicate on the "hash" field.
func HashLT(v string) predicate.Locker {
	return predicate.Locker(sql.FieldLT(FieldHash, v))
}

// HashLTE applies the LTE predicate on the "hash" field.
func HashLTE(v string) predicate.Locker {
	return predicate.Locker(sql.FieldLTE(FieldHash, v))
}

// HashContains applies the Contains predicate on the "hash" field.
func HashContains(v string) predicate.Locker {
	return predicate.Locker(sql.FieldContains(FieldHash, v))
}

// HashHasPrefix applies the HasPrefix predicate on the "hash" field.
func HashHasPrefix(v string) predicate.Locker {
	return predicate.Locker(sql.FieldHasPrefix(FieldHash, v))
}

// HashHasSuffix applies the HasSuffix predicate on the "hash" field.
func HashHasSuffix(v string) predicate.Locker {
	return predicate.Locker(sql.FieldHasSuffix(FieldHash, v))
}

// HashEqualFold applies the EqualFold predicate on the "hash" field.
func HashEqualFold(v string) predicate.Locker {
	return predicate.Locker(sql.FieldEqualFold(FieldHash, v))
}

// HashContainsFold applies the ContainsFold predicate on the "hash" field.
func HashContainsFold(v string) predicate.Locker {
	return predicate.Locker(sql.FieldContainsFold(FieldHash, v))
}

// LevelEQ applies the EQ predicate on the "level" field.
func LevelEQ(v int32) predicate.Locker {
	return predicate.Locker(sql.FieldEQ(FieldLevel, v))
}

// LevelNEQ applies the NEQ predicate on the "level" field.
func LevelNEQ(v int32) predicate.Locker {
	return predicate.Locker(sql.FieldNEQ(FieldLevel, v))
}

// LevelIn applies the In predicate on the "level" field.
func LevelIn(vs ...int32) predicate.Locker {
	return predicate.Locker(sql.FieldIn(FieldLevel, vs...))
}

// LevelNotIn applies the NotIn predicate on the "level" field.
func LevelNotIn(vs ...int32) predicate.Locker {
	return predicate.Locker(sql.FieldNotIn(FieldLevel, vs...))
}

// LevelGT applies the GT predicate on the "level" field.
func LevelGT(v int32) predicate.Locker {
	return predicate.Locker(sql.FieldGT(FieldLevel, v))
}

// LevelGTE applies the GTE predicate on the "level" field.
func LevelGTE(v int32) predicate.Locker {
	return predicate.Locker(sql.FieldGTE(FieldLevel, v))
}

// LevelLT applies the LT predicate on the "level" field.
func LevelLT(v int32) predicate.Locker {
	return predicate.Locker(sql.FieldLT(FieldLevel, v))
}

// LevelLTE applies the LTE predicate on the "level" field.
func LevelLTE(v int32) predicate.Locker {
	return predicate.Locker(sql.FieldLTE(FieldLevel, v))
}

// EncryptedIDEQ applies the EQ predicate on the "encrypted_id" field.
func EncryptedIDEQ(v string) predicate.Locker {
	return predicate.Locker(sql.FieldEQ(FieldEncryptedID, v))
}

// EncryptedIDNEQ applies the NEQ predicate on the "encrypted_id" field.
func EncryptedIDNEQ(v string) predicate.Locker {
	return predicate.Locker(sql.FieldNEQ(FieldEncryptedID, v))
}

// EncryptedIDIn applies the In predicate on the "encrypted_id" field.
func EncryptedIDIn(vs ...string) predicate.Locker {
	return predicate.Locker(sql.FieldIn(FieldEncryptedID, vs...))
}

// EncryptedIDNotIn applies the NotIn predicate on the "encrypted_id" field.
func EncryptedIDNotIn(vs ...string) predicate.Locker {
	return predicate.Locker(sql.FieldNotIn(FieldEncryptedID, vs...))
}

// EncryptedIDGT applies the GT predicate on the "encrypted_id" field.
func EncryptedIDGT(v string) predicate.Locker {
	return predicate.Locker(sql.FieldGT(FieldEncryptedID, v))
}

// EncryptedIDGTE applies the GTE predicate on the "encrypted_id" field.
func EncryptedIDGTE(v string) predicate.Locker {
	return predicate.Locker(sql.FieldGTE(FieldEncryptedID, v))
}

// EncryptedIDLT applies the LT predicate on the "encrypted_id" field.
func EncryptedIDLT(v string) predicate.Locker {
	return predicate.Locker(sql.FieldLT(FieldEncryptedID, v))
}

// EncryptedIDLTE applies the LTE predicate on the "encrypted_id" field.
func EncryptedIDLTE(v string) predicate.Locker {
	return predicate.Locker(sql.FieldLTE(FieldEncryptedID, v))
}

// EncryptedIDContains applies the Contains predicate on the "encrypted_id" field.
func EncryptedIDContains(v string) predicate.Locker {
	return predicate.Locker(sql.FieldContains(FieldEncryptedID, v))
}

// EncryptedIDHasPrefix applies the HasPrefix predicate on the "encrypted_id" field.
func EncryptedIDHasPrefix(v string) predicate.Locker {
	return predicate.Locker(sql.FieldHasPrefix(FieldEncryptedID, v))
}

// EncryptedIDHasSuffix applies the HasSuffix predicate on the "encrypted_id" field.
func EncryptedIDHasSuffix(v string) predicate.Locker {
	return predicate.Locker(sql.FieldHasSuffix(FieldEncryptedID, v))
}

// EncryptedIDEqualFold applies the EqualFold predicate on the "encrypted_id" field.
func EncryptedIDEqualFold(v string) predicate.Locker {
	return predicate.Locker(sql.FieldEqualFold(FieldEncryptedID, v))
}

// EncryptedIDContainsFold applies the ContainsFold predicate on the "encrypted_id" field.
func EncryptedIDContainsFold(v string) predicate.Locker {
	return predicate.Locker(sql.FieldContainsFold(FieldEncryptedID, v))
}

// EncryptedBodyEQ applies the EQ predicate on the "encrypted_body" field.
func EncryptedBodyEQ(v string) predicate.Locker {
	return predicate.Locker(sql.FieldEQ(FieldEncryptedBody, v))
}

// EncryptedBodyNEQ applies the NEQ predicate on the "encrypted_body" field.
func EncryptedBodyNEQ(v string) predicate.Locker {
	return predicate.Locker(sql.FieldNEQ(FieldEncryptedBody, v))
}

// EncryptedBodyIn applies the In predicate on the "encrypted_body" field.
func EncryptedBodyIn(vs ...string) predicate.Locker {
	return predicate.Locker(sql.FieldIn(FieldEncryptedBody, vs...))
}

// EncryptedBodyNotIn applies the NotIn predicate on the "encrypted_body" field.
func EncryptedBodyNotIn(vs ...string) predicate.Locker {
	return predicate.Locker(sql.FieldNotIn(FieldEncryptedBody, vs...))
}

// EncryptedBodyGT applies the GT predicate on the "encrypted_body" field.
func EncryptedBodyGT(v string) predicate.Locker {
	return predicate.Locker(sql.FieldGT(FieldEncryptedBody, v))
}

// EncryptedBodyGTE applies the GTE predicate on the "encrypted_body" field.
func EncryptedBodyGTE(v string) predicate.Locker {
	return predicate.Locker(sql.FieldGTE(FieldEncryptedBody, v))
}

// EncryptedBodyLT applies the LT predicate on the "encrypted_body" field.
func EncryptedBodyLT(v string) predicate.Locker {
	return predicate.Locker(sql.FieldLT(FieldEncryptedBody, v))
}

// EncryptedBodyLTE applies the LTE predicate on the "encrypted_body" field.
func EncryptedBodyLTE(v string) predicate.Locker {
	return predicate.Locker(sql.FieldLTE(FieldEncryptedBody, v))
}

// EncryptedBodyContains applies the Contains predicate on the "encrypted_body" field.
func EncryptedBodyContains(v string) predicate.Locker {
	return predicate.Locker(sql.FieldContains(FieldEncryptedBody, v))
}

// EncryptedBodyHasPrefix applies the HasPrefix predicate on the "encrypted_body" field.
func EncryptedBodyHasPrefix(v string) predicate.Locker {
	return predicate.Locker(sql.FieldHasPrefix(FieldEncryptedBody, v))
}

// EncryptedBodyHasSuffix applies the HasSuffix predicate on the "encrypted_body" field.
func EncryptedBodyHasSuffix(v string) predicate.Locker {
	return predicate.Locker(sql.FieldHasSuffix(FieldEncryptedBody, v))
}

// EncryptedBodyEqualFold applies the EqualFold predicate on the "encrypted_body" field.
func EncryptedBodyEqualFold(v string) predicate.Locker {
	return predicate.Locker(sql.FieldEqualFold(FieldEncryptedBody, v))
}

// EncryptedBodyContainsFold applies the ContainsFold predicate on the "encrypted_body" field.
func EncryptedBodyContainsFold(v string) predicate.Locker {
	return predicate.Locker(sql.FieldContainsFold(FieldEncryptedBody, v))
}

// HasAccount applies the HasEdge predicate on the "account" edge.
func HasAccount() predicate.Locker {
	return predicate.Locker(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, AccountTable, AccountColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasAccountWith applies the HasEdge predicate on the "account" edge with a given conditions (other predicates).
func HasAccountWith(preds ...predicate.Account) predicate.Locker {
	return predicate.Locker(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(AccountInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, AccountTable, AccountColumn),
		)
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.Locker) predicate.Locker {
	return predicate.Locker(func(s *sql.Selector) {
		s1 := s.Clone().SetP(nil)
		for _, p := range predicates {
			p(s1)
		}
		s.Where(s1.P())
	})
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.Locker) predicate.Locker {
	return predicate.Locker(func(s *sql.Selector) {
		s1 := s.Clone().SetP(nil)
		for i, p := range predicates {
			if i > 0 {
				s1.Or()
			}
			p(s1)
		}
		s.Where(s1.P())
	})
}

// Not applies the not operator on the given predicate.
func Not(p predicate.Locker) predicate.Locker {
	return predicate.Locker(func(s *sql.Selector) {
		p(s.Not())
	})
}
