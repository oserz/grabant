package main

import (
	"github.com/oserz/grabant/Crawler"
	"github.com/oserz/grabant/RuleCfg"
	"github.com/oserz/grabant/Util"

	"flag"
	"fmt"
	"os"
	"runtime"
)

var (
	rulefile = flag.String("rule", "", "Path of rule File.")
)

func main() {

	flag.Usage = usage
	flag.Parse()
	//	narg := flag.NArg()
	//	if narg < 1 {
	//		fmt.Println("too few args")
	//		usage()
	//	}
	fmt.Println("GrabAnt Start, args:", *rulefile)
	runtime.GOMAXPROCS(runtime.NumCPU())
	Util.Mainlogger.Info("Log start")

	Crawler.InitCfg()
	RuleCfg.MainCfg.InitRules(Util.Mainlogger, *rulefile)

}

func usage() {
	flag.PrintDefaults()
	os.Exit(2)
}
