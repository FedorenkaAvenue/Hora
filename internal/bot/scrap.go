package bot

import (
	"sync"
)

type scrapResult struct {
	value  scrapResultValue
	failed bool
}

type scrapResultValue []string

func (b Bot) scrap() {
	ch := make(chan scrapResult)
	var wg sync.WaitGroup

	for _, target := range b.config.Targets {
		wg.Add(1)
		go b.parse(target, ch)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for res := range ch {
		wg.Done()

		if !res.failed {
			for _, v := range res.value {
				go b.notifier.Post(v)
			}
		}
	}
}
