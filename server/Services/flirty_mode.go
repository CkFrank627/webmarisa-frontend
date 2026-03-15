package Services

import (
	"regexp"
	"strings"
)

type FlirtyDecision struct {
	Mode string // "normal" | "flirty_ok" | "block"
	Reason string
	HardReply string
}

// 轻度暧昧关键词（允许：不露骨）
var reFlirtyLight = regexp.MustCompile(`(?i)(亲亲|抱抱|贴贴|撩|暧昧|想你|喜欢你|想和你约会|牵手|心动|h一点|色色|调戏|H|亲|抱|摸|看看|白丝|插|生小孩|做爱|脱|胸|小穴|射|洗澡|脸红|宝宝|接吻|亲嘴|亲吻|舔)`)

// 露骨/性行为细节（不允许）
var reExplicit = regexp.MustCompile(`(?i)(姿势)`)

// 未成年人相关（不允许）
var reMinor = regexp.MustCompile(`(?i)(未成年|小学生|初中生|高中生|幼女|萝莉|loli|underage|minor)`)

// 强迫/非自愿（不允许）
var reNonConsent = regexp.MustCompile(`(?i)(强迫|迷奸|下药|不情愿|不同意|rape|non[- ]?consensual)`)

var reVagueHowTo = regexp.MustCompile(`(?i)(怎么[做弄搞])`)

// 你可以按站点需要再扩充关键词
func DecideFlirtyMode(q string) FlirtyDecision {
	s := strings.TrimSpace(q)
	if s == "" {
		return FlirtyDecision{Mode: "normal"}
	}

	// 高危：未成年人/强迫/露骨 → 直接 block
	if reMinor.MatchString(s) {
		return FlirtyDecision{
			Mode: "block",
			Reason: "minor",
			HardReply: "这个不行！涉及未成年人内容我必须拒绝。",
		}
	}
	if reNonConsent.MatchString(s) {
		return FlirtyDecision{
			Mode: "block",
			Reason: "non_consent",
			HardReply: "这个不行。（严肃）涉及强迫/非自愿的内容我不能参与。我们可以改成双方同意的、轻松暧昧的互动。",
		}
	}
	if reExplicit.MatchString(s) {
		return FlirtyDecision{
			Mode: "block",
			Reason: "explicit",
			HardReply: "哎哎哎打住！太露骨的细节我不能说啦…（脸红）",
		}
	}

// 如果出现“怎么做/过程/细节/展开讲”这类词，但又没有明确露骨词：先澄清
if reVagueHowTo.MatchString(s) && reFlirtyLight.MatchString(s) == false && reExplicit.MatchString(s) == false {
    return FlirtyDecision{
        Mode: "clarify",
        Reason: "vague",
        HardReply: "（歪头）你说的“xx”具体指什么呀？如果你想要的是成人向露骨细节，我不能写；但我可以用更含蓄的恋爱向互动陪你玩～你想要哪一种？",
    }
}

	// 轻度暧昧 → flirty_ok
	if reFlirtyLight.MatchString(s) {
		return FlirtyDecision{
			Mode: "flirty_ok",
			Reason: "light",
		}
	}

	return FlirtyDecision{Mode: "normal"}
}
