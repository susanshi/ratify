/*
Copyright The Ratify Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"os"
	"testing"

	configv1beta1 "github.com/ratify-project/ratify/api/v1beta1"
	"github.com/ratify-project/ratify/pkg/controllers"
	"github.com/ratify-project/ratify/pkg/customresources/verifiers"
	test "github.com/ratify-project/ratify/pkg/utils"
	vc "github.com/ratify-project/ratify/pkg/verifier/config"
	"github.com/ratify-project/ratify/pkg/verifier/types"
)

const (
	verifierName = "verifierName"
	verifierType = "verifierType"
)

func TestUpsertVerifier(t *testing.T) {
	dirPath, err := test.CreatePlugin(verifierName)
	if err != nil {
		t.Fatalf("createPlugin() expected no error, actual %v", err)
	}
	defer os.RemoveAll(dirPath)

	tests := []struct {
		name           string
		address        string
		namespace      string
		objectName     string
		verifierConfig vc.VerifierConfig
		expectedErr    bool
	}{
		{
			name:           "empty config",
			verifierConfig: vc.VerifierConfig{},
			expectedErr:    true,
		},
		{
			name:    "empty address",
			address: dirPath,
			verifierConfig: vc.VerifierConfig{
				"name": verifierName,
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetVerifierMap()

			err = UpsertVerifier("", tt.address, tt.namespace, tt.objectName, tt.verifierConfig)
			if tt.expectedErr != (err != nil) {
				t.Fatalf("UpsertVerifier() expected error %v, actual %v", tt.expectedErr, err)
			}
		})
	}
}

func TestSpecToVerifierConfig(t *testing.T) {
	tests := []struct {
		name                 string
		raw                  []byte
		verifierName         string
		verifierType         string
		artifactTypes        string
		source               *configv1beta1.PluginSource
		expectedErr          bool
		expectedVerifierName string
	}{
		{
			name:                 "empty raw",
			raw:                  []byte{},
			verifierName:         verifierName,
			verifierType:         verifierType,
			source:               &configv1beta1.PluginSource{},
			expectedVerifierName: verifierName,
			expectedErr:          false,
		},
		{
			name:                 "invalid raw",
			raw:                  []byte("test\n"),
			expectedVerifierName: "",
			expectedErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := SpecToVerifierConfig(tt.raw, tt.verifierName, tt.verifierType, tt.artifactTypes, tt.source)
			if tt.expectedErr != (err != nil) {
				t.Fatalf("SpecToVerifierConfig() expected error %v, actual %v", tt.expectedErr, err)
			}
			if _, ok := config[types.Name]; !ok {
				config[types.Name] = ""
			}
			if config[types.Name] != tt.expectedVerifierName {
				t.Fatalf("SpecToVerifierConfig() expected verifier name %s, actual %s", tt.expectedVerifierName, config[types.Name])
			}
		})
	}
}

func resetVerifierMap() {
	controllers.NamespacedVerifiers = verifiers.NewActiveVerifiers()
}
