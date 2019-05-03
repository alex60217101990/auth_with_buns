package models

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"auth_service_template/logger"

	"github.com/gocql/gocql"
)

type DB struct {
	conn interface{}
	log  *logger.Log
}

func NewDB(glob_env *map[string]string, logs *logger.Log) *DB {
	db := DB{
		log: logs,
	}
	if (*glob_env)["DB_DRIVER"] == "cassandra" {
		// Provide the cassandra cluster instance here.
		cluster := gocql.NewCluster((*glob_env)["CASSANDRA_CLUSTER"])

		cluster.Keyspace = "system"
		cluster.Timeout = 20 * time.Second
		sys_session, err := cluster.CreateSession()
		if err != nil {
			logs.Println(logger.NewError(
				"cassandra",
				fmt.Sprintf("CreateSession error: %v", err),
				logger.FATAL,
			))
		}

		err = sys_session.Query(fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s
    		WITH replication = {
        	'class' : 'SimpleStrategy',
        	'replication_factor' : %d
		}`, (*glob_env)["CASSANDRA_KEYSPACE"], 2)).Exec()
		if err != nil {
			logs.Println(logger.NewError(
				"cassandra",
				fmt.Sprintf("Create Keyspace error: %v", err),
				logger.FATAL,
			))
		}
		sys_session.Close()

		// gocql requires the keyspace to be provided before the session is created.
		// In future there might be provisions to do this later.
		cluster.Keyspace = (*glob_env)["CASSANDRA_KEYSPACE"]

		// This is time after which the creation of session call would timeout.
		// This can be customised as needed.
		if i32, err := strconv.Atoi((*glob_env)["CASSANDRA_TIMEOUT"]); err == nil {
			cluster.Timeout = time.Duration(i32) * time.Second
		}

		if i32, err := strconv.Atoi((*glob_env)["CASSANDRA_PROTO_VERSION"]); err == nil {
			cluster.ProtoVersion = i32
		}

		session, err := cluster.CreateSession()
		if err != nil {
			logs.Println(logger.NewError(
				"cassandra",
				fmt.Sprintf("Could not connect to cassandra cluster: %v", err),
				logger.FATAL,
			))
		}

		// Check if the table already exists. Create if table does not exist
		keySpaceMeta, _ := session.KeyspaceMetadata((*glob_env)["CASSANDRA_KEYSPACE"])

		if _, exists := keySpaceMeta.Tables["users"]; exists != true {
			// Create a table
			err := session.Query(`CREATE TABLE users (
				id uuid, first_name text, last_name text, 
				login text, phones list<text>, age int, 
				password text, emails list<text>, PRIMARY KEY (id, login)
				) WITH CLUSTERING ORDER BY (login ASC);`).Exec()
			if err != nil {
				logs.Println(logger.NewError(
					"cassandra",
					fmt.Sprintf("Could not create Users table: %v", err),
					logger.FATAL,
				))
			}
			// Create a index
			err = session.Query(`CREATE INDEX IF NOT EXISTS emails_index ON users (emails);`).Exec()
			if err != nil {
				logs.Println(logger.NewError(
					"cassandra",
					fmt.Sprintf("Could not create Users emails index: %v", err),
					logger.FATAL,
				))
			}
		}
		db.conn = session
	}
	return &db
}

func (db *DB) Close() {
	switch c := db.conn.(type) {
	case *sql.DB:
		c.Close()
	case *gocql.Session:
		c.Close()
	}
}
