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

package dataset

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/utils"
	fp "github.com/piprate/metalocker/utils/fingerprint"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/rs/zerolog/log"
)

const (
	trapdoorLength = 8

	CopyModeNone    = "none"
	CopyModeShallow = "shallow"
	CopyModeDeep    = "deep"
)

type LeaseBuilderBackend interface {
	Load(id string, opts ...LoadOption) (model.DataSet, error)

	Submit(lease *model.Lease, cleartext bool, lockerID string, sender *model.LockerParticipant, headName ...string) RecordFuture
}

func GenerateLeaseID() (string, error) {
	randBuffer := make([]byte, 32)
	_, err := rand.Read(randBuffer)
	if err != nil {
		return "", err
	}
	return base58.Encode(randBuffer), nil
}

func GenerateTrapdoor() (string, error) {
	randBuffer := make([]byte, trapdoorLength)
	_, err := rand.Read(randBuffer)
	if err != nil {
		return "", err
	}
	return base58.Encode(randBuffer), nil
}

type builderOptions struct {
	creator        *model.DID
	vaultName      string
	didMethod      string
	assetID        string
	parentRecordID string
	parentLockerID string
	cleartext      bool
	blobCopyMode   string
	detachments    []string
	detachResource bool
	contentType    string
	heads          []string
	timeStamp      *time.Time
}

// BuilderOption is for defining optional parameters for Builder
type BuilderOption func(opts *builderOptions) error

func WithAsset(assetID string) BuilderOption {
	return func(opts *builderOptions) error {
		opts.assetID = assetID
		return nil
	}
}

func WithDIDMethod(didMethod string) BuilderOption {
	return func(opts *builderOptions) error {
		opts.didMethod = didMethod
		return nil
	}
}

func WithContentType(contentType string) BuilderOption {
	return func(opts *builderOptions) error {
		opts.contentType = contentType
		return nil
	}
}

func WithTimestamp(ts time.Time) BuilderOption {
	return func(opts *builderOptions) error {
		opts.timeStamp = &ts
		return nil
	}
}

func WithCreator(creator *model.DID) BuilderOption {
	return func(opts *builderOptions) error {
		opts.creator = creator
		return nil
	}
}

func WithParent(parentRecordID, parentLockerID, blobCopyMode string, detachments []string, detachResource bool) BuilderOption {
	return func(opts *builderOptions) error {
		opts.parentRecordID = parentRecordID
		opts.parentLockerID = parentLockerID
		opts.blobCopyMode = blobCopyMode
		opts.detachments = detachments
		opts.detachResource = detachResource
		return nil
	}
}

func WithVault(vaultName string) BuilderOption {
	return func(opts *builderOptions) error {
		opts.vaultName = vaultName
		return nil
	}
}

func SetHeads(headName ...string) BuilderOption {
	return func(opts *builderOptions) error {
		opts.heads = headName
		return nil
	}
}

func AsCleartext() BuilderOption {
	return func(opts *builderOptions) error {
		opts.cleartext = true
		return nil
	}
}

type LeaseBuilder struct {
	backend     LeaseBuilderBackend
	blobManager model.BlobManager

	lockerID          string
	senderParticipant *model.LockerParticipant
	didMethod         string
	vaultName         string
	cleartext         bool
	creator           *model.DID
	sender            *model.DID
	recipientID       string
	imp               *model.Impression
	timestampOverride *time.Time
	resources         map[string]*model.StoredResource
	sourceRecordID    string
	sourceLease       *model.Lease
	sharingMode       bool

	headNames []string

	parentProvenance map[string]map[string]any
	provenance       map[string]any
	metaProvTemplate map[string]any
}

var _ Builder = (*LeaseBuilder)(nil)

func NewLeaseBuilder(c LeaseBuilderBackend, blobManager model.BlobManager, locker *model.Locker, creator *model.DID, opts ...BuilderOption) (*LeaseBuilder, error) {

	// process builder options

	var options builderOptions
	for _, fn := range opts {
		if err := fn(&options); err != nil {
			return nil, err
		}
	}

	if options.creator != nil {
		creator = options.creator
	}

	imp := model.NewBlankImpression()
	imp.WasAttributedTo = creator.ID

	if options.assetID != "" {
		imp.Asset = options.assetID
	}

	b := &LeaseBuilder{
		backend:           c,
		blobManager:       blobManager,
		lockerID:          locker.ID,
		senderParticipant: locker.Us(),
		creator:           creator,
		didMethod:         options.didMethod,
		vaultName:         options.vaultName,
		cleartext:         options.cleartext,
		imp:               imp,
		sharingMode:       false,
		provenance:        make(map[string]any),
		parentProvenance:  make(map[string]map[string]any),
		resources:         make(map[string]*model.StoredResource),
		headNames:         options.heads,
		timestampOverride: options.timeStamp,
	}

	if options.parentRecordID != "" {
		if err := b.setParent(
			options.parentRecordID,
			options.parentLockerID,
			options.blobCopyMode,
			options.detachments,
			options.detachResource); err != nil {
			return nil, err
		}
	}

	if !locker.IsUnilocker() {
		// if it's not a 'private' locker with one identity, automatically add share provenance
		err := b.AddShareProvenance(nil, locker.Them().ID)
		if err != nil {
			return nil, err
		}
	}

	return b, nil
}

