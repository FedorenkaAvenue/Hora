package main

import (
	"errors"
	localdb "hora/pkg/localDB"
	"hora/pkg/logger"
	"hora/pkg/notifier"
	"hora/tools"
	"io"
	"maps"
	"net/http"
	"time"

	goDom "github.com/bringmetheaugust/goDOM"
)

type config struct {
	Params struct {
		DBPath          string `yaml:"dbPath"`
		ParsingInterval int    `yaml:"parsingInterval"`
		MaxItemAmount   int    `yaml:"maxItemAmount"`
	}
	Recievers notifier.Recievers
	Targets   []target
}

type target struct {
	Url               string
	Query             string // query for goDOM querySelectorAll method
	Attr              string // which attribute take from searched element
	LinkWithoutSchema bool   `yaml:"linkWithoutSchema"` // if parsed Attr is link and without schema (http/https)
}

type scrapItems localdb.DBItems

type bot struct {
	config   config
	log      logger.Logger
	notifier notifier.Notifier
	db       localdb.LocalDB
}

func (b *bot) New() *bot {
	var db localdb.LocalDB

	b.config = tools.ParseYamlFile[config](config{}, "./config.yaml")
	b.notifier = notifier.Notifier{
		Recievers: b.config.Recievers,
	}
	b.db = db.New(b.config.Params.DBPath)

	return b
}

func (b bot) Run() {
	timeout := time.Duration(b.config.Params.ParsingInterval)

	for {
		for _, t := range b.config.Targets {
			go b.scrap(t)
		}

		<-time.After(time.Second * timeout)
	}
}

func (b *bot) scrap(t target) {
	res, err := b.parse(t)

	if err != nil {
		return
	}

	a, err := b.filter(t, res)

	if err != nil {
		panic(err)
	}

	for k := range a {
		go b.notifier.Post(k)
	}
}

func (b bot) parse(t target) (scrapItems, error) {
	resp, err := http.Get(t.Url)

	if err != nil {
		b.log.Error("During http request.", err, t)
		return nil, errors.New("")
	}

	defer resp.Body.Close()

	bytes, _ := io.ReadAll(resp.Body)
	document, err := goDom.Create(bytes)

	if err != nil {
		b.log.Error("During create document.", err, t)
		return nil, errors.New("")
	}

	elements, err := document.QuerySelectorAll(t.Query)

	if err != nil {
		b.log.Warning("Elements not found. ", t)
		return nil, errors.New("")
	}

	res := make(scrapItems)

	switch amountCount := b.config.Params.MaxItemAmount; {
	case amountCount == 0:
		break
	case len(elements) > amountCount:
		elements = elements[:amountCount]
	}

	now := time.Now()
	formattedDate := now.Format("01/02/2006")

	for _, el := range elements {
		attr, err := el.GetAttribute(t.Attr)

		if err != nil {
			b.log.Warning("Attribute not found.", el)
			continue
		}

		if t.LinkWithoutSchema {
			attr = resp.Request.URL.Host + attr
		}

		res[attr] = formattedDate
	}

	if len(res) == 0 {
		b.log.Warning("Attributes are empty. ", t)
		return nil, errors.New("")
	}

	return res, nil
}

func (b *bot) filter(t target, newV scrapItems) (scrapItems, error) {
	tData, ok := b.db.Data[t.Url]

	if !ok {
		err := b.db.Update(t.Url, localdb.DBItems(newV))

		if err != nil {
			return nil, err
		}

		return newV, nil
	} else {
		news := make(scrapItems)

		for k, v := range newV {
			if _, ok := tData[k]; !ok {
				news[k] = v
			}
		}

		if len(news) == 0 {
			b.log.Info("No new adds: ", t)
		} else {
			maps.Copy(tData, news)
			err := b.db.Update(t.Url, tData)

			if err != nil {
				return nil, err
			}
		}

		return news, nil
	}
}

func main() {
	b := bot{}
	b.New().Run()
}
