package Controllers

import (
	"encoding/json"
	"github.com/kataras/iris"
	"github.com/kataras/iris/hero"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLibraryHTTPPermissionsAndPublicRead(t *testing.T) {
	t.Setenv("RAG_DB_PATH", "file:library_http_test?mode=memory&cache=shared")
	app := iris.New()
	app.Get("/library", hero.Handler(LibraryList))
	app.Post("/library", hero.Handler(LibraryPublish))
	app.Get("/library/{id:int64}", hero.Handler(LibraryDetail))
	app.Put("/library/{id:int64}/like", hero.Handler(LibraryLike))
	app.Post("/library/{id:int64}/comments", hero.Handler(LibraryComment))
	if err := app.Build(); err != nil {
		t.Fatal(err)
	}
	token, err := makeToken(701, "author")
	if err != nil {
		t.Fatal(err)
	}
	request := func(method, path, body, auth string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if auth != "" {
			req.Header.Set("Authorization", "Bearer "+auth)
		}
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, req)
		return rec
	}
	for _, endpoint := range []struct{ method, path string }{{"POST", "/library"}, {"PUT", "/library/1/like"}, {"POST", "/library/1/comments"}} {
		if rec := request(endpoint.method, endpoint.path, `{}`, ""); rec.Code != 401 {
			t.Fatalf("guest write: %d %s", rec.Code, rec.Body.String())
		}
	}
	if rec := request("POST", "/library", `{broken`, token); rec.Code != 400 {
		t.Fatal("malformed body accepted")
	}
	rec := request("POST", "/library", `{"title":"公开测试","author":"forged","messages":[{"name":"You","content":"hello"}]}`, token)
	if rec.Code != 200 {
		t.Fatalf("publish: %d %s", rec.Code, rec.Body.String())
	}
	var result struct {
		Data struct {
			ID json.Number `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	rec = request("GET", "/library/"+result.Data.ID.String(), "", "")
	var detail struct {
		Data struct {
			Author   string `json:"author"`
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &detail); err != nil {
		t.Fatal(err)
	}
	if rec.Code != 200 || detail.Data.Author != "author" || len(detail.Data.Messages) != 1 || detail.Data.Messages[0].Content != "hello" {
		t.Fatalf("public detail or author attribution failed: %s", rec.Body.String())
	}
	if rec := request("GET", "/library/99999999", "", ""); rec.Code != 404 {
		t.Fatal("missing record status", rec.Code)
	}
}
