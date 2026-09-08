package Controllers

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/kataras/iris"
)

func TestAuditMiddlewareRestoresBodyAndRecordsGuest(t *testing.T) {
	dbPath := "file:input_audit_test?mode=memory&cache=shared"
	t.Setenv("RAG_DB_PATH", dbPath)
	app := iris.New()
	want := url.Values{"keyword": {strings.Repeat("中文🙂\n", 1000)}}.Encode()
	app.Post("/Reply", AuditUserInput, func(ctx iris.Context) {
		body, err := io.ReadAll(ctx.Request().Body)
		if err != nil || string(body) != want {
			t.Error("middleware changed request body")
		}
		ctx.StatusCode(204)
	})
	if err := app.Build(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/Reply", strings.NewReader(want))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)
	if rec.Code != 204 {
		t.Fatalf("unexpected status %d: %s", rec.Code, rec.Body.String())
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		db.Exec("DROP TRIGGER IF EXISTS reject_audit")
		db.Close()
	}()
	var raw string
	var uid int64
	if err := db.QueryRow("SELECT user_id, raw_body FROM app_input_audit").Scan(&uid, &raw); err != nil {
		t.Fatal(err)
	}
	if uid != 0 || raw != want {
		t.Fatal("guest input not recorded intact")
	}
	// A database write failure must stop the downstream handler.
	if _, err := db.Exec("CREATE TRIGGER reject_audit BEFORE INSERT ON app_input_audit BEGIN SELECT RAISE(FAIL, 'test failure'); END"); err != nil {
		t.Fatal(err)
	}
	rec = httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("POST", "/Reply", strings.NewReader(want)))
	if rec.Code != 503 {
		t.Fatalf("write failure did not stop request: %d", rec.Code)
	}
}

func TestReadableInputPreservesFormValues(t *testing.T) {
	want := "  " + strings.Repeat("中文🙂\n<script>'&+", 500) + "  "
	body := url.Values{"keyword": {want}}.Encode()
	var got map[string][]string
	if err := json.Unmarshal([]byte(readableInput([]byte(body), "application/x-www-form-urlencoded; charset=UTF-8")), &got); err != nil {
		t.Fatal(err)
	}
	if len(got["keyword"]) != 1 || got["keyword"][0] != want {
		t.Fatal("input changed")
	}
}

func TestReadableInputPreservesMalformedBody(t *testing.T) {
	body := "{invalid\n中文"
	if readableInput([]byte(body), "application/json") != body {
		t.Fatal("malformed input lost")
	}
}
