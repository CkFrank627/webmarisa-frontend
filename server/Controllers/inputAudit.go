package Controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/url"

	"github.com/kataras/iris"
	"server/Services"
)

// Only register on content endpoints, never login/register: passwords and
// Authorization headers must not be copied into the audit table.
func AuditUserInput(ctx iris.Context) {
	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		ctx.StatusCode(400)
		ctx.JSON(ModelAndView{Code: 400, Data: "无法完整读取输入，请重试"})
		return
	}
	ctx.Request().Body = io.NopCloser(bytes.NewReader(body))
	uid, username, authErr := parseToken(ctx.GetHeader("Authorization"))
	if authErr != nil {
		uid, username = 0, ""
	}
	if err := Services.RecordUserInput(uid, username, ctx.Method(), ctx.Path(),
		ctx.GetHeader("Content-Type"), string(body), readableInput(body, ctx.GetHeader("Content-Type"))); err != nil {
		ctx.Application().Logger().Errorf("input audit write failed: %v", err)
		ctx.StatusCode(503)
		ctx.JSON(ModelAndView{Code: 503, Data: "输入记录暂不可用，请稍后重试"})
		return
	}
	ctx.Next()
}

func readableInput(body []byte, contentType string) string {
	mediaType, _, _ := mime.ParseMediaType(contentType)
	if mediaType == "application/x-www-form-urlencoded" {
		if values, err := url.ParseQuery(string(body)); err == nil {
			decoded, _ := json.MarshalIndent(values, "", "  ")
			return string(decoded)
		}
	}
	var pretty bytes.Buffer
	if json.Indent(&pretty, body, "", "  ") == nil {
		return pretty.String()
	}
	return string(body)
}
