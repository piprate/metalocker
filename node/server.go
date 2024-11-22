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

package node

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/knadh/koanf"
	"github.com/piprate/metalocker/contexts"
	"github.com/piprate/metalocker/ledger"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/node/api"
	"github.com/piprate/metalocker/node/api/admin"
	"github.com/piprate/metalocker/node/vaultapi"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/sdk/cmdbase"
	"github.com/piprate/metalocker/services/notification"
	"github.com/piprate/metalocker/storage"
	"github.com/piprate/metalocker/utils"
	"github.com/piprate/metalocker/utils/security"
	"github.com/piprate/metalocker/vaults"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type (
	MetaLockerServer struct {
		ServerControls api.ServerControls
		ConfigDir      string
		Warden         *utils.GracefulWarden
		Resolver       *cmdbase.SecureParameterResolver
		JWTMiddleware  *apibase.GinJWTMiddleware
		Level1AuthFn   gin.HandlerFunc
		Level2AuthFn   gin.HandlerFunc

		IdentityBackend storage.IdentityBackend
		OffChainVault   vaults.Vault
		Ledger          model.Ledger
		BlobManager     *vaults.LocalBlobManager
		NS              notification.Service
		Router          *gin.Engine

		httpServer *http.Server

		protocol string
		host     string
		port     int
		baseURI  string
	}
)

func NewMetaLockerServer(configDir string) *MetaLockerServer {
	mls := &MetaLockerServer{
		ServerControls: api.ServerControls{
			Status:          "ok",
			MaintenanceMode: false,
		},
		ConfigDir: configDir,
	}

	return mls
}

func (mls *MetaLockerServer) InitServices(ctx context.Context, cfg *koanf.Koanf, debugMode bool) error {
	mls.Warden = utils.NewGracefulWarden(120)

	prodMode := cfg.Bool("production") && !debugMode

	// set up logging
	logWriter, err := apibase.SetupLogging(cfg, prodMode)
	if err != nil {
		log.Err(err).Msg("Failed to configure logging")
		return cli.Exit(err, 1)
	}
	mls.Warden.CloseOnShutdown(logWriter)

	// preload JSON-LD context into memory
	_ = contexts.PreloadContextsIntoMemory()

	if cfg.Bool("https") {
		mls.protocol = "https"
	} else {
		mls.protocol = "http"
	}

	mls.host = cfg.String("host")
	mls.port = cfg.Int("port")
	mls.baseURI = cfg.String("nodeURI")

	log.Info().Strs("args", os.Args).Str("go_version", runtime.Version()).
		Str("os", runtime.GOOS).Str("arch", runtime.GOARCH).Str("uri", mls.baseURI).
		Bool("prod_mode", prodMode).
		Msg("Starting new instance")

	mls.Resolver, err = cmdbase.ConfigureParameterResolver(cfg, mls.ConfigDir)
	if err != nil {
		log.Err(err).Msg("Failed to create parameter resolver")
		return cli.Exit(err, 1)
	}
	mls.Warden.CloseOnShutdown(mls.Resolver)

	// initialise identity backend

	mls.IdentityBackend, err = InitIdentityBackend(cfg, mls.Resolver)
	if err != nil {
		return err
	}
	mls.Warden.CloseOnShutdown(mls.IdentityBackend)

	// initialise notification service

	mls.NS = notification.NewLocalNotificationService(0)

	// initialise off-chain storage

	mls.OffChainVault, err = InitOffChainStorage(cfg, mls.Resolver)
	if err != nil {
		return err
	}
	mls.Warden.CloseOnShutdown(mls.OffChainVault)

	// initialise ledger connector

	mls.Ledger, err = InitLedger(ctx, cfg, mls.Resolver, mls.NS)
	if err != nil {
		return err
	}
	mls.Warden.CloseOnShutdown(mls.Ledger)

	// initialise vaults

	mls.BlobManager, err = InitVaults(cfg, mls.Resolver, mls.Ledger, mls.Warden)
	if err != nil {
		return err
	}

	// initialise router

	mls.Router = InitRouter(
		DefaultCORSConfig(
			cfg.Strings("allowedHttpOrigins"),
		),
	)

	return nil
}

