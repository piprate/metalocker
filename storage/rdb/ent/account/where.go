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

package entaccount

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/piprate/metalocker/storage/rdb/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id int) predicate.Account {
	return predicate.Account(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id int) predicate.Account {
	return predicate.Account(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id int) predicate.Account {
	return predicate.Account(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...int) predicate.Account {
	return predicate.Account(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...int) predicate.Account {
	return predicate.Account(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id int) predicate.Account {
	return predicate.Account(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id int) predicate.Account {
	return predicate.Account(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id int) predicate.Account {
	return predicate.Account(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id int) predicate.Account {
	return predicate.Account(sql.FieldLTE(FieldID, id))
}

// Did applies equality check predicate on the "did" field. It's identical to DidEQ.
func Did(v string) predicate.Account {
	return predicate.Account(sql.FieldEQ(FieldDid, v))
}

// State applies equality check predicate on the "state" field. It's identical to StateEQ.
func State(v string) predicate.Account {
	return predicate.Account(sql.FieldEQ(FieldState, v))
}

// Email applies equality check predicate on the "email" field. It's identical to EmailEQ.
func Email(v string) predicate.Account {
	return predicate.Account(sql.FieldEQ(FieldEmail, v))
}

// ParentAccount applies equality check predicate on the "parent_account" field. It's identical to ParentAccountEQ.
func ParentAccount(v string) predicate.Account {
	return predicate.Account(sql.FieldEQ(FieldParentAccount, v))
}

// DidEQ applies the EQ predicate on the "did" field.
func DidEQ(v string) predicate.Account {
	return predicate.Account(sql.FieldEQ(FieldDid, v))
}

// DidNEQ applies the NEQ predicate on the "did" field.
func DidNEQ(v string) predicate.Account {
	return predicate.Account(sql.FieldNEQ(FieldDid, v))
}

// DidIn applies the In predicate on the "did" field.
func DidIn(vs ...string) predicate.Account {
	return predicate.Account(sql.FieldIn(FieldDid, vs...))
}

// DidNotIn applies the NotIn predicate on the "did" field.
func DidNotIn(vs ...string) predicate.Account {
	return predicate.Account(sql.FieldNotIn(FieldDid, vs...))
}

// DidGT applies the GT predicate on the "did" field.
func DidGT(v string) predicate.Account {
	return predicate.Account(sql.FieldGT(FieldDid, v))
}

// DidGTE applies the GTE predicate on the "did" field.
func DidGTE(v string) predicate.Account {
	return predicate.Account(sql.FieldGTE(FieldDid, v))
}

// DidLT applies the LT predicate on the "did" field.
func DidLT(v string) predicate.Account {
	return predicate.Account(sql.FieldLT(FieldDid, v))
}

// DidLTE applies the LTE predicate on the "did" field.
func DidLTE(v string) predicate.Account {
	return predicate.Account(sql.FieldLTE(FieldDid, v))
}

// DidContains applies the Contains predicate on the "did" field.
func DidContains(v string) predicate.Account {
	return predicate.Account(sql.FieldContains(FieldDid, v))
}

// DidHasPrefix applies the HasPrefix predicate on the "did" field.
func DidHasPrefix(v string) predicate.Account {
	return predicate.Account(sql.FieldHasPrefix(FieldDid, v))
}

// DidHasSuffix applies the HasSuffix predicate on the "did" field.
func DidHasSuffix(v string) predicate.Account {
	return predicate.Account(sql.FieldHasSuffix(FieldDid, v))
}

// DidEqualFold applies the EqualFold predicate on the "did" field.
func DidEqualFold(v string) predicate.Account {
	return predicate.Account(sql.FieldEqualFold(FieldDid, v))
}

// DidContainsFold applies the ContainsFold predicate on the "did" field.
func DidContainsFold(v string) predicate.Account {
	return predicate.Account(sql.FieldContainsFold(FieldDid, v))
}

// StateEQ applies the EQ predicate on the "state" field.
func StateEQ(v string) predicate.Account {
	return predicate.Account(sql.FieldEQ(FieldState, v))
}

// StateNEQ applies the NEQ predicate on the "state" field.
func StateNEQ(v string) predicate.Account {
	return predicate.Account(sql.FieldNEQ(FieldState, v))
}

// StateIn applies the In predicate on the "state" field.
func StateIn(vs ...string) predicate.Account {
	return predicate.Account(sql.FieldIn(FieldState, vs...))
}

// StateNotIn applies the NotIn predicate on the "state" field.
func StateNotIn(vs ...string) predicate.Account {
	return predicate.Account(sql.FieldNotIn(FieldState, vs...))
}

// StateGT applies the GT predicate on the "state" field.
func StateGT(v string) predicate.Account {
	return predicate.Account(sql.FieldGT(FieldState, v))
}

// StateGTE applies the GTE predicate on the "state" field.
func StateGTE(v string) predicate.Account {
	return predicate.Account(sql.FieldGTE(FieldState, v))
}

// StateLT applies the LT predicate on the "state" field.
func StateLT(v string) predicate.Account {
	return predicate.Account(sql.FieldLT(FieldState, v))
}

// StateLTE applies the LTE predicate on the "state" field.
func StateLTE(v string) predicate.Account {
	return predicate.Account(sql.FieldLTE(FieldState, v))
}

// StateContains applies the Contains predicate on the "state" field.
func StateContains(v string) predicate.Account {
	return predicate.Account(sql.FieldContains(FieldState, v))
}

// StateHasPrefix applies the HasPrefix predicate on the "state" field.
func StateHasPrefix(v string) predicate.Account {
	return predicate.Account(sql.FieldHasPrefix(FieldState, v))
}

// StateHasSuffix applies the HasSuffix predicate on the "state" field.
func StateHasSuffix(v string) predicate.Account {
	return predicate.Account(sql.FieldHasSuffix(FieldState, v))
}

// StateEqualFold applies the EqualFold predicate on the "state" field.
func StateEqualFold(v string) predicate.Account {
	return predicate.Account(sql.FieldEqualFold(FieldState, v))
}

// StateContainsFold applies the ContainsFold predicate on the "state" field.
func StateContainsFold(v string) predicate.Account {
	return predicate.Account(sql.FieldContainsFold(FieldState, v))
}

// EmailEQ applies the EQ predicate on the "email" field.
func EmailEQ(v string) predicate.Account {
	return predicate.Account(sql.FieldEQ(FieldEmail, v))
}

// EmailNEQ applies the NEQ predicate on the "email" field.
func EmailNEQ(v string) predicate.Account {
	return predicate.Account(sql.FieldNEQ(FieldEmail, v))
}

// EmailIn applies the In predicate on the "email" field.
func EmailIn(vs ...string) predicate.Account {
	return predicate.Account(sql.FieldIn(FieldEmail, vs...))
}

// EmailNotIn applies the NotIn predicate on the "email" field.
func EmailNotIn(vs ...string) predicate.Account {
	return predicate.Account(sql.FieldNotIn(FieldEmail, vs...))
}

// EmailGT applies the GT predicate on the "email" field.
func EmailGT(v string) predicate.Account {
	return predicate.Account(sql.FieldGT(FieldEmail, v))
}

// EmailGTE applies the GTE predicate on the "email" field.
func EmailGTE(v string) predicate.Account {
	return predicate.Account(sql.FieldGTE(FieldEmail, v))
}

// EmailLT applies the LT predicate on the "email" field.
func EmailLT(v string) predicate.Account {
	return predicate.Account(sql.FieldLT(FieldEmail, v))
}

// EmailLTE applies the LTE predicate on the "email" field.
func EmailLTE(v string) predicate.Account {
	return predicate.Account(sql.FieldLTE(FieldEmail, v))
}

// EmailContains applies the Contains predicate on the "email" field.
func EmailContains(v string) predicate.Account {
	return predicate.Account(sql.FieldContains(FieldEmail, v))
}

// EmailHasPrefix applies the HasPrefix predicate on the "email" field.
func EmailHasPrefix(v string) predicate.Account {
	return predicate.Account(sql.FieldHasPrefix(FieldEmail, v))
}

// EmailHasSuffix applies the HasSuffix predicate on the "email" field.
func EmailHasSuffix(v string) predicate.Account {
	return predicate.Account(sql.FieldHasSuffix(FieldEmail, v))
}

// EmailIsNil applies the IsNil predicate on the "email" field.
func EmailIsNil() predicate.Account {
	return predicate.Account(sql.FieldIsNull(FieldEmail))
}

// EmailNotNil applies the NotNil predicate on the "email" field.
func EmailNotNil() predicate.Account {
	return predicate.Account(sql.FieldNotNull(FieldEmail))
}

// EmailEqualFold applies the EqualFold predicate on the "email" field.
func EmailEqualFold(v string) predicate.Account {
	return predicate.Account(sql.FieldEqualFold(FieldEmail, v))
}

// EmailContainsFold applies the ContainsFold predicate on the "email" field.
func EmailContainsFold(v string) predicate.Account {
	return predicate.Account(sql.FieldContainsFold(FieldEmail, v))
}

// ParentAccountEQ applies the EQ predicate on the "parent_account" field.
func ParentAccountEQ(v string) predicate.Account {
	return predicate.Account(sql.FieldEQ(FieldParentAccount, v))
}

// ParentAccountNEQ applies the NEQ predicate on the "parent_account" field.
func ParentAccountNEQ(v string) predicate.Account {
	return predicate.Account(sql.FieldNEQ(FieldParentAccount, v))
}

// ParentAccountIn applies the In predicate on the "parent_account" field.
func ParentAccountIn(vs ...string) predicate.Account {
	return predicate.Account(sql.FieldIn(FieldParentAccount, vs...))
}

// ParentAccountNotIn applies the NotIn predicate on the "parent_account" field.
func ParentAccountNotIn(vs ...string) predicate.Account {
	return predicate.Account(sql.FieldNotIn(FieldParentAccount, vs...))
}

// ParentAccountGT applies the GT predicate on the "parent_account" field.
func ParentAccountGT(v string) predicate.Account {
	return predicate.Account(sql.FieldGT(FieldParentAccount, v))
}

// ParentAccountGTE applies the GTE predicate on the "parent_account" field.
func ParentAccountGTE(v string) predicate.Account {
	return predicate.Account(sql.FieldGTE(FieldParentAccount, v))
}

// ParentAccountLT applies the LT predicate on the "parent_account" field.
func ParentAccountLT(v string) predicate.Account {
	return predicate.Account(sql.FieldLT(FieldParentAccount, v))
}

// ParentAccountLTE applies the LTE predicate on the "parent_account" field.
func ParentAccountLTE(v string) predicate.Account {
	return predicate.Account(sql.FieldLTE(FieldParentAccount, v))
}

// ParentAccountContains applies the Contains predicate on the "parent_account" field.
func ParentAccountContains(v string) predicate.Account {
	return predicate.Account(sql.FieldContains(FieldParentAccount, v))
}

// ParentAccountHasPrefix applies the HasPrefix predicate on the "parent_account" field.
func ParentAccountHasPrefix(v string) predicate.Account {
	return predicate.Account(sql.FieldHasPrefix(FieldParentAccount, v))
}

// ParentAccountHasSuffix applies the HasSuffix predicate on the "parent_account" field.
func ParentAccountHasSuffix(v string) predicate.Account {
	return predicate.Account(sql.FieldHasSuffix(FieldParentAccount, v))
}

// ParentAccountIsNil applies the IsNil predicate on the "parent_account" field.
func ParentAccountIsNil() predicate.Account {
	return predicate.Account(sql.FieldIsNull(FieldParentAccount))
}

// ParentAccountNotNil applies the NotNil predicate on the "parent_account" field.
func ParentAccountNotNil() predicate.Account {
	return predicate.Account(sql.FieldNotNull(FieldParentAccount))
}

// ParentAccountEqualFold applies the EqualFold predicate on the "parent_account" field.
func ParentAccountEqualFold(v string) predicate.Account {
	return predicate.Account(sql.FieldEqualFold(FieldParentAccount, v))
}

// ParentAccountContainsFold applies the ContainsFold predicate on the "parent_account" field.
func ParentAccountContainsFold(v string) predicate.Account {
	return predicate.Account(sql.FieldContainsFold(FieldParentAccount, v))
}

// HasRecoveryCodes applies the HasEdge predicate on the "recovery_codes" edge.
func HasRecoveryCodes() predicate.Account {
	return predicate.Account(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, RecoveryCodesTable, RecoveryCodesColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasRecoveryCodesWith applies the HasEdge predicate on the "recovery_codes" edge with a given conditions (other predicates).
func HasRecoveryCodesWith(preds ...predicate.RecoveryCode) predicate.Account {
	return predicate.Account(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(RecoveryCodesInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, RecoveryCodesTable, RecoveryCodesColumn),
		)
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasAccessKeys applies the HasEdge predicate on the "access_keys" edge.
func HasAccessKeys() predicate.Account {
	return predicate.Account(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, AccessKeysTable, AccessKeysColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasAccessKeysWith applies the HasEdge predicate on the "access_keys" edge with a given conditions (other predicates).
func HasAccessKeysWith(preds ...predicate.AccessKey) predicate.Account {
	return predicate.Account(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(AccessKeysInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, AccessKeysTable, AccessKeysColumn),
		)
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasIdentities applies the HasEdge predicate on the "identities" edge.
func HasIdentities() predicate.Account {
	return predicate.Account(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, IdentitiesTable, IdentitiesColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasIdentitiesWith applies the HasEdge predicate on the "identities" edge with a given conditions (other predicates).
func HasIdentitiesWith(preds ...predicate.Identity) predicate.Account {
	return predicate.Account(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(IdentitiesInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, IdentitiesTable, IdentitiesColumn),
		)
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasLockers applies the HasEdge predicate on the "lockers" edge.
func HasLockers() predicate.Account {
	return predicate.Account(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, LockersTable, LockersColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasLockersWith applies the HasEdge predicate on the "lockers" edge with a given conditions (other predicates).
func HasLockersWith(preds ...predicate.Locker) predicate.Account {
	return predicate.Account(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(LockersInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, LockersTable, LockersColumn),
		)
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasProperties applies the HasEdge predicate on the "properties" edge.
func HasProperties() predicate.Account {
	return predicate.Account(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, PropertiesTable, PropertiesColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasPropertiesWith applies the HasEdge predicate on the "properties" edge with a given conditions (other predicates).
func HasPropertiesWith(preds ...predicate.Property) predicate.Account {
	return predicate.Account(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(PropertiesInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, PropertiesTable, PropertiesColumn),
		)
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.Account) predicate.Account {
	return predicate.Account(func(s *sql.Selector) {
		s1 := s.Clone().SetP(nil)
		for _, p := range predicates {
			p(s1)
		}
		s.Where(s1.P())
	})
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.Account) predicate.Account {
	return predicate.Account(func(s *sql.Selector) {
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
func Not(p predicate.Account) predicate.Account {
	return predicate.Account(func(s *sql.Selector) {
		p(s.Not())
	})
}
