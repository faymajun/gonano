package timer

import (
	"fmt"
	"math"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	loopForever = -1
)
const (
	SECOND = 60
	HOUR   = 60 * SECOND
	DAY    = 24 * HOUR
)

var (
	//时间格式
	TimeFormate string = "2006-01-02 15:04:05"
	TimeYMD     string = "2006-01-02"

	SIMPLE_DAY_FORMAT string = "20060102"
	//时差修正 纳秒差值
	TimeFix int64
	//服务器使用的标准时区
	Server_Location *time.Location = time.Local
)

var (
	logger = logrus.WithField("component", "timer")

	// manager manages all timers
	manager = &struct {
		incrementId int64            // auto increment id
		timers      map[int64]*Timer // all timers

		muClosingTimer sync.Mutex
		closingTimer   []int64
		muCreatedTimer sync.Mutex
		createdTimer   []*Timer
	}{}

	// precision indicates the precision of timer, default is time.Second
	precision = time.Second
)

type (
	// TimerFunc represents a function which will be called periodically in main
	// logic gorontine.
	TimerFunc func()

	TimerCondition interface {
		Check(time.Time) bool
	}

	// Timer represents a Cron job

	Timer struct {
		id        int64         // timer id
		fn        TimerFunc     // function that execute
		createAt  int64         // timer create time
		interval  time.Duration // execution interval
		condition TimerCondition
		elapse    int64 // total elapse time
		closed    int32 // is timer closed
		counter   int   // counter
	}
)

func init() {
	manager.timers = map[int64]*Timer{}
}

// ID returns id of current timer
func (t *Timer) ID() int64 {
	return t.id
}

// Stop turns off a timer. After Stop, fn will not be called forever
func (t *Timer) Stop() {
	if atomic.AddInt32(&t.closed, 1) != 1 {
		return
	}

	t.counter = 0
}

// execute job function with protection
func safecall(id int64, fn TimerFunc) {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("Call timer function error, TimerID=%d, Error=%v", id, err)
			fmt.Fprintln(logger.Logger.Out, string(debug.Stack()))
		}
	}()

	fn()
}

func TestDo() {
	for _, t := range manager.timers {
		safecall(t.id, t.fn)
	}
}

func Cron() {
	if len(manager.createdTimer) > 0 {
		manager.muCreatedTimer.Lock()
		for _, t := range manager.createdTimer {
			manager.timers[t.id] = t
		}
		manager.createdTimer = manager.createdTimer[:0]
		manager.muCreatedTimer.Unlock()
	}

	if len(manager.timers) < 1 {
		return
	}

	now := time.Now()
	unn := now.UnixNano()
	for id, t := range manager.timers {
		if t.counter == loopForever || t.counter > 0 {
			// condition timer
			if t.condition != nil {
				if t.condition.Check(now) {
					safecall(t.id, t.fn)
				}
				continue
			}

			// execute job
			if t.createAt+t.elapse <= unn {
				safecall(id, t.fn)
				t.elapse += int64(t.interval)

				// update timer counter
				if t.counter != loopForever && t.counter > 0 {
					t.counter--
				}
			}
		}

		// prevent chClosingTimer exceed
		if t.counter == 0 {
			manager.muClosingTimer.Lock()
			manager.closingTimer = append(manager.closingTimer, t.id)
			manager.muClosingTimer.Unlock()
			continue
		}
	}

	if len(manager.closingTimer) > 0 {
		manager.muClosingTimer.Lock()
		for _, id := range manager.closingTimer {
			delete(manager.timers, id)
		}
		manager.closingTimer = manager.closingTimer[:0]
		manager.muClosingTimer.Unlock()
	}
}

// NewTimer returns a new Timer containing a function that will be called
// with a period specified by the duration argument. It adjusts the intervals
// for slow receivers.
// The duration d must be greater than zero; if not, NewTimer will panic.
// Stop the timer to release associated resources.
func NewTimer(interval time.Duration, fn TimerFunc) *Timer {
	return NewCountTimer(interval, loopForever, fn)
}

