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

package certificatestores

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/ratify-project/ratify/internal/constants"
	ctxUtils "github.com/ratify-project/ratify/internal/context"
	"github.com/ratify-project/ratify/pkg/utils"
	vu "github.com/ratify-project/ratify/pkg/verifier/utils"
)

// ActiveCertStores implements the CertStoreManager interface
type ActiveCertStores struct {
	// scopedCertStores is mapping from cert store name to certificate list.
	// The certificate store name is prefixed with the namespace.
	// Example:
	// {
	//   "namespace1/store1": []*x509.Certificate,
	//   "namespace2/store2": []*x509.Certificate
	// }
	scopedCertStores sync.Map
}

func NewActiveCertStores() CertStoreManager {
	return &ActiveCertStores{}
}

// GetCertStores fulfills the CertStoreManager interface.
// It returns a list of certificates in the given store.
func (c *ActiveCertStores) GetCertsFromStore(ctx context.Context, storeName string) ([]*x509.Certificate, error) {
	prependedName, prepended := prependNamespaceToStoreName(storeName)
	if !prepended {
		return []*x509.Certificate{}, fmt.Errorf("The given store name %s is not namespaced", storeName)
	}

	if !hasAccessToStore(ctx, storeName) {
		return []*x509.Certificate{}, fmt.Errorf("namespace: [%s] does not have access to certificate store: %s", ctxUtils.GetNamespace(ctx), storeName)
	}
	if certs, ok := c.scopedCertStores.Load(prependedName); ok {
		return certs.([]*x509.Certificate), nil
	}
	return []*x509.Certificate{}, fmt.Errorf("failed to access non-existent certificate store: %s", storeName)
}

// AddStore fulfills the CertStoreManager interface.
// It adds the given certificate under cert store.
func (c *ActiveCertStores) AddStore(storeName string, cert []*x509.Certificate) {
	c.scopedCertStores.Store(storeName, cert)
}

// DeleteStore fulfills the CertStoreManager interface.
// It deletes the given cert store.
func (c *ActiveCertStores) DeleteStore(storeName string) {
	c.scopedCertStores.Delete(storeName)
}

// A namespaced verification request could access certStores in the same namespace.
// A cluster-wide (context namespace is "") verification request could access certStores across all namespaces.
// Note: the cluster-wide behavior is different from KMP as we need to keep the behavior backward compatible.
func hasAccessToStore(ctx context.Context, storeName string) bool {
	namespace := ctxUtils.GetNamespace(ctx)
	if namespace == constants.EmptyNamespace {
		return true
	}
	return strings.HasPrefix(storeName, namespace+constants.NamespaceSeperator)
}

// prependNamespaceToStoreName prepends namespace to store name if not already present.
// If the namespace where Ratify deployed is not set, prepended would be set to false.
func prependNamespaceToStoreName(storeName string) (prependedName string, prepended bool) {
	if vu.IsNamespacedNamed(storeName) {
		return storeName, true
	}
	if ns, found := os.LookupEnv(utils.RatifyNamespaceEnvVar); found {
		return ns + constants.NamespaceSeperator + storeName, true
	}
	return storeName, false
}
