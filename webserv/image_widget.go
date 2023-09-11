package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// This general type of listener can serve as a validator as well
type ImageURLProvider func() string

type ImageWidgetObj struct {
	BaseWidgetObj
	URLProvider ImageURLProvider
}

type ImageWidget = *ImageWidgetObj

func NewImageWidget(id string) ImageWidget {
	t := &ImageWidgetObj{}
	t.BaseId = id
	return t
}

func (w ImageWidget) RenderTo(m MarkupBuilder, state JSMap) {
	Todo("Have support for scaling down requested image")

	pr := PrIf(false)
	pr("rendering:", w.Id())

	if !w.Visible() {
		m.RenderInvisible(w)
		return
	}

	m.Comment("image")

	// The outermost element must have the widget's id!  Or chaos happens during repainting.

	m.OpenTag(`div id="`, w.BaseId, `"`)

	{
		var imageSource string

		if w.URLProvider != nil {
			imageSource = w.URLProvider()
			pr("url provider returned image source:", imageSource)
			if imageSource == "" {
				imageSource = "https://upload.wikimedia.org/wikipedia/en/a/a9/Example.jpg"
			}
		} else {
			pr("no URLProvider!")
		}

		// The outermost element must have the widget's id!  Or chaos happens during repainting.
		//
		m.VoidTag(`img src="`, imageSource, `" class="img-fluid" alt="uploaded image"`)
	}
	m.CloseTag()
	pr("done render")
}
