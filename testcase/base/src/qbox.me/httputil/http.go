package httputil 

import (
    "io"
)

type Http interface {
	CallWithForm(ret interface{}, url string, param map[string][]string) (code int, err error)

    CallWith(ret interface{}, url string, bodyType string, body io.Reader, bodyLength int64) (code int, err error)

    Call(ret interface{}, url string) (code int, err error) 
}

