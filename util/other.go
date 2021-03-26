package util

import (
	"net/url"
	"strconv"
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

func GetVersionNum(ver string) int {
	digs := strings.Split(ver, ".")
	a , _ := strconv.Atoi(digs[0])
	b , _ := strconv.Atoi(digs[1])
	return a*10 + b
}
