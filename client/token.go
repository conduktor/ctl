package client

type Token struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	CreatedAt    string `json:"createdAt"`
	LastTimeUsed string `json:"lastTimeUsed"`
}

type CreatedToken struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	Token     string `json:"token"`
}