func (mls *MetaLockerServer) InitAuthentication(ctx context.Context, cfg *koanf.Koanf) error {
	authMiddleware, publicKeyBytes, err := InitAuthMiddleware(cfg, "MetaLocker", //nolint:contextcheck
		mls.Resolver, mls.ConfigDir, mls.IdentityBackend)
	if err != nil {
		return err
	}
	mls.JWTMiddleware = authMiddleware
	mls.ServerControls.JWTPublicKey = string(publicKeyBytes)

	mls.Level1AuthFn = apibase.AccessKeyMiddleware(mls.IdentityBackend, authMiddleware.MiddlewareFunc()) //nolint:contextcheck

	if cfg.Exists("apiKeys") {
		mls.Level2AuthFn, err = apibase.NewStaticAPIKeyAuthenticationHandler(ctx, cfg, "apiKeys", mls.Level1AuthFn,
			mls.Resolver, mls.IdentityBackend)
		if err != nil {
			return cli.Exit(err, 1)
		}
	} else {
		mls.Level2AuthFn = mls.Level1AuthFn
	}

	return nil
}

func (mls *MetaLockerServer) InitStandardRoutes(ctx context.Context, cfg *koanf.Koanf) error {
	r := mls.Router

	if cfg.Exists("administration") {
		adminAuthFunc, err := apibase.NewAdminAuthenticationHandler(cfg, "administration", mls.Resolver)
		if err != nil {
			return cli.Exit(err, 1)
		}
		admin.InitRoutes(r, "/v1/admin", adminAuthFunc, mls.IdentityBackend)
	}

	api.InitRegisterRoute(r, "/v1/register", cfg, mls.JWTMiddleware, mls.IdentityBackend, mls.Ledger)

	r.POST("/v1/authenticate", mls.JWTMiddleware.LoginHandler)
	r.GET("/v1/refresh-token", mls.JWTMiddleware.RefreshHandler)
	r.POST("/v1/validate-request", api.ValidateRequestSignatureHandler(mls.IdentityBackend))

	r.GET("/v1/recovery-code", api.GetRecoveryCodeHandler(mls.IdentityBackend))
	r.POST("/v1/recover-account", api.RecoverAccountHandler(mls.IdentityBackend))

	r.GET("/v1/status", api.GetStatusHandler(&mls.ServerControls, mls.Ledger))

	v1 := r.Group("/v1")
	v1.Use(mls.Level1AuthFn)
	v1.Use(apibase.ContextLoggerHandler)

	api.InitAccountRoutes(v1, mls.IdentityBackend)
	api.InitLedgerRoutes(v1, mls.Ledger, mls.OffChainVault)
	api.InitDIDRoutes(v1, mls.IdentityBackend)

	v1.GET("/notifications", api.NotificationChannelHandler(mls.NS))

	// initialise vaults

	vaultGrp := r.Group("/v1/vault")
	vaultGrp.Use(mls.Level2AuthFn)
	vaultGrp.Use(apibase.ContextLoggerHandler)

	vaultapi.InitRoutes(ctx, vaultGrp, mls.BlobManager)

	// serve JSON-LD contexts which are compatible with the current MetaLocker implementation.
	// This includes third-party contexts to avoid unexpected changes and round-trips over network.
	r.StaticFS("/static/model", contexts.HTTPFileSystem())

	return nil
}

