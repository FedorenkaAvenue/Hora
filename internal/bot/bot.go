package bot

import (
	"hora/pkg/logger"
	"hora/pkg/notifier"
	"hora/tools"
	"time"
)

type Bot struct {
	config   config
	log      logger.Logger
	notifier notifier.Notifier
}

func (b *Bot) New() *Bot {
	b.config = tools.ParseYamlFile[config](config{}, "./configs/bot.yaml")
	b.notifier = notifier.Notifier{}

	return b
}

func (b Bot) Run() {
	timeoutDuration := time.Duration(b.config.Params.ParsingInterval)

	for {
		go b.scrap()
		<-time.After(time.Second * timeoutDuration)
	}
}
