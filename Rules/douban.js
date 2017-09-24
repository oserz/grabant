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
