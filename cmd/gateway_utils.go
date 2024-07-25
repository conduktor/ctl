package cmd

import (
	"github.com/conduktor/ctl/resource"
	"github.com/conduktor/ctl/schema"
	"strings"
)

func isGatewayKind(kind schema.Kind) bool {
	return strings.Contains(kind.GetLatestKindVersion().ListPath, "gateway")
}

func isGatewayResource(resource resource.Resource, kinds schema.KindCatalog) bool {
	kind, ok := kinds[resource.Kind]
	return ok && isGatewayKind(kind)
}

func isResourceIdentifiedByName(resource resource.Resource) bool {
	return isIdentifiedByName(resource.Kind)
}

func isResourceIdentifiedByNameAndVCluster(resource resource.Resource) bool {
	return isIdentifiedByNameAndVCluster(resource.Kind)
}

func isKindIdentifiedByNameAndVCluster(kind schema.Kind) bool {
	return isIdentifiedByNameAndVCluster(kind.GetName())
}

func isIdentifiedByNameAndVCluster(kind string) bool {
	return strings.Contains(strings.ToLower(kind), "aliastopic") ||
		strings.Contains(strings.ToLower(kind), "serviceaccount") ||
		strings.Contains(strings.ToLower(kind), "concentrationrule")
}

func isIdentifiedByName(kind string) bool {
	return strings.Contains(strings.ToLower(kind), "vcluster") ||
		strings.Contains(strings.ToLower(kind), "group")
}

func isResourceInterceptor(resource resource.Resource) bool {
	return strings.Contains(strings.ToLower(resource.Kind), "interceptor")
}

func isKindInterceptor(kind schema.Kind) bool {
	return strings.Contains(strings.ToLower(kind.GetName()), "interceptor")
}