func NewLeaseBuilderForSharing(source model.DataSet, c LeaseBuilderBackend, blobManager model.BlobManager, blobCopyMode string, creator, sender *model.DID, recipientID, vaultName string, timeStamp *time.Time) (*LeaseBuilder, error) {

	if sender == nil {
		sender = creator
	}

	sourceLease := source.Lease()
	b := &LeaseBuilder{
		backend:           c,
		blobManager:       blobManager,
		vaultName:         vaultName,
		creator:           creator,
		sender:            sender,
		recipientID:       recipientID,
		imp:               sourceLease.Impression,
		sharingMode:       true,
		resources:         make(map[string]*model.StoredResource),
		sourceRecordID:    source.ID(),
		sourceLease:       sourceLease,
		timestampOverride: timeStamp,
	}

	// copy embedded blobs

	if err := b.copyBlobs(source, blobCopyMode, nil, false); err != nil {
		return nil, err
	}

	return b, nil
}

func (b *LeaseBuilder) CreatorID() string {
	return b.creator.ID
}

func (b *LeaseBuilder) AddResource(r io.Reader, opts ...BuilderOption) (string, error) {
	var options builderOptions
	for _, fn := range opts {
		if err := fn(&options); err != nil {
			return "", err
		}
	}

	// identify the vault to use

	vaultName := b.vaultName
	if options.vaultName != "" {
		vaultName = options.vaultName
	}
	if vaultName == "" {
		return "", errors.New("vault name not provided")
	}

	// save blob

	res, err := b.blobManager.SendBlob(r, b.cleartext, vaultName)
	if err != nil {
		return "", err
	}

	if _, found := b.resources[res.Asset]; found {
		// resource already saved, purge the latest copy
		if err = b.blobManager.PurgeBlob(res); err != nil {
			return "", err
		}
	} else {
		b.resources[res.Asset] = res
	}

	return res.Asset, nil
}

func (b *LeaseBuilder) ImportResource(res *model.StoredResource) error {
	if _, found := b.resources[res.Asset]; found {
		// resource already saved
		return nil
	}

	// validate storage config

	if res.ID == "" && res.Params == nil {
		return errors.New("both id and params fields are missing in blob storage config")
	}
	if res.Asset == "" {
		return errors.New("asset field is missing in blob storage config")
	}
	if res.Vault == "" {
		return errors.New("vault field is missing in blob storage config")
	}
	if res.Method == "" {
		return errors.New("method field is missing in blob storage config")
	}

	b.resources[res.Asset] = res

	return nil
}

func (b *LeaseBuilder) AddMetaResource(meta any, opts ...BuilderOption) (string, error) {

	var options builderOptions
	for _, fn := range opts {
		if err := fn(&options); err != nil {
			return "", err
		}
	}

	// marshal, if not []byte

	var data []byte
	var err error
	var ok bool

	if data, ok = meta.([]byte); !ok {
		if f, isFile := meta.(io.Reader); isFile {
			data, err = io.ReadAll(f)
			if err != nil {
				return "", err
			}
		} else {
			data, err = jsonw.Marshal(meta)
			if err != nil {
				return "", err
			}
		}
	}

	// identify the vault to use

	vaultName := b.vaultName
	if options.vaultName != "" {
		vaultName = options.vaultName
	}
	if vaultName == "" {
		return "", errors.New("vault name not provided")
	}

	// identify content type and asset ID

	contentType := options.contentType
	assetID := b.imp.Asset
	if options.assetID != "" {
		assetID = options.assetID
	}

	if contentType == "" || assetID == "" {
		discoveredAssetID, discoveredContentType := utils.DiscoverMetaGraph(data)
		if contentType == "" {
			contentType = discoveredContentType
		}
		if assetID == "" {
			assetID = discoveredAssetID
		}
	}

	if contentType == "" {
		return "", errors.New("unable to identify body content type")
	}
	if assetID != "" {
		b.imp.Asset = assetID
	}

	fpBytes, err := fp.GetFingerprint(bytes.NewReader(data), fp.AlgoSha256)
	if err != nil {
		return "", err
	}

	fpString := base64.StdEncoding.EncodeToString(fpBytes)

	// save blob

	assetID, err = b.AddResource(bytes.NewReader(data), WithVault(vaultName))
	if err != nil {
		return "", err
	}

	b.imp.MetaResource = &model.MetaResource{
		Asset:                assetID,
		ContentType:          contentType,
		Fingerprint:          fpString,
		FingerprintAlgorithm: fp.AlgoSha256,
	}

	if b.metaProvTemplate != nil {
		b.metaProvTemplate["id"] = assetID
		b.provenance[assetID] = b.metaProvTemplate
	}

	return assetID, nil
}

