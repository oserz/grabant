package Crawler

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/oserz/grabant/RuleCfg"
	"github.com/oserz/grabant/Util"

	"github.com/PuerkitoBio/goquery"
	"github.com/bitly/go-simplejson"
	xmlpath "gopkg.in/xmlpath.v2"
)

const (
	PageStart = iota
	PageHepler
	PageContent
)

type outType interface {
	InitOutput(head []string)
	Out(js *simplejson.Json)
	//	Wait()
}

type Page struct {
	domain    string
	Url       string
	PageType  int
	HandleOk  bool
	FetchLink []string
	Content   string
}

type FiledStash struct {
	name         string
	selector     string
	selectorType int
	required     bool
	xPath        *xmlpath.Path
	reg          *regexp.Regexp
}

type CrawlerCore struct {
	scanURL    []string
	Domains    []string
	helperURL  []string
	contentURL []string
	helperReg  []*regexp.Regexp
	contentReg []*regexp.Regexp
	FiledArray []FiledStash
	//	CfgRule        *RuleCfg.SmallStash
	goroutineNum   int
	handlequeueNum int
	interval       int
	mapHandledURL  map[string]bool
	fetchlinks     chan *Page
	worklist       chan []*Page
	logger         *Util.Logger
	output         outType
}

func New() *CrawlerCore {
	return &CrawlerCore{worklist: make(chan []*Page),
		fetchlinks:    make(chan *Page),
		mapHandledURL: make(map[string]bool)}
}

func (this *CrawlerCore) InitConfig(logger *Util.Logger) {
	var err error
	this.logger = logger
	this.logger.Info("Init Crawler")
	scanConfig, _ := RuleCfg.MapRuleConfig[RuleCfg.CfgScanUrls].([]string)
	scanDomains, _ := RuleCfg.MapRuleConfig[RuleCfg.CfgDomain].([]string)
	for _, v := range scanConfig {
		this.logger.Info(v)
		this.scanURL = append(this.scanURL, strings.ToLower(v))
	}
	for _, v := range scanDomains {
		this.logger.Info(v)
		this.Domains = append(this.Domains, strings.ToLower(v))
	}

	this.helperURL, _ = RuleCfg.MapRuleConfig[RuleCfg.CfgHelperUrl].([]string)
	if len(this.helperURL) == 0 {
		this.logger.Error("No helperUrlRegexes")
	}
	this.contentURL, _ = RuleCfg.MapRuleConfig[RuleCfg.CfgContentUrl].([]string)
	if len(this.contentURL) == 0 {
		this.logger.Error("Get contentUrlRegexes error:", err)
	}

	if len(this.helperURL) != 0 {
		for _, v := range this.helperURL {
			regc, err := regexp.Compile(v)
			if err != nil {
				this.logger.Error("hepler Url Regexp compile error:", err)
			} else {
				this.helperReg = append(this.helperReg, regc)
			}
		}
	}
	if len(this.contentURL) != 0 {
		for _, v := range this.contentURL {
			regc, err := regexp.Compile(v)
			if err != nil {
				this.logger.Error("content Url Regexp compile error:", err)
			} else {
				this.contentReg = append(this.contentReg, regc)
			}
		}
	}

	this.output = Util.NewOutputCSV(100, "output.csv")
	var headsname []string
	FiledArry, _ := RuleCfg.MapRuleConfig[RuleCfg.CfgFiled].([]interface{})
	for _, v := range FiledArry {
		if m, ok := v.(map[string]interface{}); ok {
			var tmpFiled FiledStash

			tmpFiled.required = false
			for k, vv := range m {
				k = strings.ToLower(k)
				switch k {
				case RuleCfg.CfgFiledName:
					tmpFiled.name, _ = vv.(string)
					headsname = append(headsname, tmpFiled.name)
				case RuleCfg.CfgFiledSelector:
					tmpFiled.selector, _ = vv.(string)
				case RuleCfg.CfgFiledSelectorType:
					tmpFiled.selectorType, _ = vv.(int)
				case RuleCfg.CfgFiledreq:
					tmpFiled.required, _ = vv.(bool)
				}
			}
			var CompileErr error
			switch tmpFiled.selectorType {
			case RuleCfg.FiledSelecotrTypeXpath:
				tmpFiled.xPath, CompileErr = xmlpath.Compile(tmpFiled.selector)
			case RuleCfg.FiledSelecotrTypeRegex:
				tmpFiled.reg, CompileErr = regexp.Compile(tmpFiled.selector)
			}
			if CompileErr == nil {
				this.FiledArray = append(this.FiledArray, tmpFiled)
			}
		}
	}
	this.interval, _ = RuleCfg.MapRuleConfig[RuleCfg.CfgInterval].(int)
	headsname = append(headsname, "URL")
	this.output.InitOutput(headsname)
	if this.interval == 0 {
		this.goroutineNum = 100
	} else {
		this.goroutineNum = 1
	}
	//	this.CfgRule = ru
	this.logger.Info("Init Crawler OK!")
}

