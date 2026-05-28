package llm

import "testing"

func TestAnthropicConvertMessages_SystemSeparated(t *testing.T) {
	system, out := anthropicConvertMessages([]Message{
		{Role: "system", Content: "sys1"},
		{Role: "user", Content: "hi"},
		{Role: "system", Content: "sys2"},
		{Role: "assistant", Content: "hello"},
	})
	if system != "sys1\n\nsys2" {
		t.Fatalf("system = %q", system)
	}
	if len(out) != 2 {
		t.Fatalf("len(out) = %d", len(out))
	}
	if out[0].Role != "user" || out[0].Content[0].Text != "hi" {
		t.Fatalf("out[0] = %#v", out[0])
	}
	if out[1].Role != "assistant" || out[1].Content[0].Text != "hello" {
		t.Fatalf("out[1] = %#v", out[1])
	}
}

func TestAnthropicConvertMessages_UnknownRoleDefaultsToUser(t *testing.T) {
	_, out := anthropicConvertMessages([]Message{
		{Role: "", Content: "x"},
		{Role: "tool", Content: "y"},
	})
	if len(out) != 2 {
		t.Fatalf("len(out) = %d", len(out))
	}
	if out[0].Role != "user" || out[1].Role != "user" {
		t.Fatalf("roles = %q, %q", out[0].Role, out[1].Role)
	}
}
