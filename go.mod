module github.com/piprate/metalocker

go 1.18

replace github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.3

require (
	github.com/aphistic/golf v0.0.0-20180712155816-02c07f170c5a
	github.com/btcsuite/btcd v0.23.4
	github.com/btcsuite/btcd/btcec/v2 v2.3.2
	github.com/btcsuite/btcd/btcutil v1.1.3
	github.com/bytedance/sonic v1.6.0
	github.com/claudiu/gocron v0.0.0-20151103142354-980c96bf412b
	github.com/cskr/pubsub v1.0.2
	github.com/fergusstrange/embedded-postgres v1.19.0
	github.com/gabriel-vasile/mimetype v1.4.1
	github.com/gin-contrib/cors v1.4.0
	github.com/gin-gonic/gin v1.8.2
	github.com/gobuffalo/packr/v2 v2.8.3
	github.com/golang-jwt/jwt/v4 v4.4.3
	github.com/golang-migrate/migrate/v4 v4.15.0
	github.com/golang/mock v1.6.0
	github.com/gorilla/websocket v1.5.0
	github.com/hashicorp/vault/api v1.1.0
	github.com/jamesruan/sodium v0.0.0-20181216154042-9620b83ffeae
	github.com/muesli/cache2go v0.0.0-20221011235721-518229cd8021
	github.com/multiformats/go-multihash v0.0.16
	github.com/olekukonko/tablewriter v0.0.5
	github.com/piprate/json-gold v0.5.0
	github.com/piprate/restgate v0.0.0-20190903092639-61855bc1bc5f
	github.com/rs/zerolog v1.28.0
	github.com/spf13/viper v1.14.0
	github.com/stretchr/testify v1.8.1
	github.com/tyler-smith/go-bip39 v1.1.0
	github.com/urfave/cli/v2 v2.23.7
	github.com/xlab/treeprint v1.1.0
	go.etcd.io/bbolt v1.3.6
	golang.org/x/crypto v0.4.0
	golang.org/x/term v0.3.0
)

require (
	github.com/aphistic/sweet v0.3.0 // indirect
	github.com/btcsuite/btcd/chaincfg/chainhash v1.0.1 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-playground/validator/v10 v10.11.1 // indirect
	github.com/gobuffalo/logger v1.0.6 // indirect
	github.com/gobuffalo/packd v1.0.1 // indirect
	github.com/goccy/go-json v0.9.11 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.0 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/vault/sdk v0.1.14-0.20200519221838-e0cfd64bc267 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/karrick/godirwalk v1.16.1 // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lib/pq v1.10.4 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/markbates/errx v1.1.0 // indirect
	github.com/markbates/oncer v1.0.0 // indirect
	github.com/markbates/safe v1.0.1 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-varint v0.0.6 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.6 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pjebs/jsonerror v0.0.0-20190614034432-63ef9a8df848 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/afero v1.9.2 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.4.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.7 // indirect
	github.com/unrolled/render v1.0.1 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/arch v0.0.0-20210923205945-b76863e36670 // indirect
	golang.org/x/net v0.4.0 // indirect
	golang.org/x/sys v0.3.0 // indirect
	golang.org/x/text v0.5.0 // indirect
	golang.org/x/time v0.0.0-20220609170525-579cf78fd858 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
