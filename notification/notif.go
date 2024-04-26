package notification

type INotifService interface {
	NotifyWithReTry(title string, text string)
	Notify(title string, text string) error
	Enable()
	Disable()
}
