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

package actions

import (
	"fmt"
	"strings"
	"time"

	"github.com/piprate/metalocker/cmd/metalo/datatypes"
	"github.com/piprate/metalocker/cmd/metalo/operations"
	"github.com/piprate/metalocker/model/expiry"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func checkLeaseDuration(s string) error {
	_, err := expiry.FromDateErr(time.Now(), s)
	if err != nil {
		return cli.Exit(err, OperationFailed)
	}
	return nil
}

func StoreDataSet(c *cli.Context) error {
	if c.Args().Len() != 1 {
		fmt.Print("Please specify the path to file or folder.\n\n")
		return cli.Exit("please specify the path to file or folder", InvalidParameter)
	}

	vaultName := c.String("vault")
	lockerID := c.String("locker")
	provPath := c.String("prov")
	provMapping := c.String("prov-mapping")
	parentRecordID := extractRecordID(c.String("parent"))
	metaType := c.String("type")
	leaseDuration := c.String("expiration")
	waitForConfirmation := c.Bool("wait")

	if lockerID == "" {
		return cli.Exit("locker not specified", InvalidParameter)
	}

	if err := checkLeaseDuration(leaseDuration); err != nil {
		return err
	}

	dw, err := LoadRemoteDataWallet(c, true)
	if err != nil {
		return err
	}

	recID, impID, err := operations.StoreDataSet(c.Context, dw.DataStore(), c.Args().Get(0), metaType, vaultName, lockerID, provPath, provMapping,
		parentRecordID, leaseDuration, waitForConfirmation)
	if err != nil {
		log.Err(err).Msg("Data set upload failed")
		return cli.Exit(err, OperationFailed)
	} else {
		fmt.Printf("%s,%s\n", recID, impID)
	}
	return nil
}

func StoreDataSets(c *cli.Context) error {
	if c.Args().Len() != 1 {
		fmt.Print("Please specify the path to folder.\n\n")
		return cli.Exit("please specify the path to folder", InvalidParameter)
	}

	vaultName := c.String("vault")
	lockerID := c.String("locker")
	provPath := c.String("prov")
	provMapping := c.String("prov-mapping")
	metaType := c.String("type")
	leaseDuration := c.String("expiration")
	waitForConfirmation := c.Bool("wait")

	if err := checkLeaseDuration(leaseDuration); err != nil {
		return err
	}

	dw, err := LoadRemoteDataWallet(c, true)
	if err != nil {
		return err
	}

	idList, err := operations.StoreDataSets(c.Context, dw.DataStore(), c.Args().Get(0), metaType, vaultName,
		lockerID, provPath, provMapping, leaseDuration, waitForConfirmation)
	if err != nil {
		log.Err(err).Msg("Data set upload failed")
		return cli.Exit(err, OperationFailed)
	} else {
		for _, id := range idList {
			fmt.Printf("%s\n", id)
		}
	}
	return nil
}

func ShareDataSet(c *cli.Context) error {
	if c.Args().Len() != 1 {
		return cli.Exit("please specify the record id to share", InvalidParameter)
	}

	vaultName := c.String("vault")
	lockerID := c.String("locker")
	leaseDuration := c.String("expiration")
	waitForConfirmation := c.Bool("wait")

	if err := checkLeaseDuration(leaseDuration); err != nil {
		return err
	}

	dw, err := LoadRemoteDataWallet(c, true)
	if err != nil {
		return err
	}

	recID := extractRecordID(c.Args().Get(0))

	sourceDS, err := dw.DataStore().Load(c.Context, recID)
	if err != nil {
		return err
	}

	locker, err := dw.GetLocker(c.Context, lockerID)
	if err != nil {
		return err
	}

	f := dw.DataStore().Share(c.Context, sourceDS, locker, vaultName, expiry.FromNow(leaseDuration))
	if waitForConfirmation {
		err = f.Wait(60 * time.Second)
	} else {
		err = f.Error()
	}
	if err != nil {
		log.Err(err).Msg("Data set share failed")
		return cli.Exit(err, OperationFailed)
	} else {
		fmt.Printf("%s\n", f.ID())
	}
	return nil
}

func RevokeLease(c *cli.Context) error {
	if c.Args().Len() != 1 {
		return cli.Exit("please specify record id", InvalidParameter)
	}

	waitForConfirmation := c.Bool("wait")

	dw, err := LoadRemoteDataWallet(c, true)
	if err != nil {
		return err
	}

	recID := c.Args().Get(0)

	f := dw.DataStore().Revoke(c.Context, recID)

	if waitForConfirmation {
		err = f.Wait(60 * time.Second)
	} else {
		err = f.Error()
	}
	if err != nil {
		log.Err(err).Msg("Lease revocation failed")
		return cli.Exit(err, OperationFailed)
	}
	return nil
}

func GetDataSet(c *cli.Context) error {
	if c.Args().Len() != 1 {
		return cli.Exit("please specify record id", InvalidParameter)
	}

	recordID := extractRecordID(c.Args().Get(0))
	dest := c.String("dest")
	writeMetaData := c.Bool("metadata")
	syncIndex := c.Bool("sync")

	dw, err := LoadRemoteDataWallet(c, syncIndex)
	if err != nil {
		return err
	}

	err = operations.GetDataSet(c.Context, dw, recordID, dest, writeMetaData)
	if err != nil {
		log.Err(err).Msg("Data set retrieval failed")
		return cli.Exit(err, OperationFailed)
	}
	return nil
}

func ListSupportedDataTypes(c *cli.Context) error {
	for _, dt := range datatypes.SupportedDataTypes() {
		println(dt)
	}
	return nil
}

func extractRecordID(id string) string {
	splitID := strings.Split(id, ",")
	if len(splitID) == 2 {
		id = splitID[0]
	}
	return id
}
