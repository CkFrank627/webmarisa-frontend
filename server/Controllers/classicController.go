//classicController.go

package Controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"github.com/kataras/iris"
	"server/Services"
)

type ClassicReq struct {
	Keyword string `json:"keyword" form:"keyword"`
	Answer  string `json:"answer"  form:"answer"`
}

// 经典版专用解析器：
// 1) 直接读原始 body
// 2) 去 BOM
// 3) json.Unmarshal
// 4) 失败再读 form
// 5) 再兜底 FormValue
func readClassicReq(ctx iris.Context) ClassicReq {
	var r ClassicReq

	// 读原始 body
	b, _ := io.ReadAll(ctx.Request().Body)
	ctx.Request().Body = io.NopCloser(bytes.NewBuffer(b))

	// 去 UTF-8 BOM
	b = bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF})
	bb := bytes.TrimSpace(b)

	// 优先 JSON
	if len(bb) > 0 {
		_ = json.Unmarshal(bb, &r)
	}

	// fallback form
	if r.Keyword == "" && r.Answer == "" {
		_ = ctx.ReadForm(&r)
	}

	// 最后兜底
	if r.Keyword == "" {
		r.Keyword = ctx.FormValue("keyword")
	}
	if r.Answer == "" {
		r.Answer = ctx.FormValue("answer")
	}
	return r
}

func ClassicStatus(ctx iris.Context) ModelAndView {
	Services.IncRequest()
	n, err := Services.ClassicCount()
	if err != nil {
		return ModelAndView{Code: 500, Data: map[string]interface{}{"error": err.Error()}}
	}
	return ModelAndView{Code: 200, Data: map[string]interface{}{"count": n}}
}

func ClassicList(ctx iris.Context) ModelAndView {
	Services.IncRequest()
	limit, _ := ctx.URLParamInt("limit")
	items, err := Services.ClassicList(limit)
	if err != nil {
		return ModelAndView{Code: 500, Data: map[string]interface{}{"error": err.Error()}}
	}
	return ModelAndView{Code: 200, Data: items}
}

func ClassicTeach(ctx iris.Context) ModelAndView {
	Services.IncRequest()

	r := readClassicReq(ctx)
	q := strings.TrimSpace(r.Keyword)
	a := strings.TrimSpace(r.Answer)

	// 支持 answer="Q`A"
	if strings.Contains(a, "`") {
		parts := strings.SplitN(a, "`", 2)
		q = strings.TrimSpace(parts[0])
		a = strings.TrimSpace(parts[1])
	}

	if q == "" || a == "" {
		return ModelAndView{Code: 400, Data: map[string]interface{}{"error": "need Q/A (answer=\"Q`A\" or keyword+answer)"}}
	}

	id, err := Services.ClassicAddQA(q, a)
	if err != nil {
		return ModelAndView{Code: 400, Data: map[string]interface{}{"error": err.Error()}}
	}
	return ModelAndView{Code: 200, Data: map[string]interface{}{"ok": true, "id": id}}
}

func ClassicForget(ctx iris.Context) ModelAndView {
	Services.IncRequest()

	r := readClassicReq(ctx)
	arg := strings.TrimSpace(r.Answer)
	if arg == "" {
		arg = strings.TrimSpace(r.Keyword)
	}

	if arg != "" {
		if id, e := strconv.ParseInt(arg, 10, 64); e == nil && id > 0 {
			ok, err := Services.ClassicDeleteByID(id)
			if err != nil {
				return ModelAndView{Code: 500, Data: map[string]interface{}{"error": err.Error()}}
			}
			return ModelAndView{Code: 200, Data: map[string]interface{}{"ok": ok}}
		}

		ok, err := Services.ClassicDeleteByQ(arg)
		if err != nil {
			return ModelAndView{Code: 500, Data: map[string]interface{}{"error": err.Error()}}
		}
		return ModelAndView{Code: 200, Data: map[string]interface{}{"ok": ok}}
	}

	ok, err := Services.ClassicDeleteLast()
	if err != nil {
		return ModelAndView{Code: 500, Data: map[string]interface{}{"error": err.Error()}}
	}
	return ModelAndView{Code: 200, Data: map[string]interface{}{"ok": ok}}
}

func ClassicReply(ctx iris.Context) ModelAndView {
	Services.IncRequest()

	r := readClassicReq(ctx)
	q := strings.TrimSpace(r.Keyword)
	if q == "" {
		q = strings.TrimSpace(r.Answer)
	}
	if q == "" {
		return ModelAndView{Code: 400, Data: map[string]interface{}{"answer": "empty keyword"}}
	}

	ans, err := Services.ClassicReplyText(q)
	if err != nil {
		return ModelAndView{Code: 500, Data: map[string]interface{}{"answer": "classic reply error: " + err.Error()}}
	}

	Services.IncReply()
	return ModelAndView{Code: 200, Data: map[string]interface{}{"answer": ans}}
}
