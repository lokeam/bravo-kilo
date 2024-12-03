package httputils

import "net/http"

func IsHeaderWritten(w http.ResponseWriter) bool {
	if rw, ok := w.(interface{ Written() bool }); ok {
			return rw.Written()
	}
	return false
}