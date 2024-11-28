package rdb_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/piprate/metalocker/storage"
	. "github.com/piprate/metalocker/storage/rdb"
	"github.com/piprate/metalocker/storage/rdb/ent"
	"github.com/piprate/metalocker/storage/rdb/ent/enttest"
	"github.com/piprate/metalocker/storage/rdb/ent/migrate"
	"github.com/piprate/metalocker/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Stamp})
}

func Test_RelationalBackend_CreateAccount(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	testAcct := createAccount(t, "test@example.com", "Test User")

	// happy path

	err := be.CreateAccount(ctx, testAcct)
	require.NoError(t, err)

	acct, err := be.GetAccount(ctx, testAcct.ID)
	require.NoError(t, err)
	assert.Equal(t, testAcct.ID, acct.ID)

	// try creating the same account again

	err = be.CreateAccount(ctx, testAcct)
	require.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrAccountExists))

	// try creating an account with the same email

	testAcct2 := createAccount(t, "test@example.com", "Test User 2")

	err = be.CreateAccount(ctx, testAcct2)
	require.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrAccountExists))

	testChild1 := createAccount(t, "", "Test Child 1")
	testChild1.ParentAccount = acct.ID
	testChild2 := createAccount(t, "", "Test Child 2")
	testChild2.ParentAccount = acct.ID

	err = be.CreateAccount(ctx, testChild1)
	require.NoError(t, err)

	err = be.CreateAccount(ctx, testChild2)
	require.NoError(t, err)

	_ = client.Close()
	err = be.CreateAccount(ctx, testAcct)
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_GetAccount(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	testAcct := createAccount(t, "test@example.com", "Test User")

	err := be.CreateAccount(ctx, testAcct)
	require.NoError(t, err)

	acct, err := be.GetAccount(ctx, testAcct.ID)
	require.NoError(t, err)
	assert.Equal(t, testAcct.ID, acct.ID)

	acct, err = be.GetAccount(ctx, testAcct.Email)
	require.NoError(t, err)
	assert.Equal(t, testAcct.ID, acct.ID)

	_ = client.Close()
	_, err = be.GetAccount(ctx, testAcct.Email)
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_UpdateAccount(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	testAcct := createAccount(t, "test@example.com", "Test User")

	err := be.CreateAccount(ctx, testAcct)
	require.NoError(t, err)

	// fail if account doesn't exist

	err = be.UpdateAccount(ctx, &account.Account{
		ID: "bad-id",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrAccountNotFound))

	// happy path

	updatedAcct := testAcct
	updatedAcct.Name = "Test User #2"
	err = be.UpdateAccount(ctx, updatedAcct)
	require.NoError(t, err)

	acct, err := be.GetAccount(ctx, updatedAcct.Email)
	require.NoError(t, err)
	assert.Equal(t, "Test User #2", acct.Name)

	_ = client.Close()
	err = be.UpdateAccount(ctx, updatedAcct)
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_DeleteAccount(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	testAcct := createAccount(t, "test@example.com", "Test User")

	err := be.CreateAccount(ctx, testAcct)
	require.NoError(t, err)

	acct, err := be.GetAccount(ctx, testAcct.ID)
	require.NoError(t, err)
	assert.Equal(t, testAcct.ID, acct.ID)

	// try deleting non-existent account

	err = be.DeleteAccount(ctx, "wrong-id")
	require.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrAccountNotFound))

	// try deleting a real account

	err = be.DeleteAccount(ctx, testAcct.ID)
	require.NoError(t, err)

	_, err = be.GetAccount(ctx, testAcct.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrAccountNotFound))

	_ = client.Close()
	err = be.DeleteAccount(ctx, testAcct.ID)
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_ListAccounts(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	testAcct := createAccount(t, "test@example.com", "Test User")

	err := be.CreateAccount(ctx, testAcct)
	require.NoError(t, err)

	testAcct2 := createAccount(t, "test2@example.com", "Test User")
	testAcct2.State = account.StateSuspended
	testAcct2.ParentAccount = testAcct.ID

	err = be.CreateAccount(ctx, testAcct2)
	require.NoError(t, err)

	list, err := be.ListAccounts(ctx, "", "")
	require.NoError(t, err)
	assert.Equal(t, 2, len(list))

	list, err = be.ListAccounts(ctx, testAcct.ID, "")
	require.NoError(t, err)
	assert.Equal(t, 1, len(list))

	list, err = be.ListAccounts(ctx, testAcct.ID, account.StateSuspended)
	require.NoError(t, err)
	assert.Equal(t, 1, len(list))

	list, err = be.ListAccounts(ctx, testAcct.ID, account.StateDeleted)
	require.NoError(t, err)
	assert.Equal(t, 0, len(list))

	list, err = be.ListAccounts(ctx, "", account.StateSuspended)
	require.NoError(t, err)
	assert.Equal(t, 1, len(list))

	list, err = be.ListAccounts(ctx, "", account.StateDeleted)
	require.NoError(t, err)
	assert.Equal(t, 0, len(list))

	_ = client.Close()
	_, err = be.ListAccounts(ctx, "", "")
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_HasAccountAccess(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	testAcct := createAccount(t, "test@example.com", "Test User")
	err := be.CreateAccount(ctx, testAcct)
	require.NoError(t, err)

	testAcct2 := createAccount(t, "test2@example.com", "Test User 2")
	testAcct2.ParentAccount = testAcct.ID
	err = be.CreateAccount(ctx, testAcct2)
	require.NoError(t, err)

	testAcct3 := createAccount(t, "test3@example.com", "Test User 3")
	testAcct3.ParentAccount = testAcct.ID
	err = be.CreateAccount(ctx, testAcct3)
	require.NoError(t, err)

	res, err := be.HasAccountAccess(ctx, testAcct.ID, testAcct3.ID)
	require.NoError(t, err)
	assert.True(t, res)

	res, err = be.HasAccountAccess(ctx, testAcct3.ID, testAcct2.ID)
	require.NoError(t, err)
	assert.False(t, res)

	_, err = be.HasAccountAccess(ctx, testAcct.ID, "bad-account")
	require.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrAccountNotFound))

	res, err = be.HasAccountAccess(ctx, "another-account", testAcct3.ID)
	require.NoError(t, err)
	assert.False(t, res)

	_ = client.Close()
	_, err = be.HasAccountAccess(ctx, testAcct3.ID, testAcct2.ID)
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_StoreIdentity(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	acctTemplate := &account.Account{
		Email:        "test@example.com",
		Name:         "Test User",
		AccessLevel:  model.AccessLevelHosted,
		DefaultVault: testbase.TestVaultName,
	}
	resp, err := account.GenerateAccount(acctTemplate, account.WithPassphraseAuth("pass123"))
	require.NoError(t, err)

	err = be.CreateAccount(ctx, resp.Account)
	require.NoError(t, err)

	// happy path

	for _, envelope := range resp.EncryptedIdentities {
		err = be.StoreIdentity(ctx, resp.Account.ID, envelope)
		require.NoError(t, err)
	}

	// fail if account doesn't exist

	err = be.StoreIdentity(ctx, "bad-account", resp.EncryptedIdentities[0])
	require.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrAccountNotFound))

	_ = client.Close()
	err = be.StoreIdentity(ctx, resp.Account.ID, resp.EncryptedIdentities[0])
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_GetIdentity(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	acctTemplate := &account.Account{
		Email:        "test@example.com",
		Name:         "Test User",
		AccessLevel:  model.AccessLevelHosted,
		DefaultVault: testbase.TestVaultName,
	}
	resp, err := account.GenerateAccount(acctTemplate, account.WithPassphraseAuth("pass123"))
	require.NoError(t, err)

	err = be.CreateAccount(ctx, resp.Account)
	require.NoError(t, err)

	for _, envelope := range resp.EncryptedIdentities {
		err = be.StoreIdentity(ctx, resp.Account.ID, envelope)
		require.NoError(t, err)
	}

	// happy path

	envelope, err := be.GetIdentity(ctx, resp.Account.ID, resp.EncryptedIdentities[0].Hash)
	require.NoError(t, err)
	assert.Equal(t, resp.EncryptedIdentities[0].EncryptedID, envelope.EncryptedID)

	// fail if account doesn't exist

	_, err = be.GetIdentity(ctx, "bad-account", resp.EncryptedIdentities[0].Hash)
	require.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrIdentityNotFound))

	_ = client.Close()
	_, err = be.GetIdentity(ctx, resp.Account.ID, resp.EncryptedIdentities[0].Hash)
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_ListIdentities(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	acctTemplate := &account.Account{
		Email:        "test@example.com",
		Name:         "Test User",
		AccessLevel:  model.AccessLevelHosted,
		DefaultVault: testbase.TestVaultName,
	}
	resp, err := account.GenerateAccount(acctTemplate, account.WithPassphraseAuth("pass123"))
	require.NoError(t, err)

	err = be.CreateAccount(ctx, resp.Account)
	require.NoError(t, err)

	for _, envelope := range resp.EncryptedIdentities {
		err = be.StoreIdentity(ctx, resp.Account.ID, envelope)
		require.NoError(t, err)
	}

	// happy paths

	list, err := be.ListIdentities(ctx, resp.Account.ID, model.AccessLevelManaged)
	require.NoError(t, err)
	assert.Equal(t, 1, len(list))

	list, err = be.ListIdentities(ctx, resp.Account.ID, model.AccessLevelHosted)
	require.NoError(t, err)
	assert.Equal(t, 0, len(list))

	_ = client.Close()
	_, err = be.ListIdentities(ctx, resp.Account.ID, model.AccessLevelHosted)
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_CreateDIDDocument(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	acctTemplate := &account.Account{
		Email:        "test@example.com",
		Name:         "Test User",
		AccessLevel:  model.AccessLevelHosted,
		DefaultVault: testbase.TestVaultName,
	}
	resp, err := account.GenerateAccount(acctTemplate, account.WithPassphraseAuth("pass123"))
	require.NoError(t, err)

	err = be.CreateAccount(ctx, resp.Account)
	require.NoError(t, err)

	// happy path

	var dDoc *model.DIDDocument
	for _, e := range resp.RootIdentities {
		dDoc, err = model.SimpleDIDDocument(e.DID, e.Created)
		require.NoError(t, err)

		err = be.CreateDIDDocument(ctx, dDoc)
		require.NoError(t, err)
	}

	// try saving the same DID doc again, should succeed.

	err = be.CreateDIDDocument(ctx, dDoc)
	require.NoError(t, err)

	// try saving a DID doc with a different proof, should fail.

	dDoc.Proof.Value = "another-proof"

	err = be.CreateDIDDocument(ctx, dDoc)
	require.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrDIDExists))

	_ = client.Close()
	err = be.CreateDIDDocument(ctx, dDoc)
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_GetDIDDocument(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	acctTemplate := &account.Account{
		Email:        "test@example.com",
		Name:         "Test User",
		AccessLevel:  model.AccessLevelHosted,
		DefaultVault: testbase.TestVaultName,
	}
	resp, err := account.GenerateAccount(acctTemplate, account.WithPassphraseAuth("pass123"))
	require.NoError(t, err)

	err = be.CreateAccount(ctx, resp.Account)
	require.NoError(t, err)

	require.Equal(t, 1, len(resp.RootIdentities))

	var dDoc *model.DIDDocument
	for _, e := range resp.RootIdentities {
		dDoc, err = model.SimpleDIDDocument(e.DID, e.Created)
		require.NoError(t, err)

		err = be.CreateDIDDocument(ctx, dDoc)
		require.NoError(t, err)
	}

	// happy path

	_, err = be.GetDIDDocument(ctx, resp.RootIdentities[0].ID())
	require.NoError(t, err)

	// try getting a non-existent DID doc

	dDoc.Proof.Value = "another-proof"

	_, err = be.GetDIDDocument(ctx, "bad-id")
	require.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrDIDNotFound))

	_ = client.Close()
	_, err = be.GetDIDDocument(ctx, resp.RootIdentities[0].ID())
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_ListDIDDocuments(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	acctTemplate := &account.Account{
		Email:        "test@example.com",
		Name:         "Test User",
		AccessLevel:  model.AccessLevelHosted,
		DefaultVault: testbase.TestVaultName,
	}
	resp, err := account.GenerateAccount(acctTemplate, account.WithPassphraseAuth("pass123"))
	require.NoError(t, err)

	err = be.CreateAccount(ctx, resp.Account)
	require.NoError(t, err)

	require.Equal(t, 1, len(resp.RootIdentities))

	var dDoc *model.DIDDocument
	for _, e := range resp.RootIdentities {
		dDoc, err = model.SimpleDIDDocument(e.DID, e.Created)
		require.NoError(t, err)

		err = be.CreateDIDDocument(ctx, dDoc)
		require.NoError(t, err)
	}

	// happy path

	list, err := be.ListDIDDocuments(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(list))

	// fail on database errors

	_ = client.Close()
	_, err = be.ListDIDDocuments(ctx)
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_StoreLocker(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	acctTemplate := &account.Account{
		Email:        "test@example.com",
		Name:         "Test User",
		AccessLevel:  model.AccessLevelHosted,
		DefaultVault: testbase.TestVaultName,
	}
	resp, err := account.GenerateAccount(acctTemplate, account.WithPassphraseAuth("pass123"))
	require.NoError(t, err)

	err = be.CreateAccount(ctx, resp.Account)
	require.NoError(t, err)

	// happy path

	for _, envelope := range resp.EncryptedLockers {
		err = be.StoreLocker(ctx, resp.Account.ID, envelope)
		require.NoError(t, err)
	}

	// fail if account doesn't exist

	err = be.StoreLocker(ctx, "bad-account", resp.EncryptedLockers[0])
	require.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrAccountNotFound))

	_ = client.Close()
	err = be.StoreLocker(ctx, resp.Account.ID, resp.EncryptedLockers[0])
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_GetLocker(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	acctTemplate := &account.Account{
		Email:        "test@example.com",
		Name:         "Test User",
		AccessLevel:  model.AccessLevelHosted,
		DefaultVault: testbase.TestVaultName,
	}
	resp, err := account.GenerateAccount(acctTemplate, account.WithPassphraseAuth("pass123"))
	require.NoError(t, err)

	err = be.CreateAccount(ctx, resp.Account)
	require.NoError(t, err)

	for _, envelope := range resp.EncryptedLockers {
		err = be.StoreLocker(ctx, resp.Account.ID, envelope)
		require.NoError(t, err)
	}

	// happy path

	envelope, err := be.GetLocker(ctx, resp.Account.ID, resp.EncryptedLockers[0].Hash)
	require.NoError(t, err)
	assert.Equal(t, resp.EncryptedIdentities[0].EncryptedID, envelope.EncryptedID)

	// fail if account doesn't exist

	_, err = be.GetLocker(ctx, "bad-account", resp.EncryptedLockers[0].Hash)
	require.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrLockerNotFound))

	_ = client.Close()
	_, err = be.GetLocker(ctx, resp.Account.ID, resp.EncryptedLockers[0].Hash)
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_ListLockers(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	acctTemplate := &account.Account{
		Email:        "test@example.com",
		Name:         "Test User",
		AccessLevel:  model.AccessLevelHosted,
		DefaultVault: testbase.TestVaultName,
	}
	resp, err := account.GenerateAccount(acctTemplate, account.WithPassphraseAuth("pass123"))
	require.NoError(t, err)

	err = be.CreateAccount(ctx, resp.Account)
	require.NoError(t, err)

	for _, envelope := range resp.EncryptedLockers {
		err = be.StoreLocker(ctx, resp.Account.ID, envelope)
		require.NoError(t, err)
	}

	// happy paths

	list, err := be.ListLockers(ctx, resp.Account.ID, model.AccessLevelManaged)
	require.NoError(t, err)
	assert.Equal(t, 1, len(list))

	list, err = be.ListLockers(ctx, resp.Account.ID, model.AccessLevelHosted)
	require.NoError(t, err)
	assert.Equal(t, 1, len(list))

	_ = client.Close()
	_, err = be.ListLockers(ctx, resp.Account.ID, model.AccessLevelHosted)
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_StoreProperty(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	acctTemplate := &account.Account{
		Email:        "test@example.com",
		Name:         "Test User",
		AccessLevel:  model.AccessLevelHosted,
		DefaultVault: testbase.TestVaultName,
	}
	resp, err := account.GenerateAccount(acctTemplate, account.WithPassphraseAuth("pass123"))
	require.NoError(t, err)

	err = be.CreateAccount(ctx, resp.Account)
	require.NoError(t, err)

	testPropEnvelope := &account.DataEnvelope{
		Hash:          "Hash",
		AccessLevel:   model.AccessLevelLocal,
		EncryptedID:   "EncryptedID",
		EncryptedBody: "EncryptedBody",
	}

	// happy path

	err = be.StoreProperty(ctx, resp.Account.ID, testPropEnvelope)
	require.NoError(t, err)

	// fail if account doesn't exist

	err = be.StoreProperty(ctx, "bad-account", testPropEnvelope)
	require.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrAccountNotFound))

	_ = client.Close()
	err = be.StoreProperty(ctx, resp.Account.ID, testPropEnvelope)
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_GetProperty(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	acctTemplate := &account.Account{
		Email:        "test@example.com",
		Name:         "Test User",
		AccessLevel:  model.AccessLevelHosted,
		DefaultVault: testbase.TestVaultName,
	}
	resp, err := account.GenerateAccount(acctTemplate, account.WithPassphraseAuth("pass123"))
	require.NoError(t, err)

	err = be.CreateAccount(ctx, resp.Account)
	require.NoError(t, err)

	testPropEnvelope := &account.DataEnvelope{
		Hash:          "Hash",
		AccessLevel:   model.AccessLevelLocal,
		EncryptedID:   "EncryptedID",
		EncryptedBody: "EncryptedBody",
	}

	err = be.StoreProperty(ctx, resp.Account.ID, testPropEnvelope)
	require.NoError(t, err)

	// happy path

	envelope, err := be.GetProperty(ctx, resp.Account.ID, testPropEnvelope.Hash)
	require.NoError(t, err)
	assert.Equal(t, testPropEnvelope.EncryptedID, envelope.EncryptedID)

	// fail if account doesn't exist

	_, err = be.GetProperty(ctx, "bad-account", testPropEnvelope.Hash)
	require.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrPropertyNotFound))

	_ = client.Close()
	_, err = be.GetProperty(ctx, resp.Account.ID, testPropEnvelope.Hash)
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_ListProperties(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	acctTemplate := &account.Account{
		Email:        "test@example.com",
		Name:         "Test User",
		AccessLevel:  model.AccessLevelHosted,
		DefaultVault: testbase.TestVaultName,
	}
	resp, err := account.GenerateAccount(acctTemplate, account.WithPassphraseAuth("pass123"))
	require.NoError(t, err)

	err = be.CreateAccount(ctx, resp.Account)
	require.NoError(t, err)

	testPropEnvelope := &account.DataEnvelope{
		Hash:          "Hash",
		AccessLevel:   model.AccessLevelLocal,
		EncryptedID:   "EncryptedID",
		EncryptedBody: "EncryptedBody",
	}

	err = be.StoreProperty(ctx, resp.Account.ID, testPropEnvelope)
	require.NoError(t, err)

	// happy paths

	list, err := be.ListProperties(ctx, resp.Account.ID, model.AccessLevelLocal)
	require.NoError(t, err)
	assert.Equal(t, 1, len(list))

	list, err = be.ListProperties(ctx, resp.Account.ID, model.AccessLevelHosted)
	require.NoError(t, err)
	assert.Equal(t, 0, len(list))

	_ = client.Close()
	_, err = be.ListProperties(ctx, resp.Account.ID, model.AccessLevelHosted)
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_CreateRecoveryCode(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	testAcct := createAccount(t, "test@example.com", "Test User")

	err := be.CreateAccount(ctx, testAcct)
	require.NoError(t, err)

	recCode, err := account.NewRecoveryCode(testAcct.ID, 120)
	require.NoError(t, err)

	// happy path

	err = be.CreateRecoveryCode(ctx, recCode)
	require.NoError(t, err)

	// fail if account not found

	recCode.UserID = "bad-account-id"
	err = be.CreateRecoveryCode(ctx, recCode)
	require.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrAccountNotFound))

	_ = client.Close()
	err = be.CreateRecoveryCode(ctx, recCode)
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_GetRecoveryCode(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	testAcct := createAccount(t, "test@example.com", "Test User")

	err := be.CreateAccount(ctx, testAcct)
	require.NoError(t, err)

	recCode, err := account.NewRecoveryCode(testAcct.ID, 120)
	require.NoError(t, err)

	err = be.CreateRecoveryCode(ctx, recCode)
	require.NoError(t, err)

	// happy path

	rc, err := be.GetRecoveryCode(ctx, recCode.Code)
	require.NoError(t, err)
	assert.Equal(t, recCode.ExpiresAt, rc.ExpiresAt)
	assert.Equal(t, testAcct.ID, rc.UserID)

	// fail if code not found

	_, err = be.GetRecoveryCode(ctx, "bad-code")
	require.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrRecoveryCodeNotFound))

	_ = client.Close()
	_, err = be.GetRecoveryCode(ctx, recCode.Code)
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_DeleteRecoveryCode(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	testAcct := createAccount(t, "test@example.com", "Test User")

	err := be.CreateAccount(ctx, testAcct)
	require.NoError(t, err)

	recCode, err := account.NewRecoveryCode(testAcct.ID, 120)
	require.NoError(t, err)

	err = be.CreateRecoveryCode(ctx, recCode)
	require.NoError(t, err)

	// happy path

	err = be.DeleteRecoveryCode(ctx, recCode.Code)
	require.NoError(t, err)

	// fail if code not found

	err = be.DeleteRecoveryCode(ctx, "bad-code")
	require.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrRecoveryCodeNotFound))

	_ = client.Close()
	err = be.DeleteRecoveryCode(ctx, recCode.Code)
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_StoreAccessKey(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	testAcct := createAccount(t, "test@example.com", "Test User")

	err := be.CreateAccount(ctx, testAcct)
	require.NoError(t, err)

	key, err := model.GenerateAccessKey(testAcct.ID, model.AccessLevelHosted)
	require.NoError(t, err)

	// happy path

	err = be.StoreAccessKey(ctx, key)
	require.NoError(t, err)

	// fail if account not found

	badKey, err := model.GenerateAccessKey("bad-account-id", model.AccessLevelHosted)
	require.NoError(t, err)

	err = be.StoreAccessKey(ctx, badKey)
	require.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrAccountNotFound))

	_ = client.Close()
	err = be.StoreAccessKey(ctx, key)
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_GetAccessKey(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	testAcct := createAccount(t, "test@example.com", "Test User")

	err := be.CreateAccount(ctx, testAcct)
	require.NoError(t, err)

	key, err := model.GenerateAccessKey(testAcct.ID, model.AccessLevelHosted)
	require.NoError(t, err)

	err = be.StoreAccessKey(ctx, key)
	require.NoError(t, err)

	// happy path

	rc, err := be.GetAccessKey(ctx, key.ID)
	require.NoError(t, err)
	assert.Equal(t, key.ManagementKey, rc.ManagementKey)

	// fail if key not found

	_, err = be.GetAccessKey(ctx, "bad-id")
	require.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrAccessKeyNotFound))

	_ = client.Close()
	_, err = be.GetAccessKey(ctx, key.ID)
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_DeleteAccessKey(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	testAcct := createAccount(t, "test@example.com", "Test User")

	err := be.CreateAccount(ctx, testAcct)
	require.NoError(t, err)

	key, err := model.GenerateAccessKey(testAcct.ID, model.AccessLevelHosted)
	require.NoError(t, err)

	err = be.StoreAccessKey(ctx, key)
	require.NoError(t, err)

	// happy path

	err = be.DeleteAccessKey(ctx, key.ID)
	require.NoError(t, err)

	// fail if key not found

	err = be.DeleteAccessKey(ctx, "bad-id")
	require.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrAccessKeyNotFound))

	_ = client.Close()
	err = be.DeleteAccessKey(ctx, key.ID)
	assertFailOnDatabaseError(t, err)
}

