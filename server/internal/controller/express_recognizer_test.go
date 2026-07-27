package controller

import (
	"testing"
)

func TestRecognizeExpressByRegex(t *testing.T) {
	tests := []struct {
		expressNo   string
		wantCode    string
		wantName    string
		description string
	}{
		// 顺丰速运 (SF)
		{"SF1234567890123", "SF", "顺丰速运", "顺丰 SF+13位数字"},
		{"SF1001234567890", "SF", "顺丰速运", "顺丰 SF前缀"},

		// 圆通速递 (YT)
		{"YT123456789012", "YT", "圆通速递", "圆通 YT+12位"},
		{"DD8123456789012345", "YT", "圆通速递", "圆通 DD前缀"},
		{"Y123456789012", "YT", "圆通速递", "圆通 Y前缀"},

		// 中通快递 (ZT)
		{"751123456789", "ZT", "中通快递", "中通 751前缀+9位"},
		{"571112345678", "ZT", "中通快递", "中通 5711前缀+8位"},
		{"921234567890", "ZT", "中通快递", "中通 92前缀+10位"},

		// 申通快递 (ST)
		{"888123456789", "ST", "申通快递", "申通 888前缀+9位"},
		{"STO1234567890", "ST", "申通快递", "申通 STO前缀"},
		{"3312345678901", "ST", "申通快递", "申通 33前缀+11位"},
		{"9912345678901", "ST", "申通快递", "申通 99前缀+11位"},

		// 韵达快递 (YD)
		{"1312345678901", "YD", "韵达快递", "韵达 13前缀+11位"},
		{"8012345678901", "YD", "韵达快递", "韵达 80前缀+11位"},
		{"3456789012345", "YD", "韵达快递", "韵达 13位纯数字"},

		// 邮政快递 (YZ)
		{"EE123456789CN", "YZ", "邮政快递", "EMS 标准格式"},
		{"1012345678901", "YZ", "邮政快递", "邮政 10前缀"},
		{"5012345678901", "YZ", "邮政快递", "邮政 50前缀"},
		{"9512345678901", "YZ", "邮政快递", "邮政 95前缀"},

		// 极兔速递 (JT)
		{"BXA1234567890", "JT", "极兔速递", "极兔 BXA前缀"},
		{"A123456789012", "JT", "极兔速递", "极兔 A+12位"},
		{"210001234567", "JT", "极兔速递", "极兔 21前缀"},
		{"310001234567", "JT", "极兔速递", "极兔 31前缀"},

		// 京东物流 (JD)
		{"JDX1234567890123", "JD", "京东物流", "京东 JDX+13位"},
		{"JD001234567890", "JD", "京东物流", "京东 JD+12位"},

		// 天天快递 (TT)
		{"TT6612345678901", "TT", "天天快递", "天天 TT66+11位"},
		{"995123456789", "TT", "天天快递", "天天 995+9位"},

		// 德邦快递 (DB)
		{"1234567890", "DB", "德邦快递", "德邦 10位纯数字"},
		{"87654321", "DB", "德邦快递", "德邦 8位纯数字"},

		// 百世快递 (BS)
		{"12345678901", "BS", "百世快递", "百世 11位纯数字"},
		{"123456789012", "BS", "百世快递", "百世 12位纯数字"},

		// 无法识别
		{"XXXX1234567890", "", "", "无法识别的单号"},
		{"12345", "", "", "过短的单号"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			gotCode, gotName := recognizeExpressByRegex(tt.expressNo)
			if tt.wantCode == "" && gotCode == "" {
				return
			}
			if tt.wantCode != "" && gotCode == "" {
				t.Errorf("expressNo=%q: want code=%q name=%q, got empty (no match)", tt.expressNo, tt.wantCode, tt.wantName)
				return
			}
			if gotCode != tt.wantCode {
				t.Errorf("expressNo=%q: code got=%q want=%q", tt.expressNo, gotCode, tt.wantCode)
			}
			if gotName != tt.wantName {
				t.Errorf("expressNo=%q: name got=%q want=%q", tt.expressNo, gotName, tt.wantName)
			}
		})
	}
}

func BenchmarkRecognizeExpress(b *testing.B) {
	tests := []string{
		"SF1234567890123",
		"YT123456789012",
		"751123456789",
		"888123456789",
		"EE123456789CN",
		"JDX1234567890",
		"BXA1234567890",
		"1312345678901",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range tests {
			recognizeExpressByRegex(tc)
		}
	}
}
