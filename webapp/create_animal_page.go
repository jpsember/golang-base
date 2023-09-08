package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
	"strings"
)

const (
	id_animal_name    = "a_name"
	id_animal_summary = "a_summary"
	id_animal_details = "a_details"
	id_add            = "a_add"
)

type CreateAnimalPageStruct struct {
	BasicPage
}

type CreateAnimalPage = *CreateAnimalPageStruct

func NewCreateAnimalPage(sess Session, parentWidget Widget) AbstractPage {
	t := &CreateAnimalPageStruct{
		NewBasicPage(sess, parentWidget),
	}
	t.devLabel = "create_animal_page"
	return t
}

func (p CreateAnimalPage) Generate() {
	// SetWidgetDebugRendering()

	m := p.GenerateHeader()

	Todo("!Have ajax listener that can show advice without an actual error, e.g., if user left some fields blank")
	m.Label("Create New Animal Record").Size(SizeLarge).AddHeading()
	m.Col(6).Open()
	{
		m.Col(12)
		m.Label("Name").Id(id_animal_name).Listener(ValidateAnimalName).AddInput()

		m.Label("Summary").Id(id_animal_summary).AddInput()
		m.Size(SizeTiny).Label("A brief paragraph to appear in the 'card' view.").AddText()
		m.Label("Details").Id(id_animal_details).AddInput()
		m.Size(SizeTiny).Label("Additional paragraphs to appear on the 'details' view.").AddText()
		
		m.Listener(p.addListener)
		m.Id(id_add).Label("Create").AddButton()
	}
	m.Close()
}

func (p CreateAnimalPage) addListener(sess Session, widget Widget) {
	if Todo("CreateAnimal") {

	}
}

func ValidateAnimalName(s Session, widget Widget) {
	errStr := ""
	n := s.GetValueString()
	n = strings.TrimSpace(n)
	for {
		ln := len(n)
		if ln < 3 || ln > 20 {
			errStr = "Length should be 3...20 characters"
			break
		}
		break
	}
	if errStr != "" {
		s.SetWidgetProblem(widget, errStr)
	}
}
