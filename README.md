# Hora

Simple bot for site parsing.

## Config file

`./config.yaml` is for example

```yaml
params:
  dbPath: string # local DB store location
  parsingInterval: int # parsing interval in seconds
  maxItemAmount: int # max items amount parsed
recievers: # array of messanger configs
  - type: const string # type of messanger (check in pkg/notifier/notifier.go)
    token: string # messanger token 
    chatID: int # chat id
  - ... # abother reciever
targets: # array of sites
  - url: string # site url
    query: string # query string for querySelectorAll method
    attr: string # which attribute get from parsed element
    linkWithoutSchema: bool # if target.attr is link and it has no schema (http/https). Bot will add schema to the responce message
  - ... # another target

```

## Commands

 * `make install` install all dependencies (dev and prod)
 * `make dev` run app in dev environment
 * `make build` build app
 * `make run` run build app
 * `make help` commands help
