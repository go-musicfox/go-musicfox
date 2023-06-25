# Timex

Provides an enhanced time.Time implementation, and add more commonly used functional methods.

## Install

```go
go get github.com/gookit/goutil/timex
```

## Usage

### Create timex instance

```go
now := timex.Now()

// from time.Time
tx := timex.New(time.Now())
tx := timex.FromTime(time.Now())

// from time unix
tx := timex.FromUnix(1647411580)
```

Create from datetime string:

```go
// auto match layout by datetime
tx, err  := timex.FromString("2022-04-20 19:40:34")
// custom set the datetime layout
tx, err  := timex.FromString("2022-04-20 19:40:34", "2006-01-02 15:04:05")
// use date template as layout
tx, err  := timex.FromDate("2022-04-20 19:40:34", "Y-m-d H:I:S")
```

### Use timex instance

```go
tx := timex.Now()
```

**Change time**:

```go
tx.Yesterday()
tx.Tomorrow()

tx.DayStart() // get time at Y-m-d 00:00:00
tx.DayEnd() // get time at Y-m-d 23:59:59
tx.HourStart() // get time at Y-m-d H:00:00
tx.HourEnd() // get time at Y-m-d H:59:59

tx.AddDay(2)
tx.AddHour(1)
tx.AddMinutes(15)
tx.AddSeconds(120)
```

**Compare time**:

```go
// before compare
tx.IsBefore(u time.Time)
tx.IsBeforeUnix(1647411580)
// after compare
tx.IsAfter(u time.Time)
tx.IsAfterUnix(1647411580)
```

### Helper functions

```go
ts := timex.NowUnix() // current unix timestamp

t := NowAddDay(1) // from now add 1 day
t := NowAddHour(1) // from now add 1 hour
t := NowAddMinutes(3) // from now add 3 minutes
t := NowAddSeconds(180) // from now add 180 seconds
```

### Convert time to date by template

**Template Chars**:

```text
 Y,y - year
  Y - year 2006
  y - year 06
 m - month 01-12
 d - day 01-31
 H,h - hour
  H - hour 00-23
  h - hour 01-12
 I,i - minute
  I - minute 00-59
  i - minute 0-59
 S,s - second
  S - second 00-59
  s - second 0-59
... ...
```

> More, please see [charMap](./template.go)

Examples, use timex format date:

```go
tx := timex.Now()
date := tx.DateFormat("Y-m-d H:I:S") // Output: 2022-04-20 19:09:03
date = tx.DateFormat("y-m-d h:i:s") // Output: 22-04-20 07:9:3
```

**Format time.Time**:

```go
tx := time.Now()
date := timex.DateFormat(tx, "Y-m-d H:I:S") // Output: 2022-04-20 19:40:34
```

More usage:

```go
ts := timex.NowUnix() // current unix timestamp

date := FormatUnix(ts, "2006-01-02 15:04:05") // Get: 2022-04-20 19:40:34
date := FormatUnixByTpl(ts, "Y-m-d H:I:S") // Get: 2022-04-20 19:40:34
```

## Functions

```go
func AddDay(t time.Time, day int) time.Time
func AddHour(t time.Time, hour int) time.Time
func AddMinutes(t time.Time, minutes int) time.Time
func AddSeconds(t time.Time, seconds int) time.Time
func Date(t time.Time, template string) string
func DateFormat(t time.Time, template string) string
func DayEnd(t time.Time) time.Time
func DayStart(t time.Time) time.Time
func Format(t time.Time) string
func FormatBy(t time.Time, layout string) string
func FormatByTpl(t time.Time, template string) string
func FormatUnix(sec int64) string
func FormatUnixBy(sec int64, layout string) string
func FormatUnixByTpl(sec int64, template string) string
func HourEnd(t time.Time) time.Time
func HourStart(t time.Time) time.Time
func HowLongAgo(sec int64) string
func NowAddDay(day int) time.Time
func NowAddHour(hour int) time.Time
func NowAddMinutes(minutes int) time.Time
func NowAddSeconds(seconds int) time.Time
func NowHourEnd() time.Time
func NowHourStart() time.Time
func NowUnix() int64
func SetLocalByName(tzName string) error
func ToDuration(s string) (time.Duration, error)
func ToLayout(template string) string
func TodayEnd() time.Time
func TodayStart() time.Time

// for create timex.Time
    func FromDate(s string, template ...string) (*Time, error)
    func FromString(s string, layouts ...string) (*Time, error)
    func FromTime(t time.Time) *Time
    func FromUnix(sec int64) *Time
    func Local() *Time
    func LocalByName(tzName string) *Time
    func New(t time.Time) *Time
    func Now() *Time
    func Wrap(t time.Time) *Time
```

### Methods in timex.Time

```go
func (t *Time) AddDay(day int) *Time
func (t *Time) AddHour(hours int) *Time
func (t *Time) AddMinutes(minutes int) *Time
func (t *Time) AddSeconds(seconds int) *Time
func (t *Time) CustomHMS(hour, min, sec int) *Time
func (t *Time) DateFormat(template string) string
func (t *Time) Datetime() string
func (t *Time) DayAfter(day int) *Time
func (t *Time) DayAgo(day int) *Time
func (t *Time) DayEnd() *Time
func (t *Time) DayStart() *Time
func (t Time) Diff(u time.Time) time.Duration
func (t Time) DiffSec(u time.Time) int
func (t *Time) Format(layout string) string
func (t *Time) HourEnd() *Time
func (t *Time) HourStart() *Time
func (t Time) HowLongAgo(before time.Time) string
func (t *Time) IsAfter(u time.Time) bool
func (t *Time) IsAfterUnix(ux int64) bool
func (t *Time) IsBefore(u time.Time) bool
func (t *Time) IsBeforeUnix(ux int64) bool
func (t *Time) SubDay(day int) *Time
func (t *Time) SubHour(hours int) *Time
func (t *Time) SubMinutes(minutes int) *Time
func (t *Time) SubSeconds(seconds int) *Time
func (t Time) SubUnix(u time.Time) int
func (t Time) T() time.Time
func (t Time) Timestamp() int64
func (t *Time) Tomorrow() *Time
func (t *Time) TplFormat(template string) string
func (t *Time) UnmarshalJSON(data []byte) error
func (t *Time) UnmarshalText(data []byte) error
func (t *Time) Yesterday() *Time
```

## Code Check & Testing

```bash
gofmt -w -l ./
golint ./...
```

**Testing**:

```shell
go test -v ./timex/...
```

**Test limit by regexp**:

```shell
go test -v -run ^TestSetByKeys ./timex/...
```
