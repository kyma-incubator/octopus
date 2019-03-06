/*

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

package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"

	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestStorageTestDefinition(t *testing.T) {
	key := types.NamespacedName{
		Name:      "foo",
		Namespace: "default",
	}
	created := &TestDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		}}

	// Test Create
	fetched := &TestDefinition{}
	require.NoError(t, c.Create(context.TODO(), created))

	require.NoError(t, c.Get(context.TODO(), key, fetched))
	assert.Equal(t, created, fetched)

	// Test Updating the Labels
	updated := fetched.DeepCopy()
	updated.Labels = map[string]string{"hello": "world"}
	require.NoError(t, c.Update(context.TODO(), updated))

	require.NoError(t, c.Get(context.TODO(), key, fetched))
	assert.Equal(t, updated, fetched)

	// Test Delete
	require.NoError(t, c.Delete(context.TODO(), fetched))
	err := c.Get(context.TODO(), key, fetched)
	assert.True(t, errors.IsNotFound(err))
}
