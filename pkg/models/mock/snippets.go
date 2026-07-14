package mock

import (
	"github.com/aitumik/snippetbox/pkg/models"
	"time"
)

var mockSnippet = &models.Snippet{
	ID:      1,
	Title:   "Dr Strange in the multiverse of madness!",
	Content: "I never meant for any of this to happen",
	Created: time.Now(),
	Expires: time.Now(),
	Author:  "Test User",
}

type SnippetModel struct{}

func (m *SnippetModel) Insert(title, content, expires string, tagIDs []int, userID int) (int, error) {
	return 2, nil
}

func (m *SnippetModel) Get(id int) (*models.Snippet, error) {
	switch id {
	case 1:
		return mockSnippet, nil
	default:
		return nil, models.ErrNoRecord
	}
}

func (m *SnippetModel) Latest() ([]*models.Snippet, error) {
	return []*models.Snippet{mockSnippet}, nil
}

func (m *SnippetModel) GetByUser(userID int) ([]*models.Snippet, error) {
	return []*models.Snippet{mockSnippet}, nil
}

func (m *SnippetModel) GetByTag(tagID int) ([]*models.Snippet, error) {
	return []*models.Snippet{mockSnippet}, nil
}

func (m *SnippetModel) GetExpiringToday() ([]*models.Snippet, error) {
	return []*models.Snippet{mockSnippet}, nil
}

func (m *SnippetModel) ExtendExpiry(id int) error {
	return nil
}