func Test_RelationalBackend_ListAccessKeys(t *testing.T) {
	client := newClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	be := NewRelationalBackend(client)

	testAcct := createAccount(t, "test@example.com", "Test User")

	err := be.CreateAccount(ctx, testAcct)
	require.NoError(t, err)

	testAcct2 := createAccount(t, "test2@example.com", "Test User")
	testAcct2.State = account.StateSuspended
	testAcct2.ParentAccount = testAcct.ID

	err = be.CreateAccount(ctx, testAcct2)
	require.NoError(t, err)

	key, err := model.GenerateAccessKey(testAcct.ID, model.AccessLevelHosted)
	require.NoError(t, err)

	err = be.StoreAccessKey(ctx, key)
	require.NoError(t, err)

	key, err = model.GenerateAccessKey(testAcct2.ID, model.AccessLevelHosted)
	require.NoError(t, err)

	err = be.StoreAccessKey(ctx, key)
	require.NoError(t, err)

	// happy paths

	list, err := be.ListAccessKeys(ctx, testAcct.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(list))

	list, err = be.ListAccessKeys(ctx, testAcct2.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(list))

	_ = client.Close()
	_, err = be.ListAccessKeys(ctx, testAcct2.ID)
	assertFailOnDatabaseError(t, err)
}

