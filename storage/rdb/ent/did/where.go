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

package did

import (
	"entgo.io/ent/dialect/sql"
	"github.com/piprate/metalocker/storage/rdb/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id int) predicate.DID {
	return predicate.DID(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id int) predicate.DID {
	return predicate.DID(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id int) predicate.DID {
	return predicate.DID(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...int) predicate.DID {
	return predicate.DID(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...int) predicate.DID {
	return predicate.DID(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id int) predicate.DID {
	return predicate.DID(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id int) predicate.DID {
	return predicate.DID(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id int) predicate.DID {
	return predicate.DID(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id int) predicate.DID {
	return predicate.DID(sql.FieldLTE(FieldID, id))
}

// Did applies equality check predicate on the "did" field. It's identical to DidEQ.
func Did(v string) predicate.DID {
	return predicate.DID(sql.FieldEQ(FieldDid, v))
}

// DidEQ applies the EQ predicate on the "did" field.
func DidEQ(v string) predicate.DID {
	return predicate.DID(sql.FieldEQ(FieldDid, v))
}

// DidNEQ applies the NEQ predicate on the "did" field.
func DidNEQ(v string) predicate.DID {
	return predicate.DID(sql.FieldNEQ(FieldDid, v))
}

// DidIn applies the In predicate on the "did" field.
func DidIn(vs ...string) predicate.DID {
	return predicate.DID(sql.FieldIn(FieldDid, vs...))
}

// DidNotIn applies the NotIn predicate on the "did" field.
func DidNotIn(vs ...string) predicate.DID {
	return predicate.DID(sql.FieldNotIn(FieldDid, vs...))
}

// DidGT applies the GT predicate on the "did" field.
func DidGT(v string) predicate.DID {
	return predicate.DID(sql.FieldGT(FieldDid, v))
}

// DidGTE applies the GTE predicate on the "did" field.
func DidGTE(v string) predicate.DID {
	return predicate.DID(sql.FieldGTE(FieldDid, v))
}

// DidLT applies the LT predicate on the "did" field.
func DidLT(v string) predicate.DID {
	return predicate.DID(sql.FieldLT(FieldDid, v))
}

// DidLTE applies the LTE predicate on the "did" field.
func DidLTE(v string) predicate.DID {
	return predicate.DID(sql.FieldLTE(FieldDid, v))
}

// DidContains applies the Contains predicate on the "did" field.
func DidContains(v string) predicate.DID {
	return predicate.DID(sql.FieldContains(FieldDid, v))
}

// DidHasPrefix applies the HasPrefix predicate on the "did" field.
func DidHasPrefix(v string) predicate.DID {
	return predicate.DID(sql.FieldHasPrefix(FieldDid, v))
}

// DidHasSuffix applies the HasSuffix predicate on the "did" field.
func DidHasSuffix(v string) predicate.DID {
	return predicate.DID(sql.FieldHasSuffix(FieldDid, v))
}

// DidEqualFold applies the EqualFold predicate on the "did" field.
func DidEqualFold(v string) predicate.DID {
	return predicate.DID(sql.FieldEqualFold(FieldDid, v))
}

// DidContainsFold applies the ContainsFold predicate on the "did" field.
func DidContainsFold(v string) predicate.DID {
	return predicate.DID(sql.FieldContainsFold(FieldDid, v))
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.DID) predicate.DID {
	return predicate.DID(func(s *sql.Selector) {
		s1 := s.Clone().SetP(nil)
		for _, p := range predicates {
			p(s1)
		}
		s.Where(s1.P())
	})
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.DID) predicate.DID {
	return predicate.DID(func(s *sql.Selector) {
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
func Not(p predicate.DID) predicate.DID {
	return predicate.DID(func(s *sql.Selector) {
		p(s.Not())
	})
}
