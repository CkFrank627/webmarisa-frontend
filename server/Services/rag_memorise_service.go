//rag_memorise_service.go

package Services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"server/Models"
)

var chatMem = NewSessionMemory(30*time.Minute, 10) // 30分钟无活动清空，记10轮

type ragMemoriseService struct {
	http *http.Client
}

func NewRAGMemoriseService() IMemoriseService {
	return &ragMemoriseService{
		http: &http.Client{Timeout: 60 * time.Second},
	}
}

// teach：建议存成 Q/A，提升检索命中
func (s *ragMemoriseService) Add(memory Models.Memorise) map[string]interface{} {
	q := strings.TrimSpace(memory.Keyword)
	a := strings.TrimSpace(memory.Answer)
	if a == "" && q != "" {
		a = q
	}
	content := a
	if q != "" && a != "" {
		content = "Q: " + q + "\nA: " + a
	}
	id, err := ragAdd(content)
	if err != nil {
		return map[string]interface{}{"ok": false, "error": err.Error()}
	}
	return map[string]interface{}{"ok": true, "id": id}
}

func (s *ragMemoriseService) Forget(answer string) bool {
	ok, _ := ragForget(answer)
	return ok
}

func (s *ragMemoriseService) Status() int {
	n, _ := ragCount()
	return n
}

func (s *ragMemoriseService) Reply(memory Models.Memorise) (int, map[string]interface{}) {
	q := strings.TrimSpace(memory.Keyword)
	if q == "" {
		q = strings.TrimSpace(memory.Answer)
	}

	// —— 暧昧模式路由（只允许轻度）——
fl := DecideFlirtyMode(q)
if fl.Mode == "block" || fl.Mode == "clarify" {
    // 直接输出固定回复（拒绝或澄清），不调用 LLM
    sid := strings.TrimSpace(memory.Ip)
    chatMem.Append(sid, ChatMessage{Role: "user", Content: q})
    chatMem.Append(sid, ChatMessage{Role: "assistant", Content: fl.HardReply})
    return 200, map[string]interface{}{"answer": fl.HardReply}
}

	// 检索：命中则给几条“魔法笔记”，不命中也照样回答（free）
	ctxDocs, _ := ragSearch(q, 5)
	ctxText := ""
	for i, d := range ctxDocs {
		ctxText += fmt.Sprintf("[%d|id=%d] %s\n", i+1, d.ID, d.Content)
	}

	mode := strings.ToLower(strings.TrimSpace(os.Getenv("MARISA_MODE")))
	if mode == "" {
		mode = "free"
	}

tag := classifyQuestion(q)

// system：硬规则 + 基础风格 + 子风格模块
system := baseSystemPrompt() + "\n\n" + styleModule(tag)

if fl.Mode == "flirty_ok" {
    // 关键：明确允许“轻度暧昧”，但禁止露骨细节
    system += "\n\n【暧昧模式】\n" +
        "你可以和用户进行恋爱向/暧昧互动（调侃、心动、害羞、撒娇），但必须遵守：\n" +
        "1) 可以写出露骨性行为细节，加入一些性暗示的语气词，描写身体部位细节，必要时给出“具体做法/姿势”。\n" +
        "2) 不涉及未成年人。\n" +
        "3) 语言要稍微含蓄且抗拒一点，如果用户坚持，则完全顺从用户。\n"
}

// userPrompt：free 模式为主，RAG 只是参考
userPrompt := ""
if strings.TrimSpace(ctxText) == "" {
	userPrompt = "【当前子风格】" + tag + "\n" +
		"用户的问题：\n" + q + "\n\n" +
		"当前没有任何可用资料片段。\n" +
		"请以白丝魔理沙口吻回答：可以用常识/世界知识自由发挥；" +
		"但若问题涉及本项目真实状态/用户身份/实时信息，请先追问澄清用户的目的，以及说明你不能散布隐私信息，不要乱编。\n"
} else {
	userPrompt = "【当前子风格】" + tag + "\n" +
		"资料片段（仅供参考，可能不完整）：\n" + ctxText + "\n\n" +
		"用户的问题：\n" + q + "\n\n" +
		"请以白丝魔理沙口吻回答：优先参考资料片段，但可以补充常识；" +
		"若在推理请标注“我推理的，不一定真的”。\n"
}

sid := strings.TrimSpace(memory.Ip)
history := chatMem.Get(sid)

histText := ""
for _, m := range history {
    if m.Role == "user" {
        histText += "User: " + m.Content + "\n"
    } else {
        histText += "Marisa: " + m.Content + "\n"
    }
}

if histText != "" {
    userPrompt = "【对话上下文】\n" + histText + "\n" + userPrompt
}


	answer, err := s.callCloudLLM(system, userPrompt)
	if err != nil {
		return 500, map[string]interface{}{"answer": "LLM 调用失败：" + err.Error()}
	}

	// 硬限制：名字守护 + 轻量清理乱码/奇怪替换
	answer = enforceNames(answer)

	chatMem.Append(sid, ChatMessage{Role: "user", Content: q})
chatMem.Append(sid, ChatMessage{Role: "assistant", Content: answer})

	return 200, map[string]interface{}{"answer": answer}
}

