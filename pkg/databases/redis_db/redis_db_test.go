package redis_db

import (
	"fmt"
	"testing"
)

func TestPrintTable(t *testing.T) {
	var tests = []struct {
		raw  string
		want bool
	}{
		{"a", true},
		{"d2", false},
		{"c", true},
	}

	myHash := NewStringHash()
	myHash.Add("a")
	myHash.Add("b")
	myHash.Add("c")

	//first, test only the table output
	for _, tt := range tests {
		testname := fmt.Sprintf("%v", tt.raw)
		t.Run(testname, func(t *testing.T) {
			res := myHash.exists(tt.raw)
			if res != tt.want {
				t.Errorf("\ngot\n %v\n\nwant\n %v", res, tt.want)
			}
		})
	}
}