func isUrlInDomain(domains string, url string) (retOK bool) {
	/*
		var newurl, newdomains string
		retOK = true
		if strings.HasPrefix(url, "http") == false {
			newurl = "http://" + url
		} else {
			newurl = url
		}
		if strings.HasPrefix(domains, "http") == false {
			newdomains = newurl[:strings.Index(newurl, "/")+2] + domains
		} else {
			newdomains = domains
		}
		if strings.HasPrefix(newurl, newdomains) == false {
			retOK = false
		}
		return
	*/
	newdomain := domains
	reqdomain, _ := http.NewRequest("GET", domains, nil)
	req, _ := http.NewRequest("GET", url, nil)
	if reqdomain.Host != "" {
		newdomain = reqdomain.Host
	}
	cmp1 := strings.Split(newdomain, ".")
	cmp2 := strings.Split(req.Host, ".")
	m := len(cmp2) - 1
	n := len(cmp1) - 1
	if m < n {
		return false
	}
	retOK = true
	for ; n >= 0; n-- {
		if strings.ToLower(cmp1[n]) != strings.ToLower(cmp2[m]) {
			retOK = false
			break
		}
		m = m - 1
	}
	return
}

func UniqList(list *[]string) []string {
	var x []string
	for _, ll := range *list {
		if len(x) == 0 {
			x = append(x, ll)
		} else {
			for k, v := range x {
				if ll == v {
					break
				}
				if k == len(x)-1 {
					x = append(x, ll)
				}
			}
		}
	}
	return x
}

func (this *CrawlerCore) absUrl(domains string, relativeUrl string, oldUrl string) (newUrl string) {
	var (
		urlprefix    string
		relativePath string
	)
	/*
		defer func() {
			if isUrlInDomain(domains, newUrl) == false {
				newUrl = ""
				return
			}
		}()
	*/
	if strings.HasPrefix(oldUrl, "http://") || strings.HasPrefix(oldUrl, "https://") {
		return oldUrl
	}
	if strings.Contains(oldUrl, ":") {
		return ""
	}

	if strings.HasPrefix(relativeUrl, "https://") || strings.HasPrefix(relativeUrl, "http://") {
		index := strings.LastIndex(relativeUrl, "/")
		if index != -1 && index != 6 && index != 7 {
			relativePath = relativeUrl[:index+1]
		} else {
			relativePath = relativeUrl + "/"
		}
		urlprefix = relativeUrl[:strings.Index(relativePath, "/")+2]
	} else {
		index := strings.LastIndex(relativeUrl, "/")
		if index != -1 {
			relativePath = "http://" + relativeUrl[:index+1]
		} else {
			relativePath = "http://" + relativeUrl + "/"
		}
		urlprefix = "http://"
	}

	var newdomains string
	if string(domains[len(domains)-1]) == "/" {
		newdomains = domains[:len(domains)-1]
	} else {
		newdomains = domains
	}
	if string(oldUrl[0]) == "/" {
		newUrl = urlprefix + newdomains + oldUrl
	} else {
		newUrl = relativePath + oldUrl
	}

	return
}

