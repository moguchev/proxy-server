package repeater

import (
	"crypto/tls"
	"fmt"
	"github.com/moguchev/proxy-server/internal/models"
	"github.com/moguchev/proxy-server/internal/request_handle"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func (server *Server) MakeRequest(request *models.RequestJSON, w http.ResponseWriter) {
	if request.IsHTTPS {
		server.makeHTTPSRequest(request, w)
	} else  {
		server.makeHTTPRequest(request, w)
	}
}

func (server *Server) makeHTTPSRequest(request *models.RequestJSON, w http.ResponseWriter) {
	req, err := request_handle.ConvertModelToRequest(*request.Req, request.IsHTTPS)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = server.db.SaveRequest(request)
	if err != nil {
		log.Printf("oops, request wasn't saved: %v", err)
	}
	req.RequestURI = ""
	req.URL.Scheme = strings.ToLower(strings.Split(req.Proto, "/")[0])
	req.URL.Host = req.Host
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	binary, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	for key, values := range resp.Header {
		w.Header().Set(key, strings.Join(values, ", "))
	}
	_, err = w.Write(binary)
	if err != nil {
		log.Println(err)
	}
}

func (server *Server) makeHTTPRequest(request *models.RequestJSON, w http.ResponseWriter) {
	r, err := request_handle.ConvertModelToRequest(*request.Req, request.IsHTTPS)
	if err != nil {
	}
	fmt.Println(r)
	server.saveRequest(r, false)
	response, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	if response != nil {
		defer func() {
			_ = response.Body.Close()
		}()
		binary, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Println(err)
		}
		for key, values := range response.Header {
			w.Header().Set(key, strings.Join(values, ", "))
		}
		_, err = w.Write(binary)
		if err != nil {
			log.Println(err)
		}
	}

}


func (server *Server) saveRequest(request *http.Request, isHTTPS bool) {
	model, err := request_handle.ConvertRequestToModel(request, isHTTPS)
	if err != nil {
		log.Printf("Request wasn't saved: %s", err)
		return
	}

	modelToSave := &models.RequestJSON{
		Req:     model,
		Path:    model.Host + model.URL,
		IsHTTPS: isHTTPS,
	}
	_, err = server.db.SaveRequest(modelToSave)

	if err != nil {
		log.Printf("Request wasn't saved: %s", err)
	}
}

func (server *Server) saveRawRequest(request []byte, isHTTPS bool) {
	model, err := request_handle.ConvertRawRequestToModel(request, isHTTPS)
	if err != nil {
		log.Printf("Request wasn't saved: %s", err)
		return
	}

	modelToSave := &models.RequestJSON{
		Req:     model,
		Path:    model.Host + model.URL,
		IsHTTPS: isHTTPS,
	}
	_, err = server.db.SaveRequest(modelToSave)
	if err != nil {
		log.Printf("Request wasn't saved: %s", err)
	}
}
