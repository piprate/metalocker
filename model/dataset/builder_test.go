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

package dataset_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/piprate/metalocker/contexts"
	"github.com/piprate/metalocker/model"
	. "github.com/piprate/metalocker/model/dataset"
	"github.com/piprate/metalocker/model/dataset/testdata/builder/mock"
	"github.com/piprate/metalocker/model/expiry"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testID = "Gm8tPkhqFzDoXNLTaCEfcHxdSKUmaWwRnQfy8V38NUab"
)

func init() {
	_ = contexts.PreloadContextsIntoMemory()
	testbase.SetupLogFormat()
}

func TestLeaseBuilderSimple(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilderBackend := mock.NewMockLeaseBuilderBackend(ctrl)
	bm := testbase.TestBlobManager(t, true, nil)
	creator := testbase.TestDID(t)
	locker := testbase.TestUniLocker(t)

	file, err := os.Open("testdata/builder/simple/meta.json")
	require.NoError(t, err)

	ts := time.Unix(1000, 0).UTC()

	lb, err := NewLeaseBuilder(context.Background(), mockBuilderBackend, bm, locker, creator,
		WithVault(testbase.TestVaultName),
		WithTimestamp(ts))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(file)
	require.NoError(t, err)

	res, err := lb.Build(expiry.Months(12))
	require.NoError(t, err)

	assert.NotEmpty(t, res.ExpiresAt)

	res.ID = testID
	res.ExpiresAt = &ts

	expectedConfigsBytes, err := os.ReadFile("testdata/builder/_results/simple_lease.json")
	require.NoError(t, err)

	testbase.AssertEqualJSON(t, expectedConfigsBytes, res)
}

func TestLeaseBuilderNoExpiry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilderBackend := mock.NewMockLeaseBuilderBackend(ctrl)
	bm := testbase.TestBlobManager(t, true, nil)
	creator := testbase.TestDID(t)
	locker := testbase.TestUniLocker(t)

	file, err := os.Open("testdata/builder/simple/meta.json")
	require.NoError(t, err)

	ts := time.Unix(1000, 0).UTC()

	lb, err := NewLeaseBuilder(context.Background(), mockBuilderBackend, bm, locker, creator,
		WithVault(testbase.TestVaultName),
		WithTimestamp(ts))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(file)
	require.NoError(t, err)

	res, err := lb.Build(expiry.Never())
	require.NoError(t, err)

	assert.Empty(t, res.ExpiresAt)
}

func TestLeaseBuilder_CustomMethod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilderBackend := mock.NewMockLeaseBuilderBackend(ctrl)
	bm := testbase.TestBlobManager(t, true, nil)
	creator := testbase.TestDID(t)
	locker := testbase.TestUniLocker(t)

	ts := time.Unix(1000, 0).UTC()

	lb, err := NewLeaseBuilder(context.Background(), mockBuilderBackend, bm, locker, creator,
		WithVault(testbase.TestVaultName),
		WithDIDMethod("example"),
		WithTimestamp(ts))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"type": "Test",
		"name": "Test Dataset",
	})
	require.NoError(t, err)

	res, err := lb.Build(expiry.Months(12))
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(res.Impression.Asset, "did:example:"))
}

