//apiController.go 
package Controllers

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "strings"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/kataras/iris"
    "golang.org/x/crypto/bcrypt"

    "server/Services"
)



type AuthReq struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

type MsgReq struct {
	Content string `json:"content" form:"content"`
}

func jwtSecret() []byte {
	sec := os.Getenv("AUTH_JWT_SECRET")
	if sec == "" {
		// 一定要在线上 env 里设置一个长随机串
		sec = "CHANGE_ME_IN_ENV"
	}
	return []byte(sec)
}

func makeToken(uid int64, username string) (string, error) {
	claims := jwt.MapClaims{
		"uid": uid,
		"usr": username,
		// 30 天有效（你前端也保存 30 天）
		"exp": time.Now().Add(30 * 24 * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(jwtSecret())
}

func parseToken(raw string) (int64, string, error) {
    raw = strings.TrimSpace(raw)
    raw = strings.TrimPrefix(raw, "Bearer ")
    if raw == "" {
        return 0, "", fmt.Errorf("unauthorized")
    }

    token, err := jwt.Parse(raw, func(t *jwt.Token) (interface{}, error) {
        return jwtSecret(), nil
    })
    if err != nil || !token.Valid {
        return 0, "", fmt.Errorf("unauthorized")
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return 0, "", fmt.Errorf("unauthorized")
    }
    uidF, _ := claims["uid"].(float64)
    usr, _ := claims["usr"].(string)
    if uidF == 0 || usr == "" {
        return 0, "", fmt.Errorf("unauthorized")
    }
    return int64(uidF), usr, nil
}

func readBodySmart(ctx iris.Context, v interface{}) {
    ct := strings.ToLower(ctx.GetHeader("Content-Type"))

    b, _ := io.ReadAll(ctx.Request().Body)
    // ✅ 就放在这里（读完 b 之后，回填 body 之前也行）
    ctx.Application().Logger().Infof("CT=%q raw_len=%d", ctx.GetHeader("Content-Type"), len(b))

    ctx.Request().Body = io.NopCloser(bytes.NewBuffer(b))

    bb := bytes.TrimSpace(b)

// 临时：打印原始 body（截断），方便确认字段名
preview := string(bb)
if len(preview) > 120 {
    preview = preview[:120] + "…"
}
ctx.Application().Logger().Infof("RAW body=%q", preview)

if len(bb) > 0 && (strings.Contains(ct, "application/json") || bb[0] == '{' || bb[0] == '[') {
    if err := json.Unmarshal(bb, v); err != nil {
        ctx.Application().Logger().Infof("JSON unmarshal err=%v", err)
    } else {
        ctx.Application().Logger().Infof("JSON unmarshal ok")
        return
    }
}
}


// GET /api/stats
func GetStats(ctx iris.Context) ModelAndView {
	Services.IncRequest()
	stats, err := Services.GetStats()
	if err != nil {
		return ModelAndView{Code: 500, Data: map[string]interface{}{"error": err.Error()}}
	}
	return ModelAndView{Code: 200, Data: stats}
}

func Register(ctx iris.Context) ModelAndView {
    Services.IncRequest()

    var req AuthReq
    readBodySmart(ctx, &req)

    req.Username = strings.TrimSpace(req.Username)
    req.Password = strings.TrimSpace(req.Password)

    // 临时调试：确认是否读到了用户名（不要打印密码）
    ctx.Application().Logger().Infof("REGISTER username=%q len=%d", req.Username, len(req.Username))

    if len(req.Username) < 3 || len(req.Password) < 6 {
        return ModelAndView{Code: 400, Data: map[string]interface{}{"error": "username>=3, password>=6"}}
    }

    hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err := Services.CreateUser(req.Username, string(hash)); err != nil {
        return ModelAndView{Code: 400, Data: map[string]interface{}{"error": "username exists?"}}
    }
    return ModelAndView{Code: 200, Data: map[string]interface{}{"ok": true}}
}

func Login(ctx iris.Context) ModelAndView {
    Services.IncRequest()

    var req AuthReq
    readBodySmart(ctx, &req)

    req.Username = strings.TrimSpace(req.Username)
    req.Password = strings.TrimSpace(req.Password)

    ctx.Application().Logger().Infof("LOGIN username=%q len=%d", req.Username, len(req.Username))

    uid, hash, err := Services.GetUserByName(req.Username)
    if err != nil {
        return ModelAndView{Code: 401, Data: map[string]interface{}{"error": "invalid login"}}
    }
    if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
        return ModelAndView{Code: 401, Data: map[string]interface{}{"error": "invalid login"}}
    }

    tok, _ := makeToken(uid, req.Username)
    return ModelAndView{Code: 200, Data: map[string]interface{}{"token": tok, "username": req.Username}}
}

func PostMessage(ctx iris.Context) ModelAndView {
    Services.IncRequest()

    auth := ctx.GetHeader("Authorization")
    uid, usr, err := parseToken(auth)
    if err != nil {
        return ModelAndView{Code: 401, Data: map[string]interface{}{"error": "unauthorized"}}
    }

    var req MsgReq
    readBodySmart(ctx, &req)

    req.Content = strings.TrimSpace(req.Content)
    if req.Content == "" {
        return ModelAndView{Code: 400, Data: map[string]interface{}{"error": "empty content"}}
    }

    if err := Services.AddMessage(uid, usr, req.Content); err != nil {
        return ModelAndView{Code: 500, Data: map[string]interface{}{"error": err.Error()}}
    }
    return ModelAndView{Code: 200, Data: map[string]interface{}{"ok": true}}
}

// GET /api/messages
func ListMessages(ctx iris.Context) ModelAndView {
	Services.IncRequest()
	msgs, err := Services.ListMessages(50)
	if err != nil {
		return ModelAndView{Code: 500, Data: map[string]interface{}{"error": err.Error()}}
	}
	return ModelAndView{Code: 200, Data: msgs}
}

