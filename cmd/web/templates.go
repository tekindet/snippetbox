package main

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"time"

	"github.com/aitumik/snippetbox/pkg/forms"
	"github.com/aitumik/snippetbox/pkg/models"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
)

type TemplateData struct {
	Snippet           *models.Snippet
	Snippets          []*models.Snippet
	ExpiringToday     []*models.Snippet
	CurrentYear       int
	Flash             string
	Form              *forms.Form
	AuthenticatedUser *models.User
	CSRFToken         string
	Tag               *models.Tag
	Tags              []*models.Tag
	User              *models.User
}

// NewTemplateCache returns a new TemplateCache
func NewTemplateCache(dir string) (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}
	for _, page := range pages {
		// Extract the file name(like 'home.page.tmpl') from the full path
		// and assign it to the name variable
		name := filepath.Base(page)

		// The template.FuncMap must be registered with the template set before the ParseFiles
		// method
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}

// humanDate function returns a nicely formatted string representation of
// time.Time value
func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

func renderMarkdown(md string) template.HTML {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(md), &buf); err != nil {
		return template.HTML(md)
	}
	policy := bluemonday.UGCPolicy()
	policy.AllowElements("pre", "code", "kbd")
	return template.HTML(policy.Sanitize(buf.String()))
}

func relativeTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	duration := time.Until(t)
	if duration < 0 {
		// Past time
		past := -duration
		switch {
		case past < time.Minute:
			return "just now"
		case past < time.Hour:
			mins := int(past.Minutes())
			if mins == 1 {
				return "1 minute ago"
			}
			return fmt.Sprintf("%d minutes ago", mins)
		case past < 24*time.Hour:
			hours := int(past.Hours())
			if hours == 1 {
				return "1 hour ago"
			}
			return fmt.Sprintf("%d hours ago", hours)
		default:
			days := int(past.Hours() / 24)
			if days == 1 {
				return "1 day ago"
			}
			return fmt.Sprintf("%d days ago", days)
		}
	}
	// Future time
	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		mins := int(duration.Minutes())
		if mins == 1 {
			return "in 1 minute"
		}
		return fmt.Sprintf("in %d minutes", mins)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "in 1 hour"
		}
		return fmt.Sprintf("in %d hours", hours)
	default:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "in 1 day"
		}
		return fmt.Sprintf("in %d days", days)
	}
}

var functions = template.FuncMap{
	"humanDate":      humanDate,
	"renderMarkdown": renderMarkdown,
	"relativeTime":   relativeTime,
}
