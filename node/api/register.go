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

package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/gin-gonic/gin"
	"github.com/knadh/koanf"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/storage"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/wallet"
)

type NewAccountForm struct {
	Account          *account.Account `json:"account"`
	RegistrationCode string           `json:"registrationCode"`
}

func RegisterHandler(registrationCodes []string, defaultVault string, secondLevelRecoveryKey []byte,
	entropyFunc account.EntropyFunction,
	hashFunction account.PasswordHashFunction,
	timeFunc func() time.Time, jwtMW *apibase.GinJWTMiddleware,
	identityBackend storage.IdentityBackend, ledger model.Ledger) func(c *gin.Context) {

	if timeFunc == nil {
		timeFunc = time.Now
	}

	if entropyFunc == nil {
		entropyFunc = account.DefaultEntropyFunction()
	}

	registrationCodeMap := make(map[string]bool, len(registrationCodes))
	for _, code := range registrationCodes {
		registrationCodeMap[code] = true
	}

	registrationCodeEnabled := len(registrationCodes) > 0

	// we don't need to provide all services for this operation
	factory, err := wallet.NewLocalFactory(ledger, nil, nil, identityBackend, nil, nil, hashFunction)
	if err != nil {
		panic(err)
	}

	return func(c *gin.Context) {
		log := apibase.CtxLogger(c)

		buf, _ := c.GetRawData()

		var accountForm NewAccountForm
		err := jsonw.Unmarshal(buf, &accountForm)
		if err != nil || accountForm.Account == nil || len(accountForm.Account.Email) == 0 || len(accountForm.Account.EncryptedPassword) == 0 {
			log.Err(err).Str("body", string(buf)).Msg("Bad register request body")
			apibase.AbortWithError(c, http.StatusBadRequest, "Bad register request")
			return
		}

		accountForm.Account.Email = strings.ToLower(accountForm.Account.Email)

		acct := accountForm.Account

		if registrationCodeEnabled && !registrationCodeMap[accountForm.RegistrationCode] {
			log.Error().Str("body", string(buf)).Str("code", accountForm.RegistrationCode).
				Msg("Bad registration code")
			apibase.AbortWithError(c, http.StatusBadRequest, "Bad registration code")
			return
		}

		log.Info().Str("userID", acct.Email).Msg("New account registration requested")

		if acct.AccessLevel == model.AccessLevelNone {
			log.Error().Str("body", string(buf)).Msg("Account access level not provided")
			apibase.AbortWithError(c, http.StatusBadRequest, "account access level not provided")
			return
		} else if acct.AccessLevel == model.AccessLevelRestricted {
			log.Error().Str("body", string(buf)).Msg("Can't register a restricted account")
			apibase.AbortWithError(c, http.StatusBadRequest, "restricted account registration not supported")
			return
		}

		acct.State = account.StateActive

		// set registration time
		registeredAt := timeFunc()
		acct.RegisteredAt = &registeredAt

		hashedPassword := acct.EncryptedPassword

		if acct.DefaultVault == "" {
			acct.DefaultVault = defaultVault
		}

		var recoveryPhrase string
		var secondLevelRecoveryCode string
		if acct.AccessLevel == model.AccessLevelManaged {
			var recDetails *wallet.RecoveryDetails
			var dw wallet.DataWallet
			dw, recDetails, err = factory.RegisterAccount(acct,
				account.WithHashedPassphraseAuth(acct.EncryptedPassword),
				account.WithCustomEntropy(entropyFunc),
				account.WithSLRK(secondLevelRecoveryKey),
				account.WithLogger(log))
			if err == nil {
				recoveryPhrase = recDetails.RecoveryPhrase
				secondLevelRecoveryCode = recDetails.SecondLevelRecoveryCode
				acct = dw.Account()
			}
		} else {
			_, err = factory.SaveAccount(acct)
		}
		if err != nil {
			log.Err(err).Msg("Error when registering account")
			if errors.Is(err, storage.ErrAccountExists) {
				apibase.AbortWithError(c, http.StatusConflict, "Account already exists")
			} else {
				apibase.AbortWithError(c, http.StatusBadRequest, "Account creation failed")
			}
			return
		}

		var rsp gin.H

		if jwtMW != nil {
			// Create the token

			managedKey, err := acct.ExtractManagedKey(hashedPassword)
			if err != nil {
				log.Err(err).Msg("Error when extracting managed key from new account")
				apibase.AbortWithError(c, http.StatusUnauthorized, "Failed to extract managed key")
				return
			}

			data := map[string]any{
				"id":     acct.ID,
				"email":  acct.Email,
				"secret": managedKey.Base64(),
			}

			tokenString, _, err := jwtMW.TokenGenerator(data)
			if err != nil {
				apibase.AbortWithError(c, http.StatusUnauthorized, "Create JWT Token failed")
				return
			}

			rsp = gin.H{
				"status": "ok",
				"token":  tokenString,
			}

		} else {
			rsp = gin.H{
				"status": "ok",
			}
		}

		if recoveryPhrase != "" {
			rsp["recoveryPhrase"] = recoveryPhrase
		}

		if secondLevelRecoveryCode != "" {
			rsp["secondLevelRecoveryCode"] = secondLevelRecoveryCode
		}

		apibase.JSON(c, http.StatusOK, rsp)
	}
}

func InitRegisterRoute(r *gin.Engine, path string, cfg *koanf.Koanf, jwtMW *apibase.GinJWTMiddleware, identityBackend storage.IdentityBackend, ledger model.Ledger) {
	var slrKey []byte
	slrKeyStr := cfg.String("secondLevelRecoveryKey")
	if slrKeyStr != "" {
		slrKey = base58.Decode(slrKeyStr)
	}

	r.POST(path, RegisterHandler(
		cfg.Strings("registrationCode"),
		cfg.String("defaultVaultName"),
		slrKey, nil, nil, nil, jwtMW, identityBackend, ledger))
}
