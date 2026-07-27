package controller

import (
	"regexp"
	"sort"
)

type expressRule struct {
	companyCode string         // 字典值
	companyName string         // 显示名称
	regex       *regexp.Regexp // 编译后的正则
	priority    int            // 优先级，数字越大越优先匹配
}

// expressRules 基于淘宝开放平台（2024.09.27更新）的快递单号正则规则
// 已剔除过于宽泛的规则，避免误匹配
// companyCode 使用大写，与数据库 sys_dict_data.dict_value 一致
var expressRules []expressRule

func init() {
	rules := []struct {
		code     string
		name     string
		pattern  string
		priority int
	}{
		// ============================================================
		// Tier 1: 字母前缀特征，可唯一识别快递公司（最高优先级）
		// ============================================================

		// 顺丰速运 (SF) - SF 前缀
		{code: "SF", name: "顺丰速运", pattern: `^SF[0-9]{12,15}$`, priority: 1100},
		{code: "SF", name: "顺丰速运", pattern: `^SF[A-Za-z0-9-]{5,18}$`, priority: 1099},

		// 申通快递 (ST) - STO 前缀
		{code: "ST", name: "申通快递", pattern: `^STO[0-9]{10}$`, priority: 1100},

		// 京东物流 (JD) - JD 前缀
		{code: "JD", name: "京东物流", pattern: `^JD[VAX][0-9]{9,13}$`, priority: 1100},
		{code: "JD", name: "京东物流", pattern: `^JD[0-9]{10,14}$`, priority: 1099},

		// 速尔快运 - SUR 前缀
		{code: "SURE", name: "速尔快运", pattern: `^SUR[0-9]{12}$`, priority: 1100},

		// 邮政/EMS (YZ) - 标准 EMS 格式：2字母 + 9数字 + 2字母
		{code: "YZ", name: "邮政快递", pattern: `^[A-Z]{2}[0-9]{9}[A-Z]{2}$`, priority: 1100},
		// 邮政快递包裹 - 特定字母前缀
		{code: "YZ", name: "邮政快递", pattern: `^([GA]|[KQ]|[PH]){2}[0-9]{9}([2-5][0-9]|[1][1-9]|[6][0-5])$`, priority: 1099},

		// 圆通速递 (YT) - YT/DD/Y 前缀
		{code: "YT", name: "圆通速递", pattern: `^YT[0-9]{8,13}$`, priority: 1100},
		{code: "YT", name: "圆通速递", pattern: `^DD[8-9][0-9]{15}$`, priority: 1099},
		{code: "YT", name: "圆通速递", pattern: `^Y[0-9]{12}$`, priority: 1098},

		// 天天快递 (TT) - TT 前缀
		{code: "TT", name: "天天快递", pattern: `^TT(66|88|99|(5(6|7)))[0-9]{11}$`, priority: 1100},

		// 极兔速递 (JT) - 字母+数字特定格式
		{code: "JT", name: "极兔速递", pattern: `^BXA[0-9]{10}$`, priority: 1100},
		{code: "JT", name: "极兔速递", pattern: `^K8[0-9]{11}$`, priority: 1099},
		{code: "JT", name: "极兔速递", pattern: `^C0000[0-9]{8}$`, priority: 1098},
		{code: "JT", name: "极兔速递", pattern: `^A[0-9]{12}$`, priority: 1097},
		{code: "JT", name: "极兔速递", pattern: `^[BDE][0-9]{12}$`, priority: 1096},

		// 快捷快递 - 字母前缀
		{code: "FAST", name: "快捷快递", pattern: `^P330[0-9]{8}$`, priority: 1095},
		{code: "FAST", name: "快捷快递", pattern: `^D[0-9]{11}$`, priority: 1094},

		// ============================================================
		// Tier 2: 明确的数字前缀特征（高优先级）
		// ============================================================

		// 中通快递 (ZT) - 特定3位数字前缀
		{
			code: "ZT", name: "中通快递",
			pattern: `^((010|768|765|778|779|719|828|618|680|518|688|880|660|805|988|628|205|717|718|728|738|761|762|763|701|757|751|359|358|100|200|118|128|689|738|528|852)[0-9]{9})$`,
			priority: 1050,
		},
		// 中通快递 - 特定4位数字前缀
		{
			code: "ZT", name: "中通快递",
			pattern: `^((5711|2008|2009|2010|2013)[0-9]{8})$`,
			priority: 1049,
		},
		// 中通快递 - 特定2位数字前缀（避开和其他快递冲突的50/53）
		{
			code: "ZT", name: "中通快递",
			pattern: `^((91|92|93|94|95|98|36|68|39|37)[0-9]{10})$`,
			priority: 1048,
		},
		// 中通快递 - 其他优化前缀
		{code: "ZT", name: "中通快递", pattern: `^(90)[0-9]{10}$`, priority: 1047},
		{code: "ZT", name: "中通快递", pattern: `^(120)[0-9]{9}$`, priority: 1046},
		{code: "ZT", name: "中通快递", pattern: `^(780)[0-9]{9}$`, priority: 1045},
		{code: "ZT", name: "中通快递", pattern: `^(881)[0-9]{9}$`, priority: 1044},
		{code: "ZT", name: "中通快递", pattern: `^(882|885)[0-9]{9}$`, priority: 1043},
		{code: "ZT", name: "中通快递", pattern: `^(54|55|56)[0-9]{10}$`, priority: 1042},
		{code: "ZT", name: "中通快递", pattern: `^(960)[0-9]{9}$`, priority: 1041},
		{code: "ZT", name: "中通快递", pattern: `^(665|666)[0-9]{9}$`, priority: 1040},
		{code: "ZT", name: "中通快递", pattern: `^(63)[0-9]{10}$`, priority: 1039},
		{code: "ZT", name: "中通快递", pattern: `^(64)[0-9]{10}$`, priority: 1038},
		{code: "ZT", name: "中通快递", pattern: `^(72)[0-9]{10}$`, priority: 1037},

		// 中通快递 - 2[1-9]前缀（避开21-39给极兔，但21-29会冲突）
		// 极兔的21-39有更高优先级，中通的2[1-9]作为兜底
		{code: "ZT", name: "中通快递", pattern: `^2[1-9][0-9]{10}$`, priority: 600},

		// 申通快递 (ST) - 特定3位数字前缀
		{code: "ST", name: "申通快递", pattern: `^(268|888|588|688|368|468|568|668|768|868|968)[0-9]{9}$`, priority: 1000},
		// 申通快递 - 特定2位数字前缀（更高精度）
		{code: "ST", name: "申通快递", pattern: `^(33)[0-9]{11}$`, priority: 999},
		{code: "ST", name: "申通快递", pattern: `^(99)[0-9]{11}$`, priority: 998},
		// 申通快递 - 带可选前缀的10位格式
		{code: "ST", name: "申通快递", pattern: `^(11|22|40|268|888|588|688|368|468|568|668|768|868|968)[0-9]{10}$`, priority: 997},

		// 天天快递 (TT) - 纯数字格式
		{code: "TT", name: "天天快递", pattern: `^(995|998)[0-9]{9}$`, priority: 900},

		// 极兔速递 (JT) - 21-39, 61, 63前缀（优先于中通的2[1-9]）
		{code: "JT", name: "极兔速递", pattern: `^((21|22|23|24|25|26|27|28|29|30|31|32|33|34|35|36|37|38|39|61|63)[0-9]{10})$`, priority: 850},
		{code: "JT", name: "极兔速递", pattern: `^(02[0-9]{11})$`, priority: 700},
		{code: "JT", name: "极兔速递", pattern: `^(000[0-9]{10})$`, priority: 699},
		{code: "JT", name: "极兔速递", pattern: `^7[0-9]{13}$`, priority: 650},
		{code: "JT", name: "极兔速递", pattern: `^6[0-9]{13}$`, priority: 649},
		{code: "JT", name: "极兔速递", pattern: `^58[0-9]{14}$`, priority: 648},

		// 快捷快递 - 数字前缀
		{code: "FAST", name: "快捷快递", pattern: `^(319)[0-9]{11}$`, priority: 500},
		{code: "FAST", name: "快捷快递", pattern: `^(536)[0-9]{9}$`, priority: 499},
		{code: "FAST", name: "快捷快递", pattern: `^(56)[0-9]{10}$`, priority: 498},

		// ============================================================
		// Tier 3: 可能存在重叠的数字特征（中优先级）
		// ============================================================

		// 韵达快递 (YD) - 特定2位前缀
		{code: "YD", name: "韵达快递", pattern: `^(13|14|15|16|17|19|18|58|80|31|39)[0-9]{11}$`, priority: 800},
		// 韵达 - 以下前缀与其他公司有重叠，降低优先级
		{code: "YD", name: "韵达快递", pattern: `^(88)[0-9]{11}$`, priority: 400},

		// 邮政 (YZ) - 与韵达共享部分前缀
		{code: "YZ", name: "邮政快递", pattern: `^(95|97)[0-9]{11}$`, priority: 750},
		{code: "YZ", name: "邮政快递", pattern: `^(96)[0-9]{11}$`, priority: 749},
		{code: "YZ", name: "邮政快递", pattern: `^(98)[0-9]{11}$`, priority: 748},
		{code: "YZ", name: "邮政快递", pattern: `^(99)[0-9]{11}$`, priority: 747},

		// 邮政/EMS - 数字前缀（注意：与韵达有10/11/12前缀重叠，邮政优先级更高）
		{code: "YZ", name: "邮政快递", pattern: `^(10|11|12)[0-9]{11}$`, priority: 900},
		{code: "YZ", name: "邮政快递", pattern: `^(50|51)[0-9]{11}$`, priority: 899},

		// 极兔 - 50/51前缀（与韵达、邮政共存）
		{code: "JT", name: "极兔速递", pattern: `^((50|51)[0-9]{12})$`, priority: 700},

		// ============================================================
		// Tier 4: 通用长度型特征（低优先级，作为兜底）
		// ============================================================

		// 德邦快递 (DB) - 8-10位纯数字
		{code: "DB", name: "德邦快递", pattern: `^[0-9]{8,10}$`, priority: 200},

		// 百世快递 (BS) - 11-12位纯数字
		{code: "BS", name: "百世快递", pattern: `^[0-9]{11,12}$`, priority: 201},

		// 韵达 - 13位纯数字（兜底）
		{code: "YD", name: "韵达快递", pattern: `^[0-9]{13}$`, priority: 202},

		// 宅急送 - 通用模式
		{code: "ZJS", name: "宅急送", pattern: `^(42|16)[0-9]{8}$`, priority: 195},
		{code: "ZJS", name: "宅急送", pattern: `^A[0-9]{12}$`, priority: 194},

		// 联邦快递
		{code: "FEDEX", name: "联邦快递", pattern: `^[0-9]{12}$`, priority: 55},

		// 速尔快运 - 12位通用兜底
		{code: "SURE", name: "速尔快运", pattern: `^[0-9]{12}$`, priority: 54},
	}

	for _, r := range rules {
		expressRules = append(expressRules, expressRule{
			companyCode: r.code,
			companyName: r.name,
			regex:       regexp.MustCompile(r.pattern),
			priority:    r.priority,
		})
	}

	// 按优先级降序排列
	sort.Slice(expressRules, func(i, j int) bool {
		return expressRules[i].priority > expressRules[j].priority
	})
}

func recognizeExpressByRegex(expressNo string) (companyCode, companyName string) {
	// 快递单号总长度不应超过30位（国内主流快递单号最长为25位左右）
	if len(expressNo) > 30 {
		return "", ""
	}
	for _, rule := range expressRules {
		if rule.regex.MatchString(expressNo) {
			return rule.companyCode, rule.companyName
		}
	}
	return "", ""
}
