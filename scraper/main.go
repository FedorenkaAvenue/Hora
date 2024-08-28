package main

import (
	"context"
	"errors"
	"fmt"
	"hora/pkg/logger"
	"hora/pkg/notifier"
	"hora/tools"
	"io"
	"net/http"
	"time"

	goDom "github.com/bringmetheaugust/goDOM"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type bot struct {
	config   config
	log      logger.Logger
	notifier notifier.Notifier
	db       *mongo.Client
}

type config struct {
	Params struct {
		ParsingInterval int   `yaml:"parsingInterval"`
		MaxItemAmount   int   `yaml:"maxItemAmount"`
		ItemLifePeriod  int64 `yaml:"itemLifePeriod"`
	}
	Recievers notifier.Recievers
	Targets   []target
}

type target struct {
	Name              string
	Url               string
	ItemLinkQuery     string `yaml:"itemLinkQuery"`
	LinkWithoutSchema bool   `yaml:"linkWithoutSchema"`
	Items             []Item
}

type Item struct {
	Href struct {
		Query string
		Attr  string
	}
	Title struct {
		Query string
		Value string
	}
	Price struct {
		Query    string
		MinValue int
	}
}

type dbItem struct {
	ID        string    `bson:"_id"`
	CreatedAt time.Time `bson:"created_at"`
}

type scrapItems []string

var ctx = context.TODO()

func (b *bot) New() *bot {
	b.config = tools.ParseYamlFile[config](config{}, "./tmp/config.yaml")
	b.notifier = notifier.Notifier{Recievers: b.config.Recievers}

	clientOptions := options.Client().ApplyURI("mongodb://hora_db")
	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		panic("Cann't connect to Mongo")
	}

	b.db = client

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
	res, err := b.getTarget(t)

	if err != nil {
		return
	}

	b.filter(t, res)

	// if err != nil {
	// 	panic(err)
	// }

	// for _, v := range a {
	// 	b.notifier.Post(v)
	// }
}

func (b bot) getTarget(t target) (scrapItems, error) {
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

	elements, err := document.QuerySelectorAll(t.ItemLinkQuery)

	if err != nil {
		b.log.Error("Elements not found. ", t)
		return nil, errors.New("")
	}

	switch amountCount := b.config.Params.MaxItemAmount; {
	case amountCount == 0:
		break
	case len(elements) > amountCount:
		elements = elements[:amountCount]
	}

	var links scrapItems

	for _, l := range elements {
		attr, aErr := l.GetAttribute("href")

		if aErr != nil {
			b.log.Error("Attribute not found", l)
			continue
		}

		links = append(links, attr)
	}

	return links, nil
}

func (b *bot) filter(t target, items scrapItems) {
	c := b.db.Database("scrapped").Collection(t.Name)

	res, _ := c.Find(ctx, bson.D{{Key: "url", Value: items[0]}})

	fmt.Println(res)

	for _, i := range items {
		// res, _ := c.Find(ctx, bson.D{{Key: "url", Value: i}})
		_, err := c.InsertOne(ctx, dbItem{ID: i})

		if err != nil {
			fmt.Printf("%v", err)
		}
	}

	// _, err := c.InsertOne(ctx, dbItem{name: "loh"})

	// tData, ok := b.db.Data[t.Name]
	// var newItems scrapItems

	// if ok {
	// 	for _, el := range newV {
	// 		linkEl, lErr := el.QuerySelector(t.ItemLinkQuery)

	// 		if lErr != nil {
	// 			b.log.Warning("Link element not found.", el)
	// 			continue
	// 		}

	// 		link, hErr := linkEl.GetAttribute("href") // var res scrapItems

	// 		if hErr != nil {
	// 			b.log.Warning("Link attribute not found.", el)
	// 			continue
	// 		}

	// 		if _, ok := tData[link]; ok {
	// 			continue
	// 		} else {
	// 			// if t.LinkWithoutSchema {
	// 			// 	attr = resp.Request.URL.Host + attr
	// 			// }

	// 			newItems = append(newItems, link)
	// 			// main
	// 		}
	// 	}

	// 	if len(newItems) == 0 {
	// 		b.log.Info("No new adds:", t.Name)
	// 	} else {
	// 		b.db.Append(t.Name, newItems)
	// 	}

	// 	return newItems, nil
	// } else {
	// 	b.db.Append(t.Name, newItems)
	// }

	// return newItems, nil
}

func main() {
	b := bot{}
	b.New().Run()
}
