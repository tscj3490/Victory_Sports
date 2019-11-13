package controllers

import (
	"fmt"
	"net/http"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/insights"
	"github.com/getsentry/raven-go"
	"github.com/go-chi/render"
)

//--
// Error response payloads & renderers
//--

// ErrResponse renderer type for handling all sorts of errors.
//
// In the best case scenario, the excellent github.com/pkg/errors package
// helps reveal information on the error, setting it on Err, and in the Render()
// method, using it to set the application-specific error code in AppCode.
type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {

	if e.HTTPStatusCode >= 500 {
		var packet *raven.Packet
		if err, ok := e.Err.(error); ok {
			packet = raven.NewPacket(e.StatusText, raven.NewException(err, raven.GetOrNewStacktrace(err, 2, 3, nil)), raven.NewHttp(r))
		} else {
			packet = raven.NewPacket(e.StatusText, raven.NewException(fmt.Errorf("%v - %v", e.HTTPStatusCode, e.StatusText), raven.NewStacktrace(2, 3, nil)), raven.NewHttp(r))
		}
		insights.Sentry.Capture(packet, nil)
	}

	render.Status(r, e.HTTPStatusCode)
	return nil
}

func (e ErrResponse) New(err error, code int, text string) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: code,
		StatusText:     text,
		ErrorText:      err.Error(),
	}
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

func ErrInternalServerError(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 500,
		StatusText:     "Internal Server Error.",
	}
}

var ErrNotFound = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}
var ErrNotImplemented = &ErrResponse{HTTPStatusCode: http.StatusBadRequest, StatusText: "Method on Resource not implemented."}

//var ErrInternalServerError = &ErrResponse{HTTPStatusCode: 500, StatusText: "Internal Server Error."}
