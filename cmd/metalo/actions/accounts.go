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

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/olekukonko/tablewriter"
	"github.com/piprate/json-gold/ld"
	"github.com/piprate/metalocker/cmd/metalo/operations"
	"github.com/piprate/metalocker/index"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/remote"
	"github.com/piprate/metalocker/storage"
	"github.com/piprate/metalocker/utils/fingerprint"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/wallet"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func createPersona(name, idType string, firstBlock int64) (*account.Identity, error) {
	did, err := model.GenerateDID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	expiryTime := now.AddDate(0, 120, 0).UTC()

	// add identity locker

	locker, err := model.GenerateLocker(model.AccessLevelHosted, name, &expiryTime, firstBlock, model.Us(did, nil))
	if err != nil {
		return nil, err
	}

	return &account.Identity{
		DID:         did,
		Created:     &now,
		Name:        name,
		Type:        idType,
		AccessLevel: model.AccessLevelHosted,
		Lockers:     []*model.Locker{locker},
	}, nil
}

func RegisterCommand(c *cli.Context) error {
	issuesFound := false
	for _, param := range []string{"email", "name"} {
		if c.String(param) == "" {
			_, _ = fmt.Fprintf(os.Stderr, "Please specify %s parameter.\n", param)
			issuesFound = true
		}
	}

	if issuesFound {
		return cli.Exit("invalid parameters", InvalidParameter)
	}

	passwd := ReadCredential(c.String("password"), "Enter password: ", true)

	if passwd == "" {
		return cli.Exit("empty password not allowed", InvalidParameter)
	}

	url := c.String("server")

	factory, err := remote.NewWalletFactory(url, WithPersonalIndexStore(), 0)
	if err != nil {
		log.Err(err).Msg("Registration failed")
		return cli.Exit(err, OperationFailed)
	}

	topBlock, err := factory.GetTopBlock()
	if err != nil {
		log.Err(err).Msg("Failed to read top block")
		return cli.Exit(err, OperationFailed)
	}

	var idy *account.Identity
	identityFileName := c.String("identity-file")
	name := c.String("name")

	if identityFileName != "" {
		var err error
		identityFile, err := os.ReadFile(identityFileName)
		if err != nil {
			log.Err(err).Str("file", identityFileName).Msg("Failed to read identity file")
			return cli.Exit(err, InvalidParameter)
		}
		var idyStruct account.Identity
		err = jsonw.Unmarshal(identityFile, &idyStruct)
		if err != nil {
			return err
		}
		idy = &idyStruct
	} else {
		idy, err = createPersona(name, account.IdentityTypePersona, topBlock)
		if err != nil {
			return err
		}
	}

	email := c.String("email")

	registrationCode := c.String("code")

	dw, recDetails, err := factory.RegisterAccount(
		c.Context,
		&account.Account{
			Email:       email,
			Name:        name,
			AccessLevel: model.AccessLevelHosted,
		},
		passwd,
		account.WithRegistrationCode(registrationCode))
	if err != nil {
		log.Err(err).Msg("Registration failed")
		return cli.Exit(err, OperationFailed)
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "\n\tYour secret recovery phrase is: %s\n\n", recDetails.RecoveryPhrase)
	}

	if err = dw.Unlock(c.Context, passwd); err != nil {
		log.Err(err).Msg("Failed to unlock new wallet")
		return cli.Exit(err, OperationFailed)
	}

	if err = dw.AddIdentity(c.Context, idy); err != nil {
		log.Err(err).Msg("Failed to add an identity to account")
		return cli.Exit(err, OperationFailed)
	} else {
		fmt.Printf("%s,%s\n", idy.DID.ID, idy.DID.VerKey)
	}
	return nil
}

func ImportIdentity(c *cli.Context) error {

	identityFileName := c.String("file")
	if identityFileName == "" {
		return cli.Exit("invalid parameters", InvalidParameter)
	}

	identityFile, err := os.ReadFile(identityFileName)
	if err != nil {
		log.Err(err).Str("file", identityFileName).Msg("Failed to read identity file")
		return cli.Exit(err, InvalidParameter)
	}

	var idy account.Identity
	err = jsonw.Unmarshal(identityFile, &idy)
	if err != nil {
		return err
	}

	dw, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	err = dw.AddIdentity(c.Context, &idy)
	if err != nil {
		log.Err(err).Msg("Identity import failed")
		return cli.Exit(err, OperationFailed)
	} else {
		fmt.Printf("%s,%s\n", idy.DID.ID, idy.DID.VerKey)
	}

	return nil
}

