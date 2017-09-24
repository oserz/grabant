package RuleCfg

import (
	"io/ioutil"
	"strconv"

	"github.com/oserz/grabant/Util"

	//	"github.com/bitly/go-simplejson"

	"github.com/robertkrimen/otto"
)

const (
	CfgDomain            string = "domains"
	CfgScanUrls          string = "scanurls"
	CfgContentUrl        string = "contenturlregexes"
	CfgHelperUrl         string = "helperurlregexes"
	CfgFiled             string = "fields"
	CfgFiledName         string = "name"
	CfgFiledSelector     string = "selector"
	CfgFiledSelectorType string = "selectortype"
	CfgFiledreq          string = "required"
	CfgInterval          string = "interval"
)

const (
	FiledSelecotrTypeXpath = iota
	FiledSelecotrTypeRegex
)

var (
	MapRuleConfig = make(map[string]interface{})
	MainCfg       = New()
)

type StCfgRule struct {
	RuleName    string
	ConfigObjet *otto.Object
	logger      *Util.Logger
	InitOK      bool
}

var OttoVM = otto.New()

func New() *StCfgRule {
	return &StCfgRule{}
}

func (this *StCfgRule) InitRules(logger *Util.Logger, rulefile string) (err error) {
	this.logger = logger
	SelectorType, _ := OttoVM.Object(`({XPath:0,Regex:1})`)
	OttoVM.Set("SelectorType", SelectorType)
	err = this.loadScript(rulefile)
	return
}

func (this *StCfgRule) initObjString(cfgString string) {
	ov, err := this.ConfigObjet.Get(cfgString)
	if err != nil {
		this.logger.Info("error can't find configs object")
		return
	}
	config, _ := ov.ToString()
	if config != "" && config != "undefined" {
		MapRuleConfig[cfgString] = config
	}
}

func (this *StCfgRule) initObjInt(cfgString string) {
	ov, err := this.ConfigObjet.Get(cfgString)
	if err != nil {
		this.logger.Info("error can't find configs object")
		return
	}
	config, _ := ov.ToInteger()
	if int(config) != 0 {
		MapRuleConfig[cfgString] = int(config)
	}
}

func (this *StCfgRule) initObjStringArray(cfgString string) {

	ov, err := this.ConfigObjet.Get(cfgString)
	if err != nil {
		this.logger.Info("error can't find configs object")
		return
	}
	config := ov.Object()
	if config == nil {
		return
	}
	len, _ := config.Get("length")
	n, _ := len.ToInteger()
	var sstr []string
	for i := 0; i < int(n); i++ {
		v, _ := config.Get(string(strconv.Itoa(i)))
		vstr, _ := v.ToString()
		sstr = append(sstr, vstr)
	}
	MapRuleConfig[cfgString] = sstr

}

func (this *StCfgRule) LazyInit(obj *otto.Object) {

	//	configsObject, err := OttoVM.Get("configs")

	//	if err == nil && configsObject.IsObject() {
	// obj := configsObject.Object()
	this.ConfigObjet = obj

	this.initObjStringArray(CfgContentUrl)
	this.initObjStringArray(CfgHelperUrl)
	this.initObjStringArray(CfgDomain)
	this.initObjStringArray(CfgScanUrls)
	this.initObjInt(CfgInterval)

	ov, err := this.ConfigObjet.Get(CfgFiled)
	if err != nil {
		this.logger.Info("error can't find configs object")
		return
	}

	var arraymapfield []interface{}
	config := ov.Object()
	len, _ := config.Get("length")
	n, _ := len.ToInteger()
	for i := 0; i < int(n); i++ {
		mapfield := make(map[string]interface{})
		v, _ := config.Get(string(strconv.Itoa(i)))
		vobj := v.Object()

		vf, err2 := vobj.Get(CfgFiledName)
		if err2 == nil {
			p, _ := vf.ToString()
			if p != "undefined" {
				mapfield[CfgFiledName] = p
			}
		}

		vf, err2 = vobj.Get(CfgFiledSelector)
		if err2 == nil {
			p, _ := vf.ToString()
			if p != "undefined" {
				mapfield[CfgFiledSelector] = p
			}
		}

		vf, err2 = vobj.Get(CfgFiledSelectorType)
		if err2 == nil {
			p, _ := vf.ToInteger()
			if p != 0 {
				mapfield[CfgFiledSelectorType] = int(p)
			}
		}

		vf, err2 = vobj.Get(CfgFiledreq)
		if err2 == nil {
			p, _ := vf.ToBoolean()
			mapfield[CfgFiledreq] = p
		}

		arraymapfield = append(arraymapfield, mapfield)
	}
	MapRuleConfig[CfgFiled] = arraymapfield

	this.InitOK = true
	//}

}

func (this *StCfgRule) loadScript(rulefile string) (err error) {
	this.logger.Info("start load Rule: " + rulefile)
	filebuffer, err := ioutil.ReadFile(rulefile)

	if err == nil {
		OttoVM.Run(filebuffer)

		/*
			js, ee := simplejson.NewJson([]byte(configstring))
			if ee == nil {
				sst := SmallStash{}
				//			sst.PlugName = info.Name()
				sst.Config = js
				this.logger.Info("load Rule ok")
				//this.SPlugin = append(this.SPlugin, sst)
			}
			return ee
		*/
	}
	if err != nil {
		this.logger.Error("load Rules fail", err)
	}
	return
}
