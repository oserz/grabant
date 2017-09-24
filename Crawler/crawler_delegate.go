package Crawler

import (
	"github.com/oserz/grabant/RuleCfg"
	"github.com/oserz/grabant/Util"
	"github.com/robertkrimen/otto"
)

var (
	CrawerObj   *otto.Object
	MainCrawler = New()
)

func newCrawler(call otto.FunctionCall) otto.Value {
	if call.Argument(0).IsObject() {
		obj := call.Argument(0).Object()
		RuleCfg.MainCfg.LazyInit(obj)
		CrawerObj, _ = RuleCfg.OttoVM.Object(`CrawerObj = {}`)
		CrawerObj.Set("start", CrawlerStart)
		retObj, _ := RuleCfg.OttoVM.ToValue(CrawerObj)
		return retObj
	}
	//	fmt.Printf("Hello, %s.\n", call.Argument(0).String())
	return otto.Value{}
}

func CrawlerStart(call otto.FunctionCall) otto.Value {
	MainCrawler.InitConfig(Util.Mainlogger)
	MainCrawler.GoRun()
	return otto.Value{}
}

func InitCfg() {
	RuleCfg.OttoVM.Set("Crawler", newCrawler)
}
