package main

import (
	"strconv"
	"strings"
	"testing"
	"time"
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

func TestTime(t *testing.T)  {
	s := 2 * time.Second.Nanoseconds()
	start := time.Now()
	time.Sleep(300*time.Microsecond)
	t.Logf("%d %d", time.Since(start).Nanoseconds(), s)
}

func TestFnv32(t *testing.T) {
	key := "ddffvfgfgf"
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	t.Log(hash)
}