func NewIdentity(c *cli.Context) error {

	idType := c.String("type")
	if idType == "" {
		return cli.Exit("please specify --type parameter", InvalidParameter)
	}

	did, err := model.GenerateDID()
	if err != nil {
		return cli.Exit(err, OperationFailed)
	}

	name := c.String("name")
	if name == "" {
		name = did.ID
	}

	canonicalIdentityType, found := map[string]string{
		"verinym":  account.IdentityTypeVerinym,
		"persona":  account.IdentityTypePersona,
		"pairwise": account.IdentityTypePairwise,
	}[idType]
	if !found {
		return cli.Exit(fmt.Errorf("unsupported value --type parameter: %s", idType), InvalidParameter)
	}

	dataWallet, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	now := time.Now()

	idy := &account.Identity{
		DID:         did,
		Created:     &now,
		Name:        name,
		Type:        canonicalIdentityType,
		AccessLevel: model.AccessLevelHosted,
	}

	if c.Bool("unilocker") {
		tb, err := dataWallet.Services().Ledger().GetTopBlock()
		if err != nil {
			return cli.Exit(err, OperationFailed)
		}

		expiryTime := time.Now().AddDate(0, 120, 0).UTC()

		// add identity locker

		locker, err := model.GenerateLocker(model.AccessLevelHosted, name, &expiryTime, tb.Number, model.Us(did, nil))
		if err != nil {
			return cli.Exit(err, OperationFailed)
		}
		idy.Lockers = append(idy.Lockers, locker)
	}

	err = dataWallet.AddIdentity(c.Context, idy)
	if err != nil {
		log.Err(err).Msg("Adding new identity failed")
		return cli.Exit(err, OperationFailed)
	} else {
		log.Debug().Str("did", idy.ID()).Msg("Created new identity")

		fmt.Printf("%s,%s\n", idy.DID.ID, idy.DID.VerKey)
	}

	return err
}

func ImportLocker(c *cli.Context) error {
	lockerFileName := c.String("file")
	if lockerFileName == "" {
		return cli.Exit("invalid parameters", InvalidParameter)
	}

	lockerBytes, err := os.ReadFile(lockerFileName)
	if err != nil {
		log.Err(err).Str("file", lockerFileName).Msg("Failed to read locker file")
		return cli.Exit(err, InvalidParameter)
	}

	var locker model.Locker
	err = jsonw.Unmarshal(lockerBytes, &locker)
	if err != nil {
		return err
	}

	dataWallet, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	var us string
	for _, p := range locker.Participants {
		_, err := dataWallet.GetIdentity(c.Context, p.ID)
		if err != nil {
			if errors.Is(err, storage.ErrIdentityNotFound) {
				continue
			} else {
				return cli.Exit(err, OperationFailed)
			}
		}
		us = p.ID
		break
	}

	_, err = dataWallet.AddLocker(c.Context, locker.Perspective(us))
	if err != nil {
		log.Err(err).Msg("Locker import failed")
		return cli.Exit(err, OperationFailed)
	}
	return err
}

func CreateLocker(c *cli.Context) error {

	var err error

	issuesFound := false
	for _, param := range []string{"us", "our-verkey", "them", "their-verkey", "name"} {
		if c.String(param) == "" {
			_, _ = fmt.Fprintf(os.Stderr, "Please specify %s parameter.\n", param)
			issuesFound = true
		}
	}
	if issuesFound {
		return cli.Exit("invalid parameter(s)", InvalidParameter)
	}

	did1 := c.String("us")
	verKey1 := c.String("our-verkey")
	did2 := c.String("them")
	verKey2 := c.String("their-verkey")
	name := c.String("name")
	months := c.Int("ttl")
	addToAccount := c.Bool("add")

	mlc := CreateHTTPCaller(c)

	controls, err := mlc.GetServerControls()
	if err != nil {
		return err
	}

	id1 := model.NewDID(did1, verKey1, "")
	id2 := model.NewDID(did2, verKey2, "")

	expiryTime := time.Now().AddDate(0, months, 0).UTC()
	locker, err := model.GenerateLocker(model.AccessLevelHosted, name, &expiryTime, controls.TopBlock,
		model.Us(id1, nil),
		model.Them(id2, nil))
	if err != nil {
		return err
	}

	log.Debug().Str("lid", locker.ID).Msg("Created new locker")

	if addToAccount {
		dataWallet, err := LoadRemoteDataWallet(c, false)
		if err != nil {
			return err
		}

		wrapper, err := dataWallet.AddLocker(c.Context, locker)
		if err != nil {
			log.Err(err).Msg("Locker import failed")
			return cli.Exit(err, OperationFailed)
		}
		locker = wrapper.Raw()
	}

	ld.PrintDocument("", locker)

	return nil
}

