package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// This general type of listener can serve as a validator as well
type ImageURLProvider func() string

type ImageWidgetObj struct {
	BaseWidgetObj
	URLProvider ImageURLProvider
	FixedSize   IPoint // If not (0,0), size to display image as
}

type ImageWidget = *ImageWidgetObj

func NewImageWidget(id string) ImageWidget {
	t := &ImageWidgetObj{}
	t.BaseId = id
	return t
}

func (w ImageWidget) RenderTo(s Session, m MarkupBuilder) {
	pr := PrIf(false)
	pr("rendering:", w.Id())

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

		m.A(`<img src="`, imageSource, `" alt="uploaded image"`)
		if w.FixedSize.IsPositive() {
			m.A(` width="`, w.FixedSize.X, `" height="`, w.FixedSize.Y, `" class="img-thumbnail"`)
		} else {
			m.A(` class="img-thumbnail image-fluid"`)
		}
		m.A(`>`)

	}
	m.CloseTag()
	pr("done render")
}
