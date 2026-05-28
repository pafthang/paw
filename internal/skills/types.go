package skills

type Skill struct {
	Name        string       `yaml:"name" json:"name"`
	Description string       `yaml:"description" json:"description"`
	Version     string       `yaml:"version" json:"version"`
	Prompts     SkillPrompts `yaml:"prompts" json:"prompts"`
	Commands    []Command    `yaml:"commands,omitempty" json:"commands,omitempty"`
}

type SkillPrompts struct {
	System string `yaml:"system,omitempty" json:"system,omitempty"`
}

type Command struct {
	Name        string         `yaml:"name" json:"name"`
	Description string         `yaml:"description,omitempty" json:"description,omitempty"`
	Tool        string         `yaml:"tool" json:"tool"`
	Input       map[string]any `yaml:"input,omitempty" json:"input,omitempty"`
}
