package proxy

import (
	"bytes"
	"crypto/tls"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/moguchev/proxy-server/internal/certificates"
	"github.com/moguchev/proxy-server/internal/config"
	"github.com/moguchev/proxy-server/internal/database"
	"github.com/moguchev/proxy-server/internal/models"
	"github.com/moguchev/proxy-server/internal/request_handle"
)

const (
	OKHeader = "HTTP/1.1 200 OK\r\n\r\n"
)

type Server struct {
	ca         tls.Certificate
	httpClient *http.Client
	db         *database.DB
	config     *config.Config
}

func NewServer(pathToConfig string) (*Server, error) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	newConfig, err := config.NewConfig(pathToConfig)
	if err != nil {
		return nil, err
	}

	server := new(Server)
	server.config = newConfig

	dbPort, err := strconv.Atoi(server.config.DBPort)
	if err != nil {
		return nil, err
	}

	server.db = database.NewDB(server.config.DBUser, server.config.DBPass, server.config.DBName,
		server.config.DBHost, uint16(dbPort))

	server.ca, err = certificates.LoadCA()
	if err != nil {
		return nil, err
	}

	server.httpClient = new(http.Client)
	server.httpClient.Timeout = 5 * time.Second
	return server, nil
}

func (server *Server) Run() error {
	err := server.db.Start()
	if err != nil {
		return err
	}

	log.Printf("Server is running on port: %s\n", server.config.HttpsPort)
	defer server.db.Close()

	err = http.ListenAndServe(":"+server.config.HttpsPort, server)

	return err
}

func (server *Server) ManageHttpRequest(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL)

	err := server.saveRequest(r, false)
	if err != nil {
		log.Printf("Request wasn't saved: %s", err)
	}

	// get response
	response, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		log.Printf("round trip error: %s", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	defer func() {
		_ = response.Body.Close()
	}()

	// Copy response`s headers
	for key, value := range response.Header {
		for _, subValue := range value {
			w.Header().Add(key, subValue)
		}
	}

	// Copy response`s body
	_, err = io.Copy(w, response.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
}

func (server *Server) LaunchSecureConnection(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL)

	leafCert, err := certificates.GenerateCert(&server.ca, r.Host)
	if err != nil {
		log.Fatalf("Error while generating certificates: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	curCert := make([]tls.Certificate, 1)
	curCert[0] = *leafCert

	curConfig := &tls.Config{
		Certificates: curCert,
		GetCertificate: func(info *tls.ClientHelloInfo) (certificate *tls.Certificate, e error) {
			return certificates.GenerateCert(&server.ca, info.ServerName)
		},
	}

	serverConn, err := tls.Dial("tcp", r.Host, curConfig)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		log.Println("Hijacking not supported")
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	conn, _, err := hijacker.Hijack()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	_, err = conn.Write([]byte(OKHeader))
	if err != nil {
		log.Printf("Unable to install conn: %v", err)
		_ = conn.Close()
		return
	}

	clientConn := tls.Server(conn, curConfig)
	err = clientConn.Handshake()
	if err != nil {
		log.Printf("Unable to handshake: %v", err)
		_ = clientConn.Close()
		_ = conn.Close()
		return
	}

	f := func(dst io.WriteCloser, src io.ReadCloser, isSaved bool) {
		if src != nil && dst != nil {
			defer func() {
				if dst != nil {
					_ = dst.Close()
				}

			}()
			defer func() {
				if src != nil {
					_ = src.Close()
				}
			}()
			buf := new(bytes.Buffer)
			multiWriter := io.MultiWriter(dst, buf)
			_, err = io.Copy(multiWriter, ioutil.NopCloser(src))
			if err != nil {
				log.Println(err)
				//return
			}
			if isSaved {
				//fmt.Println(string(buf.Bytes()))
				server.saveRawRequest(buf.Bytes(), true)
			}
		}
	}

	go f(serverConn, clientConn, true)
	go f(clientConn, serverConn, false)
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := recover()
		if err != nil {
			log.Println(err)
		}
	}()
	if r.Method == http.MethodConnect {
		server.LaunchSecureConnection(w, r)
	} else {
		server.ManageHttpRequest(w, r)
	}
}

func (server *Server) saveRequest(request *http.Request, isHTTPS bool) error {
	model, err := request_handle.ConvertRequestToModel(request, isHTTPS)
	if err != nil {
		return err
	}

	modelToSave := &models.RequestJSON{
		Req:     model,
		Path:    model.Host + model.URL,
		IsHTTPS: isHTTPS,
	}
	_, err = server.db.SaveRequest(modelToSave)

	return err
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
