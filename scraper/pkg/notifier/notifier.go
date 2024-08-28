package notifier

import (
	"fmt"
	"net/http"
)

type Notifier struct {
	Recievers
}

type Recievers []recieverConfig

type recieverConfig struct {
	Type   string
	Token  string
	ChatID string `yaml:"chatID"`
}

const (
	telegram = "telegram"
)

func (n Notifier) Post(msg string) {
	for _, r := range n.Recievers {
		switch {
		case r.Type == telegram:
			res, err := http.Get(
				fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s", r.Token, r.ChatID, msg),
			)

			if err != nil || res.StatusCode >= 400 {
				panic(fmt.Sprintf("Err: %+v\n,Res: %+v", err, res))
			}
		default:
			panic(fmt.Sprintf("Invalid config: %v.", r))
		}
	}
}
