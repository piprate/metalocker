package rdb

import (
	"context"
	"errors"
	"strings"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/storage"
	"github.com/piprate/metalocker/storage/rdb/ent"
	"github.com/piprate/metalocker/storage/rdb/ent/accesskey"
	entAccount "github.com/piprate/metalocker/storage/rdb/ent/account"
	"github.com/piprate/metalocker/storage/rdb/ent/did"
	"github.com/piprate/metalocker/storage/rdb/ent/identity"
	"github.com/piprate/metalocker/storage/rdb/ent/locker"
	"github.com/piprate/metalocker/storage/rdb/ent/predicate"
	"github.com/piprate/metalocker/storage/rdb/ent/property"
	"github.com/piprate/metalocker/storage/rdb/ent/recoverycode"
	"github.com/piprate/metalocker/utils/measure"
	"github.com/rs/zerolog/log"
)

const (
	maxAccountDepth = 10
)

type RelationalBackend struct {
	client *ent.Client
}

var _ storage.IdentityBackend = (*RelationalBackend)(nil)

func NewRelationalBackend(client *ent.Client) *RelationalBackend {
	return &RelationalBackend{
		client: client,
	}
}

func (rbe *RelationalBackend) Close() error {
	return nil
}

func (rbe *RelationalBackend) CreateAccount(ctx context.Context, acct *account.Account) error {
	defer measure.ExecTime("rdb.CreateAccount")()

	log.Debug().Str("uid", acct.ID).Msg("CreateAccount")

	existingAccount, err := rbe.client.Account.Query().Select(entAccount.FieldDid, entAccount.FieldEmail).Where(
		entAccount.Or(
			entAccount.Did(acct.ID),
			entAccount.Email(acct.Email),
		),
	).First(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return err
		}
	} else {
		if existingAccount.Did == acct.ID {
			return storage.ErrAccountExists
		} else if acct.Email != "" && existingAccount.Email == acct.Email {
			log.Error().Str("uid", acct.ID).Str("email", acct.Email).
				Msg("Email in use by another account")
			return storage.ErrAccountExists
		}
	}

	_, err = rbe.client.Account.Create().
		SetDid(acct.ID).
		SetEmail(acct.Email).
		SetState(acct.State).
		SetParentAccount(acct.ParentAccount).
		SetBody(acct).
		Save(ctx)
	return err
}

func (rbe *RelationalBackend) UpdateAccount(ctx context.Context, acct *account.Account) error {
	defer measure.ExecTime("rdb.UpdateAccount")()

	acctRow, err := rbe.client.Account.Query().Where(entAccount.Did(acct.ID)).First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return storage.ErrAccountNotFound
		} else {
			return err
		}
	}

	// FIXME: check email clash

	_, err = acctRow.Update().
		SetState(acct.State).
		SetEmail(acct.Email).
		SetParentAccount(acct.ParentAccount).
		SetBody(acct).
		Save(ctx)
	return err
}

func (rbe *RelationalBackend) GetAccount(ctx context.Context, id string) (*account.Account, error) {
	defer measure.ExecTime("rdb.GetAccount")()

	var acct *ent.Account
	var err error
	if strings.Contains(id, "@") {
		acct, err = rbe.client.Account.Query().Where(entAccount.Email(id)).First(ctx)
	} else {
		acct, err = rbe.client.Account.Query().Where(entAccount.Did(id)).First(ctx)
	}

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, storage.ErrAccountNotFound
		} else {
			return nil, err
		}
	}
	return acct.Body, nil
}

func (rbe *RelationalBackend) DeleteAccount(ctx context.Context, id string) error {
	defer measure.ExecTime("rdb.DeleteAccount")()

	rowCount, err := rbe.client.Account.Delete().Where(entAccount.Did(id)).Exec(ctx)
	if err != nil {
		return err
	}
	if rowCount == 0 {
		return storage.ErrAccountNotFound
	}
	return nil
}

