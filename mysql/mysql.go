package mysql

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"os"
	"sort"
	"strings"

	"github.com/Masterminds/squirrel"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var (
	gormConnection *gorm.DB
	debug          = false

	ErrNotFound = errors.New("not found")
)

func SetDebug(val bool) {
	debug = val
	return
}

func GetDB() (conn *gorm.DB, err error) {

	if gormConnection == nil {

		db, err := gorm.Open("mysql", os.Getenv("STEAM_MYSQL_DSN"))
		db.LogMode(debug)
		if err != nil {
			return db, nil
		}

		gormConnection = db
	}

	return gormConnection, nil
}

//
type UpdateError struct {
	err  string
	hard bool
	log  bool
}

func (e UpdateError) Error() string {
	return e.err
}

func (e UpdateError) IsHard() bool {
	return e.hard
}

func (e UpdateError) IsSoft() bool {
	return !e.hard
}

func (e UpdateError) Log() bool {
	return e.log
}

//
func Select(builder squirrel.SelectBuilder) (rows *sql.Rows, err error) {

	rawSQL, args, err := builder.ToSql()
	if err != nil {
		return rows, err
	}

	prep, err := getPrepareStatement(rawSQL)
	if err != nil {
		return rows, err
	}

	rows, err = prep.Query(args...)
	if err != nil {
		return rows, err
	}

	return rows, nil
}

func SelectFirst(builder squirrel.SelectBuilder) (row *sql.Row, err error) {

	builder.Limit(1)

	rawSQL, args, err := builder.ToSql()
	if err != nil {
		return row, err
	}

	prep, err := getPrepareStatement(rawSQL)
	if err != nil {
		return row, err
	}

	return prep.QueryRow(args...), nil
}

func Insert(builder squirrel.InsertBuilder) (result sql.Result, err error) {

	rawSQL, args, err := builder.ToSql()
	if err != nil {
		return result, err
	}

	prep, err := getPrepareStatement(rawSQL)
	if err != nil {
		return result, err
	}

	result, err = prep.Exec(args...)
	if err != nil {
		return result, err
	}

	return result, nil
}

func Update(builder squirrel.UpdateBuilder) (result sql.Result, err error) {

	rawSQL, args, err := builder.ToSql()
	if err != nil {
		return result, err
	}

	prep, err := getPrepareStatement(rawSQL)
	if err != nil {
		return result, err
	}

	result, err = prep.Exec(args...)
	if err != nil {
		return result, err
	}

	return result, nil
}

func RawQuery(query string, args []interface{}) (result sql.Result, err error) {

	prep, err := getPrepareStatement(query)
	if err != nil {
		return result, err
	}

	result, err = prep.Exec(args...)
	if err != nil {
		return result, err
	}

	return result, nil
}

func UpdateInsert(table string, data UpdateInsertData) (result sql.Result, err error) {

	query := "INSERT INTO " + table + " (" + data.formattedColumns() + ") VALUES (" + data.getMarks() + ") ON DUPLICATE KEY UPDATE " + data.getDupes() + ";"
	return RawQuery(query, data.getValues())
}

var mysqlPrepareStatements map[string]*sql.Stmt

func getPrepareStatement(query string) (statement *sql.Stmt, err error) {

	if mysqlPrepareStatements == nil {
		mysqlPrepareStatements = make(map[string]*sql.Stmt)
	}

	byteArray := md5.Sum([]byte(query))
	hash := hex.EncodeToString(byteArray[:])

	if val, ok := mysqlPrepareStatements[hash]; ok {
		if ok {
			return val, nil
		}
	}

	conn, err := getMysqlConnection()
	if err != nil {
		return statement, err
	}

	statement, err = conn.Prepare(query)
	if err != nil {
		return statement, err
	}

	mysqlPrepareStatements[hash] = statement
	return statement, nil
}

var mysqlConnection *sql.DB

func getMysqlConnection() (db *sql.DB, err error) {

	if mysqlConnection == nil {

		var err error
		mysqlConnection, err = sql.Open("mysql", os.Getenv("STEAM_MYSQL_DSN"))
		if err != nil {
			return db, err
		}
	}

	return mysqlConnection, nil
}

//
type UpdateInsertData map[string]interface{}

func (ui UpdateInsertData) sortedColumns() (columns []string) {

	var slice []string
	for k := range ui {
		slice = append(slice, k)
	}
	sort.Strings(slice)
	return slice
}

func (ui UpdateInsertData) formattedColumns() (columns string) {

	var slice []string
	for _, v := range ui.sortedColumns() {
		slice = append(slice, "`"+v+"`")
	}
	return strings.Join(ui.sortedColumns(), ", ")
}

func (ui UpdateInsertData) getDupes() (columns string) {

	var slice []string
	for _, v := range ui.sortedColumns() {
		slice = append(slice, v+"=VALUES("+v+")")
	}
	return strings.Join(slice, ", ")
}

func (ui UpdateInsertData) getValues() (columns []interface{}) {

	var slice []interface{}
	for _, v := range ui.sortedColumns() {
		slice = append(slice, ui[v])
	}
	return slice
}

func (ui UpdateInsertData) getMarks() (marks string) {

	var slice []string
	for range ui {
		slice = append(slice, "?")
	}
	return strings.Join(slice, ", ")
}
