package notification

import "fmt"

type DebugNotifier struct {
}

func MakeDebugNotifier() *DebugNotifier {
	notifier := DebugNotifier{}
	return &notifier
}

func (notifier *DebugNotifier) NotifyWithReTry(title string, text string) {
	fmt.Println("Title: ", title)
	fmt.Println("Text: ", text)
}

func (notifier *DebugNotifier) Notify(title string, text string) error {
	fmt.Println("Title: ", title)
	fmt.Println("Text: ", text)
	return nil
}

func (notifier *DebugNotifier) Enable() {

}

func (notifier *DebugNotifier) Disable() {

}