func (rbe *RelationalBackend) ListAccounts(ctx context.Context, parentAccountID, stateFilter string) ([]*account.Account, error) {
	defer measure.ExecTime("rdb.ListAccounts")()

	var clauses []predicate.Account
	if parentAccountID != "" {
		clauses = append(clauses, entAccount.ParentAccount(parentAccountID))
	}
	if stateFilter != "" {
		clauses = append(clauses, entAccount.State(stateFilter))
	}
	rows, err := rbe.client.Account.Query().Where(clauses...).All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*account.Account, len(rows))

	for i, row := range rows {
		result[i] = row.Body
	}

	return result, nil
}

func (rbe *RelationalBackend) HasAccountAccess(ctx context.Context, accountID, targetAccountID string) (bool, error) {
	for depth := 0; depth < maxAccountDepth; depth++ {
		if targetAccountID == accountID {
			return true, nil
		}

		acct, err := rbe.client.Account.Query().Where(entAccount.Did(targetAccountID)).First(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				log.Error().Str("id", targetAccountID).Msg("Account not found while looking for master account")
				return false, storage.ErrAccountNotFound
			} else {
				return false, err
			}
		}
		if acct.ParentAccount == "" {
			log.Warn().Str("id", targetAccountID).Msg("NIL parent")
			return false, nil
		}

		targetAccountID = acct.ParentAccount
	}

	return false, errors.New("max account depth exceeded")
}

func (rbe *RelationalBackend) CreateDIDDocument(ctx context.Context, ddoc *model.DIDDocument) error {
	defer measure.ExecTime("rdb.CreateDIDDocument")()

	row, err := rbe.client.DID.Query().Where(did.Did(ddoc.ID)).First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// save new doc
			_, err = rbe.client.DID.Create().
				SetDid(ddoc.ID).
				SetBody(ddoc).
				Save(ctx)
			return err
		} else {
			return err
		}
	}

	if row.Body.Equals(ddoc) {
		return nil
	} else {
		return storage.ErrDIDExists
	}
}

func (rbe *RelationalBackend) GetDIDDocument(ctx context.Context, iid string) (*model.DIDDocument, error) {
	defer measure.ExecTime("rdb.GetDIDDocument")()

	row, err := rbe.client.DID.Query().Where(did.Did(iid)).First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, storage.ErrDIDNotFound
		} else {
			return nil, err
		}
	}
	return row.Body, nil
}

func (rbe *RelationalBackend) ListDIDDocuments(ctx context.Context) ([]*model.DIDDocument, error) {
	defer measure.ExecTime("rdb.ListDIDDocuments")()

	rows, err := rbe.client.DID.Query().All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*model.DIDDocument, len(rows))

	for i, row := range rows {
		result[i] = row.Body
	}

	return result, nil
}

func (rbe *RelationalBackend) ListAccessKeys(ctx context.Context, accountID string) ([]*model.AccessKey, error) {
	defer measure.ExecTime("rdb.ListAccessKeys")()

	rows, err := rbe.client.AccessKey.Query().
		Where(
			accesskey.HasAccountWith(entAccount.Did(accountID)),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*model.AccessKey, len(rows))

	for i, row := range rows {
		result[i] = row.Body
	}

	return result, nil
}

func (rbe *RelationalBackend) StoreAccessKey(ctx context.Context, accessKey *model.AccessKey) error {
	defer measure.ExecTime("rdb.StoreAccessKey")()

	id, err := rbe.client.Account.Query().Where(entAccount.Did(accessKey.AccountID)).OnlyID(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return storage.ErrAccountNotFound
		} else {
			return err
		}
	}

	_, err = rbe.client.AccessKey.Create().
		SetAccountID(id).
		SetDid(accessKey.ID).
		SetBody(accessKey).
		Save(ctx)

	return err
}

func (rbe *RelationalBackend) GetAccessKey(ctx context.Context, keyID string) (*model.AccessKey, error) {
	defer measure.ExecTime("rdb.GetAccessKey")()

	row, err := rbe.client.AccessKey.Query().Where(accesskey.Did(keyID)).First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, storage.ErrAccessKeyNotFound
		} else {
			return nil, err
		}
	}
	return row.Body, nil
}

func (rbe *RelationalBackend) DeleteAccessKey(ctx context.Context, keyID string) error {
	defer measure.ExecTime("rdb.GetAccessKey")()

	rowCount, err := rbe.client.AccessKey.Delete().
		Where(accesskey.Did(keyID)).Exec(ctx)
	if err != nil {
		return err
	}
	if rowCount == 0 {
		return storage.ErrAccessKeyNotFound
	}
	return nil
}

