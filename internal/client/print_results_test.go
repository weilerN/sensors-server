package client

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

const (
	LineSeperator = 226 //uint8 of '├'
)

var (
	d1   = ""
	d2   = "sensor_1,Tuesday,30,10,20,"
	d3   = "sensor_1,Sunday,-,-,-,sensor_1,monday,-,-,-,sensor_1,Tuesday,30,10,20,sensor_1,Wednesday,-,-,-,sensor_1,Thursday,-,-,-,sensor_1,Friday,-,-,-,sensor_1,Saturday,-,-,-,1"
	res1 = "┌─────────┬─────┬─────┬─────┬─────┐\n│ #SERIAL  │ DAY       │ MIN │ MAX │ AVG │\n├──────────┼───────────┼─────┼─────┼─────┤\n└──────────┴───────────┴─────┴─────┴─────┘\n"
	res2 = "┌──────────┬───────────┬─────┬─────┬─────┐\n│ #SERIAL  │ DAY       │ MIN │ MAX │ AVG │\n├──────────┼───────────┼─────┼─────┼─────┤\n| sensor_1 | Tuesday | 30  | 10  | 20  |\n└──────────┴───────────┴─────┴─────┴─────┘\n"
	res3 = "┌──────────┬───────────┬─────┬─────┬─────┐\n│ #SERIAL  │ DAY       │ MIN │ MAX │ AVG │\n├──────────┼───────────┼─────┼─────┼─────┤\n│ sensor_1 │ Sunday    │ -   │ -   │ -   │\n├──────────┼───────────┼─────┼─────┼─────┤\n│ sensor_1 │ monday    │ -   │ -   │ -   │\n├──────────┼───────────┼─────┼─────┼─────┤\n│ sensor_1 │ Tuesday   │ 30  │ 10  │ 20  │\n├──────────┼───────────┼─────┼─────┼─────┤\n│ sensor_1 │ Wednesday │ -   │ -   │ -   │\n├──────────┼───────────┼─────┼─────┼─────┤\n│ sensor_1 │ Thursday  │ -   │ -   │ -   │\n├──────────┼───────────┼─────┼─────┼─────┤\n│ sensor_1 │ Friday    │ -   │ -   │ -   │\n├──────────┼───────────┼─────┼─────┼─────┤\n│ sensor_1 │ Saturday  │ -   │ -   │ -   │\n└──────────┴───────────┴─────┴─────┴─────┘\n"
)

func execToString(f func(s string), args string) string {
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f(args)

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		if err != nil {
			panic(err)
		}
		outC <- buf.String()
	}()

	// back to normal state
	if err := w.Close(); err != nil {
		panic(err)
	}
	os.Stdout = old // restoring the real stdout
	out := <-outC
	return out
}

/*
	Test description:
	first:
		sanity test - first test that the function is working as expected.
	then:
		create DB and simulate measures sends to it
		get the res from the DB
		compare the tables
*/
func TestPrintTable(t *testing.T) {
	var tests = []struct {
		raw  string
		want string
	}{
		{d1, res1},
		{d2, res2},
		{d3, res3},
	}

	//first, test only the table output
	for _, tt := range tests {
		testname := fmt.Sprintf("%v", tt.raw)
		t.Run(testname, func(t *testing.T) {

			res := execToString(PrintResult, tt.raw)
			if compareTables(res, tt.want) {
				t.Errorf("\ngot\n %v\n\nwant\n %v", res, tt.want)
			}
		})
	}
}

//return value true -> not identical , fail test
func compareTables(t1, t2 string) bool {
	table1 := strings.Split(t1, "\n")
	table2 := strings.Split(t2, "\n")
	if len(table1) != len(table2) {
		fmt.Println("Not identical len:", len(table1), len(table2))
		return true
	}
	for index := range table1 {
		if index == 0 || index == len(table1)-1 {
			continue
		}
		trimedt1 := removeSpaces(table1[index])
		trimedt2 := removeSpaces(table2[index])
		if trimedt1 != trimedt2 && trimedt1[0] != LineSeperator {
			fmt.Println("Not identical lines:\nIn Result:\n", trimedt1, "\nExpected:\n", trimedt2)
			return true
		}
	}
	return false
}

func removeSpaces(s string) string {
	return strings.Join(strings.Fields(s), "")
}
