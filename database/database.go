package database

import (
	"database/sql"
	"fmt"

	"github.com/dreamspawn/ribbon-server/config"
	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

func Connect() {
	db_config := mysql.Config{
		User:   config.Get("db_user"),
		Passwd: config.Get("db_pass"),
		Net:    "unix",
		Addr:   "/var/run/mysqld/mysqld.sock",
		DBName: config.Get("db_name"),
	}

	handle, err := sql.Open("mysql", db_config.FormatDSN())
	if err != nil {
		fmt.Println("Could not connect to database")
		panic(err)
	}

	db = handle
}

func Query(statement string, args []any) ([]map[string]interface{}, error) {
	rows, err := db.Query(statement, args...)
	if err != nil {
		return nil, db_error(statement, args, err)
	}

	cols, err := rows.ColumnTypes()
	if err != nil {
		return nil, db_error(statement, args, err)
	}

	var result []map[string]interface{}
	fields := make([]interface{}, len(cols))
	field_pointers := make([]interface{}, len(cols))

	for i := range fields {
		field_pointers[i] = &fields[i]
	}

	for rows.Next() {
		err := rows.Scan(field_pointers...)
		if err != nil {
			return nil, db_error(statement, args, err)
		}

		row_map := make(map[string]interface{})
		for i, column := range cols {
			if data, ok := fields[i].(string); ok {
				row_map[column.Name()] = data
			} else if data, ok := fields[i].([]byte); ok {
				row_map[column.Name()] = string(data)
			} else {
				row_map[column.Name()] = fields[i]
			}
		}
		result = append(result, row_map)
	}

	return result, nil
}

func db_error(statement string, args []any, err error) error {
	return fmt.Errorf("failed to exceute query:\n%s\nwith the arguments:\n%+v\nerror:\n%v", statement, args, err)
}
