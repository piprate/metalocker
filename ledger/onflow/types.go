package onflow

import (
	"encoding/base64"
	"errors"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/common"
	"github.com/onflow/flow-go-sdk"
	"github.com/piprate/metalocker/model"
)

func RecordFromCadence(val cadence.Value) (*model.Record, error) {
	if opt, ok := val.(cadence.Optional); ok {
		if opt.Value == nil {
			return nil, nil
		}
		val = opt.Value
	}

	if val == nil {
		return nil, nil
	}

	valStruct, ok := val.(cadence.Struct)
	if !ok || valStruct.StructType.QualifiedIdentifier != "MetaLocker.Record" || len(valStruct.FieldsMappedByName()) != 20 {
		return nil, errors.New("bad Record value")
	}

	allFields := valStruct.FieldsMappedByName()

	res := &model.Record{
		ID:                        string(allFields["id"].(cadence.String)),
		RoutingKey:                string(allFields["routingKey"].(cadence.String)),
		KeyIndex:                  uint32(allFields["keyIndex"].(cadence.UInt32)),
		Operation:                 model.OpType(allFields["operationType"].(cadence.UInt32)),
		OperationAddress:          string(allFields["address"].(cadence.String)),
		Flags:                     uint32(allFields["flags"].(cadence.UInt32)),
		AuthorisingCommitment:     cadenceArrayToBase64String(allFields["ac"].(cadence.Array)),
		AuthorisingCommitmentType: uint8(allFields["acType"].(cadence.UInt8)),
		RequestingCommitment:      cadenceArrayToBase64String(allFields["rc"].(cadence.Array)),
		RequestingCommitmentType:  uint8(allFields["rcType"].(cadence.UInt8)),
		ImpressionCommitment:      cadenceArrayToBase64String(allFields["ic"].(cadence.Array)),
		ImpressionCommitmentType:  uint8(allFields["icType"].(cadence.UInt8)),
		SubjectRecord:             string(allFields["subjectRecord"].(cadence.String)),
		HeadID:                    string(allFields["headID"].(cadence.String)),
		HeadBody:                  string(allFields["headBody"].(cadence.String)),
		Signature:                 string(allFields["signature"].(cadence.String)),
		Status:                    model.RecordStatus(allFields["status"].(cadence.String)),
	}

	dataAssets, ok := allFields["dataAssets"].(cadence.Array)
	if !ok {
		return nil, errors.New("bad dataAssets value")
	}
	for _, stringVal := range dataAssets.Values {
		res.DataAssets = append(res.DataAssets, string(stringVal.(cadence.String)))
	}
	revocationProof, ok := allFields["revocationProof"].(cadence.Array)
	if !ok {
		return nil, errors.New("bad revocationProof value")
	}
	for _, uint8ArrayVal := range revocationProof.Values {
		res.RevocationProof = append(res.RevocationProof, cadenceArrayToBase64String(uint8ArrayVal.(cadence.Array)))
	}

	return res, nil
}

func stringArrayToCadence(array []string) cadence.Array {
	res := make([]cadence.Value, len(array))
	for i, val := range array {
		res[i] = cadence.String(val)
	}
	return cadence.NewArray(res)
}

func base64StringToCadence(s string) cadence.Array {
	valBytes, _ := base64.StdEncoding.DecodeString(s)
	res := make([]cadence.Value, len(valBytes))
	for i, val := range valBytes {
		res[i] = cadence.UInt8(val)
	}
	return cadence.NewArray(res)
}

func cadenceArrayToBase64String(val cadence.Array) string {
	byteArray := make([]byte, len(val.Values))
	for i, int8Val := range val.Values {
		byteArray[i] = byte(int8Val.(cadence.UInt8))
	}
	return base64.StdEncoding.EncodeToString(byteArray)
}

func base64StringArrayToCadence(sa []string) cadence.Array {
	res := make([]cadence.Value, len(sa))
	for i, s := range sa {
		res[i] = base64StringToCadence(s)
	}

	return cadence.NewArray(res)
}

