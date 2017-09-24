var configs = {
    domains: ["leiphone.com"],// 网站域名，设置域名后只处理这些域名下的网页
    scanurls: ["https://www.leiphone.com/search?s=vr&site=article"],// 入口页链接，分别从这些链接开始爬取
    contenturlregexes: [
        "https://www\\.leiphone\\.com/news/\\d+/.+\\.html"
    ],// 内容页url的正则，符合这些正则的页面会被当作内容页处理
    helperurlregexes: [
        "https://www\\.leiphone\\.com/search\\?s=vr&site=article(&page=\\d+)?"
    ],// 列表页url的正则，符合这些正则的页面会被当作列表页处理
    fields: [
        {
            // 抽取内容页的文章标题
            name: "article_title",
            selector: "//h1[contains(@class,'headTit')]",
            required: true  // required为true表示该项数据不能为空
        },
        {
            // 抽取内容页的文章内容
            name: "article_content",
            selector: "//div[contains(@class,'lph-article-comView')]",
            required: true
        },
        {
            // 抽取内容页的文章发布日期
            name: "article_publish_time",
            selector: "//td[contains(@class,'time')]",
            required: true
        },
        {
            // 抽取内容页的文章作者
            name: "article_author",
            selector: "//td[contains(@class,'aut')]/a",
            required: true
        }
    ]
};

 
// 使用以上配置创建一个爬虫对象
var crawler = new Crawler(configs);
// 启动该爬虫
crawler.start();

