// Controllers/affinityController.go
package Controllers

import (
	"strconv"

	"github.com/kataras/iris"
	"server/Services"
)

func AffinityMe(ctx iris.Context) ModelAndView {
	auth := ctx.GetHeader("Authorization")
	uid, _, err := parseToken(auth)
	if err != nil || uid <= 0 {
		return ModelAndView{Code: 200, Data: map[string]interface{}{
			"logged_in":  false,
			"score":      nil,
			"show_login": true,
		}}
	}

	score, err := Services.GetAffinity(uid)
	if err != nil {
		return ModelAndView{Code: 500, Data: map[string]interface{}{"error": err.Error()}}
	}

	return ModelAndView{Code: 200, Data: map[string]interface{}{
		"logged_in":   true,
		"score":       score,
		"level":       Services.AffinityLevel(score),
		"can_intimate": Services.CanIntimate(score),
	}}
}

func AffinityLogs(ctx iris.Context) ModelAndView {
	// 你可以先不做严格鉴权，或只允许你自己的 uid。这里给你一个“最低限度”版本：
	auth := ctx.GetHeader("Authorization")
	uid, _, err := parseToken(auth)
	if err != nil || uid <= 0 {
		return ModelAndView{Code: 401, Data: map[string]interface{}{"error": "unauthorized"}}
	}

	// 简单限制：默认只能看自己的日志；如果你要“全局巡查”，你可以临时传 user_id=0 且要求 uid 是管理员
	userID := uid
	if v := ctx.URLParam("user_id"); v != "" {
		if x, e := strconv.ParseInt(v, 10, 64); e == nil && x >= 0 {
			// TODO: 如果你要管理员全局巡查，这里加 isAdmin(uid) 判断
			userID = x
		}
	}

	limit, _ := ctx.URLParamInt("limit")
	items, err := Services.ListAffinityLogs(userID, limit)
	if err != nil {
		return ModelAndView{Code: 500, Data: map[string]interface{}{"error": err.Error()}}
	}
	return ModelAndView{Code: 200, Data: items}
}
