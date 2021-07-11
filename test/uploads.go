package test

import (
	"net/http"
	"mime/multipart"
	"strings"
	"io"
	"os"
	"path/filepath"
	"bytes"
)

func CreateFormDataUploadRequest(uri string, method string,params map[string][]string, files map[string][]string, headers map[string]string) (*http.Request, error){
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	//write string parameters
	for key, values := range params {
		for _, value := range values {
			_ = writer.WriteField(key, value)
		}
	}
	//add files
	for key, paths := range files {
		for _, path := range paths {
			//get the file in the filepath
			file, err := os.Open(path)
			if err != nil {
				return nil, err
			}
			defer file.Close()

			part, err := writer.CreateFormFile(key, filepath.Base(path))
			if err != nil {
				return nil, err
			}
			_, err = io.Copy(part, file)
			if err != nil {
				return nil, err
			}
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	//set headers
	for key, value := range headers {
		if strings.ToLower(key) == "content-type" {
			continue
		}
		req.Header.Set(key, value)
	}
	return req, nil
}