func (this *CrawlerCore) BuildLinkPage(oldpage *Page) (newpage *Page, err error) {
	if oldpage.HandleOk == false {
		newpage, err = this.BuildPage(oldpage.Url, oldpage.domain, oldpage.PageType, true)
	}
	return
}

func (this *CrawlerCore) BuildNoLinkPage(oldpage *Page) (newpage *Page, err error) {
	if oldpage.HandleOk == true || oldpage.PageType != PageContent {
		return
	}
	newpage, err = this.BuildPage(oldpage.Url, oldpage.domain, oldpage.PageType, true)
	return
}

func (this *CrawlerCore) BuildPage(url string, domains string, urlType int, needLinks bool) (urlpage *Page, err error) {
	var resp *http.Response
	tmpPage := &Page{Url: url, domain: domains, PageType: urlType}
	if isUrlInDomain(domains, url) == false {
		return
	}
	if needLinks {
		resp, err = http.Get(url)
		if this.interval != 0 {
			time.Sleep(time.Duration(100 * this.interval))
		}
		if resp != nil && resp.Body != nil {
			defer resp.Body.Close()
		} else {
			this.logger.Error("ERROR " + url + " return nil ")
		}

		if resp == nil || resp.Body == nil || err != nil || resp.StatusCode != http.StatusOK {
			this.logger.Error("ERROR "+url, err)
			return
		}

		var buf []byte
		buf, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}

		if urlType != PageContent {
			root, err1 := goquery.NewDocumentFromReader(strings.NewReader(string(buf)))
			if err1 != nil {
				err = err1
				return
			}

			req, _ := http.NewRequest("GET", url, nil)

			root.Find("a").Each(func(i int, aa *goquery.Selection) {
				href, IsExist := aa.Attr("href")
				if IsExist == true {
					if len(href) > 2 {
						newLink := this.absUrl(req.Host, url, href)
						if newLink != "" {
							tmpPage.FetchLink = append(tmpPage.FetchLink, newLink)
						} //else {
						//	Util.LogDebugger.Println("absUrl none -- domains: " + domains + " url:" + url + " href:" + href)
						//}
					}
				}
			})
			tmpPage.FetchLink = UniqList(&tmpPage.FetchLink)
		}

		tmpPage.Content = string(buf)
	}

	tmpPage.HandleOk = needLinks
	urlpage = tmpPage
	return
}

func (this *CrawlerCore) doHelperURL(pg *Page) (heplerUrl []string) {
	if this.helperReg == nil {
		return pg.FetchLink
	}
	for _, ss := range pg.FetchLink {
		for _, reg := range this.helperReg {
			findss := reg.FindString(ss)
			if findss != "" && len(findss) >= len(ss) && findss[0:len(ss)] == ss {
				heplerUrl = append(heplerUrl, ss)
			}
		}
	}
	return
}

func (this *CrawlerCore) doContentURL(pg *Page) (contentUrl []string) {
	if this.contentReg == nil {
		return pg.FetchLink
	}
	for _, ss := range pg.FetchLink {
		for _, reg := range this.contentReg {
			if reg.FindString(ss) == ss {
				contentUrl = append(contentUrl, ss)
			}
		}
	}
	return

}

