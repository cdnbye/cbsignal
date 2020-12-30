package util

import (
	"net/url"
	"strings"
)

// 获取域名（不包含端口）
func GetDomain(uri string) string {
	parsed, err := url.Parse(uri)
	if err != nil {
		return ""
	}
	a := strings.Split(parsed.Host, ":")
	return a[0]
}
