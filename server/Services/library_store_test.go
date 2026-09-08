package Services

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
)

func TestLibrarySnapshotLikesCommentsAndPagination(t *testing.T) {
	path := filepath.Join(t.TempDir(), "library.db")
	t.Setenv("RAG_DB_PATH", path)
	previous := ragDB
	ragDB = nil
	t.Cleanup(func() {
		if ragDB != nil {
			ragDB.Close()
		}
		ragDB = previous
	})
	messages := []ConversationMessage{{Name: "You", Content: "中文\n<script>literal</script>"}, {Name: "魔理沙", Content: "你好"}}
	id, err := PublishLibrary(7, "reader", "故事", messages)
	if err != nil {
		t.Fatal(err)
	}
	messages[0].Content = "changed privately"
	if err := ReplaceUserConversations(7, nil); err != nil {
		t.Fatal(err)
	}
	e, err := GetLibrary(id, 0)
	if err != nil || e.Messages[0].Content != "中文\n<script>literal</script>" {
		t.Fatalf("snapshot changed: %+v %v", e, err)
	}
	for i := 0; i < 2; i++ {
		if err := SetLibraryLike(id, 8, true); err != nil {
			t.Fatal(err)
		}
	}
	if err := SetLibraryLike(id, 9, true); err != nil {
		t.Fatal(err)
	}
	e, err = GetLibrary(id, 8)
	if err != nil || e.Likes != 2 || !e.Liked {
		t.Fatalf("likes not idempotent: %+v %v", e, err)
	}
	if err := SetLibraryLike(id, 8, false); err != nil {
		t.Fatal(err)
	}
	e, _ = GetLibrary(id, 8)
	if e.Likes != 1 || e.Liked {
		t.Fatal("unlike failed")
	}
	if err := AddLibraryComment(id, 8, "reader2", "  留言\n第二行  "); err != nil {
		t.Fatal(err)
	}
	comments, err := ListLibraryComments(id, 0)
	if err != nil || len(comments) != 1 || comments[0].Content != "留言\n第二行" {
		t.Fatalf("comments: %+v %v", comments, err)
	}
	if err := AddLibraryComment(id, 8, "reader2", "  "); err != ErrLibraryInput {
		t.Fatal("empty comment accepted")
	}
	if err := AddLibraryComment(id, 8, "reader2", strings.Repeat("字", 1001)); err != ErrLibraryInput {
		t.Fatal("oversize comment accepted")
	}
	if err := SetLibraryLike(id+999, 8, true); err != sql.ErrNoRows {
		t.Fatal("missing entry liked")
	}
	if err := AddLibraryComment(id+999, 8, "reader2", "hello"); err != sql.ErrNoRows {
		t.Fatal("missing entry commented")
	}
	if _, err := PublishLibrary(0, "guest", "title", messages); err != ErrLibraryInput {
		t.Fatal("guest published")
	}
	if _, err := PublishLibrary(7, "reader", "title", nil); err != ErrLibraryInput {
		t.Fatal("empty snapshot published")
	}
	for i := 0; i < 31; i++ {
		if _, err := PublishLibrary(7, "reader", "another", messages); err != nil {
			t.Fatal(err)
		}
	}
	first, err := ListLibrary(0, 0)
	if err != nil || len(first) != 30 || first[0].ID <= first[29].ID {
		t.Fatal("first page failed", err)
	}
	second, err := ListLibrary(0, 30)
	if err != nil || len(second) != 2 || second[1].ID != id || second[1].Comments != 1 {
		t.Fatal("second page failed", err)
	}
	ragDB.Close()
	ragDB = nil
	e, err = GetLibrary(id, 9)
	if err != nil || e.Likes != 1 || e.Comments != 1 || !e.Liked {
		t.Fatal("persistence failed", err)
	}
}
