# Grabant
## go语言爬虫框架，自定义javascript规则

Grabant是一个用go语言开发，用javascript语法书写规则的爬虫框架
规则开发和神箭手的爬虫规则类似，开发的目的是兼容大部分神箭手已有的规则

## 说明

一些相关解释及说明如下：
* **扫描页**是直接配置的一级页面，这个页面扫描的结果是抓取列表页;
* **列表页**是扫描后的二级页面，这个页面是为了抓取内容页;
* **内容页**才是真正匹配我们需要内容的规则的页面，内容页会匹配我们需要的内容并输出;
* 每次扫描页面，会抽取页面的链接放入队列进行规则匹配，列表页规则如果为空则列表页抽取链接全放入列表队列，同理内容页，从内容页中抽取的数据是以xpath或者正则匹配的规则

举个栗子：
如下是一个简易的爬豆瓣电影评分的规则
```javascript
var configs = {
    domains: ["movie.douban.com"],
    interval: 3000,
    scanurls: ["https://movie.douban.com/cinema/nowplaying/shenzhen/"],
    helperurlregexes: ["https://movie\\.douban\\.com/subject/\\d+/\\?from=playing_poster"],
    fields: [
        {
            name: "film Name",
            selector: "//*[@id=\"content\"]/h1/span[1]",
            required: true
        },
        {
            name: "Rank",
            selector: "//*[@id=\"interest_sectl\"]/div[1]/div[2]/strong"
            required: true
        }
    ]
};

// 使用以上配置创建一个爬虫对象
var crawler = new Crawler(configs);
// 启动该爬虫
crawler.start();
```
configs是一个json配置对应的字段意义如下：

* domains

定义应用爬取哪些域名下的网页

* interval

爬取页面需要的延时，毫秒

* scanUrls

定义入口页url, 从入口页url开始爬取数据

* contenturlregexes

设置内容页url的正则表达式

* helperurlregexes

设置列表页url的正则表达式

* field

定义一个从内容页中抽取数据的抽取项,包括以下内容

1) name
    
    抽取项的名称

2) selector
    
    抽取项的匹配规则，可以是正则表达式或者Xpath

3) selectortype

    定义抽取区的类型,正则为SelectorType.Regex, XPath为SelectorType.XPath,如果不设置默认为XPath

3) required
    
    bool类型，如果为true，则此项必须存在才爬取此数据

## 使用方法
   
   grabant -rule /路径/规则文件

## 编译
   
    go get github.com/robertkrimen/otto
    
    go get github.com/bitly/go-simplejson

    go build

### v0.01

此版本是一个原型版本，目前还有很多判断不严密的可能引起crash的问题;

和神箭手不同之出目前有两个，一是configs项区分大小写，需要全为小写，二fields中selector只支持字符正则;

现在此版本还未实现多个对象及回调方法，只支持最简单的json规则.