func TestLeaseBuilder_WithProvenance_File(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilderBackend := mock.NewMockLeaseBuilderBackend(ctrl)
	bm := testbase.TestBlobManager(t, true, nil)
	creator := testbase.TestDID(t)
	locker := testbase.TestUniLocker(t)

	ts := time.Unix(1000, 0).UTC()

	lb, err := NewLeaseBuilder(context.Background(), mockBuilderBackend, bm, locker, creator,
		WithVault(testbase.TestVaultName),
		WithTimestamp(ts))
	require.NoError(t, err)

	file, err := os.Open("testdata/builder/prov_file/file1.txt")
	require.NoError(t, err)

	genTime := time.Unix(100000, 0).UTC()
	prov1 := &model.ProvEntity{
		ID:              "did:piprate:GGKW4B2zLbpxPJXE8TcD1okHFg4ymHvYbDfbHY2BthYc",
		Type:            model.ProvTypeEntity,
		GeneratedAtTime: &genTime,
		WasGeneratedBy: &model.ProvActivity{
			Type:              model.ProvTypeActivity,
			WasAssociatedWith: creator.ID,
			Algorithm:         "Capture",
		},
	}

	resID, err := lb.AddResource(file)
	require.NoError(t, err)
	assert.Equal(t, prov1.ID, resID)

	err = lb.AddProvenance(prov1.ID, prov1, false)
	require.NoError(t, err)

	file, err = os.Open("testdata/builder/prov_file/meta.json")
	require.NoError(t, err)

	prov2 := &model.ProvEntity{
		ID:              "did:piprate:45XawDeqWUAAKbgiZYpFh44yqqpbw5GxE8o21LRSctwD",
		Type:            model.ProvTypeEntity,
		GeneratedAtTime: &genTime,
		WasGeneratedBy: &model.ProvActivity{
			Type:              model.ProvTypeActivity,
			WasAssociatedWith: creator.ID,
			Algorithm:         "FileExtraction",
			Used:              prov1.ID,
			QualifiedUsage: []*model.ProvUsage{
				{
					Type:   model.ProvTypeUsage,
					Entity: prov1.ID,
					HadRole: &model.ProvRole{
						Type:  model.ProvTypeRole,
						Label: "file",
					},
				},
			},
		},
	}

	resID, err = lb.AddMetaResource(file)
	require.NoError(t, err)
	assert.Equal(t, prov2.ID, resID)

	err = lb.AddProvenance(prov2.ID, prov2, false)
	require.NoError(t, err)

	err = lb.AddProvenance("", []any{
		&model.ProvAgent{
			ID:   creator.ID,
			Type: model.ProvTypeAgent,
		},
	}, false)
	require.NoError(t, err)

	res, err := lb.Build(expiry.Months(12))
	require.NoError(t, err)

	res.ID = testID
	res.ExpiresAt = &ts

	expectedConfigsBytes, err := os.ReadFile("testdata/builder/_results/prov_file_lease.json")
	require.NoError(t, err)

	testbase.AssertEqualJSON(t, expectedConfigsBytes, res)
}

func TestLeaseBuilder_WithProvenanceTemplate_File(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilderBackend := mock.NewMockLeaseBuilderBackend(ctrl)
	bm := testbase.TestBlobManager(t, true, nil)
	creator := testbase.TestDID(t)
	locker := testbase.TestUniLocker(t)

	ts := time.Unix(1000, 0).UTC()

	lb, err := NewLeaseBuilder(context.Background(), mockBuilderBackend, bm, locker, creator,
		WithVault(testbase.TestVaultName),
		WithTimestamp(ts))
	require.NoError(t, err)

	genTime := time.Unix(100000, 0).UTC()

	prov1 := &model.ProvEntity{
		ID:              "did:piprate:GGKW4B2zLbpxPJXE8TcD1okHFg4ymHvYbDfbHY2BthYc",
		Type:            model.ProvTypeEntity,
		GeneratedAtTime: &genTime,
		WasGeneratedBy: &model.ProvActivity{
			Type:              model.ProvTypeActivity,
			WasAssociatedWith: creator.ID,
			Algorithm:         "Capture",
		},
	}

	err = lb.AddProvenance("", []any{
		prov1,
		&model.ProvEntity{
			ID:              "%%resource%%",
			Type:            model.ProvTypeEntity,
			GeneratedAtTime: &genTime,
			WasGeneratedBy: &model.ProvActivity{
				Type:              model.ProvTypeActivity,
				WasAssociatedWith: creator.ID,
				Algorithm:         "FileExtraction",
				Used:              prov1.ID,
				QualifiedUsage: []*model.ProvUsage{
					{
						Type:   model.ProvTypeUsage,
						Entity: prov1.ID,
						HadRole: &model.ProvRole{
							Type:  model.ProvTypeRole,
							Label: "file",
						},
					},
				},
			},
		},
		&model.ProvAgent{
			ID:   creator.ID,
			Type: model.ProvTypeAgent,
		},
	}, true)
	require.NoError(t, err)

	file, err := os.Open("testdata/builder/prov_file/file1.txt")
	require.NoError(t, err)

	resID, err := lb.AddResource(file)
	assert.Equal(t, prov1.ID, resID)
	require.NoError(t, err)

	err = lb.AddProvenance(prov1.ID, prov1, false)
	require.NoError(t, err)

	file, err = os.Open("testdata/builder/prov_file/meta.json")
	require.NoError(t, err)

	metaAssetID := "did:piprate:45XawDeqWUAAKbgiZYpFh44yqqpbw5GxE8o21LRSctwD"

	resID, err = lb.AddMetaResource(file)
	require.NoError(t, err)
	assert.Equal(t, metaAssetID, resID)

	res, err := lb.Build(expiry.Months(12))
	require.NoError(t, err)

	res.ID = testID
	res.ExpiresAt = &ts

	expectedConfigsBytes, err := os.ReadFile("testdata/builder/_results/prov_file_lease.json")
	require.NoError(t, err)

	testbase.AssertEqualJSON(t, expectedConfigsBytes, res)
}

