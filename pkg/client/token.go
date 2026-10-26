package client

//nolint:staticcheck
type Token struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	CreatedAt    string `json:"createdAt"`
	LastTimeUsed string `json:"lastTimeUsed"`
	ExpiresAt    string `json:"expiresAt,omitempty"`
	Status       string `json:"status,omitempty"`
}

//nolint:staticcheck
type CreatedToken struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	Token     string `json:"token"`
}
