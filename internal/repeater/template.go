package repeater

import (
	"github.com/moguchev/proxy-server/internal/models"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

func (server *Server) GetIndexTemplate(w http.ResponseWriter, r *http.Request) {
	paramsMap, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rowsStr := paramsMap.Get("rows")
	pageStr := paramsMap.Get("page")
	page := 1
	rows := 10
	if len(rowsStr) != 0 && len(pageStr) != 0 {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		rows, err = strconv.Atoi(rowsStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	offset := (page - 1) * rows
	requests, err := server.db.GetRequests(offset, rows)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	err = tmpl.Execute(w, struct {
		Requests []*models.RequestJSON
		Prev int
		Next int
		Len int
	}{
		requests,
		page - 1,
		page + 1,
		len(requests),
	})
	if err != nil {
		log.Println(err)
	}
}

func (server *Server) ChangeRequestTemplate(w http.ResponseWriter, r *http.Request) {
	req := server.dealGettingRequest(w, r)
	tmpl := template.Must(template.ParseFiles("templates/request.html"))
	err := tmpl.Execute(w, struct {
		Request *models.RequestJSON
	}{
		req,
	})
	if err != nil {
		log.Println(err)
	}
}