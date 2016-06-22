package gorpx

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/go-gorp/gorp"
	"github.com/zew/awis/config"
	"github.com/zew/awis/logx"
	"github.com/zew/awis/mdl"
	"github.com/zew/awis/util"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

var dbmap *gorp.DbMap
var db *sql.DB

func DB() *sql.DB {
	if db == nil {
		DBMap()
	}
	return db
}

func DBMap(dbName ...string) *gorp.DbMap {

	if dbmap != nil && db != nil {
		return dbmap
	}

	sh := config.Config.SQLHosts[util.Env()]
	var err error
	// param docu at https://github.com/go-sql-driver/mysql
	paramsJoined := "?"
	for k, v := range sh.ConnectionParams {
		paramsJoined = fmt.Sprintf("%s%s=%s&", paramsJoined, k, v)
	}

	if len(dbName) > 0 {
		sh.DBName = dbName[0]
	}

	if config.Config.SQLite {
		db, err = sql.Open("sqlite3", "./main.sqlite")
		util.CheckErr(err)
	} else {
		connStr2 := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s%s", sh.User, util.EnvVar("SQL_PW"), sh.Host, sh.Port, sh.DBName, paramsJoined)
		logx.Printf("gorp conn: %v", connStr2)
		db, err = sql.Open("mysql", connStr2)
		util.CheckErr(err)
	}

	err = db.Ping()
	util.CheckErr(err)
	logx.Printf("gorp database connection up")

	// construct a gorp DbMap
	if config.Config.SQLite {
		dbmap = &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	} else {
		dbmap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	}

	t1 := dbmap.AddTable(mdl.Site{})
	t1.ColMap("domain_name").SetUnique(true)
	dbmap.AddTable(mdl.Detail{})

	dbmap.TraceOn("gx-", logx.Get())
	err = dbmap.CreateTables()
	if err != nil {
		logx.Printf("tables already exist: %v", err)
	} else {
		err = dbmap.CreateIndex()
		if err != nil {
			logx.Printf("error creating indize: %v", err)
		}
		// CreateRumpData()
	}
	dbmap.TraceOff()

	return dbmap

}

func CreateRumpData() {

	pg1 := mdl.Site{}
	pg1.Id = 1 // ignored anyway
	pg1.Name = "dummy.org"
	pg1.Label = "Some dummy label"
	err := DBMap().Insert(&pg1)
	util.CheckErr(err)

}

// checkRes is checking the error *and* the sql result
// of a sql query.
func CheckRes(sqlRes sql.Result, err error) {
	defer logx.SL().Incr().Decr()
	defer logx.SL().Incr().Decr()
	util.CheckErr(err)
	liId, err := sqlRes.LastInsertId()
	util.CheckErr(err)
	affected, err := sqlRes.RowsAffected()
	util.CheckErr(err)
	if affected > 0 && liId > 0 {
		logx.Printf("%d row(s) affected ; lastInsertId %d ", affected, liId)
	} else if affected > 0 {
		logx.Printf("%d row(s) affected", affected)
	} else if liId > 0 {
		logx.Printf("%d liId", liId)
	}
}

func TableName(i interface{}) string {
	t := reflect.TypeOf(i)
	if table, err := dbmap.TableFor(t, false); table != nil && err == nil {
		return dbmap.Dialect.QuoteField(table.TableName)
	}
	return t.Name()
}