// NewCountTimer returns a new Timer containing a function that will be called
// with a period specified by the duration argument. After count times, timer
// will be stopped automatically, It adjusts the intervals for slow receivers.
// The duration d must be greater than zero; if not, NewTimer will panic.
// Stop the timer to release associated resources.
func NewCountTimer(interval time.Duration, count int, fn TimerFunc) *Timer {
	if fn == nil {
		panic("timer: nil timer function")
	}
	if interval <= 0 {
		panic("non-positive interval for NewTimer")
	}

	id := atomic.AddInt64(&manager.incrementId, 1)
	t := &Timer{
		id:       id,
		fn:       fn,
		createAt: time.Now().UnixNano(),
		interval: interval,
		elapse:   int64(interval), // first execution will be after interval
		counter:  count,
	}

	manager.muCreatedTimer.Lock()
	manager.createdTimer = append(manager.createdTimer, t)
	manager.muCreatedTimer.Unlock()
	return t
}

func NewAfterTimer(interval time.Duration, fn TimerFunc) *Timer {
	return NewCountTimer(interval, 1, fn)
}

// type TestCondition byte
//
// // Every year
// func(t TestCondition) Check(now time.Time)bool {
// 	return now.Year() == 0
// }
//
// NewCondTimer(TestCondition{}, func(){})
func NewCondTimer(condition TimerCondition, fn TimerFunc) *Timer {
	if condition == nil {
		panic("timer: nil condition")
	}

	t := NewCountTimer(time.Duration(math.MaxInt64), loopForever, fn)
	t.condition = condition

	return t
}

// SetPrecision set the ticker precision, and time precision can not less
// than a Millisecond, and can not change after application running. The default
// precision is time.Second
func SetPrecision(p time.Duration) {
	if p < time.Millisecond {
		panic("time p can not less than a Millisecond")
	}
	precision = p
}

func Precision() time.Duration {
	return precision
}

func String2Unix(strTime string) int64 {
	loc, _ := time.LoadLocation("Local") //重要：获取时区
	t, err := time.ParseInLocation("2006-01-02 15:04:05", strTime, loc)
	if err != nil {
		logger.Errorf(err.Error())
		return 0
	}
	return t.Unix()
}

func NowTime() int64 {
	return time.Now().Unix()
}

//检测是否时间是否是同一天
func CheckOneDay(t1, t2 int64) bool {
	year1, month1, day1 := time.Unix(t1, 0).Local().Date()
	year2, month2, day2 := time.Unix(t2, 0).Local().Date()
	return year1 == year2 && month1 == month2 && day1 == day2
}

//检测是不是同一个月
func CheckOneMonth(t1, t2 int64) bool {
	year1, month1, _ := time.Unix(t1, 0).Local().Date()
	year2, month2, _ := time.Unix(t2, 0).Local().Date()
	return year1 == year2 && month1 == month2
}

// hms 10:10:00 时分秒  当天该时分秒的时间
func GetTodayTime(hms string) int64 {
	ymd := time.Now().Format(TimeYMD)
	timeStr := ymd + " " + hms
	formatTime, _ := time.ParseInLocation(TimeFormate, timeStr, Server_Location)
	return formatTime.Unix()
}

//以今天某个时间点判断是否跨天刷新,hms 02:00:00今天凌晨 checkTime检测时间
func CheckDayRefresh(hms string, checkTime int64) bool {
	now := NowTime()                 //当前时间
	refreshTime := GetTodayTime(hms) //今日刷新时间
	zeroPoint := GetTodayTime("00:00:00")
	durationTime := refreshTime - zeroPoint
	return CheckOneDay(checkTime-durationTime, now-durationTime)

}

func CheckMonthRefresh(day int, hms string, checkTime int64) bool {
	now := NowTime()                         //当前时间
	refreshTime := GetMonthTime(day, hms)    //本月刷新时间
	zeroPoint := GetMonthTime(1, "00:00:00") //本月1号时间
	durationTime := refreshTime - zeroPoint
	return CheckOneMonth(checkTime-durationTime, now-durationTime)

}

