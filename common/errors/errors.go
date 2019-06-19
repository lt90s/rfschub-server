package errors

import (
	"encoding/json"
	"net/http"
)

type Error struct {
	Code     int    `json:"code"`
	HttpCode int    `json:"httpCode"`
	Message  string `json:"message"`
}

func New(code, httpCode int, message string) error {
	return Error{
		Code:     code,
		HttpCode: httpCode,
		Message:  message,
	}
}

func (e Error) Error() string {
	tmp, _ := json.Marshal(e)
	return string(tmp)
}

func FromError(err error) Error {
	var e Error
	ue := json.Unmarshal([]byte(err.Error()), &e)
	if ue != nil {
		e.Code = http.StatusInternalServerError
		e.HttpCode = http.StatusInternalServerError
		e.Message = err.Error()
	}
	return e
}

func NewBadRequestError(code int, message string) error {
	if code == -1 {
		code = http.StatusBadRequest
	}
	return New(code, http.StatusBadRequest, message)
}

func NewUnauthorizedError(code int, message string) error {
	if code == -1 {
		code = http.StatusUnauthorized
	}
	return New(code, http.StatusUnauthorized, message)
}

func NewInternalError(code int, message string) error {
	if code == -1 {
		code = http.StatusInternalServerError
	}
	return New(code, http.StatusInternalServerError, message)
}

func NewNotFoundError(code int, message string) error {
	if code == -1 {
		code = http.StatusNotFound
	}
	return New(code, http.StatusNotFound, message)
}

func NewServiceUnavailable(code int, message string) error {
	if code == -1 {
		code = http.StatusServiceUnavailable
	}
	return New(code, http.StatusServiceUnavailable, message)
}
