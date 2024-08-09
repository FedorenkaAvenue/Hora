package bot

import (
	"io"
	"net/http"

	goDom "github.com/bringmetheaugust/goDOM"
)

func (b Bot) parse(t target, ch chan scrapResult) {
	resp, err := http.Get(t.Url)
	defer resp.Body.Close()

	if err != nil {
		go b.log.Error("During http request.", err, t)
		ch <- scrapResult{failed: true}
		return
	}

	bytes, _ := io.ReadAll(resp.Body)
	document, err := goDom.Create(bytes)

	if err != nil {
		go b.log.Error("During create document.", err, t)
		ch <- scrapResult{failed: true}
		return
	}

	el, err := document.QuerySelector(t.Query)

	if err != nil {
		go b.log.Warning("Element not found. ", t)
		ch <- scrapResult{failed: true}
		return
	}

	attr, err := el.GetAttribute(t.Attr)

	if err != nil {
		go b.log.Warning("Attribute not found.", t)
		return
	}

	ch <- scrapResult{value: attr}
}
