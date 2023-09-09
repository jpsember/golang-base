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

		// <form method="post" enctype="multipart/form-data">
		//  <label for="file">File</label>
		//  <input id="file" name="file" type="file" />
		//  <button>Upload</button>
		//</form>
		m.OpenTag(`form id="`, formId, `" method="post" enctype="multipart/form-data"`)
		{
			labelHtml := w.Label
			if labelHtml != nil {
				m.Comment("Label")
				m.OpenTag(`label for="`, uniqueTag, `" class="form-label" style="font-size:70%"`)
				m.Escape(labelHtml)
				m.CloseTag()
			}
		}
		m.CloseTag()

		m.VoidTag(`input class="form-control" type="file"  name="file" id="`, uniqueTag, `" onchange='jsUpload("`, w.Id(), `")'`)
	}

	m.CloseTag()
}
