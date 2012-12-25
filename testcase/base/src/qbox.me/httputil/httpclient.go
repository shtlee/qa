package httputil

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"qbox.me/errcode"
	"strings"
)
 
// --------------------------------------------------------------------

type Client struct {
	*http.Client
}

func (r *Client) doPost(url, host string, bodyType string, body io.Reader, bodyLength int64) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", bodyType)
	if host != "" {
		req.Host = host
	}
	req.ContentLength = bodyLength
	return r.Do(req)
}

func doGet(url, host string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if host != "" {
		req.Host = host
	}
	resp, err = http.DefaultClient.Do(req)
	return resp, err
}

func (r *Client) doPostForm(url_, host string, data map[string][]string) (resp *http.Response, err error) {
	msg := url.Values(data).Encode()
	return r.doPost(url_, host, "application/x-www-form-urlencoded", strings.NewReader(msg), (int64)(len(msg)))
}

// ------------------------------ helpers ------------------------------

func (r *Client) CallWithFormEx(ret interface{}, url, host string, param map[string][]string) (code int, err error) {

	resp, err := r.doPostForm(url, host, param)
	if err != nil {
		return errcode.InternalError, err
	}
	return callRet(ret, resp)
}

func (r *Client) CallWithForm(ret interface{}, url string, param map[string][]string) (code int, err error) {
	return r.CallWithFormEx(ret, url, "", param)
}


func (r *Client) CallWithEx(ret interface{}, url, host string, bodyType string, body io.Reader, bodyLength int64) (code int, err error) {

	resp, err := r.doPost(url, host, bodyType, body, int64(bodyLength))
	if err != nil {
		return errcode.InternalError, err
	}
	return callRet(ret, resp)
}

func (r *Client) CallWith(ret interface{}, url string, bodyType string, body io.Reader, bodyLength int64) (code int, err error) {
	return r.CallWithEx(ret, url, "", bodyType, body, bodyLength)
}


func (r *Client) CallEx(ret interface{}, url, host string) (code int, err error) {

	resp, err := r.doPost(url, host, "application/x-www-form-urlencoded", nil, 0)
	if err != nil {
		return errcode.InternalError, err
	}
	return callRet(ret, resp)	
}

func (r *Client) Call(ret interface{}, url string) (code int, err error) {

	return r.CallEx(ret, url, "")
}

func (c *Client) DownloadEx(url, host string) (r io.ReadWriter, err error) {

	resp, err := doGet(url, host)
	defer resp.Body.Close()
	if err != nil {
		return
	}
	r = new(bytes.Buffer)
	io.Copy(r, resp.Body)
	return r, err

}

const (
	NetWorkError = 102
)


type ErrorRet struct {
	Error string "error"
}

func callRet(ret interface{}, resp *http.Response) (code int, err error) {
	defer resp.Body.Close()
	code = resp.StatusCode
	if code/100 == 2 {
		if ret == nil || resp.ContentLength == 0 {
			return
		}
		switch ret.(type) {
		case io.Writer:
			w := ret.(io.Writer)
			io.Copy(w, resp.Body)
			break
		default:
			err = json.NewDecoder(resp.Body).Decode(ret)
			if err != nil {
				code = errcode.UnexceptedResponse
			}
		}
	} else {
		if resp.ContentLength != 0 {
			if ct, ok := resp.Header["Content-Type"]; ok && ct[0] == "application/json" {
				var ret1 ErrorRet
				json.NewDecoder(resp.Body).Decode(&ret1)
				if ret1.Error != "" {
					err = errors.New(ret1.Error)
					return
				}
			}
		}
		err = errcode.Errno(code)
	}
	return
}

func (r *Client) PostMultipart(url_, host string, data map[string][]string) (resp *http.Response, err error) {

	body, ct, err := Open(data)
	if err != nil {
		return
	}
	defer body.Close()
	return r.doPost(url_, host, ct, body, -1)	
}

func (r *Client) CallWithMultipartEx(ret interface{}, url, host string, param map[string][]string) (code int, err error) {

	resp, err := r.PostMultipart(url, host, param)
	if err != nil {
		return 201, err
	}
	return callRet(ret, resp)
}

func (r *Client) CallWithMultipart(ret interface{}, url string, param map[string][]string) (code int, err error) {

	return r.CallWithMultipartEx(ret, "", url, param)
}


// ------------------------ default client helper -------------------------- //

var (
	DefaultClient = Client{http.DefaultClient}
)

func CallWithFormEx(ret interface{}, url, host string, param map[string][]string) (code int, err error) {

	return DefaultClient.CallWithFormEx(ret, url, host, param)
}

func CallWithForm(ret interface{}, url string, param map[string][]string) (code int, err error) {

	return DefaultClient.CallWithFormEx(ret, url, "", param)
}

func CallWithEx(ret interface{}, url, host string, bodyType string, body io.Reader, bodyLength int64) (code int, err error) {

	return DefaultClient.CallWithEx(ret, url, host, bodyType, body, bodyLength)
}

func CallWith(ret interface{}, url string, bodyType string, body io.Reader, bodyLength int64) (code int, err error) {

	return DefaultClient.CallWithEx(ret, url, "", bodyType, body, bodyLength)
}

func CallEx(ret interface{}, url, host string) (code int, err error) {

	return DefaultClient.CallEx(ret, url, host)
}

func Call(ret interface{}, url string) (code int, err error) {

	return DefaultClient.CallEx(ret, url, "")
}

func DownloadEx(url, host string) (r io.ReadWriter, err error) {

	return DefaultClient.DownloadEx(url, host)
}

func Download(url string) (r io.ReadWriter, err error) {

	return DefaultClient.DownloadEx(url, "")
}