package color

/*************************************************************
 * quick use color print message
 *************************************************************/

// Redp print message with Red color
func Redp(a ...interface{}) { Red.Print(a...) }

// Redf print message with Red color
func Redf(format string, a ...interface{}) { Red.Printf(format, a...) }

// Redln print message line with Red color
func Redln(a ...interface{}) { Red.Println(a...) }

// Bluep print message with Blue color
func Bluep(a ...interface{}) { Blue.Print(a...) }

// Bluef print message with Blue color
func Bluef(format string, a ...interface{}) { Blue.Printf(format, a...) }

// Blueln print message line with Blue color
func Blueln(a ...interface{}) { Blue.Println(a...) }

// Cyanp print message with Cyan color
func Cyanp(a ...interface{}) { Cyan.Print(a...) }

// Cyanf print message with Cyan color
func Cyanf(format string, a ...interface{}) { Cyan.Printf(format, a...) }

// Cyanln print message line with Cyan color
func Cyanln(a ...interface{}) { Cyan.Println(a...) }

// Grayp print message with Gray color
func Grayp(a ...interface{}) { Gray.Print(a...) }

// Grayf print message with Gray color
func Grayf(format string, a ...interface{}) { Gray.Printf(format, a...) }

// Grayln print message line with Gray color
func Grayln(a ...interface{}) { Gray.Println(a...) }

// Greenp print message with Green color
func Greenp(a ...interface{}) { Green.Print(a...) }

// Greenf print message with Green color
func Greenf(format string, a ...interface{}) { Green.Printf(format, a...) }

// Greenln print message line with Green color
func Greenln(a ...interface{}) { Green.Println(a...) }

// Yellowp print message with Yellow color
func Yellowp(a ...interface{}) { Yellow.Print(a...) }

// Yellowf print message with Yellow color
func Yellowf(format string, a ...interface{}) { Yellow.Printf(format, a...) }

// Yellowln print message line with Yellow color
func Yellowln(a ...interface{}) { Yellow.Println(a...) }

// Magentap print message with Magenta color
func Magentap(a ...interface{}) { Magenta.Print(a...) }

// Magentaf print message with Magenta color
func Magentaf(format string, a ...interface{}) { Magenta.Printf(format, a...) }

// Magentaln print message line with Magenta color
func Magentaln(a ...interface{}) { Magenta.Println(a...) }

/*************************************************************
 * quick use style print message
 *************************************************************/

// Infop print message with Info color
func Infop(a ...interface{}) { Info.Print(a...) }

// Infof print message with Info style
func Infof(format string, a ...interface{}) { Info.Printf(format, a...) }

// Infoln print message with Info style
func Infoln(a ...interface{}) { Info.Println(a...) }

// Successp print message with success color
func Successp(a ...interface{}) { Success.Print(a...) }

// Successf print message with success style
func Successf(format string, a ...interface{}) { Success.Printf(format, a...) }

// Successln print message with success style
func Successln(a ...interface{}) { Success.Println(a...) }

// Errorp print message with Error color
func Errorp(a ...interface{}) { Error.Print(a...) }

// Errorf print message with Error style
func Errorf(format string, a ...interface{}) { Error.Printf(format, a...) }

// Errorln print message with Error style
func Errorln(a ...interface{}) { Error.Println(a...) }

// Warnp print message with Warn color
func Warnp(a ...interface{}) { Warn.Print(a...) }

// Warnf print message with Warn style
func Warnf(format string, a ...interface{}) { Warn.Printf(format, a...) }

// Warnln print message with Warn style
func Warnln(a ...interface{}) { Warn.Println(a...) }
