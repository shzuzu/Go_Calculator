package repo

import (
	"database/sql"
	"fmt"
	"log"
)

type Expression struct {
	ID         string   `json:"id"`
	UserID     int64    `json:"user_id"`
	Expression string   `json:"expression"`
	Status     string   `json:"status"`
	Result     *float64 `json:"result"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(userID int64, expression string) (int64, error) {
	result, err := r.db.Exec(`INSERT INTO expressions (user_id, expression, status) VALUES (?, ?, ?)`,
		userID, expression, "pending")
	if err != nil {
		return 0, fmt.Errorf("ERROR creating expression: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil

}
func (r *Repository) UpdateStatus(id int64, status string, result *float64) error {
	var err error
	if result == nil {
		_, err = r.db.Exec(
			"UPDATE expressions SET status = ? WHERE id = ?",
			status, id,
		)
	} else {
		_, err = r.db.Exec(
			"UPDATE expressions SET status = ?, result = ? WHERE id = ?",
			status, *result, id,
		)
	}

	if err != nil {
		log.Printf("Error updating expression status: %v", err)
		return err
	}

	return nil
}

func (r *Repository) GetByID(id int64) (*Expression, error) {
	expr := &Expression{}
	var resultNull sql.NullFloat64

	err := r.db.QueryRow(
		"SELECT id, user_id, expression, status, result FROM expressions WHERE id = ?",
		id,
	).Scan(&expr.ID, &expr.UserID, &expr.Expression, &expr.Status, &resultNull)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		log.Printf("Error getting expression by ID: %v", err)
		return nil, err
	}

	if resultNull.Valid {
		val := resultNull.Float64
		expr.Result = &val
	}

	return expr, nil
}

func (r *Repository) GetByUserID(userID int64) ([]*Expression, error) {
	rows, err := r.db.Query(
		"SELECT id, user_id, expression, status, result FROM expressions WHERE user_id = ? ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		log.Printf("Error querying expressions by user ID: %v", err)
		return nil, err
	}
	defer rows.Close()

	var expressions []*Expression
	for rows.Next() {
		expr := &Expression{}
		var resultNull sql.NullFloat64

		err := rows.Scan(&expr.ID, &expr.UserID, &expr.Expression, &expr.Status, &resultNull)
		if err != nil {
			log.Printf("Error scanning expression row: %v", err)
			return nil, err
		}

		if resultNull.Valid {
			val := resultNull.Float64
			expr.Result = &val
		}

		expressions = append(expressions, expr)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating expression rows: %v", err)
		return nil, err
	}

	return expressions, nil
}
