package main

/*
   #include <stdlib.h>
*/
import "C"

import (
	"io"
	"net/http"
	"strconv"
	"unsafe"
)

func parseUid(str *C.char) int64 {
	uid, err := parseUidWithError(str)
	if err != nil {
		panic(err)
	}

	return uid
}

func parseUidWithError(str *C.char) (int64, error) {
	return strconv.ParseInt(C.GoString(str), 10, 64)
}

func uidToCString(uid int64) *C.char {
	return C.CString(strconv.FormatInt(uid, 10))
}

func freeCString(strings ...*C.char) {
	for _, str := range strings {
		C.free(unsafe.Pointer(str))
	}
}

func fetchUrl(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return readAll(resp.Body, resp.ContentLength)
}

func readAll(reader io.Reader, length int64) ([]byte, error) {
	data := make([]byte, length)
	nbytes := 0
	for nbytes < len(data) {
		n, err := reader.Read(data[nbytes:])
		if err != nil && nbytes+n < len(data) {
			return nil, err
		}

		nbytes += n
	}

	return data, nil
}
