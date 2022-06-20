package In_memo_db

import (
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"
)

var (
	GlobalDay time.Weekday
)

type sensormap struct {
	db map[string]*sensorWeekDB
	sync.RWMutex
}

//day implementation
func (s *sensorDayDB) getDayAvg() float32 {
	count := s.Count
	if count == 0 {
		return 0.0
	}
	return float32(s.Sum) / float32(count)
}

func (s *sensorDayDB) getDayRes() (int32, int32, float32) {
	return s.Max, s.Min, s.getDayAvg()
}

func (s *sensorDayDB) AddMeasure(m int32) {
	s.Count++
	s.Sum += m
	s.Min = func(a, b int32) int32 {
		if a < b {
			return a
		}
		return b
	}(s.Min, m)
	s.Max = func(a, b int32) int32 {
		if a > b {
			return a
		}
		return b
	}(s.Max, m)
}

func (s *sensorDayDB) resetDay() {
	s.Max = math.MinInt32
	s.Min = math.MaxInt32
	s.Count = 0
	s.Sum = 0
}

// AddMeasure week implementation
func (sw *sensorWeekDB) AddMeasure(m int32) {
	dayIndex := int(time.Now().Weekday()) //Sunday=0
	sw.Week[dayIndex].AddMeasure(m)
}

func (sw *sensorWeekDB) cleanDay(weekday time.Weekday) {
	d := int(weekday)
	sw.Week[d].resetDay()
}

func newSensorWeek() *sensorWeekDB {
	sw := &sensorWeekDB{Week: make([]sensorDayDB, 7)}
	sww := sw.Week
	for i := range sww {
		sww[i].resetDay()
	}
	return sw
}

func (sw *sensorWeekDB) getInfoBySensorWeek(s string, d int32) string {

	var output strings.Builder
	switch d {
	case 0, 1, 2, 3, 4, 5, 6:
		if _, err := fmt.Fprintf(&output, ",%v", buildDayString(&sw.Week[d], d)); err != nil {
			log.Println(err)
		}
	case 8: //all week
		for i, d := range sw.Week {
			if _, err := fmt.Fprintf(&output, ",%v", buildDayString(&d, int32(i))); err != nil {
				log.Println(err)
			}
		}
	case 9: //today
		today := int32(time.Now().Weekday())
		if _, err := fmt.Fprintf(&output, ",%v", buildDayString(&sw.Week[today], today)); err != nil {
			log.Println(err)
		}
	default:
		log.Println("getInfoBySensor - error: wrong day option:", d)
		return ""
	}

	return fmt.Sprintf("%s%s", s, output.String())
}

// AddMeasure - implementation of sensorDB interface
func (sm *sensormap) AddMeasure(serial string, measure int32) {
	sm.Lock()
	defer sm.Unlock()
	if _, ok := sm.db[serial]; !ok {
		sm.addSensorToMap(serial)
	}
	sm.db[serial].AddMeasure(measure)
}

func (sm *sensormap) getInfoAllSensors(day int32) string {
	sm.RLock()
	defer sm.RUnlock()

	if len(sm.db) == 0 {
		return ""
	}

	// run the query with WorkerPool
	finalResChan := make(chan string, 1)
	GetInfoWorkerPool(sm, day, finalResChan)
	output := <-finalResChan
	return output
}

func buildDayString(day *sensorDayDB, d int32) string {
	a, b, c := day.getDayRes()
	//order: sensorSerial,day,max,min,avg
	return fmt.Sprintf("%v,%v,%v,%v,", time.Weekday(d), a, b, c)
}

func (sm *sensormap) getInfoBySensor(s string, d int32) string {
	if s == "" {
		return s
	}
	sm.RLock()
	defer sm.RUnlock()
	if _, ok := sm.db[s]; !ok {
		return ""
	}

	return sm.db[s].getInfoBySensorWeek(s, d)
}

func (sm *sensormap) GetInfo(serial string, daysBefore int32) string {
	if serial == "all" {
		log.Println(sm.getInfoAllSensors(daysBefore))
		return sm.getInfoAllSensors(daysBefore)
	}
	return sm.getInfoBySensor(serial, daysBefore)
}

func (sm *sensormap) addSensorToMap(s string) {
	sw := newSensorWeek()
	sm.db[s] = sw
}

func SensorMap() *sensormap {
	GlobalDay = time.Now().Weekday() //update global
	output := &sensormap{db: make(map[string]*sensorWeekDB, 1000)}
	return output
}

/*
	Update that occur every AddMeasure and getInfo
	The design:
	Before client getInfo or sensor AddMeasure -
	Check if the day have been changed since last request
	If not - continue
	If so - need to clean the current day (run on parallel on all sensorWeekDB and tell then to reset the day)
*/
func (sm *sensormap) DayCleanup() {
	sm.Lock()
	defer sm.Unlock()
	fname := "dayCleanup"
	var wg sync.WaitGroup
	now := time.Now().Weekday()

	//if same day - no need to cleanup day from all sensors
	if GlobalDay == now {
		return
	}

	log.Println(fname, "Starting day cleanup in DB")

	for _, sensorWeek := range sm.db {
		wg.Add(1)

		go func(s *sensorWeekDB) {
			defer wg.Done()
			s.cleanDay(now)
		}(sensorWeek)
	}

	wg.Wait()
	log.Println(fname, now)
	GlobalDay = now //update global
}

func (sm *sensormap) len() int {
	sm.RLock()
	defer sm.RUnlock()
	return len(sm.db)
}

//TODO
/*
1)

*/