func TestLeaseBuilder_WithProvenance_Document(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilderBackend := mock.NewMockLeaseBuilderBackend(ctrl)
	bm := testbase.TestBlobManager(t, true, nil)
	creator := testbase.TestDID(t)
	locker := testbase.TestUniLocker(t)

	ts := time.Unix(1000, 0).UTC()

	lb, err := NewLeaseBuilder(context.Background(), mockBuilderBackend, bm, locker, creator,
		WithVault(testbase.TestVaultName),
		WithTimestamp(ts))
	require.NoError(t, err)

	file, err := os.Open("testdata/builder/prov_document/file1.txt")
	require.NoError(t, err)

	genTime := time.Unix(100000, 0).UTC()
	prov1 := &model.ProvEntity{
		ID:              "did:piprate:GGKW4B2zLbpxPJXE8TcD1okHFg4ymHvYbDfbHY2BthYc",
		Type:            model.ProvTypeEntity,
		GeneratedAtTime: &genTime,
		WasGeneratedBy: &model.ProvActivity{
			Type:              model.ProvTypeActivity,
			WasAssociatedWith: creator.ID,
			Algorithm:         "Capture",
		},
	}

	resID, err := lb.AddResource(file)
	require.NoError(t, err)
	assert.Equal(t, prov1.ID, resID)

	err = lb.AddProvenance(prov1.ID, prov1, false)
	require.NoError(t, err)

	file, err = os.Open("testdata/builder/prov_document/meta.json")
	require.NoError(t, err)

	prov2 := &model.ProvEntity{
		ID:              "did:piprate:CoDVBcdXCHxyAuiAioyTPJZVefYf4sHpK2bv5BYviu8Q",
		Type:            model.ProvTypeEntity,
		GeneratedAtTime: &genTime,
		WasGeneratedBy: &model.ProvActivity{
			Type:              model.ProvTypeActivity,
			WasAssociatedWith: creator.ID,
			Algorithm:         "DocumentGeneration",
			Used: []string{
				"_:file-meta",
				"_:deal",
			},
			QualifiedUsage: []*model.ProvUsage{
				{
					Type: model.ProvTypeUsage,
					Entity: &model.ProvEntity{
						ID:              "_:file-meta",
						Type:            model.ProvTypeEntity,
						GeneratedAtTime: &genTime,
						WasGeneratedBy: &model.ProvActivity{
							Type:      model.ProvTypeActivity,
							Algorithm: "FileExtraction",
							Used:      prov1.ID,
							QualifiedUsage: []*model.ProvUsage{
								{
									Type:   model.ProvTypeUsage,
									Entity: prov1.ID,
									HadRole: &model.ProvRole{
										Type:  model.ProvTypeRole,
										Label: "file",
									},
								},
							},
						},
					},
					HadRole: &model.ProvRole{
						Type:  model.ProvTypeRole,
						Label: "file_meta",
					},
				},
				{
					Type:   model.ProvTypeUsage,
					Entity: "deal-impression-id",
					HadRole: &model.ProvRole{
						Type:  model.ProvTypeRole,
						Label: "deal",
					},
				},
			},
		},
	}

	resID, err = lb.AddMetaResource(file)
	assert.Equal(t, prov2.ID, resID)
	require.NoError(t, err)

	err = lb.AddProvenance(prov2.ID, prov2, false)
	require.NoError(t, err)

	err = lb.AddProvenance("", []any{
		&model.ProvAgent{
			ID:   creator.ID,
			Type: model.ProvTypeAgent,
		},
	}, false)
	require.NoError(t, err)

	res, err := lb.Build(expiry.Months(12))
	require.NoError(t, err)

	res.ID = testID
	res.ExpiresAt = &ts

	expectedConfigsBytes, err := os.ReadFile("testdata/builder/_results/prov_document_lease.json")
	require.NoError(t, err)

	testbase.AssertEqualJSON(t, expectedConfigsBytes, res)
}

