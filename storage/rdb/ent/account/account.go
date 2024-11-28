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

const (
	// Label holds the string label denoting the account type in the database.
	Label = "account"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldDid holds the string denoting the did field in the database.
	FieldDid = "did"
	// FieldState holds the string denoting the state field in the database.
	FieldState = "state"
	// FieldEmail holds the string denoting the email field in the database.
	FieldEmail = "email"
	// FieldParentAccount holds the string denoting the parent_account field in the database.
	FieldParentAccount = "parent_account"
	// FieldBody holds the string denoting the body field in the database.
	FieldBody = "body"
	// EdgeRecoveryCodes holds the string denoting the recovery_codes edge name in mutations.
	EdgeRecoveryCodes = "recovery_codes"
	// EdgeAccessKeys holds the string denoting the access_keys edge name in mutations.
	EdgeAccessKeys = "access_keys"
	// EdgeIdentities holds the string denoting the identities edge name in mutations.
	EdgeIdentities = "identities"
	// EdgeLockers holds the string denoting the lockers edge name in mutations.
	EdgeLockers = "lockers"
	// EdgeProperties holds the string denoting the properties edge name in mutations.
	EdgeProperties = "properties"
	// Table holds the table name of the account in the database.
	Table = "accounts"
	// RecoveryCodesTable is the table that holds the recovery_codes relation/edge.
	RecoveryCodesTable = "recovery_codes"
	// RecoveryCodesInverseTable is the table name for the RecoveryCode entity.
	// It exists in this package in order to avoid circular dependency with the "recoverycode" package.
	RecoveryCodesInverseTable = "recovery_codes"
	// RecoveryCodesColumn is the table column denoting the recovery_codes relation/edge.
	RecoveryCodesColumn = "account"
	// AccessKeysTable is the table that holds the access_keys relation/edge.
	AccessKeysTable = "access_keys"
	// AccessKeysInverseTable is the table name for the AccessKey entity.
	// It exists in this package in order to avoid circular dependency with the "accesskey" package.
	AccessKeysInverseTable = "access_keys"
	// AccessKeysColumn is the table column denoting the access_keys relation/edge.
	AccessKeysColumn = "account"
	// IdentitiesTable is the table that holds the identities relation/edge.
	IdentitiesTable = "identities"
	// IdentitiesInverseTable is the table name for the Identity entity.
	// It exists in this package in order to avoid circular dependency with the "identity" package.
	IdentitiesInverseTable = "identities"
	// IdentitiesColumn is the table column denoting the identities relation/edge.
	IdentitiesColumn = "account"
	// LockersTable is the table that holds the lockers relation/edge.
	LockersTable = "lockers"
	// LockersInverseTable is the table name for the Locker entity.
	// It exists in this package in order to avoid circular dependency with the "locker" package.
	LockersInverseTable = "lockers"
	// LockersColumn is the table column denoting the lockers relation/edge.
	LockersColumn = "account"
	// PropertiesTable is the table that holds the properties relation/edge.
	PropertiesTable = "properties"
	// PropertiesInverseTable is the table name for the Property entity.
	// It exists in this package in order to avoid circular dependency with the "property" package.
	PropertiesInverseTable = "properties"
	// PropertiesColumn is the table column denoting the properties relation/edge.
	PropertiesColumn = "account"
)

// Columns holds all SQL columns for account fields.
var Columns = []string{
	FieldID,
	FieldDid,
	FieldState,
	FieldEmail,
	FieldParentAccount,
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
