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

package oras

import (
	"context"
	"errors"
	"fmt"
	"strings"

	oci "github.com/opencontainers/image-spec/specs-go/v1"
	re "github.com/ratify-project/ratify/errors"
	"github.com/ratify-project/ratify/pkg/common"
	"github.com/ratify-project/ratify/pkg/ocispecs"

	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"
)

const CosignArtifactType = "application/vnd.dev.cosign.artifact.sig.v1+json"
const CosignSignatureTagSuffix = ".sig"

func getCosignReferences(ctx context.Context, subjectReference common.Reference, repository registry.Repository) (*[]ocispecs.ReferenceDescriptor, error) {
	var references []ocispecs.ReferenceDescriptor
	signatureTag, err := attachedImageTag(subjectReference, CosignSignatureTagSuffix)
	if err != nil {
		return nil, err
	}

	desc, err := repository.Resolve(ctx, signatureTag)
	if err != nil {
		if errors.Is(err, errdef.ErrNotFound) {
			return nil, nil
		}
		evictOnError(ctx, err, subjectReference.Original)
		return nil, re.ErrorCodeRepositoryOperationFailure.WithError(err).WithComponentType(re.ReferrerStore)
	}

	references = append(references, ocispecs.ReferenceDescriptor{
		ArtifactType: CosignArtifactType,
		Descriptor: oci.Descriptor{
			MediaType: desc.MediaType,
			Digest:    desc.Digest,
			Size:      desc.Size,
		},
	})

	return &references, nil
}

func attachedImageTag(subjectReference common.Reference, tagSuffix string) (string, error) {
	// sha256:d34db33f -> sha256-d34db33f.suffix
	if subjectReference.Digest.String() == "" {
		return "", re.ErrorCodeReferenceInvalid.WithComponentType(re.ReferrerStore).WithDetail("Cosign subject digest is empty")
	}
	tagStr := strings.ReplaceAll(subjectReference.Digest.String(), ":", "-") + tagSuffix
	return fmt.Sprintf("%s:%s", subjectReference.Path, tagStr), nil
}
