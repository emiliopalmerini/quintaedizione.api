package shared

import (
	"fmt"
	"net/http"
)

type ErrorCommon struct {
	Code   string `json:"code"`
	Title  string `json:"title"`
	Detail string `json:"detail,omitempty"`
}

type ErrorObject struct {
	Errors []ErrorCommon `json:"errors"`
}

func NewErrorObject(code, title, detail string) ErrorObject {
	return ErrorObject{
		Errors: []ErrorCommon{
			{
				Code:   code,
				Title:  title,
				Detail: detail,
			},
		},
	}
}

func BadRequestError(detail string) ErrorObject {
	return NewErrorObject("BAD_REQUEST", "Bad Request", detail)
}

func NotFoundError(detail string) ErrorObject {
	return NewErrorObject("NOT_FOUND", "Not Found", detail)
}

func InternalServerError(detail string) ErrorObject {
	return NewErrorObject("INTERNAL_ERROR", "Internal Server Error", detail)
}

func UnauthorizedError(detail string) ErrorObject {
	return NewErrorObject("UNAUTHORIZED", "Unauthorized", detail)
}

type AppError struct {
	HTTPStatus int
	Response   ErrorObject
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	if len(e.Response.Errors) > 0 {
		return e.Response.Errors[0].Title
	}
	return "unknown error"
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(status int, response ErrorObject, err error) *AppError {
	return &AppError{
		HTTPStatus: status,
		Response:   response,
		Err:        err,
	}
}

func NewBadRequestError(detail string, err error) *AppError {
	return NewAppError(http.StatusBadRequest, BadRequestError(detail), err)
}

func NewNotFoundError(resource, id string) *AppError {
	detail := fmt.Sprintf("%s with id '%s' not found", resource, id)
	return NewAppError(http.StatusNotFound, NotFoundError(detail), nil)
}

func NewInternalError(err error) *AppError {
	return NewAppError(http.StatusInternalServerError, InternalServerError("an unexpected error occurred"), err)
}
