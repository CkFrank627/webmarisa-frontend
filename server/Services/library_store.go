package Services

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrLibraryInput = errors.New("标题需为1–80字，对话需包含1–2000条消息且不超过2MB；留言需为1–1000字")

type LibraryEntry struct {
	ID        int64                 `json:"id"`
	Title     string                `json:"title"`
	Author    string                `json:"author"`
	CreatedAt int64                 `json:"created_at"`
	Likes     int                   `json:"likes"`
	Comments  int                   `json:"comments"`
	Liked     bool                  `json:"liked"`
	Messages  []ConversationMessage `json:"messages,omitempty"`
}
type LibraryComment struct {
	ID        int64  `json:"id"`
	Author    string `json:"author"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
}

func libraryInit() error {
	if err := appInit(); err != nil {
		return err
	}
	_, err := ragDB.Exec(`
 CREATE TABLE IF NOT EXISTS app_library(id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER NOT NULL, author TEXT NOT NULL, title TEXT NOT NULL, messages_json TEXT NOT NULL, created_at INTEGER NOT NULL);
 CREATE TABLE IF NOT EXISTS app_library_likes(entry_id INTEGER NOT NULL, user_id INTEGER NOT NULL, PRIMARY KEY(entry_id,user_id));
 CREATE TABLE IF NOT EXISTS app_library_comments(id INTEGER PRIMARY KEY AUTOINCREMENT, entry_id INTEGER NOT NULL, author TEXT NOT NULL, user_id INTEGER NOT NULL, content TEXT NOT NULL, created_at INTEGER NOT NULL);
 CREATE INDEX IF NOT EXISTS idx_library_comments ON app_library_comments(entry_id,id);`)
	return err
}

func PublishLibrary(uid int64, author, title string, messages []ConversationMessage) (int64, error) {
	title = strings.TrimSpace(title)
	if uid <= 0 || title == "" || utf8.RuneCountInString(title) > 80 || len(messages) == 0 || len(messages) > 2000 {
		return 0, ErrLibraryInput
	}
	for _, m := range messages {
		if strings.TrimSpace(m.Name) == "" || strings.TrimSpace(m.Content) == "" {
			return 0, ErrLibraryInput
		}
	}
	data, err := json.Marshal(messages)
	if err != nil {
		return 0, err
	}
	if len(data) > 2*1024*1024 {
		return 0, ErrLibraryInput
	}
	if err = libraryInit(); err != nil {
		return 0, err
	}
	res, err := ragDB.Exec(`INSERT INTO app_library(user_id,author,title,messages_json,created_at) VALUES(?,?,?,?,?)`, uid, author, title, string(data), time.Now().Unix())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

const librarySelect = `SELECT e.id,e.title,e.author,e.created_at,
 (SELECT COUNT(*) FROM app_library_likes l WHERE l.entry_id=e.id),
 (SELECT COUNT(*) FROM app_library_comments c WHERE c.entry_id=e.id),
 EXISTS(SELECT 1 FROM app_library_likes l WHERE l.entry_id=e.id AND l.user_id=?) FROM app_library e `

func ListLibrary(uid int64, offset int) ([]LibraryEntry, error) {
	if err := libraryInit(); err != nil {
		return nil, err
	}
	rows, err := ragDB.Query(librarySelect+`ORDER BY e.id DESC LIMIT 30 OFFSET ?`, uid, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]LibraryEntry, 0)
	for rows.Next() {
		var e LibraryEntry
		if err := rows.Scan(&e.ID, &e.Title, &e.Author, &e.CreatedAt, &e.Likes, &e.Comments, &e.Liked); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func GetLibrary(id, uid int64) (LibraryEntry, error) {
	var e LibraryEntry
	if err := libraryInit(); err != nil {
		return e, err
	}
	err := ragDB.QueryRow(librarySelect+`WHERE e.id=?`, uid, id).Scan(&e.ID, &e.Title, &e.Author, &e.CreatedAt, &e.Likes, &e.Comments, &e.Liked)
	if err != nil {
		return e, err
	}
	var data string
	if err = ragDB.QueryRow(`SELECT messages_json FROM app_library WHERE id=?`, id).Scan(&data); err != nil {
		return e, err
	}
	err = json.Unmarshal([]byte(data), &e.Messages)
	return e, err
}

func SetLibraryLike(id, uid int64, liked bool) error {
	if _, err := GetLibrary(id, uid); err != nil {
		return err
	}
	if uid <= 0 {
		return ErrLibraryInput
	}
	if liked {
		_, err := ragDB.Exec(`INSERT OR IGNORE INTO app_library_likes(entry_id,user_id) VALUES(?,?)`, id, uid)
		return err
	}
	_, err := ragDB.Exec(`DELETE FROM app_library_likes WHERE entry_id=? AND user_id=?`, id, uid)
	return err
}

func AddLibraryComment(id, uid int64, author, content string) error {
	content = strings.TrimSpace(content)
	if uid <= 0 || content == "" || utf8.RuneCountInString(content) > 1000 {
		return ErrLibraryInput
	}
	if _, err := GetLibrary(id, uid); err != nil {
		return err
	}
	_, err := ragDB.Exec(`INSERT INTO app_library_comments(entry_id,user_id,author,content,created_at) VALUES(?,?,?,?,?)`, id, uid, author, content, time.Now().Unix())
	return err
}

func ListLibraryComments(id int64, offset int) ([]LibraryComment, error) {
	if _, err := GetLibrary(id, 0); err != nil {
		return nil, err
	}
	rows, err := ragDB.Query(`SELECT id,author,content,created_at FROM app_library_comments WHERE entry_id=? ORDER BY id DESC LIMIT 30 OFFSET ?`, id, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]LibraryComment, 0)
	for rows.Next() {
		var c LibraryComment
		if err := rows.Scan(&c.ID, &c.Author, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
