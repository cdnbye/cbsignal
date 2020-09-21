package main

import (
	"testing"
)

func TestUtilizationRate(t *testing.T)  {

	var a int64 = 1
	var b int64 = 3000
	var c float32 = float32(a)/float32(b)
	t.Logf("%f", c)
}




