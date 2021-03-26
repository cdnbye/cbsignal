package main

import (
	"strconv"
	"strings"
	"testing"
)

func TestUtilizationRate(t *testing.T)  {

	var a int64 = 1
	var b int64 = 3000
	var c float32 = float32(a)/float32(b)
	t.Logf("%f", c)
}

func TestVersion(t *testing.T) {
	ver := "0.1.1"
	digs := strings.Split(ver, ".")
	a , _ := strconv.Atoi(digs[0])
	b , _ := strconv.Atoi(digs[1])
	t.Logf("%d", a*10 + b)
}




