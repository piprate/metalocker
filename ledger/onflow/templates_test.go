package onflow_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/onflow/cadence"
	. "github.com/piprate/metalocker/ledger/onflow"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Stamp})
}

func TestNewEngine_emulator(t *testing.T) {
	client, err := NewInMemoryConnectorEmbedded(false)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = client.CreateAccountsE(ctx, "emulator-account")
	require.NoError(t, err)

	err = client.InitializeContractsE(ctx)
	require.NoError(t, err)

	_, err = NewTemplateEngine(client)
	require.NoError(t, err)
}

func TestNewEngine_emulatorWithFees(t *testing.T) {
	client, err := NewInMemoryConnectorEmbedded(true)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = client.DoNotPrependNetworkToAccountNames().CreateAccountsE(ctx, "emulator-account")
	require.NoError(t, err)

	te, err := NewTemplateEngine(client)
	require.NoError(t, err)

	adminAcct := client.Account("emulator-metalocker-admin")
	_ = te.NewTransaction("account_fund_flow").
		Argument(cadence.NewAddress(adminAcct.Address)).
		UFix64Argument("1000.0").
		SignProposeAndPayAsService().
		Test(t).
		AssertSuccess()

	err = client.InitializeContractsE(ctx)
	require.NoError(t, err)
}

func TestNewEngine_testnet(t *testing.T) {
	client, err := NewNetworkConnectorEmbedded("testnet")
	require.NoError(t, err)

	_, err = NewTemplateEngine(client)
	require.NoError(t, err)
}

func TestNewEngine_mainnet(t *testing.T) {
	client, err := NewNetworkConnectorEmbedded("mainnet")
	require.NoError(t, err)

	_, err = NewTemplateEngine(client)
	require.NoError(t, err)
}
