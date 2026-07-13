package postgres

import (
	"database/sql"

	"github.com/aitumik/snippetbox/pkg/models"
)

type TagModel struct {
	DB *sql.DB
}

func (t *TagModel) Insert(name string) (int, error) {
	var id int
	stmt := `INSERT INTO tags(name) VALUES($1) ON CONFLICT(name) DO UPDATE SET name = EXCLUDED.name RETURNING id`
	err := t.DB.QueryRow(stmt, name).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (t *TagModel) GetByName(name string) (*models.Tag, error) {
	stmt := `SELECT id, name FROM tags WHERE name = $1`
	tag := &models.Tag{}
	err := t.DB.QueryRow(stmt, name).Scan(&tag.ID, &tag.Name)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}
	return tag, nil
}

func (t *TagModel) GetForSnippet(snippetID int) ([]*models.Tag, error) {
	stmt := `SELECT t.id, t.name FROM tags t
		INNER JOIN snippet_tags st ON st.tag_id = t.id
		WHERE st.snippet_id = $1`

	rows, err := t.DB.Query(stmt, snippetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []*models.Tag
	for rows.Next() {
		tag := &models.Tag{}
		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (t *TagModel) GetAll() ([]*models.Tag, error) {
	stmt := `SELECT id, name FROM tags ORDER BY name`

	rows, err := t.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []*models.Tag
	for rows.Next() {
		tag := &models.Tag{}
		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}
