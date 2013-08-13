package gwacl

import (
    "bytes"
    "io/ioutil"
    "launchpad.net/gwacl/fork/http"
    . "launchpad.net/gwacl/logging"
)

type X509Request struct {
    APIVersion  string
    Method      string
    URL         string
    Payload     []byte
    ContentType string
}

// newX509RequestGET initializes an X509Request for a GET.  You may still need
// to set further values.
func newX509RequestGET(url, apiVersion string) *X509Request {
    return &X509Request{
        Method:     "GET",
        URL:        url,
        APIVersion: apiVersion,
    }
}

// newX509RequestPOST initializes an X509Request for a POST.  You may still
// need to set further values.
func newX509RequestPOST(url, apiVersion string, payload []byte, contentType string) *X509Request {
    return &X509Request{
        Method:      "POST",
        URL:         url,
        APIVersion:  apiVersion,
        Payload:     payload,
        ContentType: contentType,
    }
}

// newX509RequestDELETE initializes an X509Request for a DELETE.
func newX509RequestDELETE(url, apiVersion string) *X509Request {
    return &X509Request{
        Method:     "DELETE",
        URL:        url,
        APIVersion: apiVersion,
    }
}

// newX509RequestPUT initializes an X509Request for a PUT.  You may still
// need to set further values.
func newX509RequestPUT(url, apiVersion string, payload []byte, contentType string) *X509Request {
    return &X509Request{
        Method:      "PUT",
        URL:         url,
        APIVersion:  apiVersion,
        Payload:     payload,
        ContentType: contentType,
    }
}

type x509Response struct {
    StatusCode int
    // TODO: What exactly do we get back?  How will we know its encoding?
    Body   []byte
    Header http.Header
}

func performX509Request(session *x509Session, request *X509Request) (*x509Response, error) {
    Debugf("Request: %s %s", request.Method, request.URL)
    if len(request.Payload) > 0 {
        Debugf("Request body:\n%s", request.Payload)
    }

    bodyReader := ioutil.NopCloser(bytes.NewReader(request.Payload))
    httpRequest, err := http.NewRequest(request.Method, request.URL, bodyReader)
    if err != nil {
        return nil, err
    }
    httpRequest.Header.Set("Content-Type", request.ContentType)
    httpRequest.Header.Set("x-ms-version", request.APIVersion)
    httpResponse, err := session.client.Do(httpRequest)
    if err != nil {
        return nil, err
    }
    if httpResponse.Body != nil {
        defer httpResponse.Body.Close()
    }

    response := &x509Response{}
    response.StatusCode = httpResponse.StatusCode
    response.Body, err = readAndClose(httpResponse.Body)
    if err != nil {
        return nil, err
    }
    response.Header = httpResponse.Header

    Debugf("Response: %d %s", response.StatusCode, http.StatusText(response.StatusCode))
    if response.Header != nil {
        buf := bytes.Buffer{}
        response.Header.Write(&buf)
        Debugf("Response headers:\n%s", buf.String())
    }
    if len(response.Body) > 0 {
        Debugf("Response body:\n%s", response.Body)
    }

    return response, nil
}
