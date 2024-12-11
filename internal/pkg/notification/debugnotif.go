package notification

import "fmt"

type DebugNotifier struct {
}

func MakeDebugNotifier() *DebugNotifier {
	notifier := DebugNotifier{}
	return &notifier
}

func (notifier *DebugNotifier) NotifyWithReTry(title string, text string) {
	fmt.Println()
	fmt.Println("Title: ", title)
	fmt.Println("Text:\n", text)
	fmt.Println()
}

func (notifier *DebugNotifier) Notify(title string, text string) error {
	fmt.Println()
	fmt.Println(" /* ------------------------------- ")
	fmt.Println("Title: ", title)
	fmt.Println("Text:\n", text)
	fmt.Println(" ------------------------------- */ ")
	fmt.Println()
	return nil
}

func (notifier *DebugNotifier) Enable() {

}

func (notifier *DebugNotifier) Disable() {

}
