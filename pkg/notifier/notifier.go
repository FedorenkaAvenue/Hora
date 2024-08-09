package notifier

import "fmt"

type Notifier struct{}

func (n Notifier) Post(ntf any) {
	fmt.Println(ntf)
}
