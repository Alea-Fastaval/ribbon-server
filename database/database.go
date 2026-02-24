package database

import (
	"bufio"
	"database/sql"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/dreamspawn/ribbon-server/config"
	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

func Connect() {
	net := "tcp"
	address := config.Get("db_address")
	if strings.HasSuffix(address, ".sock") {
		net = "unix"
	}

	db_config := mysql.Config{
		User:   config.Get("db_user"),
		Passwd: config.Get("db_pass"),
		Net:    net,
		Addr:   address,
		DBName: config.Get("db_name"),
	}

	handle, err := sql.Open("mysql", db_config.FormatDSN())
	if err != nil {
		fmt.Println("Could not connect to database")
		panic(err)
	}

	db = handle
}

func Update() {
	fmt.Println("Checking for database updates")

	// Get DB version
	var db_version float64
	var db_version_string string
	statement := "SELECT value FROM options WHERE name='db_version'"
	row := db.QueryRow(statement)
	err := row.Scan(&db_version_string)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			db_version = 0
			query := "INSERT INTO options (name, value) VALUES('db_version', 0)"
			_, err = Exec(query, []any{})
			if err != nil {
				fmt.Print("Could not insert version into database\n")
				panic(err)
			}
		default:
			fmt.Println("Could not get version from DB")
			panic(db_error(statement, nil, err))
		}
	} else {
		db_version, err = strconv.ParseFloat(db_version_string, 64)
		if err != nil {
			fmt.Println("Could not parse version number from DB")
			panic(err)
		}
	}

	fmt.Printf("Current DB version: %.2f\n", db_version)

	// Get update folders
	var update_folders = make(map[float64]os.DirEntry, 0)
	updates_dir := config.Get("resource_dir") + "database/updates/"
	files, err := os.ReadDir(updates_dir)
	if err != nil {
		fmt.Printf("Could not read content of folder %s\n", updates_dir)
		panic(err)
	}

	// Check for new updates
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		folder_version, err := strconv.ParseFloat(file.Name(), 64)

		if err != nil || folder_version <= db_version {
			continue
		}

		update_folders[folder_version] = file
	}

	if len(update_folders) == 0 {
		fmt.Print("No new database updates found\n")
		return
	}

	// Run updates in order of folder version
	for _, key := range slices.Sorted(maps.Keys(update_folders)) {
		folder := updates_dir + update_folders[key].Name()
		script_files, err := os.ReadDir(folder)
		if err != nil {
			fmt.Printf("Could not read content of folder %s\n", folder)
			panic(err)
		}

		for _, script_file := range script_files {
			if script_file.IsDir() || filepath.Ext(script_file.Name()) != ".sql" {
				fmt.Printf("skipping file with extension: %s\n", filepath.Ext(script_file.Name()))
				continue
			}

			script, err := LoadScript(folder + "/" + script_file.Name())
			if err != nil {
				fmt.Printf("could not read script file %s\n", script_file.Name())
				panic(err)
			}
			fmt.Printf("Running update script %s\n", script_file.Name())
			fmt.Printf("Script content: %s\n", script)
			err = RunScript(script)
			if err != nil {
				fmt.Printf("Error running update script %s\n", script_file.Name())
				panic(err)
			}
		}

		query := "UPDATE options SET value=? WHERE name='db_version'"
		_, err = Exec(query, []any{key})
		if err != nil {
			fmt.Printf("Could not update version in database %s\n", key)
			panic(err)
		}
		fmt.Printf("Database updated to version %.2f\n", key)
	}
}

func LoadScript(path string) (string, error) {
	// open translation file
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	var script string
	reader := bufio.NewReader(file)
	for {
		bytes, _, err := reader.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return "", err
		}

		line := string(bytes)

		// Allow comments
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}

		script += line + "\n"
	}

	return script, nil
}

func RunScript(script string) error {
	queries := strings.Split(script, ";")
	for _, query := range queries {
		if match, _ := regexp.MatchString("\n?\\s*", query); match {
			continue
		}
		_, err := Exec(query, []any{})
		if err != nil {
			return err
		}
	}

	return nil
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

func Exec(query string, args []any) (sql.Result, error) {
	result, err := db.Exec(query, args...)
	if err != nil {
		return result, db_error(query, args, err)
	}

	return result, nil
}

func db_error(statement string, args []any, err error) error {
	return fmt.Errorf("failed to exceute query:\n%s\nwith the arguments:\n%+v\nerror:\n%v", statement, args, err)
}
