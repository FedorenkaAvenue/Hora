package bot

import (
	"sync"
)

type scrapResult struct {
	value  string
	failed bool
}

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
			go b.notifier.Post(res.value)
		}
	}
}
