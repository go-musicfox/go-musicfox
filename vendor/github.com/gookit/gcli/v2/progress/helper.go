package progress

import (
	"math/rand"
	"time"
)

func repeatRune(char rune, length int) (chars []rune) {
	for i := 0; i < length; i++ {
		chars = append(chars, char)
	}

	return
}

// CharThemes collection. can use for Progress bar, RoundTripSpinner
var CharThemes = []rune{
	CharEqual,
	CharCenter,
	CharSquare,
	CharSquare1,
	CharSquare2,
}

// GetCharTheme by index number
func GetCharTheme(index int) rune {
	if len(CharThemes) > index {
		return CharThemes[index]
	}

	return RandomCharTheme()
}

// RandomCharTheme get
func RandomCharTheme() rune {
	rand.Seed(time.Now().UnixNano())
	return CharThemes[rand.Intn(len(CharsThemes)-1)]
}

// CharsThemes collection. can use for LoadingBar, LoadingSpinner
var CharsThemes = [][]rune{
	{'卍', '卐'},
	{'☺', '☻'},
	{'░', '▒', '▓'},
	{'⊘', '⊖', '⊕', '⊗'},
	{'◐', '◒', '◓', '◑'},
	{'✣', '✤', '✥', '❉'},
	{'-', '\\', '|', '/'},
	{'▢', '■', '▢', '■'},
	[]rune("▖▘▝▗"),
	[]rune("◢◣◤◥"),
	[]rune("⌞⌟⌝⌜"),
	[]rune("◎●◯◌○⊙"),
	[]rune("◡◡⊙⊙◠◠"),
	[]rune("⇦⇧⇨⇩"),
	[]rune("✳✴✵✶✷✸✹"),
	[]rune("←↖↑↗→↘↓↙"),
	[]rune("➩➪➫➬➭➮➯➱"),
	[]rune("①②③④"),
	[]rune("㊎㊍㊌㊋㊏"),
	[]rune("⣾⣽⣻⢿⡿⣟⣯⣷"),
	[]rune("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"),
	[]rune("▉▊▋▌▍▎▏▎▍▌▋▊▉"),
	[]rune("🌍🌎🌏"),
	[]rune("☰☱☲☳☴☵☶☷"),
	[]rune("⠋⠙⠚⠒⠂⠂⠒⠲⠴⠦⠖⠒⠐⠐⠒⠓⠋"),
	[]rune("🕐🕑🕒🕓🕔🕕🕖🕗🕘🕙🕚🕛"),
}

// GetCharsTheme by index number
func GetCharsTheme(index int) []rune {
	if len(CharsThemes) > index {
		return CharsThemes[index]
	}

	return RandomCharsTheme()
}

// RandomCharsTheme get
func RandomCharsTheme() []rune {
	rand.Seed(time.Now().UnixNano())
	return CharsThemes[rand.Intn(len(CharsThemes)-1)]
}
