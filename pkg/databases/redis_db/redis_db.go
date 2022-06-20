package redis_db

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	MaxActive = 10000 //max number of connections
	MaxIdle   = 10000
)

var (
	defaultDay = sensorDayDB{Max: math.MinInt32, Min: math.MaxInt32, Count: 0, Sum: 0}
	GlobalDay  time.Weekday
	serials    *stringHash
)

type redisDB struct {
	pool *redis.Pool
}

func New() *redisDB {
	GlobalDay = time.Weekday(0)
	//GlobalDay = time.Now().Weekday()
	pool := &redis.Pool{
		MaxActive:   MaxActive,
		MaxIdle:     MaxIdle,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", ":6379")
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
	initSerials(pool)
	return &redisDB{pool}
}

func initSerials(pool *redis.Pool) {
	conn := pool.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}(conn)
	serials = NewStringHash()

	data, err := redis.Strings(conn.Do("KEYS", "*"))
	if err != nil {
		log.Println("initSerials:", err)
		return
	}
	//iterate over the data and add all values to HashSet
	for _, serial := range data {
		serials.Add(serial)
	}
}

func (rdb *redisDB) GetInfo(serial string, daysBefore int32) string {
	if serial == "all" {
		return rdb.getInfoAllSensors(int(daysBefore))
	}
	return rdb.getInfoBySensor(serial, int(daysBefore))
}

func (rdb *redisDB) buildDayString(serial string, d time.Weekday) string {
	var day sensorDayDB
	if err := rdb.getDay(serial, d, &day); err != nil {
		return ""
	}
	//order: sensorSerial,day,max,min,avg
	return fmt.Sprintf("%v,%v,%v,%v,", d, day.Max, day.Min, float32(day.Sum)/float32(day.Count))
}

func (rdb *redisDB) getInfoAllSensors(opt int) string {
	var output strings.Builder
	serials.RLock()
	serialsList := serials.stringMap //get a copy to save mutex time
	serials.RUnlock()
	for sensor := range serialsList {
		res := rdb.getInfoBySensor(sensor, opt)
		if _, err := fmt.Fprintf(&output, "%v", res); err != nil {
			log.Println(err)
		}
	}
	return output.String()
}

func (rdb *redisDB) getInfoBySensor(s string, opt int) string {

	var output strings.Builder
	switch opt {
	case 0, 1, 2, 3, 4, 5, 6:
		res := rdb.buildDayString(s, time.Weekday(opt))
		if res == "" {
			return ""
		}
		if _, err := fmt.Fprintf(&output, ",%v", res); err != nil {
			log.Println(err)
		}
	case 8: //all week
		if _, err := fmt.Fprintf(&output, "%v", rdb.getWeek(s)); err != nil {
			log.Println(err)
		}
	case 9: //today
		today := int32(time.Now().Weekday())
		res := rdb.buildDayString(s, time.Weekday(today))
		if res == "" {
			return ""
		}
		if _, err := fmt.Fprintf(&output, ",%v", res); err != nil {
			log.Println(err)
		}
	default:
		log.Println("getInfoBySensor - error: wrong day option:", opt)
		return ""
	}
	return fmt.Sprintf("%s%s", s, output.String())
}

func (rdb *redisDB) AddMeasure(serial string, measure int32) {
	//use cache to know if need to add new sensor to db
	if !serials.exists(serial) {
		rdb.addNewSensor(serial)
		serials.Add(serial)
	}
	rdb.addMeasureToday(serial, measure)
}

func (rdb *redisDB) DayCleanup() {
	return //TODO - need to think about how to run it one time, and not for all request in the same time
	//today := time.Now().Weekday()
	//if today != GlobalDay {
	//	log.Println("[DayCleanup]\ttoday:", today, "\tbefore:", GlobalDay)
	//	rdb.dayCleanup(today)
	//	GlobalDay = today
	//	log.Println("[DayCleanup]\ttoday:", today, "\tafter:", GlobalDay)
	//}
}

