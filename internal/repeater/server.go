package repeater

import (
	"crypto/tls"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/moguchev/proxy-server/internal/certificates"
	"github.com/moguchev/proxy-server/internal/config"
	"github.com/moguchev/proxy-server/internal/database"
	"log"
	"net/http"
	"strconv"
)

type Server struct {
	router *chi.Mux
	db     *database.DB
	config *config.Config
	ca tls.Certificate
}

func NewServer(pathToConfig string) (*Server, error) {
	newConfig, err := config.NewConfig(pathToConfig)
	if err != nil {
		return nil, err
	}
	server := new(Server)

	server.ca, err = certificates.LoadCA()
	if err != nil {
		return nil, err
	}


	server.config = newConfig
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// golang templates usage
	router.Get("/", server.GetIndexTemplate)
	router.Get("/repeat/{id:^[0-9]+$}", server.RepeatRequest)
	router.Get("/change/{id:^[0-9]+$}", server.ChangeRequestTemplate)

	// rest api, later maybe do contemporary frontend(no)
	subRouter := chi.NewRouter()
	subRouter.Get("/history", server.GetHistory)
	subRouter.Get("/request/{id:^[0-9]+$}", server.GetRequest)
	subRouter.Put("/request/{id:^[0-9]+$}", server.RepeatRequest)
	subRouter.Post("/request/", server.PerformRequest)
	router.Mount("/api/", subRouter)
	server.router = router
	dbPort, err := strconv.Atoi(server.config.DBPort)
	if err != nil {
		return nil, err
	}
	db := database.NewDB(server.config.DBUser, server.config.DBPass,
		server.config.DBName, server.config.DBHost, uint16(dbPort))
	server.db = db
	return server, nil
}

func (server *Server) Run() error {
	log.Printf("running https on port %s\n", server.config.HttpsPort)
	err := server.db.Start()
	if err != nil {
		log.Printf("Failed to connect to DB: %v", err)
		return err
	}
	defer server.db.Close()
	log.Fatal(http.ListenAndServe(":"+server.config.HttpsPort, server.router))
	return nil
}

