package version

import (
	"fmt"
	"runtime"
)

var (
	Version string

	//will be overwritten automatically by the build system
	GitCommit       string
	GoVersion       string
	BuildTime       string
	ExGoVersionInfo string
)

func FullVersion() string {
	return fmt.Sprintf("Version: %s \nGit commit: %s \nGo version: %s \nBuild time: %s \n",
		Version, GitCommit, GetGoVersion(), BuildTime)
}
func GetGoVersion() string {
	return fmt.Sprint(GoVersion, "(runtime: ", runtime.Version(), " ", runtime.GOOS, "/", runtime.GOARCH, ExGoVersionInfo, ")")
}
func AppVersion() string {
	return fmt.Sprintf(`
	##       ##         ##        ##   ##       ##       ## ##     ## ## ##      ## ## 
	##       ##       ## ##     ## ##  ##       ##    ##      ##      ##      ##      ##
	##       ##      ##  ##    ##  ##  ##       ##   ##               ##     ##
	##       ##     ##   ##   ##   ##  ##       ##    ## ## ##        ##     ##
	##       ##    ##    ##  ##    ##  ##       ##            ##      ##     ## 
	##       ##   ##     ## ##     ##  ##       ##  ##        ##      ##      ##      ##
	## ## ## ##  ##      ####      ##  ## ## ## ##   ## ## ##      ## ## ##    ## ## ##
		
                 %s`+"  by cnsilvan(https://github.com/cnsilvan/UnblockNeteaseMusic) \n", Version)
}
