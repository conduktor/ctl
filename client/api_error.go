package client

type ApiError struct {
	Title string
	Msg   string
	Cause string
}

func (e *ApiError) String() string {
	if e.Cause != "" {
		return e.Cause
	} else if e.Msg != "" {
		return e.Msg
	} else {
		return e.Title
	}
}