func TestLeaseBuilder_WithProvenance_Document2ndRevision(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilderBackend := mock.NewMockLeaseBuilderBackend(ctrl)
	bm := testbase.TestBlobManager(t, true, nil)
	creator := testbase.TestDID(t)
	locker := testbase.TestUniLocker(t)

	mockDataSet := &testbase.MockDataSet{}
	mockDataSet.AddMockLeaseFromFile(t, "record123", "testdata/builder/_results/prov_document_lease.json")
	mockBuilderBackend.EXPECT().Load(gomock.Any(), "record123", gomock.Any()).Return(mockDataSet, nil)

	ts := time.Unix(1000, 0).UTC()

	lb, err := NewLeaseBuilder(context.Background(), mockBuilderBackend, bm, locker, creator,
		WithParent("record123", "", CopyModeNone, nil, false),
		WithTimestamp(ts))
	require.NoError(t, err)

	file, err := os.Open("testdata/builder/prov_document_2nd_rev/file1.txt")
	require.NoError(t, err)

	genTime := time.Unix(100000, 0).UTC()
	prov1 := &model.ProvEntity{
		ID:              "did:piprate:GGKW4B2zLbpxPJXE8TcD1okHFg4ymHvYbDfbHY2BthYc",
		Type:            model.ProvTypeEntity,
		GeneratedAtTime: &genTime,
		WasGeneratedBy: &model.ProvActivity{
			Type:              model.ProvTypeActivity,
			WasAssociatedWith: creator.ID,
			Algorithm:         "Capture",
		},
	}

	resID, err := lb.AddResource(file, WithVault(testbase.TestVaultName))
	require.NoError(t, err)
	assert.Equal(t, prov1.ID, resID)

	err = lb.AddProvenance(prov1.ID, prov1, false)
	require.NoError(t, err)

	file, err = os.Open("testdata/builder/prov_document_2nd_rev/meta.json")
	require.NoError(t, err)

	prov2 := &model.ProvEntity{
		ID:              "did:piprate:A4Uap6ycAc4Y8iXeX5cdxD4oQTsfqjfVPH8g9Cw2bFSp",
		Type:            model.ProvTypeEntity,
		GeneratedAtTime: &genTime,
		WasGeneratedBy: &model.ProvActivity{
			Type:              model.ProvTypeActivity,
			WasAssociatedWith: creator.ID,
			Algorithm:         "DocumentGeneration",
			Used: []string{
				"_:file-meta",
				"_:deal",
			},
			QualifiedUsage: []*model.ProvUsage{
				{
					Type: model.ProvTypeUsage,
					Entity: &model.ProvEntity{
						ID:              "_:file-meta",
						Type:            model.ProvTypeEntity,
						GeneratedAtTime: &genTime,
						WasGeneratedBy: &model.ProvActivity{
							Type:      model.ProvTypeActivity,
							Algorithm: "FileExtraction",
							Used:      prov1.ID,
							QualifiedUsage: []*model.ProvUsage{
								{
									Type:   model.ProvTypeUsage,
									Entity: prov1.ID,
									HadRole: &model.ProvRole{
										Type:  model.ProvTypeRole,
										Label: "file",
									},
								},
							},
						},
					},
					HadRole: &model.ProvRole{
						Type:  model.ProvTypeRole,
						Label: "file_meta",
					},
				},
				{
					Type:   model.ProvTypeUsage,
					Entity: "new-deal-impression-id",
					HadRole: &model.ProvRole{
						Type:  model.ProvTypeRole,
						Label: "deal",
					},
				},
			},
		},
	}

	resID, err = lb.AddMetaResource(file, WithVault(testbase.TestVaultName))
	require.NoError(t, err)
	assert.Equal(t, prov2.ID, resID)

	err = lb.AddProvenance(prov2.ID, prov2, false)
	require.NoError(t, err)

	err = lb.AddProvenance("", []any{
		&model.ProvAgent{
			ID:   creator.ID,
			Type: model.ProvTypeAgent,
		},
	}, false)
	require.NoError(t, err)

	res, err := lb.Build(expiry.Months(12))
	require.NoError(t, err)

	res.ID = testID
	res.ExpiresAt = &ts

	expectedConfigsBytes, err := os.ReadFile("testdata/builder/_results/prov_document_2nd_rev_lease.json")
	require.NoError(t, err)

	testbase.AssertEqualJSON(t, expectedConfigsBytes, res)
}

