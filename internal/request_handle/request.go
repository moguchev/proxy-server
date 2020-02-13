package request_handle

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/moguchev/proxy-server/internal/models"
)

const (
	MAX_MEMORY = 1 * 1024 * 1024
)

func ConvertModelToRequest(request models.Request, isHTTPS bool) (*http.Request, error) {
	urlStruct := &url.URL{}
	err := urlStruct.UnmarshalBinary([]byte(request.URL))
	if err != nil {
		log.Println(err)
		return nil, err //, false
	}

	if urlStruct.Scheme == "" {
		if isHTTPS {
			urlStruct.Scheme = "https"
		} else {
			urlStruct.Scheme = "http"
		}
	}

	if urlStruct.Host == "" {
		urlStruct.Host = request.Host
	}

	form, err := url.ParseQuery(request.Form)
	if err != nil {
		return nil, err //, false
	}

	postForm, err := url.ParseQuery(request.PostForm)
	if err != nil {
		return nil, err //, false
	}

	req := &http.Request{
		Method:           request.Method,
		URL:              urlStruct,
		Proto:            request.Proto,
		ProtoMajor:       request.ProtoMajor,
		ProtoMinor:       request.ProtoMinor,
		Host:             request.Host,
		Header:           request.Header,
		Body:             ioutil.NopCloser(bytes.NewReader(request.Body)),
		ContentLength:    request.ContentLength,
		TransferEncoding: request.TransferEncoding,
		Close:            request.Close,
		Form:             form,
		PostForm:         postForm,
		MultipartForm:    request.MultipartForm,
		Trailer:          request.Trailer,
		RemoteAddr:       request.RemoteAddr,
		RequestURI:       request.RequestURI,
	}
	return req, nil // , request.IsHTTPS
}

func ConvertRequestToModel(r *http.Request, isHTTPS bool) (*models.Request, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	err = r.ParseForm()
	if err != nil {
		return nil, err
	}
	binUrl, err := r.URL.MarshalBinary()
	if err != nil {
		return nil, err
	}
	model := &models.Request{
		Method:           r.Method,
		URL:              string(binUrl),
		Proto:            r.Proto,
		ProtoMajor:       r.ProtoMajor,
		ProtoMinor:       r.ProtoMinor,
		Host:             r.Host,
		Header:           r.Header,
		Body:             body,
		ContentLength:    r.ContentLength,
		TransferEncoding: r.TransferEncoding,
		Close:            r.Close,
		Form:             r.Form.Encode(),
		PostForm:         r.PostForm.Encode(),
		Trailer:          r.Trailer,
		RemoteAddr:       r.RemoteAddr,
		RequestURI:       r.RequestURI,
	}

	err = r.ParseMultipartForm(MAX_MEMORY)
	if err != nil {
		if err.Error() == "request Content-Type isn't multipart/form-data" {
			err = r.ParseMultipartForm(MAX_MEMORY)
			model.MultipartForm = r.MultipartForm
		} else {
			return nil, err
		}
	}
	return model, nil
}

func ConvertRawRequestToModel(buf []byte, isHTTPS bool) (*models.Request, error) {
	bufReader := bufio.NewReader(bytes.NewBuffer(buf))
	request, err := http.ReadRequest(bufReader)
	if err != nil {
		return nil, err
	}
	return ConvertRequestToModel(request, isHTTPS)
}

func ConvertModelToRawRequest(request models.Request, isHTTPS bool) ([]byte, error) {
	req, err := ConvertModelToRequest(request, isHTTPS)
	if err != nil {
		return nil, err //, false
	}
	retBuf, err := httputil.DumpRequest(req, true)
	return retBuf, err //, isHTTPS
}
