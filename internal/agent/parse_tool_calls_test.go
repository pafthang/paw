package agent

import "testing"

func TestParseToolCallsEnvelope_StrictJSON(t *testing.T) {
	calls, err := ParseToolCallsEnvelope(`{"tool_calls":[{"name":"file.read","input":{"path":"README.md"}}]}`)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(calls) != 1 || calls[0].Name != "file.read" {
		t.Fatalf("calls: %#v", calls)
	}
}

func TestParseToolCallsEnvelope_FencedJSON(t *testing.T) {
	calls, err := ParseToolCallsEnvelope("```json\n{\"tool_calls\":[{\"name\":\"file.read\",\"input\":{\"path\":\"README.md\"}}]}\n```")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("calls: %#v", calls)
	}
}

func TestParseToolCallsEnvelope_TextWrapped(t *testing.T) {
	calls, err := ParseToolCallsEnvelope("I need a tool.\n{\"tool_calls\":[{\"name\":\"file.read\",\"input\":{\"path\":\"README.md\"}}]}\nThanks.")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("calls: %#v", calls)
	}
}

func TestParseToolCallsEnvelope_AlternateKey(t *testing.T) {
	calls, err := ParseToolCallsEnvelope("{\"tools\":[{\"name\":\"file.read\",\"input\":{\"path\":\"README.md\"}}]}")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("calls: %#v", calls)
	}
}

func TestParseToolCallsEnvelope_MalformedJSONReturnsError(t *testing.T) {
	_, err := ParseToolCallsEnvelope("{\"tool_calls\": [}")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseToolCallsEnvelope_NoToolCalls(t *testing.T) {
	calls, err := ParseToolCallsEnvelope("{\"answer\":\"hi\"}")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(calls) != 0 {
		t.Fatalf("calls: %#v", calls)
	}
}
