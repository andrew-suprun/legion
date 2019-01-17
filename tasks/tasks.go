package tasks

import "fmt"

type Panic struct {
	Value interface{}
}

func (p Panic) Error() string {
	return fmt.Sprint(p.Value)
}

func Start(activity func() interface{}) (resultChan chan interface{}) {
	var result interface{}
	resultChan = make(chan interface{}, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- Panic{r}
			} else {
				resultChan <- result
			}
		}()

		result = activity()
	}()
	return resultChan
}