//inner functions
func (rdb *redisDB) setDay(serial string, day time.Weekday, sd sensorDayDB) error {
	conn := rdb.pool.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}(conn)

	_, err := conn.Do("HSET", serial, day, redis.Args{}.AddFlat(&sd))
	if err != nil {
		log.Println("setDay - Error:", err)
		return err
	}
	return nil
}

func (rdb *redisDB) getDay(serial string, day time.Weekday, dest *sensorDayDB) error {
	conn := rdb.pool.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}(conn)
	data, err := redis.String(conn.Do("HGET", serial, day))
	if err != nil {
		dest.resetDay()
		return fmt.Errorf("getDay HGET error- %v", err)
	}

	err = dest.Scan(data)
	if err != nil {
		dest.resetDay()
		return fmt.Errorf("getDay ScanDay error- %v", err)
	}
	return nil
}

func (sd *sensorDayDB) resetDay() {
	sd.Min = defaultDay.Min
	sd.Max = defaultDay.Max
	sd.Sum = defaultDay.Sum
	sd.Count = defaultDay.Count
}

func (sd *sensorDayDB) Scan(data string) error {
	var err error
	var tmp int
	arr := strings.Split(data[1:len(data)-1], " ")
	if len(arr) != 8 {
		return fmt.Errorf("illegal input length. got %d, required 8", len(arr))
	}

	for i := 1; i < len(arr); i += 2 {
		tmp, err = strconv.Atoi(arr[i])
		if err != nil {
			return fmt.Errorf("strconv.Atoi:%v", err)
		}
		switch i {
		case 1:
			sd.Max = int32(tmp)
		case 3:
			sd.Min = int32(tmp)
		case 5:
			sd.Count = int32(tmp)
		case 7:
			sd.Sum = int32(tmp)
		default:
			return fmt.Errorf("error - Raw data length illegal")
		}
	}
	return nil
}

func (rdb *redisDB) getWeek(serial string) string {
	conn := rdb.pool.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}(conn)
	data, err := redis.Strings(conn.Do("HGETALL", serial))
	if err != nil {
		log.Println("getDay - Error1:", err)
		return ""
	}

	var atoi = func(str string) int {
		var output int
		var err error
		if output, err = strconv.Atoi(str); err != nil {
			log.Println(err)
		}
		return output
	}
	day := 0
	var output strings.Builder
	for i := 1; i < len(data); i += 2 {
		var avg float32
		arr := strings.Split(data[i][1:len(data[i])-1], " ")
		sum, count := float32(atoi(arr[5])), float32(atoi(arr[7]))
		if count == 0 {
			avg = 0.0
		} else {
			avg = sum / count
		}

		if _, err := fmt.Fprintf(&output, ",%s,%d,%d,%v,", time.Weekday(day), atoi(arr[1]), atoi(arr[3]), avg); err != nil {
			log.Println(err)
			return ""
		}
		day++
	}
	return output.String()
}

func (rdb *redisDB) addNewSensor(serial string) {
	for i := 0; i < 7; i++ {
		day := time.Weekday(i)
		err := rdb.setDay(serial, day, defaultDay)
		if err != nil {
			log.Fatal("addNewSensor", err)
		}
	}
}
func (rdb *redisDB) addMeasureToday(serial string, measure int32) {
	today := time.Now().Weekday()
	var sensorDay sensorDayDB
	if err := rdb.getDay(serial, today, &sensorDay); err != nil {
		log.Println("addMeasureToday:\t", err)
	}
	sensorDay.calculateMeasure(measure)
	if err := rdb.setDay(serial, today, sensorDay); err != nil {
		log.Println("addMeasureToday:\t", err)
	}
}
func (sd *sensorDayDB) calculateMeasure(m int32) {
	if sd.Max < m {
		sd.Max = m
	}
	if sd.Min > m {
		sd.Min = m
	}
	sd.Count++
	sd.Sum += m
}

func (rdb *redisDB) dayCleanup(today time.Weekday) {
	serials.RLock()
	defer serials.RUnlock()
	for serial := range serials.stringMap {
		err := rdb.setDay(serial, today, defaultDay)
		if err != nil {
			log.Println("[dayCleanup]\t", err)
		}
	}
}