func (b *LeaseBuilder) copyBlobs(ds model.DataSet, blobCopyMode string, detachments []string, detachResource bool) error {
	if blobCopyMode == CopyModeNone {
		return nil
	}

	lease := ds.Lease()
	at := lease.GenerateAccessToken(ds.ID())
NEXT:
	for _, res := range lease.Resources {
		for _, d := range detachments {
			if d == res.Asset {
				continue NEXT
			}
		}

		if detachResource && res.Asset == lease.Impression.MetaResource.Asset {
			continue NEXT
		}

		switch blobCopyMode {
		case CopyModeShallow:
			if err := b.ImportResource(res); err != nil {
				return err
			}
		case CopyModeDeep:
			r, err := b.blobManager.GetBlob(res, at)
			if err != nil {
				return err
			}
			defer r.Close()

			_, err = b.AddResource(r)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported blob copy mode: '%s'", blobCopyMode)
		}
	}
	return nil
}

func (b *LeaseBuilder) setParent(parentRecordID, parentLockerID, blobCopyMode string, detachments []string, detachResource bool) error {
	parentDS, err := b.backend.Load(parentRecordID, FromLocker(parentLockerID))
	if err != nil {
		return err
	}
	parentLease := parentDS.Lease()
	b.sourceRecordID = parentRecordID
	b.sourceLease = parentLease

	b.imp.Asset = parentLease.Impression.Asset
	b.imp.SpecializationOf = parentLease.Impression.GetVariantID()
	b.imp.RevisionNumber = parentLease.Impression.Revision() + 1
	b.imp.WasRevisionOf = parentLease.Impression.ID

	if err = b.copyBlobs(parentDS, blobCopyMode, detachments, detachResource); err != nil {
		return err
	}

	if parentLease.Impression.ProvGraph != nil {

		// read parent provenance

		parentProv, err := ProvToMaps(parentLease.Impression.ProvGraph)
		if err != nil {
			return err
		}

		for _, p := range parentProv {
			if id, found := p["id"]; found {
				b.parentProvenance[id.(string)] = p
			}
		}
	}

	return nil
}

func ProvToMaps(prov any) ([]map[string]any, error) {
	bb, err := jsonw.Marshal(prov)
	if err != nil {
		return nil, err
	}

	parentProv := make([]map[string]any, 0)
	if _, isArray := prov.([]any); isArray {
		err := jsonw.Unmarshal(bb, &parentProv)
		if err != nil {
			return nil, err
		}
	} else {
		var p map[string]any
		err = jsonw.Unmarshal(bb, &p)
		if err != nil {
			return nil, err
		}
		parentProv = []map[string]any{p}
	}
	return parentProv, nil
}

func (b *LeaseBuilder) AddProvenance(id string, provenance any, override bool) error {
	if id == "" {
		provMaps, err := ProvToMaps(provenance)
		if err != nil {
			return err
		}

		for _, p := range provMaps {
			if id, idFound := p["id"]; idFound {
				if id.(string) == "%%resource%%" {
					b.metaProvTemplate = p
				} else {
					_, found := b.provenance[id.(string)]
					if !found || override {
						b.provenance[id.(string)] = p
					}
				}
			} else {
				log.Warn().Interface("rec", p).Msg("ID not found for provenance record")
			}
		}
	} else {
		_, found := b.provenance[id]
		if !found || override {
			b.provenance[id] = provenance
		}
	}

	return nil
}

func (b *LeaseBuilder) AddShareProvenance(sender *model.DID, recipientID string) error {
	if sender == nil {
		sender = b.creator
	}
	b.sender = sender
	b.recipientID = recipientID

	return nil
}

