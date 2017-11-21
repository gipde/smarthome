package app

import (
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	//	"reflect"
	"github.com/revel/revel"
	"strconv"
)

// type MyDB sql.DB

// var DB *MyDB
var DB *sql.DB

func InitDB() {
	driver := revel.Config.StringDefault("db.driver", "sqlite3")
	connect_string := revel.Config.StringDefault("db.connect", "file::memory:?mode=memory&cache=shared")

	revel.AppLog.Info("Starting DB Connection " + driver)

	var err error
	DB, err = sql.Open(driver, connect_string)
	if err != nil {
		revel.INFO.Println("DB Error", err)
	}

	revel.AppLog.Info(strconv.FormatInt(int64(DB.Stats().OpenConnections), 10))

	fmt.Println(DB)
	//revel.AppLog.Info(reflect.TypeOf(DB))

	revel.INFO.Println("DB Connected")

	initDB()
}

// func (db *MyDB) String() string {
// 	return "Hello"
// }

// func GetDB() *sql.DB {
// 	fmt.Println("from getter")
// 	fmt.Println(DB)
// 	return DB
// }

func initDB() {
	exec("create table user(id varchar(10), name varchar(40), password varchar(20));")

	pw, _ := bcrypt.GenerateFromPassword([]byte("geheim"), bcrypt.DefaultCost)
	exec("insert into user values($1,$2,$3)", "werner", "Werner Schneider", pw)
	exec("insert into tbl1 values('zwei',2);")
}

func exec(q string, args ...interface{}) {
	revel.AppLog.Info("Executing DDL " + q)
	result, err := DB.Exec(q, args)
	if err != nil {
		revel.AppLog.Fatal(err.Error())
	}
	rows, _ := result.RowsAffected()
	lastInsertID, _ := result.LastInsertId()
	revel.AppLog.Info(fmt.Sprintf("affected: %d, last ID: %d ", rows, lastInsertID))
}

func doSQL(q string) *sql.Rows {
	revel.AppLog.Info("Query " + q)
	rows, err := DB.Query(q)
	//defer rows.Close()
	if err != nil {
		revel.AppLog.Error(err.Error())
	}
	fmt.Println(rows)
	return rows
}

func printResult(rows *sql.Rows) {
	var col []*sql.ColumnType
	revel.AppLog.Info("loop over rows")
	fmt.Println(rows)
	fmt.Println(&rows)
	firstResult := true
	for rows.Next() {
		if firstResult {
			firstResult = false
			col, _ = rows.ColumnTypes()
			for _, c := range col {
				revel.AppLog.Info("HEAD: " + c.Name())
				revel.AppLog.Info("HEAD: " + c.ScanType().String())
				revel.AppLog.Info("HEAD: " + c.DatabaseTypeName())
				var value interface{}
				rows.Scan(&value)
				fmt.Println("value: ", value)
			}

			// fmt.Printf(col.Name())
			// Columns()
			// revel.AppLog.Info("HEAD: " + strings.Join(col, ","))
		}

		//revel.AppLog.Info("TableName: " + name)
		// revel.AppLog.Info("Columns: Start")
		// revel.AppLog.Info("Columns: Ende")
	}
	rows.Close()
	revel.AppLog.Info(fmt.Sprintf("Columns: %d", len(col)))

}
