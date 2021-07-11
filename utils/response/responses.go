package response

import (
	"encoding/json"
	"fmt"
	"goventy/utils"
	"net/http"
)

func encodeResponseData(data interface{}) []byte {
	b, err := json.Marshal(data)
	if err != nil {
		utils.Log(err, "error")
	}
	return b
}

func respond(w http.ResponseWriter, data []byte, statusCode int, headers map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if headers != nil {
		for key, value := range headers {
			w.Header().Set(key, value)
		}
	}
	fmt.Fprintf(w, "%s", data)
}

func RespondSuccess(w http.ResponseWriter, data interface{}, arrayOfHeaders ...map[string]string) {
	var headers map[string]string
	headers = nil
	if len(arrayOfHeaders) > 0 {
		headers = arrayOfHeaders[0]
	}
	respond(w, encodeResponseData(data), http.StatusOK, headers)
}

func RespondUnauthorized(w http.ResponseWriter, data interface{}, arrayOfHeaders ...map[string]string) {
	var headers map[string]string
	headers = nil
	if len(arrayOfHeaders) > 0 {
		headers = arrayOfHeaders[0]
	}
	respond(w, encodeResponseData(data), http.StatusUnauthorized, headers)
}

func RespondBadRequest(w http.ResponseWriter, data interface{}, arrayOfHeaders ...map[string]string) {
	var headers map[string]string
	headers = nil
	if len(arrayOfHeaders) > 0 {
		headers = arrayOfHeaders[0]
	}
	respond(w, encodeResponseData(data), http.StatusBadRequest, headers)
}

func RespondInternalServerError(w http.ResponseWriter, data interface{}, arrayOfHeaders ...map[string]string) {
	var headers map[string]string
	headers = nil
	if len(arrayOfHeaders) > 0 {
		headers = arrayOfHeaders[0]
	}
	respond(w, encodeResponseData(data), http.StatusInternalServerError, headers)
}

func RespondRequestTooLarge(w http.ResponseWriter, data interface{}, arrayOfHeaders ...map[string]string) {
	var headers map[string]string
	headers = nil
	if len(arrayOfHeaders) > 0 {
		headers = arrayOfHeaders[0]
	}
	respond(w, encodeResponseData(data), http.StatusRequestEntityTooLarge, headers)
}
