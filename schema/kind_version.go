package schema

type KindVersion interface {
	GetListPath() string
	GetName() string
	GetParentPathParam() []string
	GetParentQueryParam() []string
	GetOrder() int
	GetListQueryParameter() map[string]FlagParameterOption
	GetApplyExample() string
}

type ConsoleKindVersion struct {
	ListPath           string
	Name               string
	ParentPathParam    []string
	ParentQueryParam   []string
	ListQueryParameter map[string]FlagParameterOption
	ApplyExample       string
	Order              int `json:1000` //same value DefaultPriority
}

func (c *ConsoleKindVersion) GetListPath() string {
	return c.ListPath
}

func (c *ConsoleKindVersion) GetApplyExample() string {
	return c.ApplyExample
}

func (c *ConsoleKindVersion) GetName() string {
	return c.Name
}

func (c *ConsoleKindVersion) GetParentPathParam() []string {
	return c.ParentPathParam
}

func (c *ConsoleKindVersion) GetParentQueryParam() []string {
	return c.ParentQueryParam
}

func (c *ConsoleKindVersion) GetOrder() int {
	return c.Order
}

func (c *ConsoleKindVersion) GetListQueryParameter() map[string]FlagParameterOption {
	return c.ListQueryParameter
}

type GatewayKindVersion struct {
	ListPath           string
	Name               string
	ParentPathParam    []string
	ParentQueryParam   []string
	ListQueryParameter map[string]FlagParameterOption
	GetAvailable       bool
	ApplyExample       string
	Order              int `json:1000` //same value DefaultPriority
}

func (g *GatewayKindVersion) GetListPath() string {
	return g.ListPath
}

func (g *GatewayKindVersion) GetName() string {
	return g.Name
}

func (g *GatewayKindVersion) GetParentPathParam() []string {
	return g.ParentPathParam
}

func (g *GatewayKindVersion) GetParentQueryParam() []string {
	return g.ParentQueryParam
}

func (g *GatewayKindVersion) GetApplyExample() string {
	return g.ApplyExample
}

func (g *GatewayKindVersion) GetOrder() int {
	return g.Order
}

func (g *GatewayKindVersion) GetListQueryParameter() map[string]FlagParameterOption {
	return g.ListQueryParameter
}
