package timex

import "time"

// some time layout or time
const (
	DatetimeLayout = "2006-01-02 15:04:05"
	LayoutWithMs3  = "2006-01-02 15:04:05.000"
	LayoutWithMs6  = "2006-01-02 15:04:05.000000"
	DateOnlyLayout = "2006-01-02"
	TimeOnlyLayout = "15:04:05"

	// ZeroUnix zero unix timestamp
	ZeroUnix int64 = -62135596800
)

var (
	// DefaultLayout template for format time
	DefaultLayout = DatetimeLayout
	// ZeroTime zero time instance
	ZeroTime = time.Time{}
)

// SetLocalByName set local by tz name. eg: UTC, PRC
func SetLocalByName(tzName string) error {
	location, err := time.LoadLocation(tzName)
	if err != nil {
		return err
	}

	time.Local = location
	return nil
}

// NowAddDay add some day time from now
func NowAddDay(day int) time.Time {
	return time.Now().AddDate(0, 0, day)
}

// NowAddHour add some hour time from now
func NowAddHour(hour int) time.Time {
	return time.Now().Add(time.Duration(hour) * OneHour)
}

// NowAddMinutes add some minutes time from now
func NowAddMinutes(minutes int) time.Time {
	return time.Now().Add(time.Duration(minutes) * OneMin)
}

// NowAddSec add some seconds time from now. alias of NowAddSeconds()
func NowAddSec(seconds int) time.Time {
	return time.Now().Add(time.Duration(seconds) * time.Second)
}

// NowAddSeconds add some seconds time from now
func NowAddSeconds(seconds int) time.Time {
	return time.Now().Add(time.Duration(seconds) * time.Second)
}

// NowHourStart time
func NowHourStart() time.Time {
	return HourStart(time.Now())
}

// NowHourEnd time
func NowHourEnd() time.Time {
	return HourEnd(time.Now())
}

// AddDay add some day time for given time
func AddDay(t time.Time, day int) time.Time {
	return t.AddDate(0, 0, day)
}

// AddHour add some hour time for given time
func AddHour(t time.Time, hour int) time.Time {
	return t.Add(time.Duration(hour) * OneHour)
}

// AddMinutes add some minutes time for given time
func AddMinutes(t time.Time, minutes int) time.Time {
	return t.Add(time.Duration(minutes) * OneMin)
}

// AddSeconds add some seconds time for given time
func AddSeconds(t time.Time, seconds int) time.Time {
	return t.Add(time.Duration(seconds) * time.Second)
}

// AddSec add some seconds time for given time. alias of AddSeconds()
func AddSec(t time.Time, seconds int) time.Time {
	return t.Add(time.Duration(seconds) * time.Second)
}

// HourStart time for given time
func HourStart(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, t.Hour(), 0, 0, 0, t.Location())
}

// HourEnd time for given time
func HourEnd(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, t.Hour(), 59, 59, int(time.Second-time.Nanosecond), t.Location())
}

// DayStart time for given time
func DayStart(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// DayEnd time for given time
func DayEnd(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
}

// TodayStart time
func TodayStart() time.Time {
	return DayStart(time.Now())
}

// TodayEnd time
func TodayEnd() time.Time {
	return DayEnd(time.Now())
}
