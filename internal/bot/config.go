package bot

type config struct {
	Params struct {
		ParsingInterval int `yaml:"parsingInterval"`
	}
	Targets targets
}

type targets []target

type target struct {
	Url   string
	Query string
	Attr  string
}
