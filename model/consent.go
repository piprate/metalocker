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

package model

import "net/url"

//  See Kantara Initiative Consent Receipt Specification:
//
//  https://kantarainitiative.org/confluence/display/infosharing/Consent+Receipt+Specification
//
//  The structures below are based on its current (as of March 16, 2018) JSON schema:
//
//  https://kantarainitiative.org/confluence/download/attachments/76447870/CR%20Schema%20v1_1_0%20DRAFT%205.json?version=1&modificationDate=1508987053000&api=v2

type DataController struct {
	OnBehalf         bool     `json:"onBehalf,omitempty"`
	Org              string   `json:"org,omitempty"`
	Contact          string   `json:"contact,omitempty"`
	Address          any      `json:"address,omitempty"`
	Email            string   `json:"email,omitempty"`
	Phone            string   `json:"phone,omitempty"`
	PIIControllerURL *url.URL `json:"piiControllerUrl,omitempty"`
}

type Service struct {
	ServiceName string `json:"serviceName,omitempty"`
}

type ConsentReceipt struct {
	Version          string          `json:"version,omitempty"`
	Jurisdiction     string          `json:"jurisdiction,omitempty"`
	ConsentTimestamp uint64          `json:"consentTimestamp,omitempty"`
	CollectionMethod string          `json:"collectionMethod,omitempty"`
	ConsentReceiptID string          `json:"consentReceiptID,omitempty"`
	Subject          string          `json:"subject,omitempty"`
	DataController   *DataController `json:"dataController,omitempty"`
	Services         []*Service      `json:"services,omitempty"`
	PolicyURL        string          `json:"policyUrl,omitempty"`
	Sensitive        bool            `json:"sensitive,omitempty"`
	SpiCat           []string        `json:"spiCat,omitempty"`
}
