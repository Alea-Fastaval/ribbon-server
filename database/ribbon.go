package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type Ribbon struct {
	ID       uint
	Category uint
	Glyph    uint
	NoWings  bool
	Ordering uint
	Special  map[string]string
}

// ----------------------------------------------------------------------------------------------------------------------
// Create
// ----------------------------------------------------------------------------------------------------------------------
func CreateRibbon(category uint, glyph uint, no_wings bool) (*Ribbon, error) {
	var ribbon_count uint
	statement := "SELECT COUNT(*) FROM ribbons WHERE category_id = ?"
	row := db.QueryRow(statement, category)
	err := row.Scan(&ribbon_count)
	if err != nil {
		args := []any{
			category,
		}
		return nil, db_error(statement, args, err)
	}

	statement = "INSERT INTO ribbons(category_id, glyph_id, no_wings, ordering) VALUES(?,?,?,?)"
	result, err := db.Exec(statement, category, glyph, no_wings, ribbon_count)
	if err != nil {
		args := []any{
			category,
			glyph,
			no_wings,
			ribbon_count,
		}
		return nil, db_error(statement, args, err)
	}

	id, err := result.LastInsertId()
	new_ribbon := Ribbon{
		uint(id),
		category,
		glyph,
		no_wings,
		ribbon_count,
		nil,
	}

	return &new_ribbon, err
}

// ----------------------------------------------------------------------------------------------------------------------
// Read
// ----------------------------------------------------------------------------------------------------------------------
func GetRibbons() ([]Ribbon, error) {
	statement := "SELECT * FROM ribbons ORDER BY ordering"
	rows, err := db.Query(statement)
	if err != nil {
		return nil, err
	}

	var result []Ribbon

	for rows.Next() {
		ribbon := Ribbon{}

		err := rows.Scan(
			&ribbon.ID,
			&ribbon.Category,
			&ribbon.Glyph,
			&ribbon.NoWings,
			&ribbon.Ordering,
		)

		ribbon.Special = getSpecial(ribbon.ID)

		if err != nil {
			return nil, db_error(statement, nil, err)
		}

		result = append(result, ribbon)
	}

	return result, nil
}

func GetRibbon(id uint) (*Ribbon, error) {
	statement := "SELECT * FROM ribbons WHERE id = ?"
	row := db.QueryRow(statement, id)

	ribbon := Ribbon{}
	err := row.Scan(
		&ribbon.ID,
		&ribbon.Category,
		&ribbon.Glyph,
		&ribbon.NoWings,
		&ribbon.Ordering,
	)

	ribbon.Special = getSpecial(id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, db_error(statement, nil, err)
	}

	return &ribbon, nil
}

func getSpecial(ribbon_id uint) map[string]string {
	query := "SELECT * FROM special_rules WHERE ribbon_id = ?"
	result, err := Query(query, []any{ribbon_id})
	if err != nil {
		log.Output(1, fmt.Sprintf("Error loading special rules from DB for ribbon %d", ribbon_id))
		return nil
	}

	if len(result) == 0 {
		return nil
	}

	special := make(map[string]string)
	for _, row := range result {
		var key, value string
		var ok bool
		if key, ok = row["name"].(string); !ok {
			continue
		}
		if value, ok = row["value"].(string); ok {
			special[key] = value
		}
	}

	if len(special) == 0 {
		return nil
	}

	return special
}

// ----------------------------------------------------------------------------------------------------------------------
// Delete
// ----------------------------------------------------------------------------------------------------------------------
func DeleteRibbon(id uint) error {
	query := "DELETE FROM ribbons WHERE id = ?"
	_, err := Exec(query, []any{id})
	return err
}

func (ribbon Ribbon) GetCategory() (*Category, error) {
	return GetCategory(ribbon.Category)
}
