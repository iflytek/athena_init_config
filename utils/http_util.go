package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"unsafe"
)

var errNIL = errors.New("Response is nil")

func HttpGet(url string, headers map[string]string, in ...interface{}) *Response {
	return doHttpRequest("GET", url, headers, nil, in...)
}

func HttpPost(url string, headers map[string]string, body interface{}, in ...interface{}) *Response {
	return doHttpRequest("POST", url, headers, body, in...)
}

type Response struct {
	Body       []byte
	StatusCode int
	Err        error
	HttpResp   *http.Response
}

func (r *Response) GetBody() []byte {
	if r == nil {
		return nil
	}
	return r.Body
}

func (r *Response) GetStatusCode() int {
	if r == nil {
		return -1
	}
	return r.StatusCode
}

func (r *Response) String() string {
	if r == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%d|%s|%v", r.StatusCode, string(r.Body), r.Err)
}

func (r *Response) GetError() error {
	if r == nil {
		return errNIL
	}
	return r.Err
}

func (r *Response) Decode(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

var client = &http.Client{
	Transport: &http.Transport{TLSClientConfig: &tls.Config{
		InsecureSkipVerify: true,
	}},
}

func doHttpRequest(method, requestUrl string, headers map[string]string, body interface{}, in ...interface{}) *Response {
	var bb io.Reader
	switch body.(type) {
	case string:
		bb = bytes.NewReader(Bytes(body.(string)))
	case []byte:
		bb = bytes.NewReader(body.([]byte))
	case nil:
		bb = nil
	case io.Reader:
		bb = body.(io.Reader)
	default:
		d, err := json.Marshal(body)
		if err != nil {
			return &Response{Err: err}
		}
		bb = bytes.NewReader(d)
	}
	//set default content-type
	if headers != nil && headers["Content-type"] == "" {
		headers["Content-type"] = "application/json"
	}
	req, err := http.NewRequest(method, requestUrl, bb)
	if err != nil {
		return &Response{Err: err}
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	rsp, err := client.Do(req)
	if err != nil {
		return &Response{Err: err}
	}

	data, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return &Response{
			Body:       nil,
			StatusCode: rsp.StatusCode,
			Err:        err,
			HttpResp:   rsp,
		}
	}

	for _, v := range in {
		if err := json.Unmarshal(data, v); err != nil {
			return &Response{
				Body:       data,
				StatusCode: rsp.StatusCode,
				Err:        err,
				HttpResp:   rsp,
			}
		}
	}

	return &Response{
		Body:       data,
		StatusCode: rsp.StatusCode,
		Err:        nil,
		HttpResp:   rsp,
	}
}

func Bytes2Str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func String(v interface{}) string {
	switch v.(type) {
	case string:
		return v.(string)
	case []byte:
		p := v.([]byte)
		return *(*string)(unsafe.Pointer(&p))
	case fmt.Stringer:
		return v.(fmt.Stringer).String()
	}
	return fmt.Sprintf("%v", v)
}

func Bytes(v string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&v))
	bh := &reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(bh))
}
