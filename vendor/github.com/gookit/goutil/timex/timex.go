// Package timex provides an enhanced time.Time implementation.
// Add more commonly used functional methods.
//
// such as: DayStart(), DayAfter(), DayAgo(), DateFormat() and more.
package timex

import (
	"time"

	"github.com/gookit/goutil/basefn"
	"github.com/gookit/goutil/strutil"
)

// provide some commonly time consts
const (
	OneSecond  = 1
	OneMinSec  = 60
	OneHourSec = 3600
	OneDaySec  = 86400
	OneWeekSec = 7 * 86400

	Microsecond = time.Microsecond
	Millisecond = time.Millisecond

	Second  = time.Second
	OneMin  = time.Minute
	Minute  = time.Minute
	OneHour = time.Hour
	Hour    = time.Hour
	OneDay  = 24 * time.Hour
	Day     = OneDay
	OneWeek = 7 * 24 * time.Hour
	Week    = OneWeek
)

// TimeX alias of Time
type TimeX = Time

// Time an enhanced time.Time implementation.
type Time struct {
	time.Time
	// Layout set the default date format layout. default use DefaultLayout
	Layout string
}

/*************************************************************
 * Create timex instance
 *************************************************************/

// Now time instance
func Now() *Time {
	return &Time{Time: time.Now(), Layout: DefaultLayout}
}

// New instance form given time
func New(t time.Time) *Time {
	return &Time{Time: t, Layout: DefaultLayout}
}

// Wrap the go time instance. alias of the New()
func Wrap(t time.Time) *Time {
	return &Time{Time: t, Layout: DefaultLayout}
}

// FromTime new instance form given time.Time. alias of the New()
func FromTime(t time.Time) *Time {
	return &Time{Time: t, Layout: DefaultLayout}
}

// Local time for now
func Local() *Time {
	return New(time.Now().In(time.Local))
}

// FromUnix create from unix time
func FromUnix(sec int64) *Time {
	return New(time.Unix(sec, 0))
}

// FromDate create from datetime string.
func FromDate(s string, template ...string) (*Time, error) {
	if len(template) > 0 && template[0] != "" {
		return FromString(s, ToLayout(template[0]))
	}
	return FromString(s)
}

// FromString create from datetime string. see strutil.ToTime()
func FromString(s string, layouts ...string) (*Time, error) {
	t, err := strutil.ToTime(s, layouts...)
	if err != nil {
		return nil, err
	}
	return New(t), nil
}

// LocalByName time for now
func LocalByName(tzName string) *Time {
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		panic(err)
	}

	return New(time.Now().In(loc))
}

/*************************************************************
 * timex usage
 *************************************************************/

// T returns the t.Time
func (t Time) T() time.Time {
	return t.Time
}

// Format returns a textual representation of the time value formatted according to the layout defined by the argument.
//
// see time.Time.Format()
func (t *Time) Format(layout string) string {
	if layout == "" {
		layout = t.Layout
	}
	return t.Time.Format(layout)
}

// Datetime use DefaultLayout format time to date. see Format()
func (t *Time) Datetime() string {
	return t.Format(t.Layout)
}

// TplFormat use input template format time to date.
//
// alias of DateFormat()
func (t *Time) TplFormat(template string) string {
	return t.DateFormat(template)
}

// DateFormat use input template format time to date.
//
// Example:
//
//	tn := timex.Now()
//	tn.DateFormat("Y-m-d H:i:s") // Output: 2019-01-01 12:12:12
//	tn.DateFormat("Y-m-d H:i") // Output: 2019-01-01 12:12
//	tn.DateFormat("Y-m-d") // Output: 2019-01-01
//	tn.DateFormat("Y-m") // Output: 2019-01
//	tn.DateFormat("y-m-d") // Output: 19-01-01
//	tn.DateFormat("ymd") // Output: 190101
//
// see ToLayout() for convert template to layout.
func (t *Time) DateFormat(template string) string {
	return t.Format(ToLayout(template))
}

// Yesterday get day ago time for the time
func (t *Time) Yesterday() *Time {
	return t.AddSeconds(-OneDaySec)
}

// DayAgo get some day ago time for the time
func (t *Time) DayAgo(day int) *Time {
	return t.AddSeconds(-day * OneDaySec)
}

// AddDay add some day time for the time
func (t *Time) AddDay(day int) *Time {
	return t.AddSeconds(day * OneDaySec)
}

// SubDay add some day time for the time
func (t *Time) SubDay(day int) *Time {
	return t.AddSeconds(-day * OneDaySec)
}

// Tomorrow time. get tomorrow time for the time
func (t *Time) Tomorrow() *Time {
	return t.AddSeconds(OneDaySec)
}

// DayAfter get some day after time for the time.
// alias of Time.AddDay()
func (t *Time) DayAfter(day int) *Time {
	return t.AddDay(day)
}

