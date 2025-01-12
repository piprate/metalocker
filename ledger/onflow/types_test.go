package onflow_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/onflow/flow-go-sdk"
	. "github.com/piprate/metalocker/ledger/onflow"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Stamp})
}

func TestRecordToCadence(t *testing.T) {
	var rec *model.Record
	require.NoError(t, jsonw.Decode(strings.NewReader(`
{
 "id": "47Gv8nkNgEr95gRZd98c4NAhi4wszmUBk6tytk4TtVCo",
 "routingKey": "db5ryf4F6JQjD3aRqMF13RdueHkeHU9WcAx9dEVfy1mq",
 "keyIndex": 4,
 "operationType": 1,
 "address": "CW1bDDn1Pp3W486rVxkrHKLUTirqytPMjBfFfXrqihU8",
 "ac": "5z//FjtYNAChMv+8dpMqIiM4uKkGv0BXWx2yu6avIgE=",
 "rc": "pVbAvIkXtnzmF7z3U6n4GVQwgjjhQgvw/bomh1Fm72o=",
 "rcType": 1,
 "dataAssets": [
   "Gp7ZX5S5NECqoMz3u7micC7TuhvWB9z8ErAkympVgt4A"
 ],
 "signature": "AN1rKvtoGVPioTYojunoBTWWx6MTdb8VHB8vN8takPu4fpe5K7LFowgqW8meuLwniFosM2xCBbHkdhqUoTovr4xDCzKF2JNZb"
}
`), &rec))
	require.NoError(t, rec.Validate())

	testAddr := flow.HexToAddress("179b6b1cb6755e31")
	c := RecordToCadence(rec, testAddr)
	require.NotNil(t, c)

	rec2, err := RecordFromCadence(c)
	require.NoError(t, err)

	require.NoError(t, rec2.Validate())
}
