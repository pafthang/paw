package missioncontrol

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pafthang/paw/internal/config"
)

type Agent struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Role          string         `json:"role"`
	Description   string         `json:"description,omitempty"`
	SessionKey    string         `json:"session_key,omitempty"`
	Backend       string         `json:"backend"`
	Status        string         `json:"status"`
	Level         string         `json:"level"`
	CurrentTaskID *string        `json:"current_task_id,omitempty"`
	Specialties   []string       `json:"specialties"`
	LastHeartbeat *string        `json:"last_heartbeat,omitempty"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
	Metadata      map[string]any `json:"metadata"`
}

type Task struct {
	ID                string         `json:"id"`
	Title             string         `json:"title"`
	Description       string         `json:"description"`
	Status            string         `json:"status"`
	Priority          string         `json:"priority"`
	AssigneeIDs       []string       `json:"assignee_ids"`
	CreatorID         *string        `json:"creator_id,omitempty"`
	ParentTaskID      *string        `json:"parent_task_id,omitempty"`
	BlockedBy         []string       `json:"blocked_by"`
	Tags              []string       `json:"tags"`
	DueDate           *string        `json:"due_date,omitempty"`
	StartedAt         *string        `json:"started_at,omitempty"`
	CompletedAt       *string        `json:"completed_at,omitempty"`
	CreatedAt         string         `json:"created_at"`
	UpdatedAt         string         `json:"updated_at"`
	Metadata          map[string]any `json:"metadata"`
	ProjectID         *string        `json:"project_id,omitempty"`
	TaskType          string         `json:"task_type"`
	Blocks            []string       `json:"blocks"`
	ActiveDescription string         `json:"active_description"`
	EstimatedMinutes  *int           `json:"estimated_minutes,omitempty"`
	Output            *string        `json:"output,omitempty"`
	RetryCount        int            `json:"retry_count"`
	MaxRetries        int            `json:"max_retries"`
	TimeoutMinutes    *int           `json:"timeout_minutes,omitempty"`
	ErrorMessage      *string        `json:"error_message,omitempty"`
}

type Message struct {
	ID          string         `json:"id"`
	TaskID      string         `json:"task_id"`
	FromAgentID string         `json:"from_agent_id"`
	Content     string         `json:"content"`
	CreatedAt   string         `json:"created_at"`
	Metadata    map[string]any `json:"metadata"`
}

type Document struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	Content   string         `json:"content"`
	Type      string         `json:"type"`
	TaskID    *string        `json:"task_id,omitempty"`
	Tags      []string       `json:"tags"`
	Version   int            `json:"version"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
	Metadata  map[string]any `json:"metadata"`
}

type Notification struct {
	ID          string         `json:"id"`
	AgentID     string         `json:"agent_id,omitempty"`
	Title       string         `json:"title"`
	Message     string         `json:"message"`
	Type        string         `json:"type"`
	Read        bool           `json:"read"`
	Delivered   bool           `json:"delivered"`
	DeliveredAt *string        `json:"delivered_at,omitempty"`
	CreatedAt   string         `json:"created_at"`
	Metadata    map[string]any `json:"metadata"`
}

type Store struct {
	baseDir string
}

func NewStore() (*Store, error) {
	dir, err := config.Dir()
	if err != nil {
		return nil, err
	}
	base := filepath.Join(dir, "mission_control")
	if err := os.MkdirAll(base, 0o700); err != nil {
		return nil, err
	}
	return &Store{baseDir: base}, nil
}

func (s *Store) ListAgents(status string) ([]Agent, error) {
	agents, err := readJSON[Agent](s.path("agents.json"))
	if err != nil {
		return nil, err
	}
	if status != "" {
		filtered := agents[:0]
		for _, agent := range agents {
			if agent.Status == status {
				filtered = append(filtered, agent)
			}
		}
		agents = filtered
	}
	sort.Slice(agents, func(i, j int) bool { return strings.ToLower(agents[i].Name) < strings.ToLower(agents[j].Name) })
	return agents, nil
}