func TestLeaseBuilder_WithProvenance_Document2ndRevisionMetaOnly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilderBackend := mock.NewMockLeaseBuilderBackend(ctrl)
	bm := testbase.TestBlobManager(t, true, nil)
	creator := testbase.TestDID(t)
	locker := testbase.TestUniLocker(t)

	mockDataSet := &testbase.MockDataSet{}
	mockDataSet.AddMockLeaseFromFile(t, "record123", "testdata/builder/_results/prov_document_lease.json")
	mockBuilderBackend.EXPECT().Load(gomock.Any(), "record123", gomock.Any()).Return(mockDataSet, nil)

	ts := time.Unix(1000, 0).UTC()

	lb, err := NewLeaseBuilder(context.Background(), mockBuilderBackend, bm, locker, creator,
		WithParent("record123", "", CopyModeNone, nil, false),
		WithVault(testbase.TestVaultName),
		WithTimestamp(ts))
	require.NoError(t, err)

	genTime := time.Unix(100000, 0).UTC()

	file, err := os.Open("testdata/builder/prov_document_2nd_rev_meta_only/meta.json")
	require.NoError(t, err)

	prov2 := &model.ProvEntity{
		ID:              "did:piprate:A4Uap6ycAc4Y8iXeX5cdxD4oQTsfqjfVPH8g9Cw2bFSp",
		Type:            model.ProvTypeEntity,
		GeneratedAtTime: &genTime,
		WasGeneratedBy: &model.ProvActivity{
			Type:              model.ProvTypeActivity,
			WasAssociatedWith: creator.ID,
			Algorithm:         "DocumentGeneration",
			Used: []string{
				"_:file-meta",
				"_:deal",
			},
			QualifiedUsage: []*model.ProvUsage{
				{
					Type: model.ProvTypeUsage,
					Entity: &model.ProvEntity{
						ID:   "_:file-meta",
						Type: model.ProvTypeEntity,

						AsInBundle: "8yAcx6qDg4VgZiDnyLeDpN5Hh1n1hz5ubayqNmxTShww",
					},
					HadRole: &model.ProvRole{
						Type:  model.ProvTypeRole,
						Label: "file_meta",
					},
				},
				{
					Type:   model.ProvTypeUsage,
					Entity: "new-deal-impression-id",
					HadRole: &model.ProvRole{
						Type:  model.ProvTypeRole,
						Label: "deal",
					},
				},
			},
		},
	}

	resID, err := lb.AddMetaResource(file)
	assert.Equal(t, prov2.ID, resID)
	require.NoError(t, err)

	err = lb.AddProvenance(prov2.ID, prov2, false)
	require.NoError(t, err)

	err = lb.AddProvenance("", []any{
		&model.ProvAgent{
			ID:   creator.ID,
			Type: model.ProvTypeAgent,
		},
	}, false)
	require.NoError(t, err)

	res, err := lb.Build(expiry.Months(12))
	require.NoError(t, err)

	res.ID = testID
	res.ExpiresAt = &ts

	expectedConfigsBytes, err := os.ReadFile("testdata/builder/_results/prov_document_2nd_rev_meta_only_lease.json")
	require.NoError(t, err)

	testbase.AssertEqualJSON(t, expectedConfigsBytes, res)
}

