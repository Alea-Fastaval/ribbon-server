package database

import (
	"fmt"
	"log"
	"strings"
)

func SetOrder(uid, ribbon uint, values map[string]uint) error {
	query := "SELECT * FROM ribbon_orders WHERE user_id = ? AND ribbon_id = ?"
	order, _ := Query(query, []any{uid, ribbon})

	if len(order) != 0 {
		return update(uid, ribbon, values, order[0]["position"].(int64))
	}

	return insert(uid, ribbon, values)
}

func insert(uid, ribbon uint, values map[string]uint) error {
	// Set default fro missing values
	args := []any{uid, ribbon}
	keys := []string{
		"grunt",
		"second",
		"leader",
	}
	for _, key := range keys {
		if value, ok := values[key]; ok {
			args = append(args, value)
		} else {
			args = append(args, 0)
		}
	}

	// Find position for new ribbon
	query := "SELECT COUNT(*) as count FROM ribbon_orders WHERE user_id = ?"
	result_count, err := Query(query, []any{uid})
	if err != nil {
		return err
	}
	poistion := result_count[0]["count"].(int64)
	args = append(args, poistion)

	// Insert ribbon order
	query = "INSERT INTO ribbon_orders (user_id, ribbon_id, grunt, second, leader, position) VALUES(?,?,?,?,?,?)"
	_, err = Exec(query, args)

	return err
}

func update(uid, ribbon uint, values map[string]uint, old_position int64) error {
	var args []any
	var value_queries []string

	// Prepare query and args for each value
	for key, value := range values {
		value_queries = append(value_queries, key+" = ?")
		args = append(args, value)

		// Update poitions of moving orders
		if key == "position" && value != uint(old_position) {
			// Check what direction we're moving things
			sign := "+"
			move_args := []any{uid}
			if values["position"] > uint(old_position) {
				sign = "-"
				move_args = append(move_args, values["position"], old_position)
			} else {
				move_args = append(move_args, old_position, values["position"])
			}

			move_query := "UPDATE ribbon_orders SET position = position " + sign + " 1 WHERE user_id = ? AND position <= ? AND position >= ?"

			// Don't set the temporary position of the order we're moving to -1 since we're using uint
			if old_position == 0 && sign == "-" {
				move_query += " AND position <> 0"
			}

			// Do the moving
			_, err := Exec(move_query, move_args)
			if err != nil {
				return err
			}
		}
	}

	query := "UPDATE ribbon_orders SET "
	query += strings.Join(value_queries, ", ")
	query += " WHERE user_id = ? AND ribbon_id = ?"
	args = append(args, uid, ribbon)

	_, err := Exec(query, args)
	return err
}

func GetOrders(uid uint) (map[string]any, error) {
	query := "SELECT * FROM ribbon_orders WHERE user_id = ?"
	result, err := Query(query, []any{uid})
	if err != nil {
		return nil, err
	}

	orders := make(map[uint]map[string]any)
	for _, row := range result {
		orders[uint(row["ribbon_id"].(int64))] = row
	}

	query = "SELECT * FROM users WHERE id = ?"
	result, err = Query(query, []any{uid})
	if err != nil {
		return nil, err
	}

	settings := make(map[string]any)
	if len(result) != 0 {
		settings["status"] = result[0]["status"]
		settings["columns"] = result[0]["columns"]
	} else {
		log.Output(1, fmt.Sprintf("Error loading order settings for user %d", uid))
	}

	return map[string]any{
		"list":     orders,
		"settings": settings,
	}, nil
}

func GetAllOrders() (map[uint]map[string]uint, error) {
	return make(map[uint]map[string]uint), nil
}

func DeleteOrder(uid, ribbon uint) error {
	query := "SELECT * FROM ribbon_orders WHERE user_id = ? AND ribbon_id = ?"
	order, _ := Query(query, []any{uid, ribbon})

	if len(order) == 0 {
		return fmt.Errorf("no ribbon order found for user %d with ribbon %d", uid, ribbon)
	}

	// Delete order
	query = "DELETE FROM ribbon_orders WHERE user_id = ? AND ribbon_id = ?"
	_, err := Exec(query, []any{uid, ribbon})
	if err != nil {
		return err
	}

	// Update positions of other orders
	query = "UPDATE ribbon_orders SET position = position -1 WHERE user_id = ? AND position > ?"
	_, err = Exec(query, []any{uid, order[0]["position"]})
	if err != nil {
		return err
	}

	return nil
}

func SetColumns(uid, columns uint) error {
	query := "UPDATE users SET columns = ? WHERE id = ?"
	_, err := Exec(query, []any{columns, uid})
	return err
}

func SetStatus(uid uint, status string) error {
	query := "UPDATE users SET status = ? WHERE id = ?"
	_, err := Exec(query, []any{status, uid})
	return err
}