func (rbe *RelationalBackend) StoreIdentity(ctx context.Context, accountID string, idy *account.DataEnvelope) error {
	defer measure.ExecTime("rdb.StoreIdentity")()

	id, err := rbe.client.Account.Query().Where(entAccount.Did(accountID)).OnlyID(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return storage.ErrAccountNotFound
		} else {
			return err
		}
	}

	_, err = rbe.client.Identity.Create().
		SetAccountID(id).
		SetHash(idy.Hash).
		SetLevel(int32(idy.AccessLevel)).
		SetEncryptedID(idy.EncryptedID).
		SetEncryptedBody(idy.EncryptedBody).
		Save(ctx)

	return err
}

func (rbe *RelationalBackend) GetIdentity(ctx context.Context, accountID string, hash string) (*account.DataEnvelope, error) {
	defer measure.ExecTime("rdb.GetIdentity")()

	row, err := rbe.client.Identity.Query().Where(
		identity.HasAccountWith(entAccount.Did(accountID)),
		identity.Hash(hash),
	).First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, storage.ErrIdentityNotFound
		} else {
			return nil, err
		}
	}
	return &account.DataEnvelope{
		Hash:          hash,
		AccessLevel:   model.AccessLevel(row.Level),
		EncryptedID:   row.EncryptedID,
		EncryptedBody: row.EncryptedBody,
	}, nil
}

func (rbe *RelationalBackend) ListIdentities(ctx context.Context, accountID string, lvl model.AccessLevel) ([]*account.DataEnvelope, error) {
	defer measure.ExecTime("rdb.ListIdentities")()

	rows, err := rbe.client.Identity.Query().
		Where(
			identity.HasAccountWith(entAccount.Did(accountID)),
			identity.Level(int32(lvl)),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*account.DataEnvelope, len(rows))

	for i, row := range rows {
		result[i] = &account.DataEnvelope{
			Hash:          row.Hash,
			AccessLevel:   lvl,
			EncryptedID:   row.EncryptedID,
			EncryptedBody: row.EncryptedBody,
		}
	}

	return result, nil
}

func (rbe *RelationalBackend) StoreLocker(ctx context.Context, accountID string, l *account.DataEnvelope) error {
	defer measure.ExecTime("rdb.StoreLocker")()

	id, err := rbe.client.Account.Query().Where(entAccount.Did(accountID)).OnlyID(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return storage.ErrAccountNotFound
		} else {
			return err
		}
	}

	_, err = rbe.client.Locker.Create().
		SetAccountID(id).
		SetHash(l.Hash).
		SetLevel(int32(l.AccessLevel)).
		SetEncryptedID(l.EncryptedID).
		SetEncryptedBody(l.EncryptedBody).
		Save(ctx)

	return err
}

func (rbe *RelationalBackend) GetLocker(ctx context.Context, accountID string, hash string) (*account.DataEnvelope, error) {
	defer measure.ExecTime("rdb.GetLocker")()

	row, err := rbe.client.Locker.Query().Where(
		locker.HasAccountWith(entAccount.Did(accountID)),
		locker.Hash(hash),
	).First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, storage.ErrLockerNotFound
		} else {
			return nil, err
		}
	}
	return &account.DataEnvelope{
		Hash:          hash,
		AccessLevel:   model.AccessLevel(row.Level),
		EncryptedID:   row.EncryptedID,
		EncryptedBody: row.EncryptedBody,
	}, nil
}

func (rbe *RelationalBackend) ListLockers(ctx context.Context, accountID string, lvl model.AccessLevel) ([]*account.DataEnvelope, error) {
	defer measure.ExecTime("rdb.ListLockers")()

	rows, err := rbe.client.Locker.Query().
		Where(
			locker.HasAccountWith(entAccount.Did(accountID)),
			locker.Level(int32(lvl)),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*account.DataEnvelope, len(rows))

	for i, row := range rows {
		result[i] = &account.DataEnvelope{
			Hash:          row.Hash,
			AccessLevel:   lvl,
			EncryptedID:   row.EncryptedID,
			EncryptedBody: row.EncryptedBody,
		}
	}

	return result, nil
}