func TestLeaseBuilder_WithProvenance_SharingFirst(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilderBackend := mock.NewMockLeaseBuilderBackend(ctrl)
	bm := testbase.TestBlobManager(t, true, nil)

	creator, err := model.GenerateDID(model.WithSeed("Steward1"))
	require.NoError(t, err)

	recipient, err := model.GenerateDID(model.WithSeed("Steward2"))
	require.NoError(t, err)

	srcDataSet := &testbase.MockDataSet{}
	srcDataSet.AddMockLeaseFromFile(t, "record123", "testdata/builder/_results/prov_document_lease.json")

	file, err := os.Open("testdata/builder/prov_document/file1.txt")
	require.NoError(t, err)

	ctx := context.Background()

	_, err = bm.SendBlob(ctx, file, false, "memory")
	require.NoError(t, err)

	file, err = os.Open("testdata/builder/prov_document/meta.json")
	require.NoError(t, err)

	_, err = bm.SendBlob(ctx, file, false, "memory")
	require.NoError(t, err)

	ts := time.Unix(100000, 0).UTC()

	lb, err := NewLeaseBuilderForSharing(context.Background(), srcDataSet, mockBuilderBackend, bm, CopyModeDeep,
		creator, nil, recipient.ID, testbase.TestVaultName, &ts)
	require.NoError(t, err)

	res, err := lb.Build(expiry.Months(12))
	require.NoError(t, err)

	res.ID = testID
	res.ExpiresAt = &ts

	expectedConfigsBytes, err := os.ReadFile("testdata/builder/_results/prov_document_share_lease.json")
	require.NoError(t, err)

	testbase.AssertEqualJSON(t, expectedConfigsBytes, res)
}

func TestLeaseBuilder_WithProvenance_SharingSecond(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilderBackend := mock.NewMockLeaseBuilderBackend(ctrl)
	bm := testbase.TestBlobManager(t, true, nil)

	creator, err := model.GenerateDID(model.WithSeed("Steward2"))
	require.NoError(t, err)

	recipient, err := model.GenerateDID(model.WithSeed("Steward3"))
	require.NoError(t, err)

	srcDataSet := &testbase.MockDataSet{}
	srcDataSet.AddMockLeaseFromFile(t, "record123", "testdata/builder/_results/prov_document_share_lease.json")

	file, err := os.Open("testdata/builder/prov_document/file1.txt")
	require.NoError(t, err)

	ctx := context.Background()

	_, err = bm.SendBlob(ctx, file, false, "memory")
	require.NoError(t, err)

	file, err = os.Open("testdata/builder/prov_document/meta.json")
	require.NoError(t, err)

	_, err = bm.SendBlob(ctx, file, false, "memory")
	require.NoError(t, err)

	ts := time.Unix(100000, 0).UTC()

	lb, err := NewLeaseBuilderForSharing(context.Background(), srcDataSet, mockBuilderBackend, bm, CopyModeDeep,
		creator, nil, recipient.ID, testbase.TestVaultName, &ts)
	require.NoError(t, err)

	res, err := lb.Build(expiry.Months(12))
	require.NoError(t, err)

	res.ID = testID
	res.ExpiresAt = &ts

	expectedConfigsBytes, err := os.ReadFile("testdata/builder/_results/prov_document_share_second_lease.json")
	require.NoError(t, err)

	testbase.AssertEqualJSON(t, expectedConfigsBytes, res)
}

