package redis_db

import "sync"

type sensorDayDB struct {
	Max   int32 `redis:"int1"`
	Min   int32 `redis:"int2"`
	Count int32 `redis:"int3"`
	Sum   int32 `redis:"int4"`
}

//type sensorWeekDB struct {
//	Week []sensorDayDB `redis:"arr1"`
//}

//Store all serials in Hash set structure
//Used as cache memory in redis_db
type stringHash struct {
	stringMap map[string]struct{}
	sync.RWMutex
}

func NewStringHash() *stringHash {
	return &stringHash{stringMap: make(map[string]struct{}, 10000)}
}

func (sh *stringHash) Add(str string) {
	sh.Lock()
	defer sh.Unlock()
	sh.stringMap[str] = struct{}{}
}

func (sh *stringHash) exists(str string) bool {
	sh.RLock()
	defer sh.RUnlock()
	_, output := sh.stringMap[str]
	return output
}