func (b *LeaseBuilder) Build(expiryTime time.Time) (*model.Lease, error) {

	now := time.Now().UTC()
	var ts *time.Time
	if b.timestampOverride != nil {
		ts = b.timestampOverride
	} else {
		ts = &now
	}

	var shareProvenance *model.ProvEntity

	if b.sharingMode {
		// generate provenance

		var wasQuotedFrom any
		if b.sourceLease.Provenance != nil {
			provCopy := b.sourceLease.Provenance.Copy()
			provCopy.Context = nil
			wasQuotedFrom = provCopy
		} else {
			wasQuotedFrom = b.imp.ID
		}

		shareProvenance = &model.ProvEntity{
			Context:         model.PiprateContextURL,
			Type:            model.ProvTypeEntity,
			WasAttributedTo: b.sender.ID,
			WasQuotedFrom:   wasQuotedFrom,
			WasAccessibleTo: b.recipientID,
			GeneratedAtTime: ts,
		}

		signKey := b.sender.SignKeyValue()
		err := shareProvenance.MerkleSign(b.sender.ID, signKey)
		if err != nil {
			return nil, err
		}

	} else {
		// finalise impression

		if b.imp.GeneratedAtTime == nil {
			b.imp.GeneratedAtTime = ts
		}

		if b.imp.Asset == "" {
			b.imp.Asset = model.NewAssetID(b.didMethod)
		}

		// generate provenance

		// fill gaps for resources

		for resID := range b.resources {
			_, found := b.provenance[resID]
			if !found {
				p, found := b.parentProvenance[resID]
				if !found {
					log.Debug().Str("resource", resID).Msg("Provenance not defined")
					continue
				}
				b.provenance[resID] = p
			}

		}

		provList := make([]any, 0)
		for _, p := range b.provenance {
			provList = append(provList, p)
		}

		if len(provList) > 0 {
			// add agent if missing

			if _, found := b.provenance[b.creator.ID]; !found {
				agent := &model.ProvAgent{
					ID:   b.creator.ID,
					Type: model.ProvTypeAgent,
				}
				b.provenance[b.creator.ID] = agent
				provList = append(provList, agent)
			}

			getID := func(v any) string {
				switch vv := v.(type) {
				case map[string]any:
					return vv["id"].(string)
				case *model.ProvEntity:
					return vv.ID
				case *model.ProvAgent:
					return vv.ID
				}
				return ""
			}
			sort.SliceStable(provList, func(i, j int) bool { return getID(provList[i]) < getID(provList[j]) })

			b.imp.ProvGraph = provList
		}

		signKey := b.creator.SignKeyValue()
		err := b.imp.MerkleSign(b.creator.ID, signKey)
		if err != nil {
			return nil, err
		}

		if b.recipientID != "" {
			shareProvenance = &model.ProvEntity{
				Context:         model.PiprateContextURL,
				Type:            model.ProvTypeEntity,
				WasAttributedTo: b.sender.ID,
				WasQuotedFrom:   b.imp.ID,
				WasAccessibleTo: b.recipientID,
				GeneratedAtTime: ts,
			}

			signKey := b.sender.SignKeyValue()
			err := shareProvenance.MerkleSign(b.sender.ID, signKey)
			if err != nil {
				return nil, err
			}
		}
	}

	// generate lease document

	id, err := GenerateLeaseID()
	if err != nil {
		return nil, err
	}

	resourceList := make([]*model.StoredResource, 0)
	for _, res := range b.resources {
		resourceList = append(resourceList, res)
	}

	sort.Slice(resourceList, func(i, j int) bool {
		ci, cj := resourceList[i], resourceList[j]
		return ci.Asset < cj.Asset
	})

	var expiresAt *time.Time
	if !expiryTime.IsZero() {
		expiresAt = &expiryTime
	}

	opRec := &model.Lease{
		ID:          id,
		Type:        "Lease",
		ExpiresAt:   expiresAt,
		Resources:   resourceList,
		DataSetType: "graph",
		Impression:  b.imp,
		Provenance:  shareProvenance,
	}

	return opRec, nil
}

func (b *LeaseBuilder) SetHeads(headName ...string) error {
	b.headNames = append(b.headNames, headName...)
	return nil
}

func (b *LeaseBuilder) Submit(expiryTime time.Time) RecordFuture {
	if !expiryTime.IsZero() {
		// check the expiry date is in the future and not too close to now
		expiryDelta := time.Until(expiryTime)
		if expiryDelta < 0 && expiryDelta < time.Second {
			return RecordFutureWithError(errors.New("invalid lease expiry time"))
		}
	}

	lease, err := b.Build(expiryTime)
	if err != nil {
		return RecordFutureWithError(err)
	}

	return b.backend.Submit(lease, b.cleartext, b.lockerID, b.senderParticipant, b.headNames...)
}

func (b *LeaseBuilder) Cancel() error {
	return errors.New("lease build cancellation not implemented yet")
}
