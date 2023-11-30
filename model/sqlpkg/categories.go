package sqlpkg

import (
	"database/sql"
	"errors"

	"forum/model"
)

/*
inserts a new comment into DB, returns an ID for the comment
*/
func (f *ForumModel) GetCategories() ([]*model.Category, error) {
	q := `SELECT id, name FROM categories ORDER BY name`
	rows, err := f.DB.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// parsing the query's result
	var categories []*model.Category
	for rows.Next() {
		category := &model.Category{}
		err = rows.Scan(&category.ID, &category.Name)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return categories, nil
}

func (f *ForumModel) GetCategoryByID(id int) (*model.Category, error) {
	q := `SELECT id, name FROM categories WHERE id=?`
	category := model.Category{}
	row := f.DB.QueryRow(q, id)
	err := row.Scan(&category.ID, &category.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNoRecord
		}
		return nil, err
	}
	return &category, nil
}