func NewAsset(c *cli.Context) error {

	filePath := ""
	if c.Args().Len() == 1 {
		filePath = c.Args().Get(0)
	}

	var id string
	var err error
	if filePath != "" {
		id, _, err = model.BuildDigitalAssetIDFromFile(filePath, fingerprint.AlgoSha256, "")
		if err != nil {
			return err
		}
	} else {
		id = model.NewAssetID("")
	}

	println(id)

	return nil
}

func PrintAccount(c *cli.Context) error {
	dataWallet, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	detailed := c.Bool("detailed")
	acct := dataWallet.Account()

	res := map[string]any{
		"account": acct,
	}

	if detailed {

		// add all identities

		idyMap, err := dataWallet.GetIdentities(c.Context)
		if err != nil {
			return err
		}
		idyList := make([]*account.Identity, 0, len(idyMap))
		for _, i := range idyMap {
			idyList = append(idyList, i.Raw())
		}

		res["identities"] = idyList

		// add all lockers

		lockers, err := dataWallet.GetLockers(c.Context)
		if err != nil {
			return err
		}

		res["lockers"] = lockers

		acct.HostedSecretStore = nil
		acct.ManagedSecretStore = nil

		// add all properties

		props, err := dataWallet.GetProperties(c.Context)
		if err != nil {
			return err
		}

		res["properties"] = props
	}

	ld.PrintDocument("", res)

	return err
}

func PrintAccountChart(c *cli.Context) error {
	dw, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	return operations.PrintWallet(c.Context, dw, "")
}

func ChangePassphrase(c *cli.Context) error {
	dataWallet, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	currentPass := ReadCredential(c.String("password"), "Enter current password: ", true)
	if currentPass == "" {
		return cli.Exit("empty current password", OperationFailed)
	}

	newPass := ReadCredential(c.String("new-password"), "Enter new password: ", true)
	if newPass == "" {
		return cli.Exit("empty new password", OperationFailed)
	}

	_, err = dataWallet.ChangePassphrase(c.Context, currentPass, newPass, false)

	if err != nil {
		log.Err(err).Msg("Passphrase change failed")
		return cli.Exit(err, OperationFailed)
	}
	return err
}

func ChangeEmail(c *cli.Context) error {
	dataWallet, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	newEmail := ReadCredential(c.String("email"), "Enter new account email: ", false)

	err = dataWallet.ChangeEmail(c.Context, newEmail)
	if err != nil {
		log.Err(err).Msg("Account email change failed")
		return cli.Exit(err, OperationFailed)
	}
	return err
}

func RecoverAccount(c *cli.Context) error {
	userID := ReadCredential(c.String("user"), "Enter account email: ", false)

	mlc := CreateHTTPCaller(c)

	recoveryPhrase := ReadCredential(c.String("rec-phrase"), "Enter recovery phrase: ", true)
	newPassphrase := ReadCredential(c.String("new-password"), "Enter new account password: ", true)

	recoveryCode, err := mlc.GetAccountRecoveryCode(c.Context, userID)
	if err != nil {
		return fmt.Errorf("failed to get recovery code from MetaLocker: %w", err)
	}

	cryptoKey, _, privKey, err := account.GenerateKeysFromRecoveryPhrase(recoveryPhrase)
	if err != nil {
		log.Err(err).Msg("Error generating keys from recovery phrase")
		return err
	}

	acct, err := mlc.RecoverAccount(userID, privKey, recoveryCode, newPassphrase)
	if err != nil {
		log.Err(err).Msg("Failed to recover account")
		return fmt.Errorf("failed to recover account: %w", err)
	}

	err = mlc.LoginWithCredentials(userID, newPassphrase)
	if err != nil {
		return err
	}

	indexClient, err := WithPersonalIndexStore()(acct.ID, mlc)
	if err != nil {
		return err
	}

	dataWallet, err := wallet.NewLocalDataWallet(acct, mlc, nil, indexClient)
	if err != nil {
		log.Err(err).Str("user", userID).Msg("Data wallet creation failed")
		return errors.New("failed to process account")
	}

	np, err := dataWallet.Recover(c.Context, cryptoKey, newPassphrase)
	if err != nil {
		return err
	}

	err = mlc.UpdateAccount(c.Context, np.Account())
	if err != nil {
		log.Err(err).Msg("Account recovery failed")
		return cli.Exit(err, OperationFailed)
	}
	return nil
}

