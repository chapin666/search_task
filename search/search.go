package search

import (
	"log"
	"sync"
)

var matchers = make(map[string]Matcher)

// Register 调用时， 会注册一个匹配，提供给后面的程序使用
func Register(feedType string, matcher Matcher) {
	if _, exists := matchers[feedType]; exists {
		log.Fatalln(feedType, "Matcher already registered")
	}

	log.Println("Register", feedType, "matcher")
	matchers[feedType] = matcher
}

// Run 执行搜索逻辑
func Run(searchTerm string) {

	// 获取搜索数据源列表
	feeds, err := RetrieveFeeds()
	if err != nil {
		log.Fatal(err)
	}

	// 创建无缓冲通道
	results := make(chan *Result)

	// 构造waitGroup
	var waitGroup sync.WaitGroup

	// 设置需要等待的处理
	// 每个数据源的goroutine数量
	waitGroup.Add(len(feeds))

	// 为每个数据源启动一个goroutine来查找结果
	for _, feed := range feeds {
		matcher, exists := matchers[feed.Type]
		if !exists {
			matcher = matchers["default"]
		}

		go func(matcher Matcher, feed *Feed) {
			Match(matcher, feed, searchTerm, results)
			waitGroup.Done()
		}(matcher, feed)
	}

	// 启动一个goroutine来监控是否所有的工作都做完了
	go func() {
		// 等待所有的任务完成
		waitGroup.Wait()
		// 关闭通道， 通知Display函数
		close(results)
	}()

	// 启动函数，显示返回的结果，并在最后一个结果显示完后返回
	Display(results)
}
