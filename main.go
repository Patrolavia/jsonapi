/*
Package jsonapi is simple wrapper for buildin net/http package.
It aims to let developers build json-based web api easier.

Create an api handler is so easy:

    // HelloArgs is data structure for arguments passed by POST body.
    type HelloArgs struct {
            Name string
            Title string
    }

    // HelloReply defines data structure this api will return.
    type HelloReply struct {
            Message string
    }

    // HelloHandler greets user with hello
    func HelloHandler(enc *json.Encoder, dec *json.Decoder, httpData *jsonapi.HTTP) {
            // Read json objct from request.
            var args HelloArgs
            if err := dec.Decode(&args); err != nil {
                    // The arguments are not passed in JSON format, do error handling here.
            }

            rep := HelloReply{fmt.Sprintf("Hello, %s %s", args,Title, args.Name)}

            // Send http response back to user.
            if err := enc.Encode(rep); err != nil {
                    // How can error occurs when JSONize our simple data structure! We must handle it here.
            }
    }

And this is how we do in main function:

    // If you used to write http.HandleFunc("/api/hello", HelloHandler)
    jsonapi.HandleFunc("/api/hello", HelloHandler)

    // If you feel more comfortable with http.Handle("/api/hello", http.HandlerFunc(HelloHandler))
    http.Handle("/api/hello", jsonapi.HTTPHandler(HelloHandler))

There is also a helper for you to write test with jsonapi.

    var data := map[string]interface{}{"Name": John Doe", "Title": "Mr."}
    resp, err := HandlerTest(HelloHandler).PostJSON("/api/hello", data)
    if err != nil {
            // Something bad happened in jsonapi's test helper!
            // Dump detailed info to find out if anything wrong in test code.
    }
    // Check response here.

*/
package jsonapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

// HTTP holds original data from http request handler.
// We keep this if you need to find out request uri, create session... etc.
type HTTP struct {
	http.ResponseWriter
	*http.Request
}

// HTTPHandler converts our json api handler to be used with net/http package.
type HTTPHandler func(*json.Encoder, *json.Decoder, *HTTP)

func (f HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := &HTTP{w, r}
	e := json.NewEncoder(w)
	d := json.NewDecoder(r.Body)
	w.Header().Add("Content-Type", "application/json")
	f(e, d, h)
	ioutil.ReadAll(r.Body) // drain data to enable socket reuse
}

// HandleFunc wraps our json api handler to http.Handle
func HandleFunc(pattern string, f func(*json.Encoder, *json.Decoder, *HTTP)) {
	http.Handle(pattern, HTTPHandler(f))
}

// HandlerTest is helper to test json api
type HandlerTest func(*json.Encoder, *json.Decoder, *HTTP)

// Get helps you to test with HTTP GET request
func (f HandlerTest) Get(uri, cookie string) (*httptest.ResponseRecorder, error) {
	ret := httptest.NewRecorder()
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	if cookie != "" {
		req.Header.Add("Cookie", cookie)
	}
	HTTPHandler(f).ServeHTTP(ret, req)
	return ret, err
}

// Post helps you to test with post data
func (f HandlerTest) Post(uri, cookie, data string) (*httptest.ResponseRecorder, error) {
	ret := httptest.NewRecorder()
	req, err := http.NewRequest("POST", uri, strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	if cookie != "" {
		req.Header.Add("Cookie", cookie)
	}
	HTTPHandler(f).ServeHTTP(ret, req)
	return ret, err
}

// PostJSON helps you to test with json encoded post data
func (f HandlerTest) PostJSON(uri, cookie string, data interface{}) (ret *httptest.ResponseRecorder, err error) {
	buf, err := json.Marshal(data)
	if err != nil {
		return
	}

	return f.Post(uri, cookie, string(buf))
}

// PostForm helps you to test with form encoded post data
func (f HandlerTest) PostForm(uri, cookie string, data url.Values) (*httptest.ResponseRecorder, error) {
	return f.Post(uri, cookie, data.Encode())
}