func RecoverAccountSecondLevel(c *cli.Context) error {
	userID := ReadCredential(c.String("user"), "Enter account email: ", false)

	generate := c.Bool("generate-password")

	mlc := CreateHTTPCaller(c)

	slrc := ReadCredential(c.String("slrc"), "Enter second level recovery code: ", true)

	// administrator enters the master private key
	masterPrivateKeyStr := ReadCredential(c.String("master-key"), "Enter master recovery key: ", true)
	masterPrivateKeyVal := base58.Decode(masterPrivateKeyStr)
	masterPrivateKey := ed25519.PrivateKey(masterPrivateKeyVal)

	var newPassphrase string
	if generate {
		newPassphrase = base58.Encode(model.NewEncryptionKey()[:])
	} else {
		newPassphrase = ReadCredential(c.String("password"), "Enter new account passphrase: ", true)
	}

	recoveryCode, err := mlc.GetAccountRecoveryCode(c.Context, userID)
	if err != nil {
		return fmt.Errorf("failed to get recovery code from MetaLocker: %w", err)
	}

	slrcBytes := base58.Decode(slrc)
	pkBytes, err := model.AnonDecrypt(slrcBytes, masterPrivateKey)
	if err != nil {
		log.Err(err).Msg("Failed to decrypt SLRC")
		return fmt.Errorf("failed to decrypt SLRC: %w", err)
	}

	privKey := ed25519.PrivateKey(pkBytes)

	acct, err := mlc.RecoverAccount(userID, privKey, recoveryCode, newPassphrase)
	if err != nil {
		log.Err(err).Msg("Failed to recover account")
		return fmt.Errorf("failed to recover account: %w", err)
	}

	if acct.AccessLevel != model.AccessLevelManaged {
		return fmt.Errorf("only managed accounts can be recovered using SLRC")
	}

	if acct.Version < 4 || acct.EncryptedRecoverySecret == "" {
		return fmt.Errorf("this account's version doesn't support SLRC recovery")
	}

	encryptedKeyBytes, err := base64.StdEncoding.DecodeString(acct.EncryptedRecoverySecret)
	if err != nil {
		return err
	}
	cryptoKeyBytes, err := model.AnonDecrypt(encryptedKeyBytes, privKey)
	if err != nil {
		return err
	}

	err = mlc.LoginWithCredentials(userID, newPassphrase)
	if err != nil {
		return err
	}

	indexClient, err := WithPersonalIndexStore()(acct.ID, mlc)
	if err != nil {
		return err
	}

	dataWallet, err := wallet.NewLocalDataWallet(acct, mlc, nil, indexClient)
	if err != nil {
		log.Err(err).Str("user", userID).Msg("Data Wallet creation failed")
		return errors.New("failed to process account")
	}

	np, err := dataWallet.Recover(c.Context, model.NewAESKey(cryptoKeyBytes), newPassphrase)
	if err != nil {
		return err
	}

	err = mlc.UpdateAccount(c.Context, np.Account())
	if err != nil {
		log.Err(err).Msg("Account recovery failed")
		return cli.Exit(err, OperationFailed)
	} else {
		fmt.Printf("New password: %s\n", newPassphrase)
	}
	return nil
}

func DeleteAccount(c *cli.Context) error {
	user := ReadCredential(c.String("user"), "Enter account email: ", false)
	password := ReadCredential(c.String("password"), "Enter password: ", true)

	mlc := CreateHTTPCaller(c)

	err := mlc.LoginWithCredentials(user, password)
	if err != nil {
		log.Err(err).Str("user", user).Msg("Authentication failed")
		return cli.Exit(err, AuthenticationFailed)
	}

	return mlc.DeleteAccount(c.Context, mlc.AuthenticatedAccountID())
}

