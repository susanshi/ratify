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

package verifier

import (
	"context"

	"github.com/ratify-project/ratify/pkg/common"
	"github.com/ratify-project/ratify/pkg/ocispecs"
	"github.com/ratify-project/ratify/pkg/referrerstore"
)

// VerifierResult describes the result of verifying a reference manifest for a subject
type VerifierResult struct { //nolint:revive // ignore linter to have unique type name
	Subject       string           `json:"subject,omitempty"`
	IsSuccess     bool             `json:"isSuccess"`
	Name          string           `json:"name,omitempty"`
	Type          string           `json:"type,omitempty"`
	Message       string           `json:"message,omitempty"`
	Extensions    interface{}      `json:"extensions,omitempty"`
	NestedResults []VerifierResult `json:"nestedResults,omitempty"`
	ArtifactType  string           `json:"artifactType,omitempty"`
}

// ReferenceVerifier is an interface that defines methods to verify a reference
// for a subject by a verifier.
type ReferenceVerifier interface {
	// Name returns the name of the verifier
	Name() string

	// Type returns the type name of the verifier
	Type() string

	// CanVerify returns if the verifier can verify the given reference
	CanVerify(ctx context.Context, referenceDescriptor ocispecs.ReferenceDescriptor) bool

	// Verify verifies the given reference of a subject and returns the result of verification
	Verify(ctx context.Context,
		subjectReference common.Reference,
		referenceDescriptor ocispecs.ReferenceDescriptor,
		referrerStore referrerstore.ReferrerStore) (VerifierResult, error)

	GetNestedReferences() []string
}
