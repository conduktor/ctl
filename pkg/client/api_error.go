package client

type APIError struct {
	Title string `json:"title"`
	Msg   string `json:"msg"`
	Cause string `json:"cause"`
}

func (e *APIError) String() string {
	if e.Cause != "" {
		return e.Cause
	} else if e.Msg != "" {
		return e.Msg
	} else {
		return e.Title
	}
}
