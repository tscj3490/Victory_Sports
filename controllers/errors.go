package controllers

// import (
// 	"fmt"
// 	"net/http"

// 	"github.com/micro/go-micro/errors"
// 	//	"github.com/micro/go-micro/errors"
// )

// // BadRequest generates a 400 error.
// func BadRequest(id string, fun interface{}, err error, format string, a ...interface{}) error {
// 	ErrorLog(id, fun, err, fmt.Sprintf(format, a...))
// 	return &errors.Error{
// 		Id:     id,
// 		Code:   400,
// 		Detail: fmt.Sprintf(format, a...),
// 		Status: http.StatusText(400),
// 	}
// }

// // Unauthorized generates a 401 error.
// func Unauthorized(id string, fun interface{}, err error, format string, a ...interface{}) error {
// 	ErrorLog(id, fun, err, fmt.Sprintf(format, a...))
// 	return &errors.Error{
// 		Id:     id,
// 		Code:   401,
// 		Detail: fmt.Sprintf(format, a...),
// 		Status: http.StatusText(401),
// 	}
// }

// // Forbidden generates a 403 error.
// func Forbidden(id string, fun interface{}, err error, format string, a ...interface{}) error {
// 	ErrorLog(id, fun, err, fmt.Sprintf(format, a...))
// 	return &errors.Error{
// 		Id:     id,
// 		Code:   403,
// 		Detail: fmt.Sprintf(format, a...),
// 		Status: http.StatusText(403),
// 	}
// }

// // NotFound generates a 404 error.
// func NotFound(id string, fun interface{}, err error, format string, a ...interface{}) error {
// 	ErrorLog(id, fun, err, fmt.Sprintf(format, a...))
// 	return &errors.Error{
// 		Id:     id,
// 		Code:   404,
// 		Detail: fmt.Sprintf(format, a...),
// 		Status: http.StatusText(404),
// 	}
// }

// // InternalServerError generates a 500 error.
// func InternalServerError(id string, fun interface{}, err error, format string, a ...interface{}) error {
// 	ErrorLog(id, fun, err, fmt.Sprintf(format, a...))
// 	return &errors.Error{
// 		Id:     id,
// 		Code:   500,
// 		Detail: fmt.Sprintf(format, a...),
// 		Status: http.StatusText(500),
// 	}
// }
