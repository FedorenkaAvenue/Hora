package main

import (
	"errors"
	localdb "hora/pkg/localDB"
	"hora/pkg/logger"
	"hora/pkg/notifier"
	"hora/tools"
	"io"
	"net/http"
	"slices"
	"time"

	goDom "github.com/bringmetheaugust/goDOM"
)

type config struct {
	Params struct {
		DBPath          string `yaml:"dbPath"`
		ParsingInterval int    `yaml:"parsingInterval"`
	}
	Recievers notifier.Recievers
	Targets   []target
}

type target struct {
	Url   string
	Query string
	Attr  string
}

type bot struct {
	config   config
	log      logger.Logger
	notifier notifier.Notifier
	db       localdb.LocalDB
}

type parseRes []string

func (b *bot) New() *bot {
	var db localdb.LocalDB

	b.config = tools.ParseYamlFile[config](config{}, "./configs/bot.yaml")
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
		return
	}

	for _, i := range a {
		go b.notifier.Post(i)
	}
}

func (b *bot) filter(t target, v parseRes) (parseRes, error) {
	tData, ok := b.db.Data[t.Url]

	if !ok {
		err := b.db.Update(t.Url, v)

		if err != nil {
			return nil, err
		}

		return v, nil
	} else {
		var news parseRes

		for _, n := range tData {
			if !slices.Contains(tData, n) {
				news = append(news, n)
			}
		}

		if len(news) == 0 {
			b.log.Info("No new adds: ", t)
		} else {
			err := b.db.Update(t.Url, news)

			if err != nil {
				return nil, err
			}
		}

		return news, nil
	}
}

func (b bot) parse(t target) (parseRes, error) {
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

	var res parseRes

	for _, el := range elements {
		attr, err := el.GetAttribute(t.Attr)

		if err != nil {
			b.log.Warning("Attribute not found.", el)
			continue
		}

		res = append(res, attr)
	}

	if len(res) == 0 {
		b.log.Warning("Attributes are empty. ", t)
		return nil, errors.New("")
	}

	return res, nil
}

func main() {
	b := bot{}
	b.New().Run()
}
