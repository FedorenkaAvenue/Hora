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
	"slices"
	"strconv"
	"strings"
	"sync"
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
		Value []string
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

type matchedItem struct {
	link    string
	matched bool
	err     interface{}
}

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
	t := time.Duration(b.config.Params.ParsingInterval)

	go func() {
		for {
			err := b.clearOldItems()

			if err != nil {
				b.log.Error(err)
			}

			<-time.After(clearItemsInterval)
		}
	}()

	for {
		for _, t := range b.config.Targets {
			go b.scrap(&t)
		}

		<-time.After(time.Second * t)
	}
}

func (b *bot) scrap(t *target) {
	defer func() {
		if err := recover(); err != nil {
			b.log.Warning("Skip target ", t.Name)
			b.config.Targets = slices.DeleteFunc(b.config.Targets, func(e target) bool {
				return e.Name == t.Name
			})
		}
	}()

	p, err := b.parseTarget(t)

	if err != nil {
		b.log.Error(t.Name, err)
		panic(err)
	}

	f, err := b.filterItems(*t, p)

	if err != nil {
		b.log.Error(t.Name, err)
		panic(err)
	}

	if len(f) == 0 {
		b.log.Info("No new adds for", t.Name)
		return
	}

	b.log.Info("new adds: ", f)

	var matched scrapItems

	if len(t.Params) > 0 {
		ch := make(chan matchedItem)
		var wg sync.WaitGroup

		go func() {
			wg.Wait()
			close(ch)
		}()

		for _, l := range f {
			wg.Add(1)
			go b.matchItem(*t, l, ch, &wg)
		}

		for r := range ch {
			if r.err == nil && r.matched {
				matched = append(matched, r.link)
			}
		}
	} else {
		matched = f
	}

	for _, l := range matched {
		go b.notifier.Post(l)
	}

	b.appendItems(t, f)
}

func (b bot) parseTarget(t *target) (scrapItems, error) {
	d, host, err := getDocument(t.Url)

	if err != nil {
		return nil, err
	}

	t.host = host
	e, err := d.QuerySelectorAll(t.ItemLinkQuery)

	if err != nil {
		return nil, fmt.Errorf("elements with '%v' not found in %v", t.ItemLinkQuery, t.Url)
	}

	var links scrapItems

	for _, l := range e {
		a, err := l.GetAttribute("href")

		if err != nil {
			return nil, fmt.Errorf("attribute href not found inside '%v in %v", t.ItemLinkQuery, t.Url)
		}

		links = append(links, a)
	}

	return links, nil
}

func (b bot) filterItems(t target, items scrapItems) (scrapItems, error) {
	c := b.db.Database(dbScrapped).Collection(t.Name)
	var news scrapItems

	for _, i := range items {
		err := c.FindOne(ctx, bson.D{{Key: "_id", Value: i}}).Decode(&dbItem{})

		if err != nil {
			news = append(news, i)
		}
	}

	return news, nil
}

func (b bot) appendItems(t *target, items scrapItems) error {
	c := b.db.Database(dbScrapped).Collection(t.Name)
	d := time.Now().Unix()
	var m []interface{}

	for _, i := range items {
		m = append(m, dbItem{ID: i, CreatedAt: d})
	}

	_, err := c.InsertMany(ctx, m)

	if err != nil {
		return fmt.Errorf("error during adding db item: %v", err)
	}

	return nil
}

func (b bot) matchItem(t target, l string, ch chan matchedItem, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()

		if err := recover(); err != nil {
			b.log.Error(err)
			ch <- matchedItem{l, false, err}
		}
	}()

	if t.LinkWithoutSchema {
		l = t.host.Scheme + "://" + t.host.Host + l
	}

	d, _, err := getDocument(l)

	if err != nil {
		panic(fmt.Errorf("cann't get document for '%v': %v", l, err))
	}

rootLoop:
	for _, i := range t.Params {
		if q, m := i.Price.Query, i.Price.MinValue; q != "" {
			p, err := d.QuerySelector(q)

			if err != nil {
				panic(fmt.Errorf("price not found: %v, %v", l, q))
			}

			re := regexp.MustCompile(`\d+`)
			match := re.FindString(p.TextContent)

			if match != "" {
				n, err := strconv.Atoi(match)

				if err != nil {
					panic(fmt.Errorf("error converting to integer: %v", err))
				} else {
					if m != 0 && n < m {
						b.log.Info("Item doesn't matched by min price", l, m)
						continue rootLoop
					}
				}
			} else {
				b.log.Error("Cann't get price number from text content: ", l, i)
				continue rootLoop
			}
		}

		if q, v := i.Title.Query, i.Title.Value; q != "" {
			t, err := d.QuerySelector(q)

			if err != nil {
				panic(fmt.Errorf("title not found in %v by %v query", l, q))
			}

			if !slices.ContainsFunc(v, func(e string) bool {
				return strings.Contains(t.TextContent, e)
			}) {
				b.log.Info("Item doesn't contain title", l, v)
				continue rootLoop
			}
		}

		ch <- matchedItem{l, true, nil}
	}

	ch <- matchedItem{l, false, nil}
}

func (b bot) clearOldItems() error {
	colls, err := b.db.Database(dbScrapped).ListCollectionNames(ctx, bson.D{})

	if err != nil {
		return fmt.Errorf("cann't get colleections list names: '%v", err)
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
			return fmt.Errorf("cann't get old items: '%v", err)
		}

		b.log.Success("Cleaned ", *res, " items.")
	}

	return nil
}

func getDocument(link string) (*goDom.Document, host, error) {
	r, err := http.Get(link)

	if err != nil || r.StatusCode >= 400 {
		return nil, nil, fmt.Errorf("during http request : %v", err)
	}

	defer r.Body.Close()

	host := r.Request.URL
	bytes, _ := io.ReadAll(r.Body)
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
