package postgres

import (
	"database/sql"
	"github.com/aitumik/snippetbox/pkg/models"
	"strconv"
	"time"
)

type SnippetModel struct {
	DB *sql.DB
}

// Insert This will insert a new snippet into the database
func (s *SnippetModel) Insert(title, content, expires string) (int, error) {
	var id int
	normExpires, err := strconv.Atoi(expires)
	if err != nil {
		return 0, err
	}

	createdAt := time.Now()
	expiresAt := createdAt.AddDate(0, 0, normExpires)
	stmt := `INSERT INTO snippets( title, content,created,expires) VALUES($1,$2,NOW(),$3) RETURNING id`

	err = s.DB.QueryRow(stmt, title, content, expiresAt).Scan(&id)
	if err != nil {
		return 0, err
	}
	// returns int64
	return int(id), nil
}

// Get this will return specific snippet based on id
func (s *SnippetModel) Get(id int) (*models.Snippet, error) {
	stmt := `SELECT id,title,content,created,expires FROM snippets WHERE expires > NOW() AND id = $1`

	// use the query row to execute the SQL statement
	row := s.DB.QueryRow(stmt, id)

	// initialize a pointer to a new zerod struct
	m := &models.Snippet{}

	err := row.Scan(&m.ID, &m.Title, &m.Content, &m.Created, &m.Expires)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}

	// If everything went okay then return the snippet object
	return m, nil
}

// Latest This will return the top 10 most recently created snippets
func (s *SnippetModel) Latest() ([]*models.Snippet, error) {
	stmt := `SELECT id,title,content,created,expires FROM snippets ORDER BY created DESC LIMIT 10`

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
		snippets = append(snippets, m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}
