package cli

import (
	"testing"

	"github.com/conduktor/ctl/pkg/resource"
	"github.com/conduktor/ctl/pkg/schema"
	"github.com/stretchr/testify/assert"
)

func TestNewDeleteHandler(t *testing.T) {
	debug := false
	rootCtx := RootContext{
		catalog: schema.Catalog{
			Kind: schema.KindCatalog{},
		},
		strict: true,
		debug:  &debug,
	}

	handler := NewDeleteHandler(rootCtx)

	assert.NotNil(t, handler)
	assert.Equal(t, rootCtx, handler.rootCtx)
}

func TestDeleteHandler_isIdentifiedByName(t *testing.T) {
	assert.True(t, isIdentifiedByName("VirtualCluster"))
	assert.True(t, isIdentifiedByName("Group"))
	assert.True(t, isIdentifiedByName("virtualcluster"))
	assert.True(t, isIdentifiedByName("group"))
	assert.False(t, isIdentifiedByName("Topic"))
}

func TestDeleteHandler_isIdentifiedByNameAndVCluster(t *testing.T) {
	assert.True(t, isIdentifiedByNameAndVCluster("AliasTopic"))
	assert.True(t, isIdentifiedByNameAndVCluster("GatewayServiceAccount"))
	assert.True(t, isIdentifiedByNameAndVCluster("ConcentrationRule"))
	assert.True(t, isIdentifiedByNameAndVCluster("aliastopic"))
	assert.False(t, isIdentifiedByNameAndVCluster("Topic"))
}

func TestDeleteHandler_isResourceInterceptor(t *testing.T) {
	resource1 := resource.Resource{Kind: "Interceptor", Name: "test"}
	resource2 := resource.Resource{Kind: "interceptor", Name: "test"}
	resource3 := resource.Resource{Kind: "Topic", Name: "test"}

	assert.True(t, isResourceInterceptor(resource1))
	assert.True(t, isResourceInterceptor(resource2))
	assert.False(t, isResourceInterceptor(resource3))
}
