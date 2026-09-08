package Controllers

import (
	"fmt"
	"os"
	"strings"

	"github.com/kataras/iris"

	"server/Services"
)

type ConversationSyncReq struct {
	Conversations []Services.ConversationItem `json:"conversations" form:"conversations"`
}

func ListConversations(ctx iris.Context) ModelAndView {
	Services.IncRequest()

	auth := ctx.GetHeader("Authorization")
	uid, _, err := parseToken(auth)
	if err != nil {
		return ModelAndView{Code: 401, Data: map[string]interface{}{"error": "unauthorized"}}
	}

	items, err := Services.ListUserConversations(uid)
	if err != nil {
		return ModelAndView{Code: 500, Data: map[string]interface{}{"error": err.Error()}}
	}
	return ModelAndView{Code: 200, Data: items}
}

func SyncConversations(ctx iris.Context) ModelAndView {
	Services.IncRequest()

	auth := ctx.GetHeader("Authorization")
	uid, _, err := parseToken(auth)
	if err != nil {
		return ModelAndView{Code: 401, Data: map[string]interface{}{"error": "unauthorized"}}
	}

	var req ConversationSyncReq
	readBodySmart(ctx, &req)

	if req.Conversations == nil {
		req.Conversations = []Services.ConversationItem{}
	}
	if len(req.Conversations) > 200 {
		return ModelAndView{Code: 400, Data: map[string]interface{}{"error": "too many conversations"}}
	}

	for i := range req.Conversations {
		req.Conversations[i].Title = strings.TrimSpace(req.Conversations[i].Title)
		req.Conversations[i].Mode = strings.TrimSpace(req.Conversations[i].Mode)
	}
	logConversationSyncPreview(ctx, uid, req.Conversations)

	if err := Services.ReplaceUserConversations(uid, req.Conversations); err != nil {
		return ModelAndView{Code: 500, Data: map[string]interface{}{"error": err.Error()}}
	}
	return ModelAndView{Code: 200, Data: map[string]interface{}{"ok": true, "count": len(req.Conversations)}}
}

func logConversationSyncPreview(ctx iris.Context, userID int64, items []Services.ConversationItem) {
	if len(items) == 0 {
		ctx.Application().Logger().Infof("SYNC conversations uid=%d count=0", userID)
		return
	}

	debugFull := os.Getenv("DEBUG_CONVERSATIONS") == "1"
	previews := make([]string, 0, len(items))
	for _, item := range items {
		msgCount := len(item.Messages)
		start := 0
		if !debugFull {
			start = msgCount - 2
			if start < 0 {
				start = 0
			}
		}

		snippets := make([]string, 0, msgCount-start)
		for _, msg := range item.Messages[start:] {
			content := strings.TrimSpace(msg.Content)
			content = strings.ReplaceAll(content, "\n", " / ")
			content = strings.ReplaceAll(content, "\r", "")
			if !debugFull && len(content) > 60 {
				content = content[:60] + "..."
			}
			if content == "" {
				content = "(empty)"
			}
			snippets = append(snippets, fmt.Sprintf("%s:%s", strings.TrimSpace(msg.Name), content))
		}

		previews = append(previews, fmt.Sprintf(
			`{%s|%s|messages=%d|tail=%q}`,
			item.Title,
			item.Mode,
			msgCount,
			strings.Join(snippets, " || "),
		))
	}

	ctx.Application().Logger().Infof("SYNC conversations uid=%d count=%d preview=%s", userID, len(items), strings.Join(previews, " "))
}
