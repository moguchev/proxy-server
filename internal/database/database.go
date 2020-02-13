package database

import (
	"encoding/json"
	"github.com/moguchev/proxy-server/internal/models"
	_ "github.com/lib/pq"
	
	"gopkg.in/jackc/pgx.v2"
	"time"
)

const (
	saveRequest   = "INSERT INTO requests(url, is_https, data) VALUES($1, $2, $3) RETURNING req_id;"
	selectRequest = "SELECT * FROM requests WHERE req_id = $1;"
	pageRequests  = "SELECT req.* FROM requests req " +
		"JOIN ( SELECT req_id FROM requests req ORDER BY req_id DESC " +
		"LIMIT $1 OFFSET $2) sub_q ON (req.req_id = sub_q.req_id) ORDER BY req_id DESC;"
)

type DB struct {
	db           *pgx.ConnPool
	user         string
	password     string
	databaseName string
	host         string
	port         uint16
}

func NewDB(user string, password string, dataBaseName string,
	host string, port uint16) *DB {
	db := new(DB)
	db.user = user
	db.databaseName = dataBaseName
	db.password = password
	db.host = host
	db.port = port
	return db
}

func (db *DB) Start() error {
	conf := pgx.ConnConfig{
		Host:     db.host,
		Port:     db.port,
		User:     db.user,
		Password: db.password,
		Database: db.databaseName,
	}
	poolConf := pgx.ConnPoolConfig{
		ConnConfig:     conf,
		MaxConnections: 80,
		AcquireTimeout: time.Duration(7 * time.Second),
	}
	dataBase, err := pgx.NewConnPool(poolConf)
	if err != nil {
		return err
	}
	db.db = dataBase
	return nil
}

func (db *DB) Close() {
	db.db.Close()
}

func (db *DB) StartTransaction() (*pgx.Tx, error) {
	return db.db.Begin()
}

func (db *DB) SaveRequest(request *models.RequestJSON) (int, error) {
	rawRequest, err := json.Marshal(request.Req)
	if err != nil {
		return 0, err
	}
	row := db.db.QueryRow(saveRequest, request.Path, request.IsHTTPS, rawRequest)
	id := 0
	err = row.Scan(&id)
	return id, err
}

func (db *DB) GetRequest(reqId int) (*models.RequestJSON, error) {
	row := db.db.QueryRow(selectRequest, reqId)
	req := new(models.RequestJSON)
	bytesReq := make([]byte, 0)
	err := row.Scan(&req.ID, &req.Path, &req.IsHTTPS, &bytesReq)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytesReq, &req.Req)

	if err != nil {
		return nil, err
	}

	return req, nil
}

func (db *DB) GetRequests(offset int, limit int) (models.Requests, error) {
	rows, err := db.db.Query(pageRequests, limit, offset)
	if err == pgx.ErrNoRows {
		requests := make(models.Requests, 0)
		return requests, nil
	}
	if err != nil {
		return nil, err
	}
	requests := make(models.Requests, 0)
	defer rows.Close()
	for rows.Next() {
		req := new(models.RequestJSON)
		bytesReq := make([]byte, 0)
		err := rows.Scan(&req.ID, &req.Path, &req.IsHTTPS, &bytesReq)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytesReq, &req.Req)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}
