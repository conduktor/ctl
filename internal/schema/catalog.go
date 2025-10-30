package schema

import (
	"encoding/json"
)

type Catalog struct {
	Kind KindCatalog
	Run  RunCatalog
}

type KindCatalog = map[string]Kind
type RunCatalog = map[string]Run

func ConsoleDefaultCatalog() *Catalog {
	return buildCatalogFromByteSchema[*ConsoleKindVersion](consoleDefaultByteSchema, CONSOLE)
}

func GatewayDefaultCatalog() *Catalog {
	return buildCatalogFromByteSchema[*GatewayKindVersion](gatewayDefaultByteSchema, GATEWAY)
}

// TODO: colision don't silently hide others.
func (catalog *Catalog) Merge(other *Catalog) Catalog {
	result := Catalog{
		Kind: make(map[string]Kind),
		Run:  make(map[string]Run),
	}
	for kindName, kind := range catalog.Kind {
		result.Kind[kindName] = kind
	}
	for kindName, kind := range other.Kind {
		result.Kind[kindName] = kind
	}
	for runName, run := range catalog.Run {
		result.Run[runName] = run
	}
	for runName, run := range other.Run {
		result.Run[runName] = run
	}
	return result
}

func buildCatalogFromByteSchema[T KindVersion](byteSchema []byte, backendType BackendType) *Catalog {
	var jsonResult CatalogGeneric[T]
	err := json.Unmarshal(byteSchema, &jsonResult)
	if err != nil {
		panic(err)
	}
	var result = Catalog{
		Kind: KindCatalog{},
		Run:  RunCatalog{},
	}
	for kindName, kindGeneric := range jsonResult.Kind {
		kind := Kind{
			Versions: make(map[int]KindVersion),
		}
		for version, kindVersion := range kindGeneric.Versions {
			kind.Versions[version] = kindVersion
		}
		result.Kind[kindName] = kind
	}
	for runName, run := range jsonResult.Run {
		run.BackendType = backendType
		result.Run[runName] = run
	}
	return &result
}

type kindGeneric[T KindVersion] struct {
	Versions map[int]T
}

type CatalogGeneric[T KindVersion] struct {
	Kind map[string]kindGeneric[T]
	Run  RunCatalog
}