func RecordToCadence(rec *model.Record, contractAddr flow.Address) cadence.Value {
	return cadence.NewStruct([]cadence.Value{
		cadence.String(rec.ID),
		cadence.String(rec.RoutingKey),
		cadence.UInt32(rec.KeyIndex),
		cadence.UInt32(rec.Operation),
		cadence.String(rec.OperationAddress),
		cadence.UInt32(rec.Flags),
		base64StringToCadence(rec.AuthorisingCommitment),
		cadence.UInt8(rec.AuthorisingCommitmentType),
		base64StringToCadence(rec.RequestingCommitment),
		cadence.UInt8(rec.RequestingCommitmentType),
		base64StringToCadence(rec.ImpressionCommitment),
		cadence.UInt8(rec.ImpressionCommitmentType),
		stringArrayToCadence(rec.DataAssets),
		cadence.String(rec.SubjectRecord),
		base64StringArrayToCadence(rec.RevocationProof),
		cadence.String(rec.HeadID),
		cadence.String(rec.HeadBody),
		cadence.String(rec.Signature),
		cadence.String(rec.Status),
		cadence.UInt64(0),
	}).WithType(cadence.NewStructType(
		common.AddressLocation{
			Address: common.Address(contractAddr),
			Name:    common.AddressLocationPrefix,
		},
		"MetaLocker.Record",
		recordCadenceFields,
		nil,
	))
}

var recordCadenceFields = []cadence.Field{
	{
		Identifier: "id",
		Type:       cadence.StringType,
	},
	{
		Identifier: "routingKey",
		Type:       cadence.StringType,
	},
	{
		Identifier: "keyIndex",
		Type:       cadence.UInt32Type,
	},
	{
		Identifier: "operationType",
		Type:       cadence.UInt32Type,
	},
	{
		Identifier: "address",
		Type:       cadence.StringType,
	},
	{
		Identifier: "flags",
		Type:       cadence.UInt32Type,
	},
	{
		Identifier: "ac",
		Type:       cadence.NewVariableSizedArrayType(cadence.UInt8Type),
	},
	{
		Identifier: "acType",
		Type:       cadence.UInt8Type,
	},
	{
		Identifier: "rc",
		Type:       cadence.NewVariableSizedArrayType(cadence.UInt8Type),
	},
	{
		Identifier: "rcType",
		Type:       cadence.UInt8Type,
	}, {
		Identifier: "ic",
		Type:       cadence.NewVariableSizedArrayType(cadence.UInt8Type),
	},
	{
		Identifier: "icType",
		Type:       cadence.UInt8Type,
	},
	{
		Identifier: "dataAssets",
		Type:       cadence.NewVariableSizedArrayType(cadence.StringType),
	},
	{
		Identifier: "subjectRecord",
		Type:       cadence.StringType,
	},
	{
		Identifier: "revocationProof",
		Type:       cadence.NewVariableSizedArrayType(cadence.NewVariableSizedArrayType(cadence.UInt8Type)),
	},
	{
		Identifier: "headID",
		Type:       cadence.StringType,
	},
	{
		Identifier: "headBody",
		Type:       cadence.StringType,
	},
	{
		Identifier: "signature",
		Type:       cadence.StringType,
	},
	{
		Identifier: "status",
		Type:       cadence.StringType,
	},
	{
		Identifier: "blockNumber",
		Type:       cadence.UInt64Type,
	},
}

func BlockFromCadence(val cadence.Value) (*model.Block, error) {
	if opt, ok := val.(cadence.Optional); ok {
		if opt.Value == nil {
			return nil, nil
		}
		val = opt.Value
	}

	valStruct, ok := val.(cadence.Struct)
	if !ok || valStruct.StructType.QualifiedIdentifier != "MetaLocker.MetaBlock" {
		return nil, errors.New("bad MetaBlock value")
	}

	allFields := valStruct.FieldsMappedByName()

	return &model.Block{
		Number:     uint64(allFields["number"].(cadence.UInt64)),
		Hash:       string(allFields["hash"].(cadence.String)),
		ParentHash: string(allFields["parentHash"].(cadence.String)),
		Status:     model.BlockStatus(allFields["status"].(cadence.UInt8)),
	}, nil
}

func BlockRecordsFromCadence(val cadence.Value) ([][]string, error) {
	if opt, ok := val.(cadence.Optional); ok {
		if opt.Value == nil {
			return nil, nil
		}
		val = opt.Value
	}

	valStruct, ok := val.(cadence.Struct)
	if !ok || valStruct.StructType.QualifiedIdentifier != "MetaLocker.MetaBlock" {
		return nil, errors.New("bad MetaBlock value")
	}

	recsField := valStruct.SearchFieldByName("records").(cadence.Array)

	records := make([][]string, len(recsField.Values))
	for i, elem := range recsField.Values {
		arrayVal := elem.(cadence.Array)
		rec := make([]string, len(arrayVal.Values))
		for j, val := range arrayVal.Values {
			rec[j] = string(val.(cadence.String))
		}
		records[i] = rec
	}

	return records, nil
}
