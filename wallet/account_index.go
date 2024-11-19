package wallet

import (
	"context"
	"errors"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/storage"
	"github.com/rs/zerolog/log"
)

type (
	// AccountIndex allows indexing information that is not available in MetaLocker
	// datasets. If you pass an index that implements AccountIndex interface to IndexUpdater,
	// it will receive updates about account components.
	AccountIndex interface {
		UpdateAccount(acct *account.Account) error
		UpdateIdentity(accountID string, idy Identity) error
		UpdateLocker(accountID string, l Locker) error
	}
)

func InitAccountIndex(ctx context.Context, ai AccountIndex, dw DataWallet) error {

	if err := ai.UpdateAccount(dw.Account()); err != nil {
		return err
	}

	rootIdy, err := dw.GetRootIdentity(ctx)
	if err != nil {
		return err
	}
	if err = ai.UpdateIdentity(dw.ID(), rootIdy); err != nil {
		return err
	}

	l, err := dw.GetRootLocker(ctx, model.AccessLevelManaged)
	if err != nil {
		return err
	}
	if err = ai.UpdateLocker(dw.ID(), l); err != nil {
		return err
	}

	if dw.LockLevel() >= model.AccessLevelHosted {
		l, err := dw.GetRootLocker(ctx, model.AccessLevelHosted)
		if err != nil {
			return err
		}
		if err = ai.UpdateLocker(dw.ID(), l); err != nil {
			return err
		}
	}

	return nil
}

func ApplyAccountUpdate(ctx context.Context, ai AccountIndex, update *AccountUpdate, dw DataWallet) error {
	for _, iid := range update.IdentitiesAdded {
		idy, err := dw.GetIdentity(ctx, iid)
		if err != nil {
			if errors.Is(err, storage.ErrIdentityNotFound) {
				continue
			} else {
				return err
			}
		}

		if err = ai.UpdateIdentity(update.AccountID, idy); err != nil {
			log.Err(err).Msg("Error when updating identity in account index")
			return err
		}
	}

	for _, lid := range update.LockersOpened {
		l, err := dw.GetLocker(ctx, lid)
		if err != nil {
			if errors.Is(err, storage.ErrLockerNotFound) {
				continue
			} else {
				return err
			}
		}

		if err = ai.UpdateLocker(update.AccountID, l); err != nil {
			log.Err(err).Msg("Error when updating locker in account index")
			return err
		}
	}

	for _, subID := range update.SubAccountsAdded {
		log.Debug().Str("subID", subID).Msg("Processing new sub-account")

		subDW, err := dw.GetSubAccountWallet(ctx, subID)
		if err != nil {
			return err
		}

		if err = InitAccountIndex(ctx, ai, subDW); err != nil {
			return err
		}
	}

	return nil
}
