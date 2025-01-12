package onflow

import (
	"embed"

	"github.com/piprate/splash"
)

//go:embed templates
var templateFS embed.FS

var (
	requiredWellKnownContracts = []string{
		"Burner", "FlowToken", "FungibleToken", "FungibleTokenMetadataViews",
		"FungibleTokenSwitchboard", "MetadataViews",
		"NonFungibleToken", "MetaLocker",
	}
)

func NewTemplateEngine(client *splash.Connector) (*splash.TemplateEngine, error) {
	return splash.NewTemplateEngine(client, templateFS, []string{"templates/transactions", "templates/scripts"}, requiredWellKnownContracts)
}
