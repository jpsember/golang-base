package webserv

type TextWidgetObj struct {
	BaseWidgetObj
	Text string
}

type TextWidget = *TextWidgetObj

func NewTextWidget() TextWidget {
	return &TextWidgetObj{}
}
