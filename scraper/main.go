package main

import (
	"context"
	"fmt"
	"hora/pkg/logger"
	"hora/pkg/notifier"
	"hora/tools"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
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
		ParsingInterval int `yaml:"parsingInterval"`
	}
	Recievers notifier.Recievers
	Targets   []target
}

type target struct {
	Name              string
	Url               string
	ItemLinkQuery     string      `yaml:"itemLinkQuery"`
	LinkWithoutSchema bool        `yaml:"linkWithoutSchema"`
	Params            []Parameter `yaml:"params"`
	host              host
}

type host *url.URL

type Parameter struct {
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
	ID        string `bson:"_id"`
	CreatedAt int64  `bson:"created_at"`
}

type scrapItems []string

const (
	dbScrapped         string        = "scrapped"
	itemLifePeriod     time.Duration = time.Hour * 24 * 30
	clearItemsInterval time.Duration = time.Hour * 24
)

var ctx = context.TODO()

func (b *bot) New() *bot {
	b.config = tools.ParseYamlFile[config](config{}, "./tmp/config.yaml")
	b.notifier = notifier.Notifier{Recievers: b.config.Recievers}

	credential := options.Credential{Username: os.Getenv("DB_USER"), Password: os.Getenv("DB_PASSWORD")}
	clientOptions := options.Client().ApplyURI("mongodb://hora_db").SetAuth(credential)
	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		panic(fmt.Sprintf("Cann't connect to db\n. %v", err))
	}

	b.db = client

	return b
}

func (b bot) Run() {
	timeout := time.Duration(b.config.Params.ParsingInterval)

	go func() {
		for {
			b.clearParsedItems()
			<-time.After(clearItemsInterval)
		}
	}()

	for {
		for _, t := range b.config.Targets {
			go b.scrap(&t)
		}

		<-time.After(time.Second * timeout)
	}
}

func (b bot) scrap(t *target) {
	l, err := b.parse(t)

	if err != nil {
		b.log.Error(t.Name, err)
		panic(err)
	}

	n := b.filter(*t, l)

	if len(n) == 0 {
		b.log.Info("No new adds for", t.Name)
		return
	}

	b.log.Info("new adds: ", n)

	for _, l := range n {
		go b.match(*t, l)
	}
}

func (b bot) parse(t *target) (scrapItems, error) {
	document, host, err := getDocument(t.Url)

	if err != nil {
		return nil, err
	}

	t.host = host
	elements, err := document.QuerySelectorAll(t.ItemLinkQuery)

	if err != nil {
		return nil, fmt.Errorf("elements with '%v' not found in %v", t.ItemLinkQuery, t.Url)
	}

	var links scrapItems

	for _, l := range elements {
		attr, err := l.GetAttribute("href")

		if err != nil {
			b.log.Error("href attribute not found.")
			panic(err)
		}

		links = append(links, attr)
	}

	return links, nil
}

func (b bot) filter(t target, items scrapItems) scrapItems {
	c := b.db.Database(dbScrapped).Collection(t.Name)
	var news scrapItems
	d := time.Now().Unix()

	for _, i := range items {
		err := c.FindOne(ctx, bson.D{{Key: "_id", Value: i}}).Decode(&dbItem{})

		if err != nil {
			_, err := c.InsertOne(ctx, dbItem{ID: i, CreatedAt: d})

			if err != nil {
				panic(fmt.Sprintf("Error during adding db item: %v", err))
			}

			news = append(news, i)
		}
	}

	return news
}

func (b bot) match(t target, l string) {
	if t.LinkWithoutSchema {
		l = t.host.Scheme + "://" + t.host.Host + l
	}

	document, _, err := getDocument(l)

	if err != nil {
		b.log.Error(err)
		panic(err)
	}

rootLoop:
	for _, i := range t.Params {
		if i.Price.Query != "" {
			price, err := document.QuerySelector(i.Price.Query)

			if err != nil {
				b.log.Error("Price not found. ", l, i.Price.Query)
				panic(err)
			}

			re := regexp.MustCompile(`\d+`)
			match := re.FindString(price.TextContent)

			if match != "" {
				number, err := strconv.Atoi(match)

				if err != nil {
					b.log.Error("Error converting to integer: ", err)
					panic(err)
				} else {
					if i.Price.MinValue != 0 && number < i.Price.MinValue {
						b.log.Info("Item doesn't matched by min price", l, i.Price.MinValue)
						continue rootLoop
					}
				}
			} else {
				b.log.Error("Cann't get price number from text content: ", l, i)
				continue rootLoop
			}
		}

		if i.Title.Query != "" {
			title, err := document.QuerySelector(i.Title.Query)

			if err != nil {
				b.log.Error("Title not found: ", l, i)
				panic(err)
			}

			if !strings.Contains(title.TextContent, i.Title.Value) {
				b.log.Info("Item doesn't contain title", l, i.Title.Value)
				continue rootLoop
			}
		}

		go b.notifier.Post(l)

		break
	}
}

func (b bot) clearParsedItems() {
	colls, err := b.db.Database(dbScrapped).ListCollectionNames(ctx, bson.D{})

	if err != nil {
		b.log.Error("Cann't get colleections list names: ", err)
		panic("Cann't get colleections list names.")
	}

	d := time.Now().Unix() - int64(itemLifePeriod)

	for _, c := range colls {
		c := b.db.Database(dbScrapped).Collection(c)
		res, err := c.DeleteMany(ctx, bson.D{{
			Key: "created_at", Value: bson.D{{
				Key: "$lte", Value: d,
			}},
		}})

		if err != nil {
			b.log.Error("Cann't get old items: ", err)
			panic("Cann't get old items.")
		}

		b.log.Success("Cleaned ", *res, " items.")
	}
}

func getDocument(link string) (*goDom.Document, host, error) {
	resp, err := http.Get(link)

	if err != nil {
		return nil, nil, fmt.Errorf("during http request : %v", err)
	}

	defer resp.Body.Close()

	host := resp.Request.URL
	bytes, _ := io.ReadAll(resp.Body)
	document, err := goDom.Create(bytes)

	if err != nil {
		return nil, nil, fmt.Errorf("during create document : %v", err)
	}

	return document, host, nil
}

func main() {
	b := bot{}
	b.New().Run()
}
