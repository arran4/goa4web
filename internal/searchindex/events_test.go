package searchindex

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

type searchQuerierStub struct {
	*db.QuerierStub
	words          map[string]int64
	added          map[int32]int32
	indexedComment int32
	createErr      error
}

func (s *searchQuerierStub) SystemCreateSearchWord(_ context.Context, word string) (int64, error) {
	if s.createErr != nil {
		return 0, s.createErr
	}
	if id, ok := s.words[word]; ok {
		return id, nil
	}
	id := int64(len(s.words) + 1)
	s.words[word] = id
	return id, nil
}

func (s *searchQuerierStub) SystemAddToForumCommentSearch(_ context.Context, arg db.SystemAddToForumCommentSearchParams) error {
	if arg.CommentID != 42 {
		return errors.New("unexpected CommentID")
	}
	s.added[arg.SearchwordlistIdsearchwordlist] = arg.WordCount
	return nil
}

func (s *searchQuerierStub) SystemSetCommentLastIndex(_ context.Context, id int32) error {
	s.indexedComment = id
	return nil
}

func TestIndexEventContract(t *testing.T) {
	data := IndexEventData{Type: TypeComment, ID: 42, Text: "A forum reply"}
	if IndexEventKey != "search_index" {
		t.Fatalf("IndexEventKey = %q, want search_index", IndexEventKey)
	}
	if TypeComment != "comment" || TypeWriting != "writing" || TypeLinker != "linker" || TypeImage != "image" {
		t.Fatalf("unexpected type constants: %q %q %q %q", TypeComment, TypeWriting, TypeLinker, TypeImage)
	}
	if data.Type != TypeComment || data.ID != 42 || data.Text == "" {
		t.Fatalf("IndexEventData = %#v", data)
	}
}

func TestBreakupTextToWords(t *testing.T) {
	got := BreakupTextToWords("Hello, world! Go-lang's 2nd test.")
	want := []string{"Hello", "world", "Go-lang's", "2nd", "test"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BreakupTextToWords = %v, want %v", got, want)
	}
}

func TestIndex(t *testing.T) {
	t.Run("indexes comment words and sets last index", func(t *testing.T) {
		stub := &searchQuerierStub{
			QuerierStub: &db.QuerierStub{},
			words:       map[string]int64{},
			added:       map[int32]int32{},
		}

		if err := Index(context.Background(), stub, IndexEventData{
			Type: TypeComment,
			ID:   42,
			Text: "Go go test!",
		}); err != nil {
			t.Fatalf("Index: %v", err)
		}
		if stub.indexedComment != 42 {
			t.Fatalf("indexedComment = %d, want 42", stub.indexedComment)
		}
		goID := int32(stub.words["go"])
		testID := int32(stub.words["test"])
		if stub.added[goID] != 2 || stub.added[testID] != 1 {
			t.Fatalf("word counts = %v (goID=%d testID=%d)", stub.added, goID, testID)
		}
	})

	t.Run("propagates search word creation error", func(t *testing.T) {
		want := errors.New("search word failed")
		stub := &searchQuerierStub{
			QuerierStub: &db.QuerierStub{},
			createErr:   want,
		}
		if err := Index(context.Background(), stub, IndexEventData{
			Type: TypeComment,
			ID:   7,
			Text: "hello",
		}); !errors.Is(err, want) {
			t.Fatalf("Index error = %v, want %v", err, want)
		}
	})
}