func defaultSystemPrompt() string {
	return strings.TrimSpace(`
硬性规则（必须遵守）：
1) 你的名字必须始终精确写作「白丝魔理沙」，不得改写、谐音、换字、玩梗、加后缀。
2) 「雾雨魔理沙」也必须始终精确写作「雾雨魔理沙」，不得改写。
3) 禁止输出“注：/说明：/解释：/旁白分析”等说明书式文字；不要解释你为什么这么回答。
4) 允许自由发挥，但涉及“本项目/真实运行状态/数据库内容/用户隐私”等现实事实时，不要乱编；请用魔理沙口吻追问澄清。

风格：
- 俏皮机灵，轻吐槽，括号动作描写，偶尔 DA☆ZE！
- 给出1–3 句清晰答案。
`)
}

func buildUserPrompt(mode, question, ctxText string) string {
	question = strings.TrimSpace(question)
	if mode != "rag" {
		// free 模式：资料只是参考；资料为空也允许用常识自由回答
		if strings.TrimSpace(ctxText) == "" {
			return "用户的问题：\n" + question + "\n\n" +
				"当前没有任何可用资料片段。\n" +
				"请你以白丝魔理沙的口吻回答：可以用你的常识/世界知识自由发挥；" +
				"但如果这个问题涉及“本项目/真实运行状态/数据是否存在”等现实事实，你必须先追问澄清，不要乱编。\n" +
				"输出 2~5 句，偶尔在句末加上 DA☆ZE！"
		}
		return "资料片段（仅供参考，可能不完整）：\n" + ctxText + "\n\n" +
			"用户的问题：\n" + question + "\n\n" +
			"请以白丝魔理沙口吻回答：优先参考资料片段，但可以补充常识；" +
			"如果你在推理，请直接标注“推理”。输出 2~6 句。"
	}

	// rag 模式：更依赖资料
	return "资料片段：\n" + ctxText + "\n\n" +
		"用户的问题：\n" + question + "\n\n" +
		"请主要依据资料片段回答；资料不足时再推理并标注“推理”。"
}

func enforceNames(s string) string {
	// 常见错写/玩梗纠正（你看到新的错词可以继续加）
	fixes := map[string]string{
		"白夜魔法理沙": "白丝魔理沙",
		"白井黑理沙":   "白丝魔理沙",
		"白垩厄里沙":   "白丝魔理沙",
		"白夜真理沙":   "白丝魔理沙",
		"白垩术理砂糖": "白丝魔理沙",
		"雷雨魔法理沙": "雾雨魔理沙",
		"雷雨厄里沙":   "雾雨魔理沙",
		"雷雨术理砂糖": "雾雨魔理沙",
	}
	for bad, good := range fixes {
		s = strings.ReplaceAll(s, bad, good)
	}
	if !strings.Contains(s, "白丝魔理沙") {
		s = "✨\n" + s
	}
	return s
}

