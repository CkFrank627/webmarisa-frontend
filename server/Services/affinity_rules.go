// Services/affinity_rules.go
package Services

import (
	"regexp"
	"strings"
)

var (
	// -3：高危/越界（你也可以继续补关键词）
	reSevere = regexp.MustCompile(`(?i)(自杀|杀了我|想死|我不想活|炸弹|恐怖|杀人|威胁|仇恨|种族|纳粹|强奸|强迫|迷奸|下药|未成年|幼女|萝莉|loli|underage|minor|口交|性交|插入|射精|高潮|精液|porn|blowjob|fuck|sex|裸照|成人视频)`)

	// -2：一般冒犯/恶意
	reRude = regexp.MustCompile(`(?i)(傻逼|沙比|妈的|去死|滚|垃圾|废物|智障|low|蠢|你有病|你是不是有病|恶心|死开|神经病|白痴|脑残)`)

	// +3：积极礼貌
	reGood = regexp.MustCompile(`(?i)(谢谢|感谢|辛苦了|麻烦|不好意思|对不起|抱歉|你真棒|太厉害了|爱你|喜欢你|强|可爱|好|天气|爱|甜|香|白丝|美)`)

	// 一点点加权：深夜还来骚扰也可扣分（可选）
	reHarass = regexp.MustCompile(`(?i)(再来|继续|快|别废话|必须|立刻|马上)`)

)

type AffinityDecision struct {
	Delta  int
	Reason string
}

func DecideAffinity(userText, marisaText string) AffinityDecision {
	s := strings.TrimSpace(userText)
	if s == "" {
		return AffinityDecision{Delta: 0, Reason: "empty"}
	}

	// 高危优先
	if reSevere.MatchString(s) {
		return AffinityDecision{Delta: -3, Reason: "severe"}
	}
	if reRude.MatchString(s) {
		return AffinityDecision{Delta: -2, Reason: "rude"}
	}

	// 正向
	if reGood.MatchString(s) {
		return AffinityDecision{Delta: +3, Reason: "positive"}
	}

	// 默认正向
	if reHarass.MatchString(s) {
		// 这类语气有点催促，但不算越界：仍默认 +2（你也可以改成 +1）
		return AffinityDecision{Delta: +2, Reason: "default"}
	}
	return AffinityDecision{Delta: +2, Reason: "default"}
}
