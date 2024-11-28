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

package migrate

import (
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// AccessKeysColumns holds the columns for the "access_keys" table.
	AccessKeysColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "did", Type: field.TypeString, Unique: true},
		{Name: "body", Type: field.TypeJSON},
		{Name: "account", Type: field.TypeInt, Nullable: true},
	}
	// AccessKeysTable holds the schema information for the "access_keys" table.
	AccessKeysTable = &schema.Table{
		Name:       "access_keys",
		Columns:    AccessKeysColumns,
		PrimaryKey: []*schema.Column{AccessKeysColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "access_keys_accounts_access_keys",
				Columns:    []*schema.Column{AccessKeysColumns[3]},
				RefColumns: []*schema.Column{AccountsColumns[0]},
				OnDelete:   schema.SetNull,
			},
		},
		Indexes: []*schema.Index{
			{
				Name:    "accesskey_did",
				Unique:  true,
				Columns: []*schema.Column{AccessKeysColumns[1]},
			},
		},
	}
	// AccountsColumns holds the columns for the "accounts" table.
	AccountsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "did", Type: field.TypeString, Unique: true},
		{Name: "state", Type: field.TypeString},
		{Name: "email", Type: field.TypeString, Nullable: true},
		{Name: "parent_account", Type: field.TypeString, Nullable: true},
		{Name: "body", Type: field.TypeJSON},
	}
	// AccountsTable holds the schema information for the "accounts" table.
	AccountsTable = &schema.Table{
		Name:       "accounts",
		Columns:    AccountsColumns,
		PrimaryKey: []*schema.Column{AccountsColumns[0]},
		Indexes: []*schema.Index{
			{
				Name:    "account_did",
				Unique:  false,
				Columns: []*schema.Column{AccountsColumns[1]},
			},
			{
				Name:    "account_state",
				Unique:  false,
				Columns: []*schema.Column{AccountsColumns[2]},
			},
			{
				Name:    "account_email",
				Unique:  false,
				Columns: []*schema.Column{AccountsColumns[3]},
			},
			{
				Name:    "account_parent_account",
				Unique:  false,
				Columns: []*schema.Column{AccountsColumns[4]},
			},
		},
	}
	// DidDocumentsColumns holds the columns for the "did_documents" table.
	DidDocumentsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "did", Type: field.TypeString, Unique: true},
		{Name: "body", Type: field.TypeJSON},
	}
	// DidDocumentsTable holds the schema information for the "did_documents" table.
	DidDocumentsTable = &schema.Table{
		Name:       "did_documents",
		Columns:    DidDocumentsColumns,
		PrimaryKey: []*schema.Column{DidDocumentsColumns[0]},
		Indexes: []*schema.Index{
			{
				Name:    "did_did",
				Unique:  true,
				Columns: []*schema.Column{DidDocumentsColumns[1]},
			},
		},
	}
	// IdentitiesColumns holds the columns for the "identities" table.
	IdentitiesColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "hash", Type: field.TypeString, Unique: true},
		{Name: "level", Type: field.TypeInt32},
		{Name: "encrypted_id", Type: field.TypeString},
		{Name: "encrypted_body", Type: field.TypeString},
		{Name: "account", Type: field.TypeInt, Nullable: true},
	}
	// IdentitiesTable holds the schema information for the "identities" table.
	IdentitiesTable = &schema.Table{
		Name:       "identities",
		Columns:    IdentitiesColumns,
		PrimaryKey: []*schema.Column{IdentitiesColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "identities_accounts_identities",
				Columns:    []*schema.Column{IdentitiesColumns[5]},
				RefColumns: []*schema.Column{AccountsColumns[0]},
				OnDelete:   schema.SetNull,
			},
		},
		Indexes: []*schema.Index{
			{
				Name:    "identity_hash",
				Unique:  true,
				Columns: []*schema.Column{IdentitiesColumns[1]},
			},
		},
	}
	// LockersColumns holds the columns for the "lockers" table.
	LockersColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "hash", Type: field.TypeString, Unique: true},
		{Name: "level", Type: field.TypeInt32},
		{Name: "encrypted_id", Type: field.TypeString},
		{Name: "encrypted_body", Type: field.TypeString},
		{Name: "account", Type: field.TypeInt, Nullable: true},
	}
	// LockersTable holds the schema information for the "lockers" table.
	LockersTable = &schema.Table{
		Name:       "lockers",
		Columns:    LockersColumns,
		PrimaryKey: []*schema.Column{LockersColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "lockers_accounts_lockers",
				Columns:    []*schema.Column{LockersColumns[5]},
				RefColumns: []*schema.Column{AccountsColumns[0]},
				OnDelete:   schema.SetNull,
			},
		},
		Indexes: []*schema.Index{
			{
				Name:    "locker_hash",
				Unique:  true,
				Columns: []*schema.Column{LockersColumns[1]},
			},
		},
	}
	// PropertiesColumns holds the columns for the "properties" table.
	PropertiesColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "hash", Type: field.TypeString, Unique: true},
		{Name: "level", Type: field.TypeInt32},
		{Name: "encrypted_id", Type: field.TypeString},
		{Name: "encrypted_body", Type: field.TypeString},
		{Name: "account", Type: field.TypeInt, Nullable: true},
	}
	// PropertiesTable holds the schema information for the "properties" table.
	PropertiesTable = &schema.Table{
		Name:       "properties",
		Columns:    PropertiesColumns,
		PrimaryKey: []*schema.Column{PropertiesColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "properties_accounts_properties",
				Columns:    []*schema.Column{PropertiesColumns[5]},
				RefColumns: []*schema.Column{AccountsColumns[0]},
				OnDelete:   schema.SetNull,
			},
		},
		Indexes: []*schema.Index{
			{
				Name:    "property_hash",
				Unique:  true,
				Columns: []*schema.Column{PropertiesColumns[1]},
			},
		},
	}
	// RecoveryCodesColumns holds the columns for the "recovery_codes" table.
	RecoveryCodesColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "code", Type: field.TypeString, Unique: true},
		{Name: "expires_at", Type: field.TypeTime, Nullable: true},
		{Name: "account", Type: field.TypeInt, Nullable: true},
	}
	// RecoveryCodesTable holds the schema information for the "recovery_codes" table.
	RecoveryCodesTable = &schema.Table{
		Name:       "recovery_codes",
		Columns:    RecoveryCodesColumns,
		PrimaryKey: []*schema.Column{RecoveryCodesColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "recovery_codes_accounts_recovery_codes",
				Columns:    []*schema.Column{RecoveryCodesColumns[3]},
				RefColumns: []*schema.Column{AccountsColumns[0]},
				OnDelete:   schema.SetNull,
			},
		},
		Indexes: []*schema.Index{
			{
				Name:    "recoverycode_code",
				Unique:  true,
				Columns: []*schema.Column{RecoveryCodesColumns[1]},
			},
		},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		AccessKeysTable,
		AccountsTable,
		DidDocumentsTable,
		IdentitiesTable,
		LockersTable,
		PropertiesTable,
		RecoveryCodesTable,
	}
)

func init() {
	AccessKeysTable.ForeignKeys[0].RefTable = AccountsTable
	DidDocumentsTable.Annotation = &entsql.Annotation{
		Table: "did_documents",
	}
	IdentitiesTable.ForeignKeys[0].RefTable = AccountsTable
	LockersTable.ForeignKeys[0].RefTable = AccountsTable
	PropertiesTable.ForeignKeys[0].RefTable = AccountsTable
	RecoveryCodesTable.ForeignKeys[0].RefTable = AccountsTable
}