func (s *Store) SaveAgent(agent Agent) (Agent, error) {
	now := nowISO()
	if agent.ID == "" {
		agent.ID = uuid.NewString()
		agent.CreatedAt = now
	}
	if agent.CreatedAt == "" {
		agent.CreatedAt = now
	}
	if agent.Backend == "" {
		agent.Backend = "go"
	}
	if agent.Status == "" {
		agent.Status = "idle"
	}
	if agent.Level == "" {
		agent.Level = "specialist"
	}
	if agent.Specialties == nil {
		agent.Specialties = []string{}
	}
	if agent.Metadata == nil {
		agent.Metadata = map[string]any{}
	}
	agent.UpdatedAt = now
	agents, err := readJSON[Agent](s.path("agents.json"))
	if err != nil {
		return agent, err
	}
	updated := false
	for i := range agents {
		if agents[i].ID == agent.ID {
			agents[i] = agent
			updated = true
			break
		}
	}
	if !updated {
		agents = append(agents, agent)
	}
	return agent, writeJSON(s.path("agents.json"), agents)
}

func (s *Store) DeleteAgent(id string) error {
	return deleteByID[Agent](s.path("agents.json"), id, func(a Agent) string { return a.ID })
}

func (s *Store) ListTasks(status string) ([]Task, error) {
	tasks, err := readJSON[Task](s.path("tasks.json"))
	if err != nil {
		return nil, err
	}
	if status != "" {
		filtered := tasks[:0]
		for _, task := range tasks {
			if task.Status == status {
				filtered = append(filtered, task)
			}
		}
		tasks = filtered
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].UpdatedAt > tasks[j].UpdatedAt })
	return tasks, nil
}

func (s *Store) RunningTasks() ([]Task, error) {
	tasks, err := s.ListTasks("")
	if err != nil {
		return nil, err
	}
	out := make([]Task, 0)
	for _, task := range tasks {
		if task.Status == "assigned" || task.Status == "in_progress" || task.Status == "review" {
			out = append(out, task)
		}
	}
	return out, nil
}

func (s *Store) SaveTask(task Task) (Task, error) {
	now := nowISO()
	if task.ID == "" {
		task.ID = uuid.NewString()
		task.CreatedAt = now
	}
	if task.CreatedAt == "" {
		task.CreatedAt = now
	}
	if task.Status == "" {
		task.Status = "inbox"
	}
	if task.Priority == "" {
		task.Priority = "medium"
	}
	if task.AssigneeIDs == nil {
		task.AssigneeIDs = []string{}
	}
	if task.BlockedBy == nil {
		task.BlockedBy = []string{}
	}
	if task.Tags == nil {
		task.Tags = []string{}
	}
	if task.Blocks == nil {
		task.Blocks = []string{}
	}
	if task.Metadata == nil {
		task.Metadata = map[string]any{}
	}
	if task.TaskType == "" {
		task.TaskType = "agent"
	}
	if task.MaxRetries == 0 {
		task.MaxRetries = 1
	}
	task.UpdatedAt = now
	tasks, err := readJSON[Task](s.path("tasks.json"))
	if err != nil {
		return task, err
	}
	updated := false
	for i := range tasks {
		if tasks[i].ID == task.ID {
			tasks[i] = task
			updated = true
			break
		}
	}
	if !updated {
		tasks = append(tasks, task)
	}
	return task, writeJSON(s.path("tasks.json"), tasks)
}

func (s *Store) DeleteTask(id string) error {
	return deleteByID[Task](s.path("tasks.json"), id, func(t Task) string { return t.ID })
}

func (s *Store) ListMessages(taskID string, limit int) ([]Message, error) {
	messages, err := readJSON[Message](s.path("messages.json"))
	if err != nil {
		return nil, err
	}
	out := make([]Message, 0)
	for _, message := range messages {
		if message.TaskID == taskID {
			out = append(out, message)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt < out[j].CreatedAt })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *Store) SaveMessage(message Message) (Message, error) {
	if message.ID == "" {
		message.ID = uuid.NewString()
	}
	if message.CreatedAt == "" {
		message.CreatedAt = nowISO()
	}
	if message.Metadata == nil {
		message.Metadata = map[string]any{}
	}
	messages, err := readJSON[Message](s.path("messages.json"))
	if err != nil {
		return message, err
	}
	messages = append(messages, message)
	return message, writeJSON(s.path("messages.json"), messages)
}

