package orderedjson

import (
	"encoding/json"

	orderedmap "github.com/wk8/go-ordered-map/v2"
	yaml "gopkg.in/yaml.v3"
)

type OrderedData struct {
	orderedMap *orderedmap.OrderedMap[string, OrderedData]
	array      *[]OrderedData
	fallback   *interface{}
}

func (orderedData *OrderedData) UnmarshalJSON(data []byte) error {
	orderedData.orderedMap = orderedmap.New[string, OrderedData]()
	err := json.Unmarshal(data, &orderedData.orderedMap)
	if err != nil {
		orderedData.orderedMap = nil
		orderedData.array = new([]OrderedData)
		err = json.Unmarshal(data, orderedData.array)
	}
	if err != nil {
		orderedData.array = nil
		orderedData.fallback = new(interface{})
		err = json.Unmarshal(data, &orderedData.fallback)
	}
	return err
}

// TODO: remove once hack in printYaml is not needed anymore
func (orderedData *OrderedData) GetMapOrNil() *orderedmap.OrderedMap[string, OrderedData] {
	return orderedData.orderedMap
}

// TODO: remove once hack in printYaml is not needed anymore
func (orderedData *OrderedData) GetArrayOrNil() *[]OrderedData {
	return orderedData.array
}

func (orderedData OrderedData) MarshalJSON() ([]byte, error) {
	if orderedData.orderedMap != nil {
		return json.Marshal(orderedData.orderedMap)
	} else if orderedData.array != nil {
		return json.Marshal(orderedData.array)
	} else if orderedData.fallback != nil {
		return json.Marshal(orderedData.fallback)
	} else {
		return json.Marshal(nil)
	}
}

func (orderedData OrderedData) MarshalYAML() (interface{}, error) {
	if orderedData.orderedMap != nil {
		return orderedData.orderedMap, nil
	} else if orderedData.array != nil {
		return orderedData.array, nil
	}
	return orderedData.fallback, nil
}

func (orderedData *OrderedData) UnmarshalYAML(value *yaml.Node) error {
	panic("Not supported")
}