func TestLeaseBuilder_WithProvenance_SharingOnly(t *testing.T) {

	// this test covers the case when we explicitly share the _first_ instance of an impression
	// with a specific recipient. We assume this may and should happen when the new impression
	// is saved directly into a shared locker.

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilderBackend := mock.NewMockLeaseBuilderBackend(ctrl)
	bm := testbase.TestBlobManager(t, true, nil)

	creator := testbase.TestDID(t)
	locker := testbase.TestUniLocker(t)

	recipient, err := model.GenerateDID(model.WithSeed("Steward3"))
	require.NoError(t, err)

	ts := time.Unix(1000, 0).UTC()

	lb, err := NewLeaseBuilder(context.Background(), mockBuilderBackend, bm, locker, creator, WithTimestamp(ts))
	require.NoError(t, err)

	file, err := os.Open("testdata/builder/prov_document/file1.txt")
	require.NoError(t, err)

	genTime := time.Unix(100000, 0).UTC()
	prov1 := &model.ProvEntity{
		ID:              "did:piprate:GGKW4B2zLbpxPJXE8TcD1okHFg4ymHvYbDfbHY2BthYc",
		Type:            model.ProvTypeEntity,
		GeneratedAtTime: &genTime,
		WasGeneratedBy: &model.ProvActivity{
			Type:              model.ProvTypeActivity,
			WasAssociatedWith: creator.ID,
			Algorithm:         "Capture",
		},
	}

	resID, err := lb.AddResource(file, WithVault(testbase.TestVaultName))
	require.NoError(t, err)
	assert.Equal(t, prov1.ID, resID)

	err = lb.AddProvenance(prov1.ID, prov1, false)
	require.NoError(t, err)

	file, err = os.Open("testdata/builder/prov_document/meta.json")
	require.NoError(t, err)

	prov2 := &model.ProvEntity{
		ID:              "did:piprate:CoDVBcdXCHxyAuiAioyTPJZVefYf4sHpK2bv5BYviu8Q",
		Type:            model.ProvTypeEntity,
		GeneratedAtTime: &genTime,
		WasGeneratedBy: &model.ProvActivity{
			Type:              model.ProvTypeActivity,
			WasAssociatedWith: creator.ID,
			Algorithm:         "Capture",
		},
	}

	resID, err = lb.AddMetaResource(file,
		WithAsset("did:piprate:3y2LHUXU84QqiGJ8u35r3XLLvGxJF7ZGYdL4DeqPZZSj"),
		WithContentType("File"),
		WithVault(testbase.TestVaultName))
	require.NoError(t, err)
	assert.Equal(t, prov2.ID, resID)

	err = lb.AddProvenance(prov2.ID, prov2, false)
	require.NoError(t, err)

	err = lb.AddShareProvenance(nil, recipient.ID)
	require.NoError(t, err)

	res, err := lb.Build(expiry.Months(12))
	require.NoError(t, err)

	res.ID = testID
	res.ExpiresAt = &ts

	expectedConfigsBytes, err := os.ReadFile("testdata/builder/_results/prov_document_lease_with_provenance.json")
	require.NoError(t, err)

	testbase.AssertEqualJSON(t, expectedConfigsBytes, res)
}

//nolint:thelper
func testExpandedForm(t *testing.T, testName string, updateExpectedFile bool) {
	_ = contexts.PreloadContextsIntoMemory()

	testFilePath := fmt.Sprintf("testdata/builder/prov_share/%s.json", testName)
	expectedFilePath := fmt.Sprintf("testdata/builder/_results/prov_share/%s_expanded.json", testName)

	testDoc, err := os.ReadFile(testFilePath)
	require.NoError(t, err)

	var pe *model.ProvEntity
	err = jsonw.Unmarshal(testDoc, &pe)
	require.NoError(t, err)

	expanded, err := model.ExpandDocument(pe.Bytes())
	require.NoError(t, err)

	expectedBytes, err := os.ReadFile(expectedFilePath)
	require.NoError(t, err)

	if ok, pretty := testbase.AssertEqualJSON(t, expectedBytes, expanded); !ok && updateExpectedFile {
		// update expected file - useful for updating the schema
		err = os.WriteFile(expectedFilePath, pretty, 0o600)
		require.NoError(t, err)
	}
}

func TestImpression_ExpandedForm(t *testing.T) {
	testExpandedForm(t, "single_share", true)
}
