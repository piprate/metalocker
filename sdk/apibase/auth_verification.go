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
	"encoding/base64"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/model"
)

var (
	ErrFailedToDecodeClientSecret  = errors.New("failed to decode client secret")
	ErrFailedToDecryptClientSecret = errors.New("failed to decrypt client secret")
)

func ExtractSecret(claims MapClaims, audiencePrivateKey []byte) (*model.AESKey, error) {

	if secretStr, exists := claims[ClaimEncryptedAudienceSecret]; exists {
		encryptedSecret, err := base64.StdEncoding.DecodeString(secretStr.(string))
		if err != nil {
			return nil, ErrFailedToDecodeClientSecret
		}
		secret, err := model.AnonDecrypt(encryptedSecret, audiencePrivateKey)
		if err != nil {
			return nil, ErrFailedToDecryptClientSecret
		}

		return model.NewAESKey(secret), nil
	} else if secretStr, exists = claims[ClaimLegacySecret]; exists {
		secret, err := base64.StdEncoding.DecodeString(secretStr.(string))
		if err != nil {
			return nil, ErrFailedToDecodeClientSecret
		}
		return model.NewAESKey(secret), nil
	}

	return nil, nil
}

func AuthorisationHandler(audiencePrivateKey []byte) func(data any, c *gin.Context) bool {
	return func(data any, c *gin.Context) bool {
		if claims, exists := c.Get("JWT_PAYLOAD"); exists {
			secret, err := ExtractSecret(claims.(MapClaims), audiencePrivateKey)
			if err != nil {
				return false
			}
			if secret != nil {
				c.Set(ClientSecretKey, secret)
			}
			return true
		} else {
			return false
		}
	}
}

func JWTMiddlewareWithTokenVerification(realm string, audiencePrivateKey string, rsaPublicKeyFile string, timeFunc func() time.Time) (*GinJWTMiddleware, error) {
	audiencePrivateKeyBytes, err := ReadAudiencePrivateKeyFromString(audiencePrivateKey)
	if err != nil {
		return nil, err
	}

	return New(&GinJWTMiddleware{
		Realm:            realm,
		SigningAlgorithm: "RS256",
		PubKeyFile:       rsaPublicKeyFile,
		MaxRefresh:       time.Hour,
		Authorizator:     AuthorisationHandler(audiencePrivateKeyBytes),
		IdentityHandler:  IdentityHandler,
		IdentityKey:      UserIDKey,
		PayloadFunc:      Payload,
		Unauthorized:     Unauthorized,
		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: timeFunc,
	})
}
