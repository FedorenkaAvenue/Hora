package main

import (
	"errors"
	localdb "hora/pkg/localDB"
	"hora/pkg/logger"
	"hora/pkg/notifier"
	"hora/tools"
	"io"
	"net/http"
	"time"

	goDom "github.com/bringmetheaugust/goDOM"
)

type config struct {
	Params struct {
		DBPath          string `yaml:"dbPath"`
		ParsingInterval int    `yaml:"parsingInterval"`
		MaxItemAmount   int    `yaml:"maxItemAmount"`
		ItemLifePeriod  int64  `yaml:"itemLifePeriod"`
	}
	Recievers notifier.Recievers
	Targets   []target
}

type target struct {
	Name              string
	Url               string
	Query             string // query for goDOM querySelectorAll method
	Attr              string // which attribute take from searched element
	LinkWithoutSchema bool   `yaml:"linkWithoutSchema"` // if parsed Attr is link and without schema (http/https)
}

type scrapItems []string

type bot struct {
	config   config
	log      logger.Logger
	notifier notifier.Notifier
	db       localdb.LocalDB
}

func (b *bot) New() *bot {
	var dbNames []string

	b.config = tools.ParseYamlFile[config](config{}, "./config.yaml")
	b.notifier = notifier.Notifier{Recievers: b.config.Recievers}
	b.db = localdb.LocalDB{
		StorePath:      b.config.Params.DBPath,
		ItemLifePeriod: b.config.Params.ItemLifePeriod,
	}

	for _, t := range b.config.Targets {
		dbNames = append(dbNames, t.Name)
	}

	b.db.Init(dbNames)

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

	for _, v := range a {
		go b.notifier.Post(v)
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

	var res scrapItems

	switch amountCount := b.config.Params.MaxItemAmount; {
	case amountCount == 0:
		break
	case len(elements) > amountCount:
		elements = elements[:amountCount]
	}

	for _, el := range elements {
		attr, err := el.GetAttribute(t.Attr)

		if err != nil {
			b.log.Warning("Attribute not found.", el)
			continue
		}

		if t.LinkWithoutSchema {
			attr = resp.Request.URL.Host + attr
		}

		res = append(res, attr)
	}

	if len(res) == 0 {
		b.log.Warning("Attributes are empty. ", t)
		return nil, errors.New("")
	}

	return res, nil
}

func (b *bot) filter(t target, newV scrapItems) (scrapItems, error) {
	tData, ok := b.db.Data[t.Name]

	if !ok {
		err := b.db.Append(t.Name, newV)

		if err != nil {
			return nil, err
		}

		return newV, nil
	} else {
		var news scrapItems

		for _, i := range newV {
			if _, ok := tData[i]; !ok {
				news = append(news, i)
			}
		}

		if len(news) == 0 {
			b.log.Info("No new adds:", t.Name)
		} else {
			err := b.db.Append(t.Name, news)

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
