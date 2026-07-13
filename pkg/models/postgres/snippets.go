package postgres

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/aitumik/snippetbox/pkg/models"
)

type SnippetModel struct {
	DB *sql.DB
}

func (s *SnippetModel) Insert(title, content, expires string, tagIDs []int) (int, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	normExpires, err := strconv.Atoi(expires)
	if err != nil {
		return 0, err
	}

	createdAt := time.Now()
	expiresAt := createdAt.AddDate(0, 0, normExpires)
	stmt := `INSERT INTO snippets(title, content, created, expires) VALUES($1, $2, NOW(), $3) RETURNING id`

	var id int
	err = tx.QueryRow(stmt, title, content, expiresAt).Scan(&id)
	if err != nil {
		return 0, err
	}

	if len(tagIDs) > 0 {
		for _, tagID := range tagIDs {
			_, err = tx.Exec(`INSERT INTO snippet_tags(snippet_id, tag_id) VALUES($1, $2)`, id, tagID)
			if err != nil {
				return 0, err
			}
		}
	}

	return id, tx.Commit()
}

func (s *SnippetModel) Get(id int) (*models.Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets WHERE expires > NOW() AND id = $1`

	row := s.DB.QueryRow(stmt, id)

	m := &models.Snippet{}

	err := row.Scan(&m.ID, &m.Title, &m.Content, &m.Created, &m.Expires)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}

	m.Tags, err = s.getTagsForSnippet(id)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (s *SnippetModel) Latest() ([]*models.Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets ORDER BY created DESC LIMIT 10`

	rows, err := s.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snippets []*models.Snippet

	for rows.Next() {
		m := &models.Snippet{}
		err = rows.Scan(&m.ID, &m.Title, &m.Content, &m.Created, &m.Expires)
		if err != nil {
			return nil, err
		}

		m.Tags, err = s.getTagsForSnippet(m.ID)
		if err != nil {
			return nil, err
		}

		snippets = append(snippets, m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}

func (s *SnippetModel) GetByUser(userID int) ([]*models.Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets WHERE user_id = $1 ORDER BY created DESC LIMIT 10`

	rows, err := s.DB.Query(stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snippets []*models.Snippet

	for rows.Next() {
		m := &models.Snippet{}
		err = rows.Scan(&m.ID, &m.Title, &m.Content, &m.Created, &m.Expires)
		if err != nil {
			return nil, err
		}

		m.Tags, err = s.getTagsForSnippet(m.ID)
		if err != nil {
			return nil, err
		}

		snippets = append(snippets, m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}

func (s *SnippetModel) getTagsForSnippet(snippetID int) ([]models.Tag, error) {
	stmt := `SELECT t.id, t.name FROM tags t
		INNER JOIN snippet_tags st ON st.tag_id = t.id
		WHERE st.snippet_id = $1`

	rows, err := s.DB.Query(stmt, snippetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

func (s *SnippetModel) GetByTag(tagID int) ([]*models.Snippet, error) {
	stmt := `SELECT s.id, s.title, s.content, s.created, s.expires FROM snippets s
		INNER JOIN snippet_tags st ON st.snippet_id = s.id
		WHERE st.tag_id = $1 AND s.expires > NOW()
		ORDER BY s.created DESC LIMIT 10`

	rows, err := s.DB.Query(stmt, tagID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snippets []*models.Snippet
	for rows.Next() {
		m := &models.Snippet{}
		err = rows.Scan(&m.ID, &m.Title, &m.Content, &m.Created, &m.Expires)
		if err != nil {
			return nil, err
		}
		m.Tags, err = s.getTagsForSnippet(m.ID)
		if err != nil {
			return nil, err
		}
		snippets = append(snippets, m)
	}
	return snippets, rows.Err()
}
