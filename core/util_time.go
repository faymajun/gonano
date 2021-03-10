package core

import (
	"time"
)

const (
	MinuteSeconds = 60
	HourSeconds   = MinuteSeconds * 60
	DaySeconds    = HourSeconds * 24
)

func Date() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func ParseTime(str string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", str)
}

/**
* @brief 获得timestamp距离下个小时的时间，单位s
*
* @return uint32_t 距离下个小时的时间，单位s
 */
func GetNextHourIntervalS() int {
	return int(3600 - (time.Now().Unix() % 3600))
}

/**
 * @brief 获得timestamp距离下个小时的时间，单位ms
 *
 * @return uint32_t 距离下个小时的时间，单位ms
 */
func GetNextHourIntervalMS() int {
	return GetNextHourIntervalS() * 1000
}

//获得timestamp距离明天的时间，单位s
func GetNextDayIntervalS(timezone int) int {
	return int(86400-time.Now().Unix()%86400) - timezone*3600
}

/**
* @brief 时间戳转换为小时，24小时制，0点用24表示
*
* @param timestamp 时间戳
* @param timezone  时区
* @return uint32_t 小时 范围 1-24
 */
func GetHour24(timestamp int64, timezone int) int {
	hour := (int((timestamp%86400)/3600) + timezone)
	if hour > 24 {
		return hour - 24
	}
	return hour
}

/**
 * @brief 时间戳转换为小时，24小时制，0点用0表示
 *
 * @param timestamp 时间戳
 * @param timezone  时区
 * @return uint32_t 小时 范围 0-23
 */
func GetHour23(timestamp int64, timezone int) int {
	hour := GetHour24(timestamp, timezone)
	if hour == 24 {
		return 0 //24点就是0点
	}
	return hour
}

func GetHour(timestamp int64, timezone int) int {
	return GetHour23(timestamp, timezone)
}

/**
* @brief 判断两个时间戳是否是同一天
*
* @param now 需要比较的时间戳
* @param old 需要比较的时间戳
* @param timezone 时区
* @return uint32_t 返回不同的天数
 */
func IsDiffDay(now, old int64, timezone int) int {
	now += int64(timezone * 3600)
	old += int64(timezone * 3600)
	return int((now / 86400) - (old / 86400))
}

/**
* @brief 判断时间戳是否处于一个小时的两边，即一个时间错大于当前的hour，一个小于
*
* @param now 需要比较的时间戳
* @param old 需要比较的时间戳
* @param hour 小时，0-23
* @param timezone 时区
* @return bool true表示时间戳是否处于一个小时的两边
 */
func IsDiffHour(now, old int64, hour, timezone int) bool {
	diff := IsDiffDay(now, old, timezone)
	if diff == 1 {
		if GetHour23(old, timezone) > hour {
			return GetHour23(now, timezone) >= hour
		}
	} else if diff >= 2 {
		return true
	}

	return (GetHour23(now, timezone) >= hour) && (GetHour23(old, timezone) < hour)
}

func IsAcrossDay(now, old int64, hour, minute int64, timezone int) bool {
	if now <= old {
		return false
	}

	if now-old >= DaySeconds {
		return true
	}

	now += int64(timezone * HourSeconds)
	old += int64(timezone * HourSeconds)

	nowAcross := now/DaySeconds*DaySeconds + HourSeconds*hour + MinuteSeconds*minute
	return old < nowAcross && now >= nowAcross
}

func GetNowDaySpecUnix(t int64, hour, min, sec int) int64 {
	t1 := time.Unix(t, 0)
	tm1 := time.Date(t1.Year(), t1.Month(), t1.Day(), hour, min, sec, 0, t1.Location())
	return tm1.Unix()
}
