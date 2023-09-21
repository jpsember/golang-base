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

		m.TgOpen(`form id=`).A(QUOTED, formId, ` enctype="multipart/form-data" method="post" `).TgContent()

		// I suspect the multipart/form-data has nothing to do with file uploads, but is for forms in general

		{
			labelHtml := w.Label
			if labelHtml != nil {
				m.Comment("Label")
				m.TgOpen(`label for=`).A(QUOTED, inputId, ` class="form-label"`).Style(`font-size:70%`).TgContent()
				m.Escape(labelHtml)
				m.TgClose()
			}
		}

		m.TgOpen(`input class="form-control" type="file" name=`)
		m.A(QUOTED, inputName, ` id=`, QUOTED, inputId, ` onchange='jsUpload(`, QUOTED, w.Id(), `)'`)
		m.TgClose()

		problemId := WidgetIdWithProblem(w.BaseId)
		problemText := s.StringValue(problemId)

		hasProblem := problemText != ""

		if hasProblem {
			m.Comment("Problem")
			m.A(`<div class="form-text text-danger" style="font-size:  70%">`)
			m.Escape(problemText).A(`</div>`)
		}

		m.TgClose()
	}

	m.TgClose()
}
