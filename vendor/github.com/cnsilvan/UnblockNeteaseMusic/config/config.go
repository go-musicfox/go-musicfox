package config

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/cnsilvan/UnblockNeteaseMusic/common"
	"github.com/cnsilvan/UnblockNeteaseMusic/utils"
	"github.com/cnsilvan/UnblockNeteaseMusic/version"
)

var (
	Addr               = flag.String("a", "0.0.0.0", "specify server listen address,such as : \"0.0.0.0\"")
	Port               = flag.Int("p", 80, "specify server port,such as : \"80\"")
	TLSPort            = flag.Int("sp", 443, "specify server tls port,such as : \"443\"")
	Source             = flag.String("o", "kuwo", "specify server source,such as : \"kuwo\"")
	CertFile           = flag.String("c", "./server.crt", "specify server cert,such as : \"server.crt\"")
	KeyFile            = flag.String("k", "./server.key", "specify server cert key ,such as : \"server.key\"")
	LogFile            = flag.String("l", "", "specify log file ,such as : \"/var/log/unblockNeteaseMusic.log\"")
	Mode               = flag.Int("m", 1, "specify running mode（1:hosts） ,such as : \"1\"")
	V                  = flag.Bool("v", false, "display version info")
	EndPoint           = flag.Bool("e", false, "enable replace song url")
	ForceBestQuality   = flag.Bool("b", false, "force the best music quality")
	SearchLimit        = flag.Int("sl", 0, "specify the number of songs searched on other platforms(the range is 0 to 3) ,such as : \"1\"")
	BlockUpdate        = flag.Bool("bu", false, "block version update message")
	BlockAds           = flag.Bool("ba", false, "block advertising requests")
	EnableLocalVip     = flag.Bool("lv", false, "enable local vip")
	UnlockSoundEffects = flag.Bool("sef", false, "unlock SoundEffects")
	QQCookieFile       = flag.String("qc", "./qq.cookie", "specify cookies file ,such as : \"qq.cookie\"")
	LogWebTraffic      = flag.Bool("wl", false, "log request url and response")
)

func ValidParams() bool {
	flag.Parse()
	if flag.NArg() > 0 {
		log.Println("--------------------Invalid Params------------------------")
		log.Printf("Invalid params=%s, num=%d\n", flag.Args(), flag.NArg())
		for i := 0; i < flag.NArg(); i++ {
			log.Printf("arg[%d]=%s\n", i, flag.Arg(i))
		}
	}
	if *V {
		// active call should use fmt
		fmt.Println(version.FullVersion())
		return false
	}
	sources := strings.Split(strings.ToLower(*Source), ":")
	if len(sources) < 1 {
		log.Printf("source param invalid: %v \n", *Source)
		return false
	}
	if *SearchLimit < 0 || *SearchLimit > 3 {
		log.Printf("searchLimit param invalid （0-3）: %v \n", *SearchLimit)
		return false
	}
	for _, source := range sources {
		common.Source = append(common.Source, source)
	}
	
	currentPath, err := utils.GetCurrentPath()
	if err != nil {
		log.Println(err)
		currentPath = ""
	}
	// log.Println(currentPath)
	certFile, _ := filepath.Abs(*CertFile)
	keyFile, _ := filepath.Abs(*KeyFile)
	_, err = os.Open(certFile)
	if err != nil {
		certFile, _ = filepath.Abs(currentPath + *CertFile)
	}
	_, err = os.Open(keyFile)
	if err != nil {
		keyFile, _ = filepath.Abs(currentPath + *KeyFile)
	}
	*CertFile = certFile
	*KeyFile = keyFile
	log.SetFlags(log.LstdFlags)
	if len(strings.TrimSpace(*LogFile)) > 0 {
		logFilePath, _ := filepath.Abs(*LogFile)
		logFile, logErr := os.OpenFile(logFilePath, os.O_CREATE|os.O_RDWR|os.O_SYNC|os.O_APPEND, 0666)
		if logErr != nil {
			// log.Println("Fail to find unblockNeteaseMusic.log start Failed")
			// panic(logErr)
			logFilePath, _ = filepath.Abs(currentPath + *LogFile)
		} else {
			logFile.Close()
		}
		*LogFile = logFilePath
		logFile, logErr = os.OpenFile(logFilePath, os.O_CREATE|os.O_RDWR|os.O_SYNC|os.O_APPEND, 0666)
		if logErr != nil {
			log.Println("Fail to find " + logFilePath + " start Failed")
			panic(logErr)
		}
		os.Stdout = logFile
		os.Stderr = logFile
		fileInfo, err := logFile.Stat()
		if err != nil {
			panic(err)
		}
		if (fileInfo.Size() >> 20) > 2 { // 2M
			logFile.Seek(0, io.SeekStart)
			logFile.Truncate(0)
		}
		log.SetOutput(logFile)
	} else {
		log.SetOutput(os.Stdout)
	}
	return true
}