func (s *Store) ListDocuments() ([]Document, error) {
	docs, err := readJSON[Document](s.path("documents.json"))
	if err != nil {
		return nil, err
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].UpdatedAt > docs[j].UpdatedAt })
	return docs, nil
}

func (s *Store) SaveDocument(doc Document) (Document, error) {
	now := nowISO()
	if doc.ID == "" {
		doc.ID = uuid.NewString()
		doc.CreatedAt = now
		doc.Version = 1
	}
	if doc.CreatedAt == "" {
		doc.CreatedAt = now
	}
	if doc.Type == "" {
		doc.Type = "draft"
	}
	if doc.Tags == nil {
		doc.Tags = []string{}
	}
	if doc.Metadata == nil {
		doc.Metadata = map[string]any{}
	}
	doc.UpdatedAt = now
	docs, err := readJSON[Document](s.path("documents.json"))
	if err != nil {
		return doc, err
	}
	updated := false
	for i := range docs {
		if docs[i].ID == doc.ID {
			doc.Version = docs[i].Version + 1
			docs[i] = doc
			updated = true
			break
		}
	}
	if !updated {
		docs = append(docs, doc)
	}
	return doc, writeJSON(s.path("documents.json"), docs)
}

func (s *Store) DeleteDocument(id string) error {
	return deleteByID[Document](s.path("documents.json"), id, func(d Document) string { return d.ID })
}

func (s *Store) ListNotifications(unreadOnly bool) ([]Notification, error) {
	notifications, err := readJSON[Notification](s.path("notifications.json"))
	if err != nil {
		return nil, err
	}
	if unreadOnly {
		filtered := notifications[:0]
		for _, n := range notifications {
			if !n.Read {
				filtered = append(filtered, n)
			}
		}
		notifications = filtered
	}
	sort.Slice(notifications, func(i, j int) bool { return notifications[i].CreatedAt > notifications[j].CreatedAt })
	return notifications, nil
}

func (s *Store) MarkNotificationRead(id string) error {
	notifications, err := readJSON[Notification](s.path("notifications.json"))
	if err != nil {
		return err
	}
	for i := range notifications {
		if notifications[i].ID == id {
			notifications[i].Read = true
			return writeJSON(s.path("notifications.json"), notifications)
		}
	}
	return os.ErrNotExist
}

func (s *Store) Stats() (map[string]any, error) {
	agents, err := s.ListAgents("")
	if err != nil {
		return nil, err
	}
	tasks, err := s.ListTasks("")
	if err != nil {
		return nil, err
	}
	docs, err := s.ListDocuments()
	if err != nil {
		return nil, err
	}
	notifications, err := s.ListNotifications(false)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"agents":        map[string]any{"total": len(agents), "by_status": countBy(agents, func(a Agent) string { return a.Status })},
		"tasks":         map[string]any{"total": len(tasks), "by_status": countBy(tasks, func(t Task) string { return t.Status })},
		"documents":     map[string]any{"total": len(docs)},
		"notifications": map[string]any{"total": len(notifications), "unread": countUnread(notifications)},
	}, nil
}

func (s *Store) path(name string) string { return filepath.Join(s.baseDir, name) }

func readJSON[T any](path string) ([]T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []T{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return []T{}, nil
	}
	var out []T
	if err := json.Unmarshal(data, &out); err != nil {
		return []T{}, nil
	}
	return out, nil
}

func writeJSON[T any](path string, values []T) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func deleteByID[T any](path string, id string, getID func(T) string) error {
	values, err := readJSON[T](path)
	if err != nil {
		return err
	}
	out := values[:0]
	found := false
	for _, value := range values {
		if getID(value) == id {
			found = true
			continue
		}
		out = append(out, value)
	}
	if !found {
		return os.ErrNotExist
	}
	return writeJSON(path, out)
}

func countBy[T any](items []T, key func(T) string) map[string]int {
	out := map[string]int{}
	for _, item := range items {
		out[key(item)]++
	}
	return out
}

func countUnread(items []Notification) int {
	count := 0
	for _, item := range items {
		if !item.Read {
			count++
		}
	}
	return count
}

func nowISO() string { return time.Now().UTC().Format(time.RFC3339Nano) }