func ListLockers(c *cli.Context) error {
	dataWallet, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	lockerIDs := make([]string, 0)
	lockers, err := dataWallet.GetLockers(c.Context)
	if err != nil {
		log.Err(err).Msg("Failed to read locker list")
		return cli.Exit(err, OperationFailed)
	}
	for _, l := range lockers {
		lockerIDs = append(lockerIDs, l.ID)
	}
	sort.Strings(lockerIDs)

	tf := "2006-01-02 15:04:05-07:00"
	data := make([][]string, 0)
	for _, lid := range lockerIDs {
		lw, err := dataWallet.GetLocker(c.Context, lid)
		if err != nil {
			return err
		}

		l := lw.Raw()

		// format participants
		var p1, p2 string
		if len(l.Participants) == 1 {
			p1 = l.Participants[0].ID
			p2 = "Self"
		} else {
			p1 = l.Participants[0].ID
			p2 = l.Participants[1].ID
			if p1 == p2 {
				p2 = "Self"
			}
		}
		var createdStr, expiresStr string
		if l.Created != nil {
			createdStr = l.Created.Format(tf)
		}
		if l.Expires != nil {
			expiresStr = l.Expires.Format(tf)
		}
		data = append(data, []string{l.ID, l.Name, p1, p2, createdStr, expiresStr})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Identity1", "Identity2", "Created", "Expires"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data) // Add Bulk Data
	table.Render()

	return nil
}

func PurgeDeletedDataAssets(c *cli.Context) error {
	locker := c.String("locker")
	maxRecords := c.Uint64("max-records")
	syncIndex := c.Bool("sync")

	dw, err := LoadRemoteDataWallet(c, syncIndex)
	if err != nil {
		return err
	}

	rootIndex, err := dw.RootIndex(c.Context)
	if err != nil {
		return err
	}

	err = rootIndex.TraverseRecords(locker, "", func(r *index.RecordState) error {
		if r.Status == model.StatusRevoked {
			if err = dw.DataStore().PurgeDataAssets(c.Context, r.ID); err != nil {
				if errors.Is(err, model.ErrBlobNotFound) {
					log.Debug().Str("rid", r.ID).Msg("Data assets already purged")
				} else {
					return err
				}
			}
		}
		return nil
	}, maxRecords)
	if err != nil {
		log.Err(err).Msg("Data wallet sync")
		return cli.Exit(err, OperationFailed)
	}

	return nil
}

func ExportWallet(c *cli.Context) error {
	if c.Args().Len() != 1 {
		fmt.Print("Please specify the path to file or folder.\n\n")
		return cli.Exit("please specify the path to file or folder", InvalidParameter)
	}

	lockerID := c.String("locker")
	participantID := c.String("participant")
	userFriendly := c.String("mode") == "user"
	forceRewrite := c.Bool("force-rewrite")

	dw, err := LoadRemoteDataWallet(c, true)
	if err != nil {
		return err
	}

	err = operations.ExportWallet(c.Context, dw, c.Args().Get(0), lockerID, participantID, userFriendly, forceRewrite)
	if err != nil {
		log.Err(err).Msg("Wallet export failed")
		return cli.Exit(err, OperationFailed)
	}
	return err
}

func ListRecords(c *cli.Context) error {
	locker := c.String("locker")
	maxRecords := c.Uint64("max-records")
	syncIndex := c.Bool("sync")
	includeRevokedLeases := c.Bool("include-revocations")

	dw, err := LoadRemoteDataWallet(c, syncIndex)
	if err != nil {
		return err
	}

	rootIndex, err := dw.RootIndex(c.Context)
	if err != nil {
		return err
	}

	data := make([][]string, 0)

	err = rootIndex.TraverseRecords(locker, "", func(r *index.RecordState) error {
		if r.Status == model.StatusPublished || includeRevokedLeases {
			data = append(data, []string{r.LockerID, r.ParticipantID, r.ID, strconv.Itoa(int(r.Operation)), r.ImpressionID, r.ContentType, string(r.Status)})
		}
		return nil
	}, maxRecords)
	if err != nil {
		log.Err(err).Msg("Data wallet sync")
		return cli.Exit(err, OperationFailed)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Locker", "Participant", "ID", "Type", "Impression", "Content Type", "Status"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data) // Add Bulk Data
	table.Render()

	return nil
}

func SyncIndex(c *cli.Context) error {
	_, err := LoadRemoteDataWallet(c, true)
	return err
}