func (this *CrawlerCore) dispatchGo() {

	var contentwg sync.WaitGroup

	for i := 0; i < this.goroutineNum; i++ {
		go func() {
			for lfetched := range this.fetchlinks {
				var slcpg []*Page
				//Util.LogDebugger.Println("test for go routine entry get:"+lfetched.Url+" Type:"+strconv.Itoa(lfetched.PageType))

				switch lfetched.PageType {
				case PageStart:
					LinkPage, err := this.BuildLinkPage(lfetched)
					if err == nil && LinkPage != nil {
						var sshelperUrl []string
						if len(this.helperURL) != 0 {
							sshelperUrl = this.doHelperURL(LinkPage)
						} else {
							sshelperUrl = LinkPage.FetchLink
						}
						for _, ss := range sshelperUrl {
							pg, err := this.BuildPage(ss, LinkPage.domain, PageHepler, false)
							if err == nil && pg != nil {
								slcpg = append(slcpg, pg)
							}
						}
						//this.worklist <- slcpg //don't use this,may dead lock
						this.AddWorkList(slcpg)
					}

				case PageHepler:
					LinkPage, err := this.BuildLinkPage(lfetched)
					if err == nil && LinkPage != nil {
						var ssContentUrl []string
						if len(this.contentURL) != 0 {
							ssContentUrl = this.doContentURL(LinkPage)
						} else {
							ssContentUrl = LinkPage.FetchLink
						}
						for _, ss := range ssContentUrl {
							pg, err := this.BuildPage(ss, LinkPage.domain, PageContent, false)
							if err == nil && pg != nil {
								slcpg = append(slcpg, pg)
							}
						}
						//this.worklist <- slcpg
						this.AddWorkList(slcpg)
					}

				case PageContent:
					contentwg.Add(1)
					NoLinkPage, err := this.BuildNoLinkPage(lfetched)
					if err == nil && NoLinkPage != nil {
						available := true
						jsContent := simplejson.New()
						for _, ff := range this.FiledArray {
							var content string
							if ff.xPath != nil {
								node, err := xmlpath.ParseHTML(strings.NewReader(NoLinkPage.Content))
								if err == nil {
									content, _ = ff.xPath.String(node)
								}
							}
							if ff.reg != nil {
								content = ff.reg.FindString(NoLinkPage.Content)
							}
							if ff.required == true && content == "" {
								available = false
							}
							if /*bok || content != ""*/ available {
								jsContent.Set(ff.name, content)
								jsContent.Set("URL", lfetched.Url)
							}
						}
						if available {
							this.output.Out(jsContent)
							this.logger.Info("page content:", jsContent)
						} else {
							this.logger.Info("no elements found, content URL is:", lfetched.Url)
						}
					}
					contentwg.Done()
				}
			}
		}()
	}

	//	for worklst := range this.worklist {
	for ; this.handlequeueNum > 0; this.handlequeueNum-- {
		worklst := <-this.worklist
		if worklst != nil {
			for _, workpage := range worklst {

				if this.mapHandledURL[strconv.Itoa(workpage.PageType)+workpage.Url] == false {
					this.mapHandledURL[strconv.Itoa(workpage.PageType)+workpage.Url] = true

					//go func() { //use this may cause insert wrong fetchlinks
					this.fetchlinks <- workpage
					//}()
					if workpage.PageType != PageContent {
						this.handlequeueNum++
					}
				}
			}
		}
	}

	contentwg.Wait()

	this.logger.Info("find all links")

}

func (this *CrawlerCore) AddWorkList(pg []*Page) {
	//this.handlequeueNum = this.handlequeueNum + 1
	go func() {
		this.worklist <- pg
	}()
}

func (this *CrawlerCore) GoRun() {
	for _, ssDomains := range this.Domains {
		for _, ssUrl := range this.scanURL {

			var slcpg []*Page
			pg, err := this.BuildPage(ssUrl, ssDomains, PageStart, false)
			if err != nil {
				this.logger.Error("ERROR "+ssUrl, err)
				return
			}

			slcpg = append(slcpg, pg)
			this.AddWorkList(slcpg)
			this.handlequeueNum = 1
			this.dispatchGo()
		}
	}
}
