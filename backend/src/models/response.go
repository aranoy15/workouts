package models

type Response struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func NewSuccessResponse(data interface{}) *Response {
	return &Response{
		Data: data,
	}
}

func NewMessageResponse(message string) *Response {
	return &Response{
		Message: message,
	}
}

func NewErrorResponse(error string) *Response {
	return &Response{
		Error: error,
	}
}
