package schema

import "encoding/json"

type KindCatalog = map[string]Kind

func ConsoleDefaultKind() KindCatalog {
	return buildKindCatalogFromByteSchema[*ConsoleKindVersion](consoleDefaultByteSchema)
}

func GatewayDefaultKind() KindCatalog {
	return buildKindCatalogFromByteSchema[*GatewayKindVersion](gatewayDefaultByteSchema)
}

func buildKindCatalogFromByteSchema[T KindVersion](byteSchema []byte) KindCatalog {
	var jsonResult map[string]kindGeneric[T]
	err := json.Unmarshal(byteSchema, &jsonResult)
	if err != nil {
		panic(err)
	}
	var result KindCatalog = make(map[string]Kind)
	for kindName, kindGeneric := range jsonResult {
		kind := Kind{
			Versions: make(map[int]KindVersion),
		}
		for version, kindVersion := range kindGeneric.Versions {
			kind.Versions[version] = kindVersion
		}
		result[kindName] = kind
	}
	return result
}

type kindGeneric[T KindVersion] struct {
	Versions map[int]T
}
