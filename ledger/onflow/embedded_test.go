package onflow_test

import (
	"context"
	"testing"

	. "github.com/piprate/metalocker/ledger/onflow"
	"github.com/stretchr/testify/require"
)

func TestNewGoWithTheFlowEmbedded(t *testing.T) {
	client, err := NewInMemoryConnectorEmbedded(false)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = client.CreateAccountsE(ctx, "emulator-account")
	require.NoError(t, err)

	err = client.InitializeContractsE(ctx)
	require.NoError(t, err)
}
