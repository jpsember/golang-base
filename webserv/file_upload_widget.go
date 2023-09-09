package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type FileUploadObj struct {
	BaseWidgetObj
	Label HtmlString
}

type FileUpload = *FileUploadObj

func NewFileUpload(id string, label HtmlString) FileUpload {
	t := &FileUploadObj{}
	t.BaseId = id
	t.Label = label
	return t
}

func (w FileUpload) RenderTo(m MarkupBuilder, state JSMap) {
	if !w.Visible() {
		m.RenderInvisible(w)
		return
	}

	uniqueTag := w.BaseId + ".aux"
	formId := w.BaseId + ".form"

	m.Comment("file upload")

	// The outermost element must have the widget's id!  Or chaos happens during repainting.

	m.OpenTag(`div id="`, w.BaseId, `" class="mb-3"`)
	{

		m.OpenTag(`form id="`, formId, `" enctype="multipart/form-data" method="post" `)

		// I suspect the multipart/form-data has nothing to do with file uploads, but is for forms in general

		{
			labelHtml := w.Label
			if labelHtml != nil {
				m.Comment("Label")
				m.OpenTag(`label for="`, uniqueTag, `" class="form-label" style="font-size:70%"`)
				m.Escape(labelHtml)
				m.CloseTag()
			}
		}

		m.VoidTag(`input class="form-control" type="file" name="file" id="`, uniqueTag, `" onchange='jsUpload("`, w.Id(), `")'`)
		Todo("Is id requred on input?")
		Todo("is multiple required?")

		m.CloseTag()

	}

	m.CloseTag()
}
