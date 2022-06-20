package In_memo_db

import (
	"SensorServer/internal/client"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"
)

var (
	emptydb = ""
	d0      = "sensor_1,Sunday,-2147483648,2147483647,0,"
	d1      = "sensor_1,Monday,-2147483648,2147483647,0,"
	d2      = "sensor_1,Tuesday,-2147483648,2147483647,0,"
	d3      = "sensor_1,Wednesday,-2147483648,2147483647,0,"
	d4      = "sensor_1,Thursday,-2147483648,2147483647,0,"
	d5      = "sensor_1,Friday,-2147483648,2147483647,0,"
	d6      = "sensor_1,Saturday,-2147483648,2147483647,0,"
	min     = math.MinInt32
	max     = math.MaxInt32
)

func TestGetInfoBySensor_emptyDB(t *testing.T) {
	testName := "TestGetInfoBySensor_emptyDB"
	mapDb := SensorMap()
	t.Run(testName, func(t *testing.T) {
		s := mapDb.getInfoBySensor("", 8)
		if s != emptydb {
			t.Errorf("got %v, want %v", s, emptydb)
		}
	})
}

func TestGetInfoBySensor_SensorsNoTraffic(t *testing.T) {
	testName := "TestGetInfoBySensor_SensorsNoTraffic"
	var tmgrpc_dbuff strings.Builder
	mapDB := SensorMap()
	for i := 1; i < 50; i++ {
		s := fmt.Sprintf("sensor_%d", i)
		mapDB.addSensorToMap(s)
		for i := 0; i < 7; i += 1 {
			if i != 0 {
				fmt.Fprintf(&tmgrpc_dbuff, "%v", fmt.Sprintf("%s,%s,%d,%d,0,", "", time.Weekday(i), min, max))
			} else {
				fmt.Fprintf(&tmgrpc_dbuff, "%v", fmt.Sprintf("%s,%s,%d,%d,0,", s, time.Weekday(i), min, max))
			}
		}
	}
	res := tmgrpc_dbuff.String()
	t.Run(testName, func(t *testing.T) {
		s := mapDB.GetInfo("all", 8)
		if len(s) != len(res) {
			t.Errorf("Different output")
			fmt.Println("got:")
			client.PrintResult(s)
			fmt.Println("\n\n------------------------\n\nwanted:")
			client.PrintResult(res)
		}
	})
}

// add measures to sernsor_1, then test all days ofsensor_1
func TestAddMeasure(t *testing.T) {
	testName, sname := "TestAddMeasure", "sensor_1"
	var tt string
	mapDB := SensorMap()
	mapDB.AddMeasure(sname, 10)
	mapDB.AddMeasure(sname, 30)

	var tests = []string{d0, d1, d2, d3, d4, d5, d6}
	now := time.Now().Weekday()
	today := int32(now)
	var i int32
	for i = 0; i < 7; i++ {
		t.Run(testName, func(t *testing.T) {
			curDay := (today + i) % 7
			s := mapDB.getInfoBySensor(sname, curDay)
			//only today is having measures
			if curDay == today {
				tt = fmt.Sprintf("sensor_1,%s,30,10,20,", now)
			} else {
				tt = tests[curDay]
			}
			if s != tt {
				t.Errorf("go t:\t%v\nwant:\t%v\n\n", s, tt)
			}
		})
	}

}

func TestAddSensorToMap(t *testing.T) {
	mapDB := SensorMap()
	if mapDB.len() != 0 {
		t.Errorf("got len=%v, want len=%v", mapDB.len(), 0)
	}
	for i := 1; i < 10; i++ {
		s := fmt.Sprintf("sensor_%d", i)
		mapDB.addSensorToMap(s)
		if mapDB.len() != i {
			t.Errorf("got len=%v, want len=%v", mapDB.len(), i)
		}
	}
}
