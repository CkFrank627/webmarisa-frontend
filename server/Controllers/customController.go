package Controllers

import (
	"fmt"
	"strings"

	"github.com/kataras/iris"

	"server/Services"
)

type PersonaCreateReq struct {
	Name   string `json:"name" form:"name"`
	Prompt string `json:"prompt" form:"prompt"`
}

type TeachReq struct {
	Q string `json:"q" form:"q"`
	A string `json:"a" form:"a"`
}

type CustomReplyReq struct {
	PersonaID int64  `json:"persona_id" form:"persona_id"`
	Message   string `json:"message" form:"message"`
}

func ListPersonas(ctx iris.Context) ModelAndView {
	auth := ctx.GetHeader("Authorization")
	uid, _, err := parseToken(auth)
	if err != nil {
		return ModelAndView{Code: 401, Data: map[string]interface{}{"error": "unauthorized"}}
	}
	ps, err := Services.ListPersonas(uid)
	if err != nil {
		return ModelAndView{Code: 500, Data: map[string]interface{}{"error": err.Error()}}
	}
	return ModelAndView{Code: 200, Data: ps}
}

func CreatePersona(ctx iris.Context) ModelAndView {
	auth := ctx.GetHeader("Authorization")
	uid, _, err := parseToken(auth)
	if err != nil {
		return ModelAndView{Code: 401, Data: map[string]interface{}{"error": "unauthorized"}}
	}
	var req PersonaCreateReq
readBodySmart(ctx, &req)

	id, err := Services.CreatePersona(uid, req.Name, req.Prompt)
	if err != nil {
		return ModelAndView{Code: 400, Data: map[string]interface{}{"error": err.Error()}}
	}
	return ModelAndView{Code: 200, Data: map[string]interface{}{"id": id}}
}

func DeletePersona(ctx iris.Context) ModelAndView {
	auth := ctx.GetHeader("Authorization")
	uid, _, err := parseToken(auth)
	if err != nil {
		return ModelAndView{Code: 401, Data: map[string]interface{}{"error": "unauthorized"}}
	}
	id, _ := ctx.Params().GetUint64("id")
	if id == 0 {
		return ModelAndView{Code: 400, Data: map[string]interface{}{"error": "bad id"}}
	}
	if err := Services.DeletePersona(uid, int64(id)); err != nil {
		return ModelAndView{Code: 400, Data: map[string]interface{}{"error": err.Error()}}
	}
	return ModelAndView{Code: 200, Data: map[string]interface{}{"ok": true}}
}

func ListPersonaTeach(ctx iris.Context) ModelAndView {
	auth := ctx.GetHeader("Authorization")
	uid, _, err := parseToken(auth)
	if err != nil { return ModelAndView{Code: 401, Data: map[string]interface{}{"error": "unauthorized"}} }

	id, _ := ctx.Params().GetUint64("id")
	items, err := Services.ListTeach(uid, int64(id), 100)
	if err != nil { return ModelAndView{Code: 400, Data: map[string]interface{}{"error": err.Error()}} }
	return ModelAndView{Code: 200, Data: items}
}

func AddPersonaTeach(ctx iris.Context) ModelAndView {
	auth := ctx.GetHeader("Authorization")
	uid, _, err := parseToken(auth)
	if err != nil { return ModelAndView{Code: 401, Data: map[string]interface{}{"error": "unauthorized"}} }

	id, _ := ctx.Params().GetUint64("id")
	var req TeachReq
	readBodySmart(ctx, &req)

	if err := Services.AddTeach(uid, int64(id), req.Q, req.A); err != nil {
		return ModelAndView{Code: 400, Data: map[string]interface{}{"error": err.Error()}}
	}
	return ModelAndView{Code: 200, Data: map[string]interface{}{"ok": true}}
}

func DeleteTeachLast(ctx iris.Context) ModelAndView {
	auth := ctx.GetHeader("Authorization")
	uid, _, err := parseToken(auth)
	if err != nil { return ModelAndView{Code: 401, Data: map[string]interface{}{"error": "unauthorized"}} }

	id, _ := ctx.Params().GetUint64("id")
	if err := Services.DeleteTeachLast(uid, int64(id)); err != nil {
		return ModelAndView{Code: 400, Data: map[string]interface{}{"error": err.Error()}}
	}
	return ModelAndView{Code: 200, Data: map[string]interface{}{"ok": true}}
}

