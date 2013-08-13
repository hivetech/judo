package gwacl

import (
    "io/ioutil"
    . "launchpad.net/gocheck"
    "net/http"
    "net/http/httptest"
)

type x509DispatcherSuite struct{}

var _ = Suite(&x509DispatcherSuite{})

type Request struct {
    *http.Request
    BodyContent []byte
}

// makeRecordingHTTPServer creates an http server (don't forget to Close() it when done)
// that serves at the given base URL, copies incoming requests into the given
// channel, and finally returns the given status code.  If body is not nil, it
// will be returned as the request body.
func makeRecordingHTTPServer(requests chan *Request, status int, body []byte, headers http.Header) *httptest.Server {
    returnRequest := func(w http.ResponseWriter, r *http.Request) {
        // Capture all the request body content for later inspection.
        requestBody, err := ioutil.ReadAll(r.Body)
        if err != nil {
            panic(err)
        }
        requests <- &Request{r, requestBody}

        for header, values := range headers {
            for _, value := range values {
                w.Header().Set(header, value)
            }
        }
        w.WriteHeader(status)
        if body != nil {
            w.Write(body)
        }
    }
    serveMux := http.NewServeMux()
    serveMux.HandleFunc("/", returnRequest)
    return httptest.NewServer(serveMux)
}

func (*x509DispatcherSuite) TestGetRequestDoesHTTPGET(c *C) {
    httpRequests := make(chan *Request, 1)
    server := makeRecordingHTTPServer(httpRequests, http.StatusOK, nil, nil)
    defer server.Close()
    // No real certificate needed since we're testing on http, not https.
    session, err := newX509Session("subscriptionid", "", "West US")
    c.Assert(err, IsNil)
    path := "/foo/bar"
    version := "test-version"
    request := newX509RequestGET(server.URL+path, version)

    response, err := performX509Request(session, request)
    c.Assert(err, IsNil)
    c.Assert(response.StatusCode, Equals, http.StatusOK)

    httpRequest := <-httpRequests
    c.Check(httpRequest.Method, Equals, "GET")
    c.Check(httpRequest.Header[http.CanonicalHeaderKey("X-Ms-Version")], DeepEquals, []string{version})
    c.Check(httpRequest.URL.String(), Equals, path)
    c.Check(httpRequest.BodyContent, HasLen, 0)
}

func (*x509DispatcherSuite) TestPostRequestDoesHTTPPOST(c *C) {
    httpRequests := make(chan *Request, 1)
    requestBody := []byte{1, 2, 3}
    responseBody := []byte{4, 5, 6}
    requestContentType := "bogusContentType"
    server := makeRecordingHTTPServer(httpRequests, http.StatusOK, responseBody, nil)
    defer server.Close()
    // No real certificate needed since we're testing on http, not https.
    session, err := newX509Session("subscriptionid", "", "West US")
    c.Assert(err, IsNil)
    path := "/foo/bar"
    version := "test-version"
    request := newX509RequestPOST(server.URL+path, version, requestBody, requestContentType)

    response, err := performX509Request(session, request)
    c.Assert(err, IsNil)
    c.Assert(response.StatusCode, Equals, http.StatusOK)
    c.Check(response.Body, DeepEquals, responseBody)

    httpRequest := <-httpRequests
    c.Check(httpRequest.Header[http.CanonicalHeaderKey("Content-Type")], DeepEquals, []string{requestContentType})
    c.Check(httpRequest.Header[http.CanonicalHeaderKey("X-Ms-Version")], DeepEquals, []string{request.APIVersion})
    c.Check(httpRequest.Method, Equals, "POST")
    c.Check(httpRequest.URL.String(), Equals, path)
    c.Check(httpRequest.BodyContent, DeepEquals, requestBody)
}

func (*x509DispatcherSuite) TestDeleteRequestDoesHTTPDELETE(c *C) {
    httpRequests := make(chan *Request, 1)
    server := makeRecordingHTTPServer(httpRequests, http.StatusOK, nil, nil)
    defer server.Close()
    // No real certificate needed since we're testing on http, not https.
    session, err := newX509Session("subscriptionid", "", "West US")
    c.Assert(err, IsNil)
    path := "/foo/bar"
    version := "test-version"
    request := newX509RequestDELETE(server.URL+path, version)

    response, err := performX509Request(session, request)
    c.Assert(err, IsNil)
    c.Assert(response.StatusCode, Equals, http.StatusOK)

    httpRequest := <-httpRequests
    c.Check(httpRequest.Method, Equals, "DELETE")
    c.Check(httpRequest.Header[http.CanonicalHeaderKey("X-Ms-Version")], DeepEquals, []string{version})
    c.Check(httpRequest.URL.String(), Equals, path)
    c.Check(httpRequest.BodyContent, HasLen, 0)
}

func (*x509DispatcherSuite) TestPutRequestDoesHTTPPUT(c *C) {
    httpRequests := make(chan *Request, 1)
    requestBody := []byte{1, 2, 3}
    responseBody := []byte{4, 5, 6}
    server := makeRecordingHTTPServer(httpRequests, http.StatusOK, responseBody, nil)
    defer server.Close()
    // No real certificate needed since we're testing on http, not https.
    session, err := newX509Session("subscriptionid", "", "West US")
    c.Assert(err, IsNil)
    path := "/foo/bar"
    version := "test-version"
    request := newX509RequestPUT(server.URL+path, version, requestBody, "application/octet-stream")

    response, err := performX509Request(session, request)
    c.Assert(err, IsNil)
    c.Assert(response.StatusCode, Equals, http.StatusOK)
    c.Check(response.Body, DeepEquals, responseBody)

    httpRequest := <-httpRequests
    c.Check(httpRequest.Method, Equals, "PUT")
    c.Check(httpRequest.Header[http.CanonicalHeaderKey("X-Ms-Version")], DeepEquals, []string{version})
    c.Check(httpRequest.URL.String(), Equals, path)
    c.Check(httpRequest.BodyContent, DeepEquals, requestBody)
}

func (*x509DispatcherSuite) TestRequestRegistersHeader(c *C) {
    customHeader := http.CanonicalHeaderKey("x-gwacl-test")
    customValue := []string{"present"}
    returnRequest := func(w http.ResponseWriter, r *http.Request) {
        w.Header()[customHeader] = customValue
        w.WriteHeader(http.StatusOK)
    }
    serveMux := http.NewServeMux()
    serveMux.HandleFunc("/", returnRequest)
    server := httptest.NewServer(serveMux)
    defer server.Close()
    session, err := newX509Session("subscriptionid", "", "West US")
    c.Assert(err, IsNil)
    path := "/foo/bar"
    request := newX509RequestGET(server.URL+path, "testversion")

    response, err := performX509Request(session, request)
    c.Assert(err, IsNil)

    c.Check(response.Header[customHeader], DeepEquals, customValue)
}
