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
	Retired  bool
	Hidden   bool
	Special  map[string]string
}

// ----------------------------------------------------------------------------------------------------------------------
// Create
// ----------------------------------------------------------------------------------------------------------------------
func CreateRibbon(category uint, glyph uint, no_wings bool) (*Ribbon, error) {
	// Get number of existing ribbons, for insert position
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

	// Create ribbon
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

	// Get ID for result
	id, err := result.LastInsertId()
	new_ribbon := Ribbon{
		uint(id),
		category,
		glyph,
		no_wings,
		ribbon_count,
		false,
		false,
		nil,
	}

	return &new_ribbon, err
}

// ----------------------------------------------------------------------------------------------------------------------
// Read
// ----------------------------------------------------------------------------------------------------------------------
func GetRibbons(include_hidden bool) ([]Ribbon, error) {
	statement := ""
	if include_hidden {
		statement = "SELECT * FROM ribbons ORDER BY ordering"
	} else {
		statement = "SELECT * FROM ribbons WHERE hidden = false ORDER BY ordering"
	}

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
			&ribbon.Retired,
			&ribbon.Hidden,
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
		&ribbon.Retired,
		&ribbon.Hidden,
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
// UPDATE
// ----------------------------------------------------------------------------------------------------------------------
func UpdateRibbon(ribbon_id, category, glyph uint, no_wings bool) (*Ribbon, error) {
	var old_category uint
	statement := "SELECT category_id FROM ribbons WHERE id = ?"
	row := db.QueryRow(statement, ribbon_id)
	err := row.Scan(&old_category)
	if err != nil {
		return nil, db_error(statement, nil, err)
	}

	statement = "UPDATE ribbons SET category_id=?, glyph_id=?, no_wings=? WHERE id = ?"
	_, err = db.Exec(statement, category, glyph, no_wings, ribbon_id)
	if err != nil {
		args := []any{
			category,
			glyph,
			no_wings,
			ribbon_id,
		}
		return nil, db_error(statement, args, err)
	}

	new_ribbon := Ribbon{
		ribbon_id,
		category,
		glyph,
		no_wings,
		0,
		false,
		false,
		nil,
	}

	if old_category == category {
		return &new_ribbon, err
	}

	// Set position to last in new category
	statement = "UPDATE ribbons as r, (SELECT COUNT(*) as total FROM ribbons WHERE category_id = ?) as new SET r.ordering = new.total WHERE id = ?"
	_, err = db.Exec(statement, category, ribbon_id)
	if err != nil {
		return nil, db_error(statement, []any{category, ribbon_id}, err)
	}

	// Move all ribbons with later position from old category up one place
	statement = "UPDATE ribbons as r, (SELECT ordering FROM ribbons WHERE id = ?) as old SET r.ordering = r.ordering - 1 WHERE r.ordering > old.ordering AND category_id = ?"
	_, err = db.Exec(statement, ribbon_id, category)
	if err != nil {
		return nil, db_error(statement, []any{category, ribbon_id}, err)
	}

	return &new_ribbon, err
}

// ----------------------------------------------------------------------------------------------------------------------
// retire/unretire ribbon
// ----------------------------------------------------------------------------------------------------------------------
func RetireRibbon(id uint, retired bool) error {
	query := "UPDATE ribbons SET retired = ? WHERE id = ?"
	_, err := Exec(query, []any{retired, id})
	return err
}

// ----------------------------------------------------------------------------------------------------------------------
// Undelete ribbon from view
// ----------------------------------------------------------------------------------------------------------------------
func ShowRibbon(id uint) error {
	query := "UPDATE ribbons SET hidden = false WHERE id = ?"
	_, err := Exec(query, []any{id})
	return err
}

// ----------------------------------------------------------------------------------------------------------------------
// Delete (from view)
// ----------------------------------------------------------------------------------------------------------------------
func HideRibbon(id uint) error {
	query := "UPDATE ribbons SET hidden = true WHERE id = ?"
	_, err := Exec(query, []any{id})
	return err
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
