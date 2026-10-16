package search

import "testing"

func TestIndexEventContract(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		data := IndexEventData{Type: TypeComment, ID: 42, Text: "A forum reply"}
		if IndexEventKey != "search_index" {
			t.Fatalf("IndexEventKey = %q, want search_index", IndexEventKey)
		}
		if data.Type != TypeComment || data.ID != 42 || data.Text == "" {
			t.Fatalf("IndexEventData = %#v", data)
		}
	})
}
