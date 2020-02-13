package repeater

import (
	"github.com/go-chi/chi"
	"github.com/moguchev/proxy-server/internal/models"
	"net/http"
	"net/url"
	"strconv"
)

func (server *Server) GetHistory(w http.ResponseWriter, r *http.Request) {
	paramsMap, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rowsStr := paramsMap.Get("rows")
	pageStr := paramsMap.Get("page")
	if len(rowsStr) == 0 || len(pageStr) == 0 {
		http.Error(w, "no rows or page in query", http.StatusBadRequest)
		return
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rows, err := strconv.Atoi(rowsStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	offset := (page - 1) * rows
	requests, err := server.db.GetRequests(offset, rows)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(requests) == 0 {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	WriteToResponse(w, http.StatusOK, requests)
}

func (server *Server) GetRequest(w http.ResponseWriter, r *http.Request) {
	request := server.dealGettingRequest(w, r)
	if request != nil {
		WriteToResponse(w, http.StatusOK, request)
	}
}

func (server *Server) PerformRequest(w http.ResponseWriter, r *http.Request) {
	request := models.RequestJSON{}
	err := ReadFromBody(r, w, &request)
	if err != nil {
		return
	}
	server.MakeRequest(&request, w)
	//WriteToResponse(w, http.StatusOK, response)
}

func (server *Server) RepeatRequest(w http.ResponseWriter, r *http.Request) {
	request := server.dealGettingRequest(w, r)
	if request != nil {
		server.MakeRequest(request, w)
		//if err != nil {
		//	http.Error(w, err.Error(), http.StatusInternalServerError)
		//	return
		//}
		//if err != nil {
		//	log.Printf("couldn't write response: %v", err)
		//}
	}
}

func (server *Server) dealGettingRequest(w http.ResponseWriter, r *http.Request) *models.RequestJSON{
	strId := chi.URLParam(r, "id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil
	}
	request, err := server.db.GetRequest(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	if request == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return nil
	}
	return request
}