// AddDur some duration time
func (t *Time) AddDur(dur time.Duration) *Time {
	return &Time{
		Time:   t.Add(dur),
		Layout: DefaultLayout,
	}
}

// AddString add duration time string.
//
// Example:
//
//	tn := timex.Now() // example as "2019-01-01 12:12:12"
//	nt := tn.AddString("1h")
//	nt.Datetime() // Output: 2019-01-01 13:12:12
func (t *Time) AddString(dur string) *Time {
	d, err := ToDuration(dur)
	if err != nil {
		panic(err)
	}
	return t.AddDur(d)
}

// AddHour add some hour time
func (t *Time) AddHour(hours int) *Time {
	return t.AddSeconds(hours * OneHourSec)
}

// SubHour add some hour time
func (t *Time) SubHour(hours int) *Time {
	return t.AddSeconds(-hours * OneHourSec)
}

// AddMinutes add some minutes time for the time
func (t *Time) AddMinutes(minutes int) *Time {
	return t.AddSeconds(minutes * OneMinSec)
}

// SubMinutes add some minutes time for the time
func (t *Time) SubMinutes(minutes int) *Time {
	return t.AddSeconds(-minutes * OneMinSec)
}

// AddSeconds add some seconds time the time
func (t *Time) AddSeconds(seconds int) *Time {
	return &Time{
		Time: t.Add(time.Duration(seconds) * time.Second),
		// with layout
		Layout: DefaultLayout,
	}
}

// SubSeconds add some seconds time the time
func (t *Time) SubSeconds(seconds int) *Time {
	return &Time{
		Time: t.Add(time.Duration(-seconds) * time.Second),
		// with layout
		Layout: DefaultLayout,
	}
}

// Diff calc diff duration for t - u.
// alias of time.Time.Sub()
func (t Time) Diff(u time.Time) time.Duration {
	return t.Sub(u)
}

// DiffSec calc diff seconds for t - u
func (t Time) DiffSec(u time.Time) int {
	return int(t.Sub(u) / time.Second)
}

// DiffUnix calc diff seconds for t.Unix() - u
func (t Time) DiffUnix(u int64) int {
	return int(t.Unix() - u)
}

// SubUnix calc diff seconds for t - u
func (t Time) SubUnix(u time.Time) int {
	return int(t.Sub(u) / time.Second)
}

// HourStart time
func (t *Time) HourStart() *Time {
	y, m, d := t.Date()
	newTime := time.Date(y, m, d, t.Hour(), 0, 0, 0, t.Location())

	return New(newTime)
}

// HourEnd time
func (t *Time) HourEnd() *Time {
	y, m, d := t.Date()
	newTime := time.Date(y, m, d, t.Hour(), 59, 59, int(time.Second-time.Nanosecond), t.Location())

	return New(newTime)
}

// DayStart get time at 00:00:00
func (t *Time) DayStart() *Time {
	y, m, d := t.Date()
	newTime := time.Date(y, m, d, 0, 0, 0, 0, t.Location())

	return New(newTime)
}

// DayEnd get time at 23:59:59
func (t *Time) DayEnd() *Time {
	y, m, d := t.Date()
	newTime := time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())

	return New(newTime)
}

// CustomHMS custom change the hour, minute, second for create new time.
func (t *Time) CustomHMS(hour, min, sec int) *Time {
	y, m, d := t.Date()
	newTime := time.Date(y, m, d, hour, min, sec, int(time.Second-time.Nanosecond), t.Location())

	return FromTime(newTime)
}

// IsBefore the given time
func (t *Time) IsBefore(u time.Time) bool {
	return t.Before(u)
}

// IsBeforeUnix the given unix timestamp
func (t *Time) IsBeforeUnix(ux int64) bool {
	return t.Before(time.Unix(ux, 0))
}

// IsAfter the given time
func (t *Time) IsAfter(u time.Time) bool {
	return t.After(u)
}

// IsAfterUnix the given unix timestamp
func (t *Time) IsAfterUnix(ux int64) bool {
	return t.After(time.Unix(ux, 0))
}

// Timestamp value. alias of t.Unix()
func (t Time) Timestamp() int64 {
	return t.Unix()
}

// HowLongAgo format diff time to string.
func (t Time) HowLongAgo(before time.Time) string {
	return basefn.HowLongAgo(t.Unix() - before.Unix())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
//
// Tip: will auto match a format by strutil.ToTime()
func (t *Time) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}

	// Fractional seconds are handled implicitly by Parse.
	tt, err := strutil.ToTime(string(data[1 : len(data)-1]))
	if err == nil {
		t.Time = tt
	}
	return err
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
//
// Tip: will auto match a format by strutil.ToTime()
func (t *Time) UnmarshalText(data []byte) error {
	// Fractional seconds are handled implicitly by Parse.
	tt, err := strutil.ToTime(string(data))
	if err == nil {
		t.Time = tt
	}
	return err
}
