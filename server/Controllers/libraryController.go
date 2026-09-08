package Controllers

import (
	"database/sql"
	"encoding/json"
	"github.com/kataras/iris"
	"io"
	"server/Services"
	"strconv"
)

func libraryResult(ctx iris.Context, data interface{}, err error) (int, ModelAndView) {
	code := 200
	if err != nil {
		code = 500
		msg := "图书馆暂时不可用，请稍后重试"
		if err == sql.ErrNoRows {
			code = 404
			msg = "这条记录不存在"
		}
		if err == Services.ErrLibraryInput {
			code = 400
			msg = err.Error()
		}
		data = map[string]interface{}{"error": msg}
	}
	ctx.StatusCode(code)
	return code, ModelAndView{Code: code, Data: data}
}
func libraryOffset(ctx iris.Context) int {
	n, _ := strconv.Atoi(ctx.URLParam("offset"))
	if n < 0 {
		return 0
	}
	return n
}
func libraryID(ctx iris.Context) int64 {
	n, _ := strconv.ParseInt(ctx.Params().Get("id"), 10, 64)
	return n
}
func libraryAuth(ctx iris.Context) (int64, string, error) {
	return parseToken(ctx.GetHeader("Authorization"))
}
func libraryUnauthorized(ctx iris.Context) (int, ModelAndView) {
	ctx.StatusCode(401)
	return 401, ModelAndView{Code: 401, Data: map[string]interface{}{"error": "请先登录"}}
}
func libraryRead(ctx iris.Context, v interface{}) error {
	return json.NewDecoder(io.LimitReader(ctx.Request().Body, 3*1024*1024)).Decode(v)
}

func LibraryList(ctx iris.Context) (int, ModelAndView) {
	uid, _, err := libraryAuth(ctx)
	if err != nil {
		uid = 0
	}
	items, e := Services.ListLibrary(uid, libraryOffset(ctx))
	return libraryResult(ctx, items, e)
}
func LibraryDetail(ctx iris.Context) (int, ModelAndView) {
	uid, _, err := libraryAuth(ctx)
	if err != nil {
		uid = 0
	}
	item, e := Services.GetLibrary(libraryID(ctx), uid)
	return libraryResult(ctx, item, e)
}
func LibraryPublish(ctx iris.Context) (int, ModelAndView) {
	uid, name, err := libraryAuth(ctx)
	if err != nil {
		return libraryUnauthorized(ctx)
	}
	var req struct {
		Title    string                         `json:"title"`
		Messages []Services.ConversationMessage `json:"messages"`
	}
	if libraryRead(ctx, &req) != nil {
		return libraryResult(ctx, nil, Services.ErrLibraryInput)
	}
	id, err := Services.PublishLibrary(uid, name, req.Title, req.Messages)
	return libraryResult(ctx, map[string]interface{}{"id": id}, err)
}
func LibraryLike(ctx iris.Context) (int, ModelAndView) {
	uid, _, err := libraryAuth(ctx)
	if err != nil {
		return libraryUnauthorized(ctx)
	}
	var req struct {
		Liked bool `json:"liked"`
	}
	if libraryRead(ctx, &req) != nil {
		return libraryResult(ctx, nil, Services.ErrLibraryInput)
	}
	if err = Services.SetLibraryLike(libraryID(ctx), uid, req.Liked); err != nil {
		return libraryResult(ctx, nil, err)
	}
	item, err := Services.GetLibrary(libraryID(ctx), uid)
	return libraryResult(ctx, item, err)
}
func LibraryComments(ctx iris.Context) (int, ModelAndView) {
	items, err := Services.ListLibraryComments(libraryID(ctx), libraryOffset(ctx))
	return libraryResult(ctx, items, err)
}
func LibraryComment(ctx iris.Context) (int, ModelAndView) {
	uid, name, err := libraryAuth(ctx)
	if err != nil {
		return libraryUnauthorized(ctx)
	}
	var req struct {
		Content string `json:"content"`
	}
	if libraryRead(ctx, &req) != nil {
		return libraryResult(ctx, nil, Services.ErrLibraryInput)
	}
	err = Services.AddLibraryComment(libraryID(ctx), uid, name, req.Content)
	return libraryResult(ctx, map[string]interface{}{"ok": err == nil}, err)
}
