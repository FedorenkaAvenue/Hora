# Hora

Simple bot for site parsing.

<!-- ## Config file

`./config.yaml` is for example

```yaml
params:
  parsingInterval: int # parsing interval in seconds
  maxItemAmount: int # max items amount parsed
  itemLifePeriod: int # item life period in seconds
recievers: # array of messanger configs
  - type: const string # type of messanger (check in pkg/notifier/notifier.go)
    token: string # messanger token 
    chatID: int # chat id
  - ... # another reciever
targets: # array of sites
  - name: # target alias
    url: string # site url
    itemLinkQuery: string # query string for DOM querySelectorAll method. get element from html page
    linkWithoutSchema: bool # if target.attr is link and it has no schema (http/https). Bot will add schema to the responce message
  - params:
      - price:
          query: ".css-90xrc0" # query string for DOM querySelectorAll method. get element from html page
          minValue: 50
        title:
          query: ".css-1kc83jo"
          value: "Boss"
      - ... # another param

``` -->

<!-- ## Commands

 * `make clear_log` clear log files -->
