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

package utils

import (
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"
	"go.etcd.io/bbolt"
)

type BoltClient struct {
	DB *bbolt.DB
}

type InstallSchemaFunc func(bc *BoltClient) error

func NewBoltClient(boltDBFile string, installSchemaFunc InstallSchemaFunc) (*BoltClient, error) {
	db, err := bbolt.Open(boltDBFile, 0o600, nil)
	if err != nil {
		log.Err(err).Msg("Failed to open Bold DB")
		return nil, err
	}

	bc := &BoltClient{db}

	if err := installSchemaFunc(bc); err != nil {
		log.Err(err).Msg("Failed to install BoltDB schema")
		return nil, err
	}

	return bc, nil
}

func (bc *BoltClient) Close() error {
	return bc.DB.Close()
}

func (bc *BoltClient) FetchBytes(bucket, key string) ([]byte, error) {
	var val []byte
	err := bc.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}

		val = b.Get([]byte(key))

		return nil
	})
	return val, err
}

func (bc *BoltClient) FetchString(bucket, key string) (string, error) {
	bval, err := bc.FetchBytes(bucket, key)
	if err != nil {
		return "", err
	}
	return string(bval), nil
}

func (bc *BoltClient) FetchInt(bucket, key string) (int, error) {
	v, err := bc.FetchString(bucket, key)
	if err != nil {
		return 0, err
	}
	if v == "" {
		return 0, nil
	} else {
		return strconv.Atoi(v)
	}
}

func (bc *BoltClient) Update(bucket, key string, value []byte) error {
	return bc.DB.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}

		err := b.Put([]byte(key), value)
		if err != nil {
			return err
		}
		return nil
	})
}

func (bc *BoltClient) UpdateInline(tx *bbolt.Tx, bucket, key string, value []byte) error {
	b := tx.Bucket([]byte(bucket))
	if b == nil {
		return fmt.Errorf("bucket %s not found", bucket)
	}

	err := b.Put([]byte(key), value)
	if err != nil {
		return err
	}
	return nil
}

func (bc *BoltClient) UpdateInt64(bucket, key string, value int64) error {
	return bc.Update(bucket, key, []byte(strconv.FormatInt(value, 10)))
}
