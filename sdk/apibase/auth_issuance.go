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

package apibase

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/storage"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

const (
	ClaimAccountID               = "id"
	ClaimEmail                   = "email"
	ClaimIssuer                  = "iss"
	ClaimAudience                = "aud"
	ClaimAudienceKeyID           = "akid"
	ClaimEncryptedAudienceSecret = "eas"
	ClaimLegacySecret            = "secret"
)

type AccountBackend interface {
	CreateAccount(account *account.Account) error

	GetAccount(email string) (account *account.Account, err error)
}

// LoginForm form structure.
type LoginForm struct {
	Username          string `form:"username" json:"username" binding:"required"`
	Password          string `form:"password" json:"password" binding:"required"`
	Audience          string `form:"audience" json:"audience"`
	AudiencePublicKey string `form:"audienceKey" json:"audienceKey"`
}

func (lf LoginForm) Bytes() []byte {
	b, _ := jsonw.Marshal(lf)
	return b
}

func AuthenticationHandler(accountBackend storage.IdentityBackend, acceptedAudiences []string, defaultAudience, defaultAudienceKey string) func(c *gin.Context) (any, error) {
	acceptedAudiencesFilter := make(map[string]bool, len(acceptedAudiences))
	for _, aud := range acceptedAudiences {
		acceptedAudiencesFilter[aud] = true
	}

	return func(c *gin.Context) (any, error) {
		var loginVals LoginForm
		if bindErr := c.BindJSON(&loginVals); bindErr != nil {
			return "", ErrMissingLoginValues
		}

		userID := loginVals.Username
		password := loginVals.Password

		if strings.Contains(userID, "@") {
			// if the username is an email, transform to lower case
			userID = strings.ToLower(userID)
		}

		acct, err := accountBackend.GetAccount(userID)
		if err != nil {
			log.Err(err).Str("ip", c.ClientIP()).Str("userID", userID).Msg("Error when retrieving account for authentication")
			return userID, ErrFailedAuthentication
		}

		if acct.State == account.StateSuspended {
			log.Error().Str("ip", c.ClientIP()).Str("userID", userID).Msg("Authentication failed: account suspended")
			return userID, ErrFailedAuthentication
		}

		if acct.State == account.StateDeleted {
			log.Error().Str("ip", c.ClientIP()).Str("userID", userID).Msg("Authentication failed: account deleted")
			return userID, ErrFailedAuthentication
		}

		err = bcrypt.CompareHashAndPassword([]byte(acct.EncryptedPassword), []byte(password))
		if err != nil {
			log.Err(err).Str("ip", c.ClientIP()).Str("userID", userID).Msg("Authentication failed")
			return userID, ErrFailedAuthentication
		}

		log.Debug().Str("userID", userID).Msg("Authentication successful")

		var audience string
		if loginVals.Audience != "" {
			if _, found := acceptedAudiencesFilter[loginVals.Audience]; !found {
				log.Err(err).Str("ip", c.ClientIP()).Str("userID", userID).
					Str("aud", loginVals.Audience).Msg("Audience not accepted")
				return userID, ErrFailedAuthentication
			}
			audience = loginVals.Audience
		} else if defaultAudience != "" {
			audience = defaultAudience
		}

		response := map[string]any{
			ClaimAccountID: acct.ID,
			ClaimEmail:     acct.Email,
			ClaimAudience:  audience,
		}

		if acct.State != account.StateRecovery {
			managedKey, err := acct.ExtractManagedKey(password)
			if err != nil {
				log.Err(err).Str("ip", c.ClientIP()).Str("userID", userID).Msg("Failed to extract managed key")
				return userID, ErrFailedAuthentication
			}

			audiencePublicKey := loginVals.AudiencePublicKey
			if audiencePublicKey == "" {
				if audience == defaultAudience {
					audiencePublicKey = defaultAudienceKey
				} else {
					log.Err(err).Str("ip", c.ClientIP()).Str("userID", userID).Str("aud", audience).Msg("Missing audience key")
					return userID, ErrFailedAuthentication
				}
			}
			if audiencePublicKey != "" {
				pubKey, err := base64.StdEncoding.DecodeString(audiencePublicKey)
				if err != nil || len(pubKey) != 32 {
					log.Err(err).Str("ip", c.ClientIP()).Str("userID", userID).Msg("Invalid audience key")
					return userID, ErrFailedAuthentication
				}

				response[ClaimEncryptedAudienceSecret] = base64.StdEncoding.EncodeToString(
					model.AnonEncrypt(managedKey.Bytes(), pubKey),
				)
				response[ClaimAudienceKeyID] = GetKeyID(pubKey)
			}
		}

		return response, nil
	}
}

func Payload(data any) MapClaims {
	return data.(map[string]any)
}

func IdentityHandler(c *gin.Context) any {
	claims := ExtractClaims(c)
	return claims[ClaimAccountID]
}

func Unauthorized(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"message": message})
}

func JWTMiddlewareWithTokenIssuance(realm, issuer string, authenticatorFn func(c *gin.Context) (any, error),
	audiencePrivateKey, rsaPrivateKeyFile, rsaPublicKeyFile string, timeout time.Duration, timeFunc func() time.Time) (*GinJWTMiddleware, error) {
	audiencePrivateKeyBytes, err := ReadAudiencePrivateKeyFromString(audiencePrivateKey)
	if err != nil {
		return nil, err
	}

	return New(&GinJWTMiddleware{
		Realm:            realm,
		Issuer:           issuer,
		SigningAlgorithm: "RS256",
		PubKeyFile:       rsaPublicKeyFile,
		PrivKeyFile:      rsaPrivateKeyFile,
		Timeout:          timeout,
		MaxRefresh:       timeout,
		Authenticator:    authenticatorFn,
		Authorizator:     AuthorisationHandler(audiencePrivateKeyBytes),
		IdentityHandler:  IdentityHandler,
		IdentityKey:      UserIDKey,
		PayloadFunc:      Payload,
		Unauthorized:     Unauthorized,
		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: timeFunc,
	})
}

func GetKeyID(key []byte) string {
	hash := sha256.Sum256(key)
	return base58.Encode(hash[0:4])
}

func ReadAudiencePrivateKeyFromString(val string) ([]byte, error) {
	audiencePrivateKeyBytes, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return nil, err
	}
	if len(audiencePrivateKeyBytes) != 64 {
		return nil, errors.New("bad audience private key")
	}

	return audiencePrivateKeyBytes, nil
}
