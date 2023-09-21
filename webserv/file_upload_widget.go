package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type FileUploadWidgetListener func(sess Session, widget FileUpload, value []byte) error

type FileUploadObj struct {
	BaseWidgetObj
	Label    HtmlString
	listener FileUploadWidgetListener
}

type FileUpload = *FileUploadObj

func NewFileUpload(id string, label HtmlString, listener FileUploadWidgetListener) FileUpload {
	t := &FileUploadObj{}
	t.InitBase(id)
	t.Label = label
	t.listener = listener
	return t
}

func (w FileUpload) RenderTo(s Session, m MarkupBuilder) {
	id := w.BaseId
	inputId := id + ".input"
	formId := id + ".form"
	inputName := id + ".input"

	m.Comment("file upload")

	// The outermost element must have the widget's id!  Or chaos happens during repainting.

	m.TgOpen(`div id=`).A(QUOTED, id, ` class="mb-3"`).TgContent()
	{

		m.OpenTag(`form id="`, formId, `" enctype="multipart/form-data" method="post" `)

		// I suspect the multipart/form-data has nothing to do with file uploads, but is for forms in general

		{
			labelHtml := w.Label
			if labelHtml != nil {
				m.Comment("Label")
				m.OpenTag(`label for="`, inputId, `" class="form-label" style="font-size:70%"`)
				m.Escape(labelHtml)
				m.CloseTag()
			}
		}

		m.VoidTag(`input class="form-control" type="file" name="`, inputName, `" id="`, inputId, `" onchange='jsUpload("`, w.Id(), `")'`)

		problemId := WidgetIdWithProblem(w.BaseId)
		problemText := s.StringValue(problemId)

		hasProblem := problemText != ""

		if hasProblem {
			m.Comment("Problem")
			m.A(`<div class="form-text text-danger" style="font-size:  70%">`)
			m.Escape(problemText).A(`</div>`)
		}

		m.CloseTag()
	}

	m.TgClose()
}
