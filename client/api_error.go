package client

type ApiError struct {
	Title string
	Msg   string
}

func (e *ApiError) String() string {
	if e.Msg == "" {
		return e.Title
	} else {
		return e.Msg
	}
}