func (mls *MetaLockerServer) Run(cfg *koanf.Koanf) error {
	listenAddr := fmt.Sprintf(":%d", mls.port)

	log.Info().Str("addr", listenAddr).Bool("secure", cfg.Bool("https")).Msg("Starting HTTP server")

	mls.httpServer = &http.Server{
		Addr:              listenAddr,
		Handler:           mls.Router.Handler(),
		ReadHeaderTimeout: time.Second * 30, // protect against Slowloris Attack
	}

	var httpError error
	defer func() {
		if httpError != nil && gin.IsDebugging() && !errors.Is(httpError, http.ErrServerClosed) {
			log.Err(httpError).Msg("Error when running HTTP(S) server")
		}
	}()

	if mls.protocol == "https" {

		certPath := path.Join(mls.ConfigDir, cfg.String("httpsCert"))
		log.Debug().Str("path", certPath).Msg("Cert file")

		keyPath := path.Join(mls.ConfigDir, cfg.String("httpsKey"))
		log.Debug().Str("path", keyPath).Msg("Key file")

		_, errCert := os.Stat(certPath)
		_, errKey := os.Stat(keyPath)
		if errCert != nil || errKey != nil {
			if os.IsNotExist(errCert) || os.IsNotExist(errKey) {
				err := security.GenerateCertificate(2048, make([]string, 0), time.Time{}, 0, mls.ConfigDir)
				if err != nil {
					log.Err(err).Msg("Could not generate self-signed TLS cert")
					return cli.Exit(err, 1)
				}
			} else {
				log.Error().AnErr("cert_error", errCert).AnErr("key_error", errKey).
					Msg("Could not stat cert or key")
				if errCert != nil {
					return cli.Exit(errCert, 1)
				} else {
					return cli.Exit(errKey, 1)
				}
			}
		}

		httpError = mls.httpServer.ListenAndServeTLS(certPath, keyPath)
	} else {
		httpError = mls.httpServer.ListenAndServe()
	}

	return nil
}

func (mls *MetaLockerServer) BaseURI() string {
	if mls.baseURI != "" {
		return mls.baseURI
	} else {
		return fmt.Sprintf("%s://%s:%d", mls.protocol, mls.host, mls.port)
	}
}

func (mls *MetaLockerServer) Close() error {
	if mls.httpServer != nil {
		return mls.httpServer.Close()
	}
	return nil
}

func (mls *MetaLockerServer) CloseOnShutdown(closer io.Closer) {
	mls.Warden.CloseOnShutdown(closer)
}

func InitIdentityBackend(cfg *koanf.Koanf, resolver cmdbase.ParameterResolver) (storage.IdentityBackend, error) {
	if cfg.Exists("accountStore") {
		var backendCfg storage.IdentityBackendConfig
		err := cfg.Unmarshal("accountStore", &backendCfg)
		if err != nil {
			log.Err(err).Msg("Failed to read account storage configuration")
			return nil, cli.Exit(err, 1)
		}

		identityBackend, err := storage.CreateIdentityBackend(&backendCfg, resolver)
		if err != nil {
			log.Err(err).Msg("Failed to create storage backend")
			return nil, cli.Exit(err, 1)
		}

		return identityBackend, nil
	} else {
		return nil, cli.Exit("account store not defined", 1)
	}
}

func InitOffChainStorage(cfg *koanf.Koanf, resolver cmdbase.ParameterResolver) (vaults.Vault, error) {
	var vaultCfg vaults.Config
	err := cfg.Unmarshal("offChainStore", &vaultCfg)
	if err != nil {
		log.Err(err).Msg("Failed to read vault configuration")
		return nil, cli.Exit(err, 1)
	}

	offchainAPI, err := vaults.CreateVault(&vaultCfg, resolver, nil)
	if err != nil {
		log.Err(err).Msg("Failed to create an offchain vault")
		os.Exit(1)
	}

	if !offchainAPI.CAS() {
		log.Err(err).Msg("Offchain operation vault should be a content-addressable storage")
		return nil, cli.Exit(err, 1)
	}

	return offchainAPI, nil
}

