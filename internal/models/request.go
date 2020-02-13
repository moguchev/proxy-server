package models

import (
	"mime/multipart"
	"net/http"
)

type Request struct {
	Method           string          `json:"method"`
	URL              string          `json:"path"`
	Proto            string          `json:"proto"`
	ProtoMajor       int             `json:"proto_major,omitempty"`
	ProtoMinor       int             `json:"proto_minor,omitempty"`
	Host             string          `json:"host"`
	Header           http.Header     `json:"header"`
	Body             []byte          `json:"body,omitempty"`
	ContentLength    int64           `json:"content_length"`
	TransferEncoding []string        `json:"transfer_encoding,omitempty"`
	Close            bool            `json:"close"`
	Form             string          `json:"form,omitempty"`
	PostForm         string          `json:"post_form,omitempty"`
	MultipartForm    *multipart.Form `json:"multipart_form,omitempty"`
	Trailer          http.Header     `json:"trailer,omitempty"`
	RemoteAddr       string          `json:"remote_addr,omitempty"`
	RequestURI       string          `json:"request_uri"`
}

type RequestJSON struct {
	ID      int      `json:"id"`
	IsHTTPS bool     `json:"is_https"`
	Path    string   `json:"path"`
	Req     *Request `json:"request"`
}

type Requests []*RequestJSON
