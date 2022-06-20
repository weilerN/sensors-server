package client

import (
	grpcdb "SensorServer/pkg/grpc_db"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

func MyPanic(e error) {
	if e != nil {
		panic(fmt.Sprintf("%v", e))
	}
}

func Debug(fname, s string) {
	if *verbose {
		log.Printf("[%s]\t%s\n", fname, s)
	}
}

func NewConnReq() *grpcdb.ConnReq {
	fields := make([]string, 2)
	for i := 0; i < 2; i++ {
		switch i {
		case 0:
			fmt.Println("enter user name")
		case 1:
			fmt.Println("enter password")
		}
		_, err := fmt.Scanf("%s", &fields[i])
		MyPanic(err)
	}
	return &grpcdb.ConnReq{UserName: fields[0], Password: fields[1]}
}

func VerifyLogin(r *grpcdb.ConnRes, err error) bool {
	res := ""
	if err != nil {
		if e := UnpackError(err); e == loginCredentialsError || e == alreadyConnectedError {
			log.Println("Error:", e)
		} else {
			log.Fatalf("Error:%s", e)
		}
	} else {
		res = r.GetRes()
		log.Printf("Res: %s", res)
	}
	return res == loginConnectedMessage
}

func UnpackError(e error) string {
	s := fmt.Sprintf("%v", e)
	return s[strings.LastIndex(s, "=")+2:]
}

func printHelper(min, max, avg string) (string, string, string) {
	if min == strconv.Itoa(math.MinInt32) {
		return "-", "-", "-"
	}
	return min, max, avg
}

func PrintResult(s string) {
	arr := strings.Split(s, ",")
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	t.AppendHeader(table.Row{"#SERIAL", "DAY", "MIN", "MAX", "AVG"})
	for i := 0; i < len(arr)-1; i += 5 {
		if arr[i] != "" && i > 0 {
			t.AppendSeparator()
		}
		a, b, c := printHelper(arr[i+2], arr[i+3], arr[i+4])
		t.AppendRows([]table.Row{
			{arr[i], arr[i+1], a, b, c},
		})
	}
	//t.AppendFooter(table.Row{"", "", "Total", 10000})
	t.Render()
}
