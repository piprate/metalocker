package onflow

import (
	"embed"

	"github.com/onflow/flowkit/v2/config"
	"github.com/piprate/splash"
)

//go:embed contracts
//go:embed flow.json
var testdataFS embed.FS

// NewNetworkConnectorEmbedded creates a new Splash Connector that uses embedded Flow configuration.
func NewNetworkConnectorEmbedded(network string) (*splash.Connector, error) {
	return splash.NewNetworkConnector([]string{config.DefaultPath}, splash.NewEmbedLoader(&testdataFS), network, splash.NewZeroLogger())
}

// NewInMemoryConnectorEmbedded creates a new Splash Connector for in-memory emulator that uses embedded Flow configuration.
func NewInMemoryConnectorEmbedded(enableTxFees bool) (*splash.Connector, error) {
	return splash.NewInMemoryConnector([]string{config.DefaultPath}, splash.NewEmbedLoader(&testdataFS), enableTxFees, splash.NewZeroLogger())
}