func createAccount(t *testing.T, email, name string) *account.Account { //nolint: thelper
	acctTemplate := &account.Account{
		Email:        email,
		Name:         name,
		AccessLevel:  model.AccessLevelHosted,
		DefaultVault: testbase.TestVaultName,
	}
	passPhrase := "pass123"
	resp, err := account.GenerateAccount(acctTemplate, account.WithPassphraseAuth(passPhrase))
	require.NoError(t, err)

	return resp.Account
}

func newSqliteDataSourceName(t *testing.T) string { //nolint: thelper
	dbName, err := utils.RandomID(8)
	require.NoError(t, err)

	return "file:" + dbName + "?mode=memory&cache=shared&_fk=1&_journal=WAL&_busy_timeout=5000"
}

func newClient(t *testing.T) *ent.Client { //nolint: thelper

	client := enttest.Open(t,
		"sqlite3",
		newSqliteDataSourceName(t),
		enttest.WithOptions(ent.Debug()),
		enttest.WithMigrateOptions(migrate.WithGlobalUniqueID(false)),
	)
	client.SetOneOpenConnection()

	return client
}

func assertFailOnDatabaseError(t *testing.T, err error) { //nolint: thelper
	require.Error(t, err)
	assert.Equal(t, "sql: database is closed", err.Error())
}
