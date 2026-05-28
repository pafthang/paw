package search

import (
	"errors"
	"strings"

	"github.com/pafthang/paw/internal/db"
	"github.com/pafthang/paw/internal/filestore"
	"github.com/pafthang/paw/internal/memory"
	"gorm.io/gorm"
)

type Result struct {
	Type    string `json:"type"`
	ID      uint   `json:"id"`
	Title   string `json:"title,omitempty"`
	Snippet string `json:"snippet,omitempty"`
}

type Response struct {
	Query   string   `json:"query"`
	Results []Result `json:"results"`
}

func Run(database *gorm.DB, query string, limit int) (Response, error) {
	if database == nil {
		return Response{}, errors.New("database is required")
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return Response{}, errors.New("query is required")
	}
	if limit <= 0 {
		limit = 50
	}

	results := make([]Result, 0, limit)

	mem, _ := memory.Search(database, query, limit)
	for _, item := range mem {
		results = append(results, Result{Type: "memory", ID: item.ID, Title: item.Type, Snippet: snippet(item.Content)})
		if len(results) >= limit {
			return Response{Query: query, Results: results[:limit]}, nil
		}
	}

	sessions, _ := db.SearchChatSessions(database, query, limit)
	for _, s := range sessions {
		results = append(results, Result{Type: "session", ID: s.ID, Title: s.Title, Snippet: snippet(s.Title)})
		if len(results) >= limit {
			return Response{Query: query, Results: results[:limit]}, nil
		}
	}

	messages, _ := searchMessages(database, query, limit)
	for _, m := range messages {
		results = append(results, Result{Type: "message", ID: m.ID, Title: m.Role, Snippet: snippet(m.Content)})
		if len(results) >= limit {
			return Response{Query: query, Results: results[:limit]}, nil
		}
	}

	files, _ := filestore.Search(database, query, limit)
	for _, f := range files {
		results = append(results, Result{Type: "file", ID: f.ID, Title: f.Name, Snippet: snippet(f.Path)})
		if len(results) >= limit {
			return Response{Query: query, Results: results[:limit]}, nil
		}
	}

	return Response{Query: query, Results: results}, nil
}

func searchMessages(database *gorm.DB, query string, limit int) ([]db.ChatMessage, error) {
	like := "%" + strings.TrimSpace(query) + "%"
	var out []db.ChatMessage
	if err := database.Where("content LIKE ?", like).Order("created_at desc, id desc").Limit(limit).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func snippet(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= 120 {
		return s
	}
	return s[:120]
}