func (rbe *RelationalBackend) StoreProperty(ctx context.Context, accountID string, prop *account.DataEnvelope) error {
	defer measure.ExecTime("rdb.StoreProperty")()

	id, err := rbe.client.Account.Query().Where(entAccount.Did(accountID)).OnlyID(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return storage.ErrAccountNotFound
		} else {
			return err
		}
	}

	_, err = rbe.client.Property.Create().
		SetAccountID(id).
		SetHash(prop.Hash).
		SetLevel(int32(prop.AccessLevel)).
		SetEncryptedID(prop.EncryptedID).
		SetEncryptedBody(prop.EncryptedBody).
		Save(ctx)

	return err
}

func (rbe *RelationalBackend) GetProperty(ctx context.Context, accountID string, hash string) (*account.DataEnvelope, error) {
	defer measure.ExecTime("rdb.GetProperty")()

	row, err := rbe.client.Property.Query().Where(
		property.HasAccountWith(entAccount.Did(accountID)),
		property.Hash(hash),
	).First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, storage.ErrPropertyNotFound
		} else {
			return nil, err
		}
	}
	return &account.DataEnvelope{
		Hash:          hash,
		AccessLevel:   model.AccessLevel(row.Level),
		EncryptedID:   row.EncryptedID,
		EncryptedBody: row.EncryptedBody,
	}, nil
}

func (rbe *RelationalBackend) ListProperties(ctx context.Context, accountID string, lvl model.AccessLevel) ([]*account.DataEnvelope, error) {
	defer measure.ExecTime("rdb.ListProperties")()

	rows, err := rbe.client.Property.Query().
		Where(
			property.HasAccountWith(entAccount.Did(accountID)),
			property.Level(int32(lvl)),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*account.DataEnvelope, len(rows))

	for i, row := range rows {
		result[i] = &account.DataEnvelope{
			Hash:          row.Hash,
			AccessLevel:   lvl,
			EncryptedID:   row.EncryptedID,
			EncryptedBody: row.EncryptedBody,
		}
	}

	return result, nil
}

func (rbe *RelationalBackend) DeleteProperty(ctx context.Context, accountID string, hash string) error {
	rowCount, err := rbe.client.Property.Delete().Where(
		property.HasAccountWith(entAccount.Did(accountID)),
		property.Hash(hash),
	).Exec(ctx)
	if err != nil {
		return err
	}
	if rowCount == 0 {
		return storage.ErrPropertyNotFound
	}
	return nil
}

func (rbe *RelationalBackend) CreateRecoveryCode(ctx context.Context, c *account.RecoveryCode) error {
	defer measure.ExecTime("rdb.CreateRecoveryCode")()

	id, err := rbe.client.Account.Query().Where(entAccount.Did(c.UserID)).OnlyID(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return storage.ErrAccountNotFound
		} else {
			return err
		}
	}

	_, err = rbe.client.RecoveryCode.Create().
		SetCode(c.Code).
		SetAccountID(id).
		SetExpiresAt(*c.ExpiresAt).
		Save(ctx)

	return err
}

func (rbe *RelationalBackend) GetRecoveryCode(ctx context.Context, code string) (*account.RecoveryCode, error) {
	defer measure.ExecTime("rdb.GetRecoveryCode")()

	row, err := rbe.client.RecoveryCode.Query().
		Where(recoverycode.Code(code)).
		WithAccount().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, storage.ErrRecoveryCodeNotFound
		} else {
			return nil, err
		}
	}

	return &account.RecoveryCode{
		Code:      code,
		UserID:    row.Edges.Account.Did,
		ExpiresAt: row.ExpiresAt,
	}, nil
}

func (rbe *RelationalBackend) DeleteRecoveryCode(ctx context.Context, code string) error {
	defer measure.ExecTime("rdb.DeleteRecoveryCode")()
	rowCount, err := rbe.client.RecoveryCode.Delete().
		Where(recoverycode.Code(code)).Exec(ctx)
	if err != nil {
		return err
	}
	if rowCount == 0 {
		return storage.ErrRecoveryCodeNotFound
	}
	return nil
}
