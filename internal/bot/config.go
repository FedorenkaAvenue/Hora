package bot

import "hora/pkg/notifier"

type config struct {
	Params struct {
		ParsingInterval int `yaml:"parsingInterval"`
	}
	Recievers notifier.Recievers
	Targets   []target
}

type target struct {
	Url   string
	Query string
	Attr  string
}