func InitLedger(ctx context.Context, cfg *koanf.Koanf, resolver cmdbase.ParameterResolver, ns notification.Service) (model.Ledger, error) {
	var ledgerAPI model.Ledger
	if cfg.Exists("ledger") {
		var ledgerCfg ledger.Config
		err := cfg.Unmarshal("ledger", &ledgerCfg)
		if err != nil {
			log.Err(err).Msg("Failed to read ledger connector configuration")
			return nil, cli.Exit(err, 1)
		}

		ledgerAPI, err = ledger.CreateLedgerConnector(ctx, &ledgerCfg, ns, resolver)
		if err != nil {
			log.Err(err).Msg("Failed to create ledger connector")
			return nil, cli.Exit(err, 1)
		}
		return ledgerAPI, nil
	} else {
		return nil, cli.Exit("ledger configuration not found", 1)
	}
}

func InitVaults(cfg *koanf.Koanf, resolver cmdbase.ParameterResolver, ledgerAPI model.Ledger, warden *utils.GracefulWarden) (*vaults.LocalBlobManager, error) {
	var vaultConfigs []*vaults.Config
	if err := cfg.Unmarshal("vaults", &vaultConfigs); err != nil {
		log.Err(err).Msg("Failed to read vault configurations")
		return nil, cli.Exit(err, 1)
	}

	lbm := vaults.NewLocalBlobManager()

	for _, cfg := range vaultConfigs {

		vault, err := vaults.CreateVault(cfg, resolver, ledgerAPI)
		if err != nil {
			log.Err(err).Msg("Failed to create a vault")
			return nil, cli.Exit(err, 1)
		}

		warden.CloseOnShutdown(vault)

		lbm.AddVault(vault, cfg)
	}

	return lbm, nil
}

func InitRouter(corsCfg *cors.Config) *gin.Engine {
	r := gin.New()
	_ = r.SetTrustedProxies(nil)

	r.RedirectTrailingSlash = true
	r.RedirectFixedPath = true

	// Middlewares

	r.Use(apibase.SetRequestLogger())
	r.Use(gin.Recovery())
	r.Use(cors.New(*corsCfg))

	return r
}

func InitAuthMiddleware(cfg *koanf.Koanf, realm string, resolver cmdbase.ParameterResolver, configDir string, identityBackend storage.IdentityBackend) (*apibase.GinJWTMiddleware, []byte, error) {
	privateKeyPath := path.Join(configDir, cfg.String("tokenPrivateKey"))
	publicKeyPath := path.Join(configDir, cfg.String("tokenPublicKey"))

	// the jwt middleware
	jwtTimeout := cfg.Int64("jwtTimeout")
	if jwtTimeout == 0 {
		// default timeout is 14 days
		jwtTimeout = 14 * 24 * 60
	}

	acceptedAudiences := cfg.Strings("acceptedAudiences")

	defaultAudience := cfg.String("defaultAudience")
	if len(acceptedAudiences) > 0 {
		if defaultAudience == "" {
			defaultAudience = acceptedAudiences[0]
		}
	} else {
		acceptedAudiences = append(acceptedAudiences, defaultAudience)
	}

	defaultAudiencePublicKey, err := resolver.ResolveString(cfg.Get("defaultAudiencePublicKey"))
	if err != nil {
		log.Error().Msg("Error reading audience public key")
		return nil, nil, cli.Exit(err, 1)
	}

	defaultAudiencePrivateKey, err := resolver.ResolveString(cfg.Get("defaultAudiencePrivateKey"))
	if err != nil {
		log.Error().Msg("Error reading audience private key")
		return nil, nil, cli.Exit(err, 1)
	}

	authMiddleware, err := apibase.JWTMiddlewareWithTokenIssuance(
		realm,
		cfg.String("issuer"),
		apibase.AuthenticationHandler(identityBackend, acceptedAudiences, defaultAudience, defaultAudiencePublicKey),
		defaultAudiencePrivateKey,
		privateKeyPath, publicKeyPath, time.Duration(jwtTimeout)*time.Minute, time.Now)
	if err != nil {
		log.Err(err).Msg("Error when initialising JWT middleware")
		return nil, nil, cli.Exit(err, 1)
	}

	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		log.Err(err).Msg("Error when reading token public key")
		return nil, nil, cli.Exit(err, 1)
	}

	return authMiddleware, publicKeyBytes, nil
}
