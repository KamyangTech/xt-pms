// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import "testing"

func TestSanitizeString(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"移除 script 标签", "<script>alert(1)</script>", "alert(1)"},
		{"移除普通标签", "<b>hello</b>", "hello"},
		{"转义特殊字符", "a & b \"c\" 'd'", "a &amp; b &quot;c&quot; &#39;d&#39;"},
		{"去除首尾空白", "  trimmed  ", "trimmed"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := SanitizeString(c.in); got != c.want {
				t.Fatalf("SanitizeString(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}
