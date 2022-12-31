// Copyright 2022 Piprate Limited
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

package actions

import "github.com/urfave/cli/v2"

var (
	AccountSet = []*cli.Command{
		{
			Name:   "register",
			Usage:  "register new user",
			Action: RegisterCommand,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "email",
					Value: "",
					Usage: "Account email",
				},
				&cli.StringFlag{
					Name:    "password",
					Value:   "",
					Usage:   "Password",
					EnvVars: []string{"PASSWD"},
				},
				&cli.StringFlag{
					Name:    "code",
					Value:   "",
					Usage:   "Registration Code (optional)",
					EnvVars: []string{"REGCODE"},
				},
				&cli.StringFlag{
					Name:  "identity-file",
					Value: "",
					Usage: "Identity file",
				},
				&cli.StringFlag{
					Name:  "name",
					Value: "",
					Usage: "Full name",
				},
			},
		},
		{
			Name:   "print-account",
			Usage:  "print account JSON",
			Action: PrintAccount,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "detailed",
					Usage: "include all identities and lockers",
				},
			},
		},
		{
			Name:   "print-account-chart",
			Usage:  "print account chart",
			Action: PrintAccountChart,
		},
		{
			Name:   "delete-account",
			Usage:  "delete account",
			Action: DeleteAccount,
		},
		{
			Name:   "change-email",
			Usage:  "change account email",
			Action: ChangeEmail,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "email",
					Value:   "",
					Usage:   "New account email",
					EnvVars: []string{"NEWUSER"},
				},
			},
		},
		{
			Name:   "change-password",
			Usage:  "change account password",
			Action: ChangePassphrase,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "new-password",
					Value:   "",
					Usage:   "New Password",
					EnvVars: []string{"NEWPASSWD"},
				},
			},
		},
		{
			Name:   "recover-account",
			Usage:  "recover account",
			Action: RecoverAccount,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "rec-phrase",
					Value:   "",
					Usage:   "Recovery phrase (received at account creation)",
					EnvVars: []string{"RECPHRASE"},
				},
				&cli.StringFlag{
					Name:    "new-password",
					Value:   "",
					Usage:   "New Password",
					EnvVars: []string{"NEWPASSWD"},
				},
			},
		},
		{
			Name:   "recover-account-with-slrc",
			Usage:  "recover account with SLRC code",
			Action: RecoverAccountSecondLevel,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "slrc",
					Value:   "",
					Usage:   "Second Level Recovery Code [SLRC] (received at account creation)",
					EnvVars: []string{"SLRC"},
				},
				&cli.StringFlag{
					Name:    "master-key",
					Value:   "",
					Usage:   "Master Recovery Key",
					EnvVars: []string{"MASTERKEY"},
				},
				&cli.StringFlag{
					Name:    "password",
					Value:   "",
					Usage:   "New Password",
					EnvVars: []string{"NEWPASSWD"},
				},
				&cli.BoolFlag{
					Name: "generate-password",
					Usage: "If this option provided, the metalo tool will generate a secure password and print it" +
						" out after successful password recovery",
				},
			},
		},
		{
			Name:  "identity",
			Usage: "commands for identity management",
			Subcommands: []*cli.Command{
				{
					Name:   "new",
					Usage:  "create new identity and add to the account",
					Action: NewIdentity,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "type",
							Value: "pairwise",
							Usage: "Identity type (default: pairwise). Allowed values: verinym, persona, pairwise",
						},
						&cli.StringFlag{
							Name:  "name",
							Value: "",
							Usage: "Full name",
						},
						&cli.BoolFlag{
							Name:  "unilocker",
							Usage: "add unilocker to the new identity",
						},
					},
				},
				{
					Name:   "import",
					Usage:  "import identity from file",
					Action: ImportIdentity,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "file",
							Value: "",
							Usage: "Identity file",
						},
					},
				},
			},
		},
		{
			Name:  "locker",
			Usage: "commands for locker management",
			Subcommands: []*cli.Command{
				{
					Name:   "new",
					Usage:  "create new locker definition",
					Action: CreateLocker,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "us",
							Value: "",
							Usage: "Our DID",
						},
						&cli.StringFlag{
							Name:  "them",
							Value: "",
							Usage: "Their DID",
						},
						&cli.StringFlag{
							Name:  "our-verkey",
							Value: "",
							Usage: "Our VerKey",
						},
						&cli.StringFlag{
							Name:  "their-verkey",
							Value: "",
							Usage: "Their VerKey",
						},
						&cli.StringFlag{
							Name:  "name",
							Value: "",
							Usage: "Locker name",
						},
						&cli.IntFlag{
							Name:  "ttl",
							Value: 12,
							Usage: "Number of months until the locker expires",
						},
						&cli.BoolFlag{
							Name:  "add",
							Usage: "If present, add the generated locker to the user's account",
						},
					},
				},
				{
					Name:   "import",
					Usage:  "import locker from file",
					Action: ImportLocker,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "file",
							Value: "",
							Usage: "Locker file",
						},
					},
				},
				{
					Name:   "ls",
					Usage:  "list all account lockers",
					Action: ListLockers,
				},
			},
		},
		{
			Name:  "prop",
			Usage: "commands for account property management",
			Subcommands: []*cli.Command{
				{
					Name:   "ls",
					Usage:  "list properties",
					Action: ListProperties,
				},
				{
					Name:   "get",
					Usage:  "get property",
					Action: GetProperty,
				},
				{
					Name:   "set",
					Usage:  "set property key and value",
					Action: SetProperty,
				},
				{
					Name:   "rm",
					Usage:  "delete property",
					Action: DeleteProperty,
				},
			},
		},
		{
			Name:  "access-key",
			Usage: "commands for access key management",
			Subcommands: []*cli.Command{
				{
					Name:   "new",
					Usage:  "generate new access key",
					Action: GenerateAccessKey,
				},
				{
					Name:   "ls",
					Usage:  "list access keys",
					Action: ListAccessKeys,
				},
				{
					Name:   "get",
					Usage:  "get access key",
					Action: GetAccessKey,
				},
				{
					Name:   "rm",
					Usage:  "delete access key",
					Action: DeleteAccessKey,
				},
			},
		},
		{
			Name:  "sub-account",
			Usage: "commands for sub-account management",
			Subcommands: []*cli.Command{
				{
					Name:   "new",
					Usage:  "create new sub-account",
					Action: CreateSubAccount,
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "new-key",
							Usage: "create new access key",
						},
						&cli.StringFlag{
							Name:  "name",
							Value: "",
							Usage: "Account name",
						},
						&cli.StringFlag{
							Name:  "sa-email",
							Value: "",
							Usage: "[optional] Email (for interactive logins)",
						},
						&cli.StringFlag{
							Name:    "sa-password",
							Value:   "",
							Usage:   "[optional] Password (for interactive logins)",
							EnvVars: []string{"SAPASS"},
						},
					},
				},
				{
					Name:   "ls",
					Usage:  "list sub-accounts",
					Action: ListSubAccounts,
				},
				{
					Name:   "get",
					Usage:  "get sub-account",
					Action: GetSubAccount,
				},
				{
					Name:   "rm",
					Usage:  "remove a sub-account",
					Action: DeleteSubAccount,
				},
				{
					Name:  "key",
					Usage: "commands for sub-account access key management",
					Subcommands: []*cli.Command{
						{
							Name:   "new",
							Usage:  "generate new access key",
							Action: GenerateSubAccountAccessKey,
						},
						{
							Name:   "ls",
							Usage:  "list access keys",
							Action: ListSubAccountAccessKeys,
						},
						{
							Name:   "get",
							Usage:  "get access key",
							Action: GetSubAccountAccessKey,
						},
						{
							Name:   "rm",
							Usage:  "delete access key",
							Action: DeleteSubAccountAccessKey,
						},
					},
				},
			},
		},
		{
			Name:  "dataset",
			Usage: "commands for dataset management",
			Subcommands: []*cli.Command{
				{
					Name:   "upload",
					Usage:  "upload dataset into MetaLocker",
					Action: StoreDataSet,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "vault",
							Value: "local",
							Usage: "Vault Name (default: local)",
						},
						&cli.StringFlag{
							Name:  "locker",
							Value: "",
							Usage: "Locker ID. If not specified, the default one will be used",
						},
						&cli.StringFlag{
							Name:  "prov",
							Value: "",
							Usage: "Provenance definition (JSON). If not specified, it will be auto-generated",
						},
						&cli.StringFlag{
							Name:  "prov-mapping",
							Value: "",
							Usage: "Entity mapping for provenance template. Optional. Format: ent->value;ent-->value;...",
						},
						&cli.StringFlag{
							Name:  "parent",
							Value: "",
							Usage: "Parent ledger record ID",
						},
						&cli.StringFlag{
							Name:  "type",
							Usage: "Defines semantic type of the provided data",
						},
						&cli.StringFlag{
							Name:  "expiration",
							Value: "1y",
							Usage: "Lease duration (i.e. 10y, 1y6m, 12d, 1h30min, 30s, never)",
						},
						&cli.BoolFlag{
							Name:  "wait",
							Usage: "If specified, wait until the data is published on the ledger",
						},
					},
				},
				{
					Name:   "bulk-upload",
					Usage:  "upload all datasets in the given folder into MetaLocker",
					Action: StoreDataSets,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "vault",
							Value: "local",
							Usage: "Vault Name (default: local)",
						},
						&cli.StringFlag{
							Name:  "locker",
							Value: "",
							Usage: "Locker ID. If not specified, the default one will be used",
						},
						&cli.StringFlag{
							Name:  "prov",
							Value: "",
							Usage: "Provenance definition (JSON). If not specified, it will be auto-generated",
						},
						&cli.StringFlag{
							Name:  "type",
							Usage: "Defines semantic type of the provided data",
						},
						&cli.StringFlag{
							Name:  "expiration",
							Value: "1y",
							Usage: "Lease duration (i.e. 10y, 1y6m, 12d, 1h30min, 30s, never)",
						},
						&cli.BoolFlag{
							Name:  "wait",
							Usage: "If specified, wait until the data is published on the ledger",
						},
					},
				},
				{
					Name:   "supported-data-types",
					Usage:  "list supported data types",
					Action: ListSupportedDataTypes,
				},
				{
					Name:   "share",
					Usage:  "share data set",
					Action: ShareDataSet,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "vault",
							Value: "local",
							Usage: "Vault Name (default: local)",
						},
						&cli.StringFlag{
							Name:  "locker",
							Value: "",
							Usage: "Locker ID. If not specified, the default one will be used",
						},
						&cli.StringFlag{
							Name:  "expiration",
							Value: "1y",
							Usage: "Lease duration (i.e. 10y, 1y6m, 12d, 1h30min, 30s, never)",
						},
						&cli.BoolFlag{
							Name:  "wait",
							Usage: "If specified, wait until the data is published on the ledger",
						},
					},
				},
				{
					Name:   "get",
					Usage:  "get data set from MetaLocker",
					Action: GetDataSet,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "dest",
							Value: "",
							Usage: "path to destination folder/file. Not required for graph data sets (will be printed out on the console)",
						},
						&cli.BoolFlag{
							Name:  "metadata",
							Usage: "save metadata artifacts into the folder",
						},
						&cli.BoolFlag{
							Name:  "sync",
							Usage: "sync data wallet before reading the data set",
						},
					},
				},
				{
					Name:   "revoke",
					Usage:  "revoke data set's lease",
					Action: RevokeLease,
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "wait",
							Usage: "If specified, wait until the data is published on the ledger",
						},
					},
				},
			},
		},
		{
			Name:  "wallet",
			Usage: "commands for wallet management",
			Subcommands: []*cli.Command{
				{
					Name:   "ls",
					Usage:  "list wallet records",
					Action: ListRecords,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "locker",
							Value: "",
							Usage: "Locker ID. If not specified, records from all lockers will be displayed",
						},
						&cli.Uint64Flag{
							Name:  "max-records",
							Value: 20,
							Usage: "Maximum number of records to be displayed",
						},
						&cli.BoolFlag{
							Name:  "sync",
							Usage: "sync data wallet before displaying its contents",
						},
						&cli.BoolFlag{
							Name:  "include-revocations",
							Usage: "include revoked leases",
						},
					},
				},
				{
					Name:   "sync",
					Usage:  "sync local data wallet with the ledger",
					Action: SyncIndex,
				},
				{
					Name:   "export",
					Usage:  "export data wallet",
					Action: ExportWallet,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "locker",
							Value: "",
							Usage: "Locker ID. If specified, only the given locker will be exported",
						},
						&cli.StringFlag{
							Name:  "participant",
							Value: "",
							Usage: "Participant ID. If specified, only the given locker participant will be exported",
						},
						&cli.StringFlag{
							Name:  "mode",
							Value: "full",
							Usage: "'user' or 'full' (default). 'User' mode skips internal metadata and prettifies JSON for graphs",
						},
						&cli.BoolFlag{
							Name:  "force-rewrite",
							Usage: "if true, export records even when they already exist on disk",
						},
					},
				},
				{
					Name:   "purge-deleted-assets",
					Usage:  "purge deleted data assets",
					Action: PurgeDeletedDataAssets,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "locker",
							Value: "",
							Usage: "Locker ID. If not specified, records from all lockers will be processed",
						},
						&cli.Uint64Flag{
							Name:  "max-records",
							Value: 20,
							Usage: "Maximum number of records to be processed",
						},
						&cli.BoolFlag{
							Name:  "sync",
							Usage: "sync data wallet before performing the operation",
						},
					},
				},
			},
		},
		{
			Name:   "new-asset",
			Usage:  "generate new asset. If file path provided as a parameter, generate a digital asset definition",
			Action: NewAsset,
		},
		{
			Name:   "sample-crypto",
			Usage:  "generate sample crypto material for MetaLocker configuration",
			Action: SampleCrypto,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "name",
					Value: "",
					Usage: "(optional) sample identity name",
				},
				&cli.BoolFlag{
					Name:  "skip-locker",
					Usage: "(optional) don't create a unilocker for the sample identity",
				},
			},
		},
	}

	AdminSet = []*cli.Command{
		{
			Name:   "export-ledger",
			Usage:  "export MetaLocker ledger into the given directory",
			Action: ExportLedger,
		},
		{
			Name:   "import-ledger",
			Usage:  "import MetaLocker ledger data from the given directory",
			Action: ImportLedger,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "import-operations",
					Usage: "import operations (be careful if swapping ledgers!)",
				},
				&cli.BoolFlag{
					Name:  "wait",
					Usage: "If specified, wait until each block is published on the ledger",
				},
			},
		},
		{
			Name:   "export-accounts",
			Usage:  "export MetaLocker accounts into the given directory",
			Action: ExportAccounts,
		},
		{
			Name:   "import-accounts",
			Usage:  "import MetaLocker accounts from the given directory",
			Action: ImportAccounts,
		},
		{
			Name:   "account-state",
			Usage:  "update account state for the given ID (email or DID)",
			Action: UpdateAccountState,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "lock",
					Usage: "lock account",
				},
				&cli.BoolFlag{
					Name:  "unlock",
					Usage: "unlock account",
				},
				&cli.BoolFlag{
					Name:  "delete",
					Usage: "mark account as deleted",
				},
			},
		},
	}

	StandardSet []*cli.Command
)

func init() {
	StandardSet = append(AccountSet, AdminSet...)
}
