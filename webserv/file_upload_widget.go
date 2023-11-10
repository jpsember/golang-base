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
	id := w.Id()
	filenameId := id + ".input"
	formId := id + ".form"
	inputName := id + ".input"

	m.Comment("file upload widget")

	// The outermost element must have the widget's id!  Or chaos happens during repainting.

	m.TgOpen(`div id=`).A(QUO, id, ` class="mb-3"`).TgContent()
	{

		m.TgOpen(`form id=`).A(QUO, formId, ` enctype="multipart/form-data" method="post" `).TgContent()

		// I suspect the multipart/form-data has nothing to do with file uploads, but is for forms in general

		{
			labelHtml := w.Label
			if labelHtml != nil {
				m.Comment("Label for the widget")
				m.TgOpen(`label for=`).A(QUO, filenameId, ` class="form-label"`).Style(`font-size:70%`).TgContent()
				m.Escape(labelHtml)
				m.TgClose()
			}
		}

		m.Comment(`The input element that Bootstrap does some magic on`)
		m.TgOpen(`input class='form-control' type='file' name=`)
		m.A(QUO, inputName, ` id=`, QUO, filenameId, ` onchange="jsUpload('`, w.Id(), `')"`)
		m.TgClose()

		problemText := s.WidgetProblem(w)

		hasProblem := problemText != ""

		if hasProblem {
			m.Comment("Problem")
			m.TgOpen(`div class="form-text text-danger"`).Style(`font-size:  70%`).TgContent()
			m.A(ESCAPED, problemText).TgClose()
		}

		m.TgClose()
	}

	m.TgClose()
}