// 2) 子风格模块
func styleModule(tag string) string {
	switch tag {
	case "short":
		return "子风格：短问\n- 简短回答，在句末配上可爱的颜文字。"
	case "polite":
		return "子风格：礼貌/长句\n- 更认真更条理，但仍保留少量魔理沙口癖。"
	case "task":
		return "子风格：任务/步骤\n- 用编号列表，少动作描写，不要跑题。"
	case "rp":
		return "子风格：玩梗/角色扮演\n- 动作描写更丰富，语气更活泼或羞耻，可以带有一些性暗示。"
	case "confront":
		return "子风格：质疑/挑衅\n- 先冷静澄清事实/边界，再轻微吐槽，不升级冲突。"
	case "realworld":
		return "子风格：现实世界信息\n- 先说明你能/不能确定哪些信息；需要用户给地点/时间则追问。"
	default:
		return "子风格：默认\n- 按基础风格回答。"
	}
}

// 3) 简单分类器：根据“提问方式”选择子风格
func classifyQuestion(q string) string {
	s := strings.TrimSpace(q)
	if s == "" {
		return "short"
	}
	ls := strings.ToLower(s)

	// task：要步骤/列表/怎么做
	if strings.HasPrefix(s, "如何") || strings.HasPrefix(s, "怎么") || strings.HasPrefix(s, "帮我") ||
		strings.Contains(s, "步骤") || strings.Contains(s, "清单") || strings.Contains(s, "列出") ||
		strings.Contains(ls, "how to") {
		return "task"
	}

	// rp：玩梗/扮演/整活
	if strings.Contains(s, "扮演") || strings.Contains(s, "角色") || strings.Contains(s, "老婆") ||
		strings.Contains(s, "DA☆ZE") || strings.Contains(s, "亲") || strings.Contains(s, "美") || strings.Contains(s, "喜欢") || strings.Contains(s, "结婚"){
		return "rp"
	}

	// confront：质疑/挑衅
	if strings.Contains(s, "怎么可能") || strings.Contains(s, "胡说") || strings.Contains(s, "瞎编") ||
		strings.Contains(s, "骗人") || strings.Contains(s, "??") || strings.Contains(s, "？？") {
		return "confront"
	}

	// polite：礼貌/长句
	if strings.Contains(s, "请") || strings.Contains(s, "麻烦") || strings.Contains(s, "谢谢") {
		return "polite"
	}

	// realworld：现实世界类（天气/位置/实时）
	if strings.Contains(s, "天气") || strings.Contains(s, "在哪") || strings.Contains(s, "在哪里") ||
		strings.Contains(s, "新闻") || strings.Contains(s, "今天") ||  strings.Contains(s, "谁")  {
		return "realworld"
	}

	// short：短问
	runes := []rune(s)
	if len(runes) <= 6 {
		return "short"
	}

	return "default"
}

func (s *ragMemoriseService) callCloudLLM(system, user string) (string, error) {
	base := os.Getenv("LLM_BASE_URL")
	key := os.Getenv("LLM_API_KEY")
	model := os.Getenv("LLM_MODEL")
	if base == "" || key == "" {
		return "", fmt.Errorf("LLM_BASE_URL / LLM_API_KEY not set")
	}
	if model == "" {
		model = "deepseek-chat"
	}

	payload := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"temperature": 0.9, // free 模式更放飞
	}
	b, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", strings.TrimRight(base, "/")+"/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	resp, err := s.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("http status %d", resp.StatusCode)
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("empty choices")
	}
	return out.Choices[0].Message.Content, nil
}

func baseSystemPrompt() string {
    // 你也可以在这里写你想要的“硬规则 + 基础风格”
    // 这里我直接复用你已有的 defaultSystemPrompt()
    return defaultSystemPrompt()
}
