package response

import "net/http"

// DefaultResponse is a default return value of http request
type DefaultResponse struct {
	Status  int         `json:"status"`  // http status code
	Message string      `json:"message"` // response message
	Data    interface{} `json:"data"`    // response data body
}

// NewOKResonseData create a new response for success response
func NewOKResonseData(data interface{}) *DefaultResponse {
	return &DefaultResponse{
		Status:  http.StatusOK,
		Message: "ok",
		Data:    data,
	}
}

// NewHTTPErrorResponse create a new response by HTTPError
func NewHTTPErrorResponse(err HTTPError, data interface{}) *DefaultResponse {
	return &DefaultResponse{
		Status:  err.Code(),
		Message: err.Error(),
		Data:    data,
	}
}

// NewErrorResponse create a new response by HTTPError
func NewErrorResponse(err error, data interface{}) *DefaultResponse {
	return &DefaultResponse{
		Status:  http.StatusInternalServerError,
		Message: err.Error(),
		Data:    data,
	}
}

// NewServerErrorResponse create a response which show internal server error
func NewServerErrorResponse(msg string, data interface{}) *DefaultResponse {
	return NewErrorResponse(NewServerError(msg), data)
}

// NewNotFoundResponse create a response which show not found
func NewNotFoundResponse(msg string) *DefaultResponse {
	return NewErrorResponse(NewNotFoundError(msg), nil)
}

// NewForbiddenResponse create a response which show forbidden
func NewForbiddenResponse(msg string) *DefaultResponse {
	return NewErrorResponse(NewForbiddenError(msg), nil)
}

// NewBadRequestResponse create a response which show bad request
func NewBadRequestResponse(msg string) *DefaultResponse {
	return NewErrorResponse(NewBadRequestError(msg), nil)
}
