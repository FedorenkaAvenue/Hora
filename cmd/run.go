package main

import bot "hora/internal/bot"

func main() {
	b := bot.Bot{}
	b.New().Run()
}
