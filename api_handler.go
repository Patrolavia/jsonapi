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
}

func (h Error) Error() string {
	ret := strconv.Itoa(h.Code)
	if h.Message != "" {
		ret += ": " + h.Message
	}

	return ret
}

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

	msg := err.Error()
	code := http.StatusInternalServerError
	if httperr, ok := err.(Error); ok {
		code = httperr.Code
	}
	httpData.WriteHeader(code)
	enc.Encode(msg)
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
