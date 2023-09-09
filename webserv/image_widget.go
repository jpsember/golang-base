package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type ImageWidgetObj struct {
	BaseWidgetObj
}

type ImageWidget = *ImageWidgetObj

func NewImageWidget(id string) ImageWidget {
	t := &ImageWidgetObj{}
	t.BaseId = id
	return t
}

func (w ImageWidget) RenderTo(m MarkupBuilder, state JSMap) {
	if !w.Visible() {
		m.RenderInvisible(w)
		return
	}

	m.Comment("image")

	// The outermost element must have the widget's id!  Or chaos happens during repainting.
	//
	var imageSource = "https://upload.wikimedia.org/wikipedia/en/a/a9/Example.jpg"

	m.OpenTag(`img src="`, imageSource, `" class="img-fluid" alt="uploaded image"`)
	m.CloseTag()
}
