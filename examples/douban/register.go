package douban

import (
	"github.com/astenir/rulecrawl/engine"
	"github.com/astenir/rulecrawl/parse/doubanbook"
	"github.com/astenir/rulecrawl/parse/doubangroup"
	"github.com/astenir/rulecrawl/parse/doubangroupjs"
)

func Register(store *engine.CrawlerStore) {
	store.AddJSTask(doubangroupjs.DoubangroupJSTask)
	store.Add(doubangroup.DoubangroupTask)
	store.Add(doubanbook.DoubanBookTask)
}