func CustomReply(ctx iris.Context) ModelAndView {
	auth := ctx.GetHeader("Authorization")
	uid, _, err := parseToken(auth)
	if err != nil { return ModelAndView{Code: 401, Data: map[string]interface{}{"error": "unauthorized"}} }

	var req CustomReplyReq
	readBodySmart(ctx, &req)
	ctx.Application().Logger().Infof("CUSTOM_REPLY persona_id=%d message_len=%d", req.PersonaID, len(strings.TrimSpace(req.Message)))

	msg := strings.TrimSpace(req.Message)
	if req.PersonaID <= 0 || msg == "" {
		return ModelAndView{Code: 400, Data: map[string]interface{}{"error": "persona_id/message required"}}
	}
	// ✅ 记录用户输入（完整对话 log：user）
_ = Services.AddPersonaChatLog(uid, req.PersonaID, "user", msg)

	// 取 persona prompt + teach 列表
	pp, err := Services.GetPersonaPrompt(uid, req.PersonaID)
	if err != nil {
		return ModelAndView{Code: 400, Data: map[string]interface{}{"error": err.Error()}}
	}
	teach, _ := Services.ListTeach(uid, req.PersonaID, 50)

	teachText := ""
	for i := 0; i < len(teach); i++ {
		t := teach[i]
		teachText += fmt.Sprintf("Q: %s\nA: %s\n", t.Q, t.A)
	}

	// teach > persona > default
	system := "你必须始终精确写作「白丝魔理沙」「雾雨魔理沙」，不得改写。禁止输出注释腔。\n"
	system += "【默认设定】你是白丝魔理沙，俏皮机灵。\n"
	system += "【人格描述】\n" + pp + "\n"

	userPrompt := "【优先使用的 Teach（若问题匹配，优先用对应 A 回复）】\n" + teachText + "\n"
	userPrompt += "规则：如果 Teach 中存在与用户问题语义相近的 Q，请优先用其 A 作答；否则按人格描述回答；再否则按默认设定回答。\n\n"
	userPrompt += "用户问题：\n" + msg

	ans, err := Services.CallCloudLLM(system, userPrompt)
	if err != nil {
		return ModelAndView{Code: 500, Data: map[string]interface{}{"answer": "LLM 调用失败：" + err.Error()}}
	}
// ✅ 记录助手输出（完整对话 log：assistant）
_ = Services.AddPersonaChatLog(uid, req.PersonaID, "assistant", ans)

	return ModelAndView{Code: 200, Data: map[string]interface{}{"answer": ans}}
}

// GET /api/custom/personas/{id}/logs?limit=100
func ListPersonaLogs(ctx iris.Context) ModelAndView {
	auth := ctx.GetHeader("Authorization")
	uid, _, err := parseToken(auth)
	if err != nil {
		return ModelAndView{Code: 401, Data: map[string]interface{}{"error": "unauthorized"}}
	}

	id, _ := ctx.Params().GetUint64("id")
	limit, _ := ctx.URLParamInt("limit")

	items, err := Services.ListPersonaChatLog(uid, int64(id), limit)
	if err != nil {
		return ModelAndView{Code: 400, Data: map[string]interface{}{"error": err.Error()}}
	}
	return ModelAndView{Code: 200, Data: items}
}

// GET /api/custom/logs?limit=200
func ListUserLogs(ctx iris.Context) ModelAndView {
	auth := ctx.GetHeader("Authorization")
	uid, _, err := parseToken(auth)
	if err != nil {
		return ModelAndView{Code: 401, Data: map[string]interface{}{"error": "unauthorized"}}
	}
	limit, _ := ctx.URLParamInt("limit")
	items, err := Services.ListUserChatLog(uid, limit)
	if err != nil {
		return ModelAndView{Code: 500, Data: map[string]interface{}{"error": err.Error()}}
	}
	return ModelAndView{Code: 200, Data: items}
}

// DELETE /api/custom/personas/{id}/logs
func ClearPersonaLogs(ctx iris.Context) ModelAndView {
	auth := ctx.GetHeader("Authorization")
	uid, _, err := parseToken(auth)
	if err != nil {
		return ModelAndView{Code: 401, Data: map[string]interface{}{"error": "unauthorized"}}
	}
	id, _ := ctx.Params().GetUint64("id")

	if err := Services.ClearPersonaChatLog(uid, int64(id)); err != nil {
		return ModelAndView{Code: 400, Data: map[string]interface{}{"error": err.Error()}}
	}
	return ModelAndView{Code: 200, Data: map[string]interface{}{"ok": true}}
}