//获得当前月几号几点的时间(day几号，hms 02:00:00)
func GetMonthTime(day int, hms string) int64 {
	year, month, _ := time.Now().Date()
	dayTime := time.Date(year, month, day, 0, 0, 0, 0, Server_Location)
	ymd := dayTime.Format(TimeYMD)
	timeStr := ymd + " " + hms
	formatTime, _ := time.ParseInLocation(TimeFormate, timeStr, Server_Location)
	return formatTime.Unix()
}

//获得当前周，周几几点的时间
func GetWeekTime(week int, hms string) int64 {
	if week > 7 || week < 0 {
		return 0
	}
	weekDay := time.Now().Weekday()
	year, month, day := time.Now().Date()
	day += (week - int(weekDay))
	dayTime := time.Date(year, month, int(day), 0, 0, 0, 0, Server_Location)
	ymd := dayTime.Format(TimeYMD)
	timeStr := ymd + " " + hms
	formatTime, _ := time.ParseInLocation(TimeFormate, timeStr, Server_Location)
	return formatTime.Unix()

}

// 获取今天0点时间
func GetNowDay0Unix(t int64) int64 {
	t1 := time.Unix(t, 0)
	tm1 := time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, t1.Location())
	return tm1.Unix()
}

func GetNowDaySpecUnix(t int64, hour, min, sec int) int64 {
	t1 := time.Unix(t, 0)
	tm1 := time.Date(t1.Year(), t1.Month(), t1.Day(), hour, min, sec, 0, t1.Location())
	return tm1.Unix()
}

//计算现在里下次周几几点的时间
func GetNextWeekTime(weekDay time.Weekday, hour, minute, second int) time.Time {
	now := time.Now()
	var next time.Time

	//计算下次执行时间
	next = time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, 0, now.Location())

	days := (int(weekDay) + 7 - int(now.Weekday())) % 7
	if days != 0 {
		next = next.Add(24 * time.Hour * time.Duration(days))
	} else {
		if next.Unix() <= now.Unix() {
			next = next.Add(7 * 24 * time.Hour)
		}
	}
	return next
}

//计算当前时间到下一个整点时间
func GetNextFullTime() (next time.Time) {
	now := time.Now()
	nowFullTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
	nextTime := nowFullTime.Unix() + HOUR
	return time.Unix(nextTime, 0)
}

//以当前时间为节点，计算最近的下一个几号几点的时间戳
func GetNextMonthDay(monthDay, hour, minute, second int) time.Time {
	now := time.Now()
	nowMonthTime := time.Date(now.Year(), now.Month(), monthDay, hour, minute, second, 0, now.Location())
	if now.Unix() < nowMonthTime.Unix() {
		return nowMonthTime
	}
	nextTime := time.Date(now.Year(), now.Month()+1, monthDay, hour, minute, second, 0, now.Location())
	return nextTime
}

//以当前时间为节点，计算最近的上一个几号几点的时间戳
func GetLastMonthDay(monthDay, hour, minute, second int) time.Time {
	now := time.Now()
	nowMonthTime := time.Date(now.Year(), now.Month(), monthDay, hour, minute, second, 0, now.Location())
	if now.Unix() > nowMonthTime.Unix() {
		return nowMonthTime
	}
	lastTime := time.Date(now.Year(), now.Month()-1, monthDay, hour, minute, second, 0, now.Location())
	return lastTime
}

//判断时间是当年的第几周
func WeekByDate(t time.Time) string {
	yearDay := t.YearDay()
	yearFirstDay := t.AddDate(0, 0, -yearDay+1)
	firstDayInWeek := int(yearFirstDay.Weekday())

	//今年第一周有几天
	firstWeekDays := 1
	if firstDayInWeek != 0 {
		firstWeekDays = 7 - firstDayInWeek + 1
	}
	var week int
	if yearDay <= firstWeekDays {
		week = 1
	} else {
		week = (yearDay-firstWeekDays)/7 + 2
	}
	return fmt.Sprintf("%dyear%dweek", t.Year(), week)
}

//判断当前是否是周末
func CheckWeekEnd() bool {
	weekDay := time.Now().Weekday()
	if weekDay == time.Sunday || weekDay == time.Saturday {
		return true
	}
	return false
}
