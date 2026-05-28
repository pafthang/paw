package tools

import (
	"fmt"
	"sort"
)

type Registry struct {
	tools map[string]Tool
}

func NewRegistry(toolList ...Tool) *Registry {
	registry := &Registry{tools: map[string]Tool{}}
	for _, tool := range toolList {
		registry.Register(tool)
	}
	return registry
}

func DefaultRegistry() *Registry {
	return NewRegistry(FileReadTool{}, FileWriteTool{}, ShellRunTool{})
}

func (r *Registry) Register(tool Tool) {
	if tool == nil {
		return
	}
	r.tools[tool.Name()] = tool
}

func (r *Registry) Get(name string) (Tool, error) {
	tool, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("unknown tool %q", name)
	}
	return tool, nil
}

func (r *Registry) List() []map[string]string {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]map[string]string, 0, len(names))
	for _, name := range names {
		tool := r.tools[name]
		out = append(out, map[string]string{"name": tool.Name(), "description": tool.Description()})
	}
	return out
}
