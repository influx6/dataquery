package sql

import (
	"database/sql"
	"log"
	"sync"
	"testing"

	_ "github.com/go-sql-driver/mysql"

	"github.com/influx6/data/query/adaptors"
	"github.com/influx6/data/query/parser"
	"github.com/influx6/flux"
)

func TestTableBuilder(t *testing.T) {
	var ws sync.WaitGroup
	ws.Add(1)

	com := adaptors.ChunkFileParser(parser.DefaultInspectionFactory)

	sql := TableBuilder(TemplatesQueries, RelQueries)

	com.Bind(sql, true)

	sql.React(func(r flux.Reactor, err error, d interface{}) {
		ws.Done()
		if err != nil {
			flux.FatalFailed(t, "Failed to parse query: %+s", err)
		}
		flux.LogPassed(t, "Received Parse query")
	}, true)

	com.Send("./../../fixtures/dataset.dq")

	ws.Wait()
	com.Close()
}

func TestEngine(t *testing.T) {
	var ws sync.WaitGroup
	ws.Add(1)

	reader := flux.FileLoader()
	qo := BasicQueroEngine()

	reader.Bind(qo, true)

	qo.React(func(r flux.Reactor, err error, d interface{}) {
		ws.Done()
		if err != nil {
			flux.FatalFailed(t, "Failed in Building sql.Statement, Error Received: %+s", err)
		}
		flux.LogPassed(t, "Successful created sql.Statement")
	}, true)

	reader.Send("./../../fixtures/dataset.dq")

	defer qo.Close()
	ws.Wait()
}

func prepareTable(t *testing.T, db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS test.users(id integer not null AUTO_INCREMENT primary key,name varchar(50),age integer,street varchar(50),stamp date)")

	if err != nil {
		flux.FatalFailed(t, "Creating sql test.users table: %+s", err)
	}

	flux.LogPassed(t, "Successfully exectuable create table to sql db: %+s", "test.users")

	_, err = db.Exec("INSERT INTO test.users(name,age,street) VALUES('alex',21,'lagos')")

	if err != nil {
		flux.FatalFailed(t, "Running sql insert into test.users table: %+s", err)
	}
	_, err = db.Exec("INSERT INTO test.users(name,age,street) VALUES('josh',32,'new york')")

	if err != nil {
		flux.FatalFailed(t, "Running sql insert into test.users table: %+s", err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS test.photos(url varchar(50),user_id integer,id int not null primary key AUTO_INCREMENT)")

	if err != nil {
		flux.FatalFailed(t, "Creating sql test.photos table: %+s", err)
	}

	_, err = db.Exec("INSERT INTO test.photos(url,user_id) VALUES('./images/sock.jpg',2)")

	if err != nil {
		flux.FatalFailed(t, "Running sql insert into test.photos table: %+s", err)
	}

	_, err = db.Exec("INSERT INTO test.photos(url,user_id) VALUES('./images/winnie.jpg',1)")

	if err != nil {
		flux.FatalFailed(t, "Running sql insert into test.photos table: %+s", err)
	}

	flux.LogPassed(t, "Successfully exectuable create table to sql db: %+s", "test.photos")
}

func dropTables(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DROP TABLE IF EXISTS test.users")

	if err != nil {
		flux.FatalFailed(t, "Failed sql drop test.users table: %+s", err)
	}

	flux.LogPassed(t, "Successfully exectuable DROP table to sql db: %+s", "test.users")

	_, err = db.Exec("DROP TABLE IF EXISTS test.photos")

	if err != nil {
		flux.FatalFailed(t, "Failed sql drop test.photos table: %+s", err)
	}

	flux.LogPassed(t, "Successfully exectuable DROP table to sql db: %+s", "test.photos")
}

func TestQuero(t *testing.T) {
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/test")

	if err != nil {
		flux.FatalFailed(t, "Creating sql connection: %+s", err)
	}

	defer db.Close()

	prepareTable(t, db)

	var ws sync.WaitGroup
	ws.Add(1)

	reader := flux.FileLoader()
	qo := Quero(db)
	reader.Bind(qo, true)

	qo.React(func(r flux.Reactor, err error, d interface{}) {
		ws.Done()
		if err != nil {
			flux.FatalFailed(t, "Failed in Building sql.Statement, Error Received: %+s", err)
		}
		log.Printf("Model data: %+q", d)
		flux.LogPassed(t, "Successful created sql.Statement")
	}, true)

	reader.Send("./../../fixtures/models.dq")

	defer qo.Close()

	ws.Wait()

	dropTables(t, db)
}
