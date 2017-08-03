package physical

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/armon/go-metrics"
	_ "github.com/denisenkom/go-mssqldb"
	log "github.com/mgutz/logxi/v1"
)

type MsSQLBackend struct {
	dbTable    string
	client     *sql.DB
	statements map[string]*sql.Stmt
	logger     log.Logger
}

func newMsSQLBackend(conf map[string]string, logger log.Logger) (Backend, error) {
	username, ok := conf["username"]
	if !ok {
		username = ""
	}

	password, ok := conf["password"]
	if !ok {
		password = ""
	}

	server, ok := conf["server"]
	if !ok || server == "" {
		return nil, fmt.Errorf("missing server")
	}

	database, ok := conf["database"]
	if !ok {
		database = "Vault"
	}

	table, ok := conf["table"]
	if !ok {
		table = "Vault"
	}

	appname, ok := conf["appname"]
	if !ok {
		appname = "Vault"
	}

	connectionTimeout, ok := conf["connectiontimeout"]
	if !ok {
		connectionTimeout = "30"
	}

	logLevel, ok := conf["loglevel"]
	if !ok {
		logLevel = "0"
	}

	schema, ok := conf["schema"]
	if !ok || schema == "" {
		schema = "dbo"
	}

	connectionString := fmt.Sprintf("server=%s;app name=%s;connection timeout=%s;log=%s", server, appname, connectionTimeout, logLevel)
	if username != "" {
		connectionString += ";user id=" + username
	}

	if password != "" {
		connectionString += ";password=" + password
	}

	db, err := sql.Open("mssql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mssql: %v", err)
	}

	if _, err := db.Exec("IF NOT EXISTS(SELECT * FROM sys.databases WHERE name = '" + database + "') CREATE DATABASE " + database); err != nil {
		return nil, fmt.Errorf("failed to create mssql database: %v", err)
	}

	dbTable := database + "." + schema + "." + table
	createQuery := "IF NOT EXISTS(SELECT 1 FROM " + database + ".INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE='BASE TABLE' AND TABLE_NAME='" + table + "' AND TABLE_SCHEMA='" + schema +
		"') CREATE TABLE " + dbTable + " (Path VARCHAR(512) PRIMARY KEY, Value VARBINARY(MAX))"

	if schema != "dbo" {
		if _, err := db.Exec("USE " + database); err != nil {
			return nil, fmt.Errorf("failed to switch mssql database: %v", err)
		}

		var num int
		err = db.QueryRow("SELECT 1 FROM sys.schemas WHERE name = '" + schema + "'").Scan(&num)

		switch {
		case err == sql.ErrNoRows:
			if _, err := db.Exec("CREATE SCHEMA " + schema); err != nil {
				return nil, fmt.Errorf("failed to create mssql schema: %v", err)
			}

		case err != nil:
			return nil, fmt.Errorf("failed to check if mssql schema exists: %v", err)
		}
	}

	if _, err := db.Exec(createQuery); err != nil {
		return nil, fmt.Errorf("failed to create mssql table: %v", err)
	}

	m := &MsSQLBackend{
		dbTable:    dbTable,
		client:     db,
		statements: make(map[string]*sql.Stmt),
		logger:     logger,
	}

	statements := map[string]string{
		"put": "IF EXISTS(SELECT 1 FROM " + dbTable + " WHERE Path = ?) UPDATE " + dbTable + " SET Value = ? WHERE Path = ?" +
			" ELSE INSERT INTO " + dbTable + " VALUES(?, ?)",
		"get":    "SELECT Value FROM " + dbTable + " WHERE Path = ?",
		"delete": "DELETE FROM " + dbTable + " WHERE Path = ?",
		"list":   "SELECT Path FROM " + dbTable + " WHERE Path LIKE ?",
	}

	for name, query := range statements {
		if err := m.prepare(name, query); err != nil {
			return nil, err
		}
	}

	return m, nil
}

func (m *MsSQLBackend) prepare(name, query string) error {
	stmt, err := m.client.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare '%s': %v", name, err)
	}

	m.statements[name] = stmt

	return nil
}

func (m *MsSQLBackend) Put(entry *Entry) error {
	defer metrics.MeasureSince([]string{"mssql", "put"}, time.Now())

	_, err := m.statements["put"].Exec(entry.Key, entry.Value, entry.Key, entry.Key, entry.Value)
	if err != nil {
		return err
	}

	return nil
}

func (m *MsSQLBackend) Get(key string) (*Entry, error) {
	defer metrics.MeasureSince([]string{"mssql", "get"}, time.Now())

	var result []byte
	err := m.statements["get"].QueryRow(key).Scan(&result)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	ent := &Entry{
		Key:   key,
		Value: result,
	}

	return ent, nil
}

func (m *MsSQLBackend) Delete(key string) error {
	defer metrics.MeasureSince([]string{"mssql", "delete"}, time.Now())

	_, err := m.statements["delete"].Exec(key)
	if err != nil {
		return err
	}

	return nil
}

func (m *MsSQLBackend) List(prefix string) ([]string, error) {
	defer metrics.MeasureSince([]string{"mssql", "list"}, time.Now())

	likePrefix := prefix + "%"
	rows, err := m.statements["list"].Query(likePrefix)

	var keys []string
	for rows.Next() {
		var key string
		err = rows.Scan(&key)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rows: %v", err)
		}

		key = strings.TrimPrefix(key, prefix)
		if i := strings.Index(key, "/"); i == -1 {
			keys = append(keys, key)
		} else if i != -1 {
			keys = appendIfMissing(keys, string(key[:i+1]))
		}
	}

	sort.Strings(keys)

	return keys, nil
}
