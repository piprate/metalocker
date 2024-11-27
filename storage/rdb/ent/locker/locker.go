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

const (
	// Label holds the string label denoting the locker type in the database.
	Label = "locker"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldHash holds the string denoting the hash field in the database.
	FieldHash = "hash"
	// FieldLevel holds the string denoting the level field in the database.
	FieldLevel = "level"
	// FieldEncryptedID holds the string denoting the encrypted_id field in the database.
	FieldEncryptedID = "encrypted_id"
	// FieldEncryptedBody holds the string denoting the encrypted_body field in the database.
	FieldEncryptedBody = "encrypted_body"
	// EdgeAccount holds the string denoting the account edge name in mutations.
	EdgeAccount = "account"
	// Table holds the table name of the locker in the database.
	Table = "lockers"
	// AccountTable is the table that holds the account relation/edge.
	AccountTable = "lockers"
	// AccountInverseTable is the table name for the Account entity.
	// It exists in this package in order to avoid circular dependency with the "entaccount" package.
	AccountInverseTable = "accounts"
	// AccountColumn is the table column denoting the account relation/edge.
	AccountColumn = "account"
)

// Columns holds all SQL columns for locker fields.
var Columns = []string{
	FieldID,
	FieldHash,
	FieldLevel,
	FieldEncryptedID,
	FieldEncryptedBody,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "lockers"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"account",
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	for i := range ForeignKeys {
		if column == ForeignKeys[i] {
			return true
		}
	}
	return false
}
