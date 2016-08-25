package jsonapi

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// these codes are inspired by http://go-talks.appspot.com/github.com/broady/talks/web-frameworks-gophercon.slide#1

// Error represents an error status of the HTTP request. Used with APIHandler.
type Error struct {
	Code    int
	Message string
	URL     string // url for 3xx redirect
}

// SetData creates a new Error instance and set the Message or URL property according to the error code
func (h Error) SetData(data string) Error {
	if h.Code >= 300 && h.Code < 400 {
		h.URL = data
		return h
	}

	h.Message = data
	return h
}

func (h Error) Error() string {
	ret := strconv.Itoa(h.Code)
	if h.Message != "" {
		ret += ": " + h.Message
	}

	return ret
}

// here are predefined error instances, you should call SetData before use it like
//
//     return nil, E404.SetData("User not found")
//
// You might noticed that here's no 500 error. You should just return a normal error
// instance instead.
//
//     return nil, errors.New("internal server error")
var (
	E301 = Error{Code: 301, Message: "Resource has been moved permanently"}
	E302 = Error{Code: 302, Message: "Resource has bee found at another location"}
	E307 = Error{Code: 307, Message: "Resource has been moved to another location temporarily"}
	E400 = Error{Code: 400, Message: "Error parsing request"}
	E401 = Error{Code: 401, Message: "You have to be authorized before accessing this resource"}
	E403 = Error{Code: 403, Message: "You have no right to access this resource"}
	E404 = Error{Code: 404, Message: "Resource not found"}
	E418 = Error{Code: 418, Message: "I'm a teapot"}
	E504 = Error{Code: 504, Message: "Service unavailable"}
)

// APIHandler is easy to use entry for API developer.
//
// Just return something, and it will be encoded to JSON format and send to client.
// Or return an Error to specify http status code and error string.
//
//     func myHandler(dec *json.Decoder, httpData *HTTP) (interface{}, error) {
//         var param paramType
//         if err := dec.Decode(&param); err != nil {
//             return nil, jsonapi.Error{http.StatusBadRequst, "You must send parameters in JSON format."}
//         }
//         return doSomething(param), nil
//     }
//
// To redirect clients, return 3xx status code and set Data property
//
//     return nil, jsonapi.Error{http.StatusBadRequst, "http://google.com"}
type APIHandler func(dec *json.Decoder, httpData *HTTP) (interface{}, error)

// Handler acts as jsonapi.Handler
func (h APIHandler) Handler(enc *json.Encoder, dec *json.Decoder, httpData *HTTP) {
	res, err := h(dec, httpData)
	if err == nil {
		if err := enc.Encode(res); err != nil {
			httpData.WriteHeader(http.StatusInternalServerError)
			enc.Encode("Cannot encode response into JSON format, please contact the administrator.")
		}
		return
	}

	code := http.StatusInternalServerError
	if httperr, ok := err.(Error); ok {
		code = httperr.Code
		if code >= 300 && code < 400 && httperr.URL != "" {
			// 3xx redirect
			http.Redirect(httpData.ResponseWriter, httpData.Request, httperr.URL, code)
			enc.Encode(err.Error())
			return
		}
	}

	httpData.WriteHeader(code)
	enc.Encode(err.Error())
}

// API denotes how a json api handler registers to a servemux
type API struct {
	Pattern    string
	APIHandler APIHandler
}

// Register helps you to register many APIHandlers to a http.ServeMux
func Register(apis []API, mux *http.ServeMux) {
	reg := http.Handle
	if mux != nil {
		reg = mux.Handle
	}

	for _, api := range apis {
		reg(api.Pattern, HTTPHandler(api.APIHandler.Handler))
	}
}
