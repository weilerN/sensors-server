package In_memo_db

type sensorDayDB struct {
	Max   int32
	Min   int32
	Count int32
	Sum   int32
}

type sensorWeekDB struct {
	Week []sensorDayDB
}
