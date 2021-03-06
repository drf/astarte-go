// Copyright © 2019 Ispirata Srl
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

package client

import (
	"fmt"
	"net/url"
	"path"
)

// PairingService is the API Client for Pairing API
type PairingService struct {
	client     *Client
	pairingURL *url.URL
}

// RegisterDevice registers a new device into the Realm.
// Returns the Credential Secret of the Device when successful.
// TODO: add support for initial_introspection
func (s *PairingService) RegisterDevice(realm string, deviceID string, token string) (string, error) {
	callURL, _ := url.Parse(s.pairingURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/%s/agent/devices", realm))

	var requestBody struct {
		HwID string `json:"hw_id"`
	}
	requestBody.HwID = deviceID

	decoder, err := s.client.genericJSONDataAPIPostWithResponse(callURL.String(), requestBody, token, 201)
	if err != nil {
		return "", err
	}

	// Decode the reply
	var responseBody struct {
		Data deviceRegistrationResponse `json:"data"`
	}
	err = decoder.Decode(&responseBody)
	if err != nil {
		return "", err
	}

	return responseBody.Data.CredentialsSecret, nil
}

// UnregisterDevice resets the registration state of a device. This makes it possible to register it again.
// All data belonging to the device will be left as is in Astarte.
func (s *PairingService) UnregisterDevice(realm string, deviceID string, token string) error {
	callURL, _ := url.Parse(s.pairingURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/%s/agent/devices/%s", realm, deviceID))

	err := s.client.genericJSONDataAPIDelete(callURL.String(), token, 204)
	if err != nil {
		return err
	}

	return nil
}

// ObtainNewMQTTv1CertificateForDevice returns a valid SSL Certificate for Devices running on astarte_mqtt_v1.
// This API is meant to be called by the device
func (s *PairingService) ObtainNewMQTTv1CertificateForDevice(realm, deviceID, credentialsSecret, csr string) (string, error) {
	callURL, _ := url.Parse(s.pairingURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/%s/devices/%s/protocols/astarte_mqtt_v1/credentials", realm, deviceID))

	var requestBody struct {
		CSR string `json:"csr"`
	}
	requestBody.CSR = csr

	decoder, err := s.client.genericJSONDataAPIPostWithResponse(callURL.String(), requestBody, credentialsSecret, 201)
	if err != nil {
		return "", err
	}

	// Decode the reply
	var responseBody struct {
		Data getMQTTv1CertificateResponse `json:"data"`
	}
	err = decoder.Decode(&responseBody)
	if err != nil {
		return "", err
	}

	return responseBody.Data.ClientCertificate, nil
}

// GetMQTTv1ProtocolInformationForDevice returns protocol information (such as the broker URL) for devices running
// on astarte_mqtt_v1.
// This API is meant to be called by the device
func (s *PairingService) GetMQTTv1ProtocolInformationForDevice(realm, deviceID, credentialsSecret string) (AstarteMQTTv1ProtocolInformation, error) {
	callURL, _ := url.Parse(s.pairingURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/%s/devices/%s", realm, deviceID))

	decoder, err := s.client.genericJSONDataAPIGET(callURL.String(), credentialsSecret, 200)
	if err != nil {
		return AstarteMQTTv1ProtocolInformation{}, err
	}

	// Decode the reply
	var responseBody struct {
		Data getDeviceProtocolStatusResponse `json:"data"`
	}
	err = decoder.Decode(&responseBody)
	if err != nil {
		return AstarteMQTTv1ProtocolInformation{}, err
	}

	return responseBody.Data.Protocols.AstarteMQTTv1, nil
}
