package progress

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/gookit/goutil/fmtutil"
)

var builtinWidgets = map[string]WidgetFunc{
	"elapsed": func(p *Progress) string { // 消耗时间
		// fmt.Sprintf("%.3f", time.Since(startTime).Seconds()*1000)
		elapsed := time.Since(p.StartedAt()).Seconds()
		return fmtutil.HowLongAgo(int64(elapsed))
	},
	"remaining": func(p *Progress) string { // 剩余时间
		step := p.Progress() // current progress

		// not set max steps OR current progress is 0
		if p.MaxSteps == 0 || step == 0 {
			return "unknown"
		}

		// get elapsed time
		elapsed := int64(time.Since(p.StartedAt()).Seconds())
		// calc remaining time
		remaining := uint(elapsed) / step * (p.MaxSteps - step)
		return fmtutil.HowLongAgo(int64(remaining))
	},
	"estimated": func(p *Progress) string { // 计算总的预计时间
		step := p.Progress() // current progress

		// not set max steps OR current progress is 0
		if p.MaxSteps == 0 || step == 0 {
			return "unknown"
		}

		// get elapsed time
		elapsed := int64(time.Since(p.StartedAt()).Seconds())
		// calc estimated time
		estimated := float32(elapsed) / float32(step) * float32(p.MaxSteps)

		return fmtutil.HowLongAgo(int64(estimated))
	},
	"memory": func(p *Progress) string { // Memory consumption
		mem := new(runtime.MemStats)
		runtime.ReadMemStats(mem)
		return fmtutil.DataSize(mem.Sys)
	},
	"max": func(p *Progress) string {
		return fmt.Sprint(p.MaxSteps)
	},
	"current": func(p *Progress) string {
		step := fmt.Sprint(p.Progress())
		width := fmt.Sprint(p.StepWidth)
		diff := len(width) - len(step)
		if diff <= 0 {
			return step
		}

		return strings.Repeat(" ", diff) + step
	},
	"percent": func(p *Progress) string {
		return fmt.Sprintf("%.1f", p.Percent()*100)
	},
}

// DynamicTextWidget dynamic text message widget for progress bar.
// for param messages: int is percent, range is 0 - 100. value is message string.
// Usage please example.
func DynamicTextWidget(messages map[int]string) WidgetFunc {
	var numbers []int
	for val := range messages {
		numbers = append(numbers, val)
	}

	// sort
	sort.Ints(numbers)

	return func(p *Progress) string {
		percent := int(p.Percent() * 100)
		for _, val := range numbers {
			if percent <= val {
				return messages[val]
			}
		}

		return " Handling ..." // Should never happen
	}
}

// LoadingWidget create a loading progress widget
func LoadingWidget(chars []rune) WidgetFunc {
	builder := loadingCharBuilder(chars)

	return func(_ *Progress) string {
		return builder()
	}
}

// RoundTripWidget create a round-trip widget for progress bar.
//
// Output like `[  ====   ]`
func RoundTripWidget(char rune, charNum, boxWidth int) WidgetFunc {
	builder := roundTripTextBuilder(char, charNum, boxWidth)

	return func(_ *Progress) string {
		return builder()
	}
}

// BarWidget create a progress bar widget.
//
// Output like `[==============>-------------]`
func BarWidget(width int, cs BarChars) WidgetFunc {
	if width < 1 {
		width = BarWidth
	}

	if cs.Completed == 0 {
		cs.Completed = CharWell
	}

	return func(p *Progress) string {
		var completeLen float32

		if p.MaxSteps > 0 { // MaxSteps is valid
			completeLen = p.percent * float32(width)
		} else { // not set MaxSteps
			completeLen = float32(p.step % uint(width))
		}

		bar := string(repeatRune(cs.Completed, int(completeLen)))

		if diff := width - int(completeLen); diff > 0 {
			bar += string(cs.Processing) + string(repeatRune(cs.Remaining, diff-1))
		}

		return bar
	}
}

func loadingCharBuilder(chars []rune) func() string {
	if len(chars) == 0 {
		chars = RandomCharsTheme()
	}

	index := 0
	length := len(chars)

	return func() string {
		char := string(chars[index])
		if index+1 == length { // reset
			index = 0
		} else {
			index++
		}

		return char
	}
}

func roundTripTextBuilder(char rune, charNum, boxWidth int) func() string {
	if char == 0 {
		char = CharEqual
	}

	if charNum < 1 {
		charNum = 4
	}

	if boxWidth < 1 {
		boxWidth = 12
	}

	cursor := string(repeatRune(char, charNum))
	// control direction. False: -> True: <->
	direction := false
	// record cursor position
	position := 0

	return func() string {
		var bar string
		if position > 0 {
			bar += strings.Repeat(" ", position)
		}

		bar += cursor + strings.Repeat(" ", boxWidth-position-charNum)

		if direction { // left <-
			if position <= 0 { // begin ->
				direction = false
			} else {
				position--
			}
		} else { // -> right
			if position+charNum >= boxWidth { // begin <-
				direction = true
			} else {
				position++
			}
		}

		return bar
	}
}
