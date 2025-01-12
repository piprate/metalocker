package emulator

import (
	"context"
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/piprate/metalocker/ledger/onflow"
	"github.com/piprate/splash"
	"github.com/stretchr/testify/require"
)

func ConfigureInMemoryEmulator(t *testing.T, client *splash.Connector, adminAccountName string, adminFlowDeposit string) {
	t.Helper()

	_, err := client.DoNotPrependNetworkToAccountNames().CreateAccountsE(context.Background(), "emulator-account")
	require.NoError(t, err)

	if adminFlowDeposit != "" {
		adminAcct := client.Account(adminAccountName)

		se, err := onflow.NewTemplateEngine(client)
		require.NoError(t, err)

		FundAccountWithFlow(t, se, adminAcct.Address, adminFlowDeposit)
	}

	err = client.InitializeContractsE(context.Background())
	require.NoError(t, err)
}

func AddLedgerPoolKeys(t *testing.T, client *splash.Connector, srcAccountName string, srcKeyIndex, numKeys int) {
	t.Helper()

	tx := client.Transaction(`
transaction(srcIndex: Int, numKeys: Int) {

    prepare(signer: auth(AddContract, AddKey) &Account) {
        // copy the main key with 0 weight multiple times
        // to create the required number of keys
        let key = signer.keys.get(keyIndex: srcIndex)!
        var count: Int = 0
        while count < numKeys {
            signer.keys.add(
                publicKey: key.publicKey,
                hashAlgorithm: key.hashAlgorithm,
                weight: 0.0
            )
            count = count + 1
        }
    }
}`).
		IntArgument(srcKeyIndex).
		IntArgument(numKeys).
		SignProposeAndPayAs(srcAccountName).
		Test(t).
		AssertSuccess()
	require.NoError(t, tx.Err)
}

func FundAccountWithFlow(t *testing.T, se *splash.TemplateEngine, receiverAddress flow.Address, amount string) {
	t.Helper()

	tx := se.NewTransaction("account_fund_flow").
		Argument(cadence.NewAddress(receiverAddress)).
		UFix64Argument(amount).
		SignProposeAndPayAsService().
		Test(t).
		AssertSuccess()
	require.NoError(t, tx.Err)
}

func GetFlowBalance(t *testing.T, se *splash.TemplateEngine, address flow.Address) float64 {
	t.Helper()

	v, err := se.NewScript("account_balance_flow").
		Argument(cadence.NewAddress(address)).
		RunReturns(context.Background())
	require.NoError(t, err)

	return splash.ToFloat64(v)
}
