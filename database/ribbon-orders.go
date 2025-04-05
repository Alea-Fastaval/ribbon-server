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

func GetOrderByID(order_id uint) (map[string]any, error) {
	var settings map[string]any

	query := "SELECT * FROM ribbon_orders WHERE id = ?"

	result, err := Query(query, []any{order_id})
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		log.Output(1, fmt.Sprintf("Error: Unknown order id %d", order_id))
		return nil, fmt.Errorf("no order with id %d", order_id)
	}

	uid, ok := result[0]["user_id"].(int64)
	if !ok {
		log.Output(1, fmt.Sprintf("Error getting user_id from order %d", order_id))
		return map[string]any{
			"order":    result[0],
			"settings": settings,
		}, nil
	}

	query = "SELECT * FROM users WHERE id = ?"
	user_result, err := Query(query, []any{uid})
	if err != nil {
		return nil, err
	}

	if len(user_result) != 0 {
		settings = make(map[string]any)
		settings["status"] = user_result[0]["status"]
		settings["columns"] = user_result[0]["columns"]
	} else {
		log.Output(1, fmt.Sprintf("Error loading order settings for user %d", uid))
	}

	return map[string]any{
		"order":    result[0],
		"settings": settings,
	}, nil
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

func GetAllOrders() (map[string]any, error) {
	query := "SELECT * FROM ribbon_orders r JOIN users u ON r.user_id = u.id"
	orders, err := Query(query, []any{})
	if err != nil {
		return nil, err
	}

	list := make(map[int64]map[int64]any)
	for _, row := range orders {
		year := row["year"].(int64)
		if _, ok := list[year]; !ok {
			list[year] = make(map[int64]any)
		}

		user_id := row["user_id"].(int64)
		var user_rows []map[string]any
		if user_orders, ok := list[year][user_id].([]map[string]any); ok {
			user_rows = user_orders
		}
		user_rows = append(user_rows, row)
		list[year][user_id] = user_rows
	}

	query = "SELECT year, COUNT(*) as users FROM users GROUP BY year"
	users, err := Query(query, []any{})
	if err != nil {
		return nil, err
	}

	user_counts := make(map[int64]any)
	for _, row := range users {
		user_counts[row["year"].(int64)] = row["users"]
	}

	query = "SELECT year, COUNT(DISTINCT(user_id)) as users FROM ribbon_orders r JOIN users u ON r.user_id = u.id GROUP BY year"
	users, err = Query(query, []any{})
	if err != nil {
		return nil, err
	}

	order_counts := make(map[int64]any)
	for _, row := range users {
		order_counts[row["year"].(int64)] = row["users"]
	}

	query = "SELECT year, COUNT(DISTINCT(user_id)) as users FROM ribbon_orders r JOIN users u ON r.user_id = u.id WHERE u.status = 'closed' GROUP BY year"
	users, err = Query(query, []any{})
	if err != nil {
		return nil, err
	}

	closed_counts := make(map[int64]any)
	for _, row := range users {
		closed_counts[row["year"].(int64)] = row["users"]
	}

	return map[string]any{
		"list": list,
		"counts": map[string]any{
			"users":  user_counts,
			"orders": order_counts,
			"closed": closed_counts,
		},
	}, nil
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
