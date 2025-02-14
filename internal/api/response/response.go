package response

type Response struct {
	Status string `json:"status"`
	Error  string `json:"errorMsg,omitempty"`
}

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

func OK() Response {
	return Response{Status: StatusOK}
}
