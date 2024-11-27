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

const (
	// Label holds the string label denoting the did type in the database.
	Label = "did"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldDid holds the string denoting the did field in the database.
	FieldDid = "did"
	// FieldBody holds the string denoting the body field in the database.
	FieldBody = "body"
	// Table holds the table name of the did in the database.
	Table = "did_documents"
)

// Columns holds all SQL columns for did fields.
var Columns = []string{
	FieldID,
	FieldDid,
	FieldBody,
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}
