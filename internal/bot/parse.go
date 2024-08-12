package bot

import (
	"io"
	"net/http"

	goDom "github.com/bringmetheaugust/goDOM"
)

func (b Bot) parse(t target, ch chan scrapResult) {
	resp, err := http.Get(t.Url)

	if err != nil {
		b.log.Error("During http request.", err, t)
		ch <- scrapResult{failed: true}
		return
	}

	defer resp.Body.Close()

	bytes, _ := io.ReadAll(resp.Body)
	document, err := goDom.Create(bytes)

	if err != nil {
		b.log.Error("During create document.", err, t)
		ch <- scrapResult{failed: true}
		return
	}

	elements, err := document.QuerySelectorAll(t.Query)

	if err != nil {
		b.log.Warning("Element not found. ", t)
		ch <- scrapResult{failed: true}
		return
	}

	var res scrapResultValue

	for _, el := range elements {
		attr, err := el.GetAttribute(t.Attr)

		if err != nil {
			b.log.Warning("Attribute not found.", el)
			continue
		}

		res = append(res, attr)
	}

	ch <- scrapResult{value: res}
}
