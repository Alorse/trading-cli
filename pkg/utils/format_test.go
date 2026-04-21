package utils

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"key": "value"}
	err := WriteJSON(&buf, data)
	if err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, `"key"`) || !strings.Contains(output, `"value"`) {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestToJSONString(t *testing.T) {
	data := map[string]int{"count": 42}
	s, err := ToJSONString(data)
	if err != nil {
		t.Fatalf("ToJSONString failed: %v", err)
	}
	if !strings.Contains(s, `"count"`) || !strings.Contains(s, "42") {
		t.Errorf("unexpected output: %s", s)
	}
}
