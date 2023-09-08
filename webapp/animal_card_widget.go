package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

// A Widget that displays editable text
type AnimalCardWidgetObj struct {
	BaseWidgetObj
	animal   Animal
	children *Array[Widget]
}

type AnimalCardWidget = *AnimalCardWidgetObj

func OpenAnimalCardWidget(m WidgetManager, baseId string, animal Animal, viewButtonListener WidgetListener) {
	widget := newAnimalCardWidget(baseId, animal)
	m.OpenContainer(widget)
	// Create a button within this card
	m.Id(baseId + "_view").Label(`View`).Listener(viewButtonListener).Size(SizeSmall).AddButton()
	m.Close()
}

func newAnimalCardWidget(widgetId string, animal Animal) AnimalCardWidget {
	w := AnimalCardWidgetObj{}
	w.Base().BaseId = widgetId
	w.animal = animal
	w.children = NewArray[Widget]()
	return &w
}

//var picCounter = 0

func (w AnimalCardWidget) RenderTo(m MarkupBuilder, state JSMap) {

	if !w.Visible() {
		m.RenderInvisible(w)
		return
	}

	// Open a bootstrap card

	m.Comments("AnimalCardWidget").OpenTag(`div class="card bg-light mb-3 animal-card"`)
	{
		imgUrl := "unknown"
		photoId := w.animal.PhotoThumbnail()
		if photoId == 0 {
			Alert("!Animal has no photo")
		} else {
			imgUrl = ReadImageIntoCache(photoId)
		}

		// Display an image
		//picCounter++
		//imgUrl = IntToString(MyMod(picCounter, 3)) + ".jpg"
		//Todo("!add support for image based on particular animal")
		//Pr("imgUrl:", imgUrl)
		m.Comment("animal image").VoidTag(`img class="card-jimg-top" src="`, imgUrl, `"`)

		// Display title and brief summary
		m.Comments("title and summary").
			OpenTag(`div class="card-body" style="max-height:8em; padding-top:.5em;  padding-bottom:.2em;"`)
		{

			m.OpenTag(`h6 class="card-title"`)
			{
				m.Escape(w.animal.Name())
			}
			m.CloseTag()

			m.OpenTag(`p class="card-text" style="font-size:75%;"`)
			{
				m.Escape(w.animal.Summary())
			}
			m.CloseTag()
		}
		m.CloseTag()

		m.Comments(`Progress towards goal, controls`).OpenTag(`div class="card-body"`)
		{
			m.Comments("progress-container").OpenTag(`div class="progress-container"`)
			{
				m.Comment("Plot grey in background, full width").OpenCloseTag(`div class="progress-bar-bgnd"`)
				m.Comment("Plot bar graph in foreground, partial width").OpenCloseTag(`div class="progress-bar" style="width: 35%;"`)
			}
			m.CloseTag()
			m.OpenTag(`div class="progress-text"`)
			{
				m.Escape(CurrencyToString(w.animal.CampaignBalance()) + ` raised of ` + CurrencyToString(w.animal.CampaignTarget()) + ` goal`)
			}
			m.CloseTag()
			m.Comments("right-justified button").OpenTag(`div class="row"`)
			{
				m.OpenTag(`div class="d-grid justify-content-md-end"`)
				{
					// Add the single child widget (a view button)
					ch := w.GetChildren()
					CheckState(len(ch) == 1, "expected single 'view' button widget")
					vb := ch[0]
					Todo("!Add ability to add style = 'width:100%; font-size:75%;' to the child button")
					Todo("!add:  <button class='btn btn-primary btn-sm'> to button")
					vb.RenderTo(m, state)
				}
				m.CloseTag()
			}
			m.CloseTag()
		}
		m.CloseTag()
	}
	m.CloseTag()
}

func (w AnimalCardWidget) GetChildren() []Widget {
	return w.children.Array()
}

const maxChildren = 1

func (w AnimalCardWidget) AddChild(c Widget, manager WidgetManager) {
	CheckState(w.children.Size() < maxChildren)
	w.children.Add(c)
}

func ReadImageIntoCache(blobId int) string {
	s := SharedWebCache
	blob := s.GetBlobWithId(blobId)
	var url string
	if blob.Id() == 0 {
		url = "missing.jpg"
	} else {
		url = "r/" + blob.Name()
	}
	//url := s.CacheMap.Get(blobId)
	//if url == "" {
	//	blob, err1 := ReadBlob(blobId)
	//	if err1 != nil {
	//		err = err1
	//		Alert("#50Trouble reading blob:", blobId)
	//	} else {
	//		// Choose a name to store this as, something obscure
	//		// For now, a simple mapping
	//		imageName := "s" + IntToString(blobId) + ".jpg"
	//		path := s.CacheDir.JoinM(imageName)
	//		Todo("we need to write to a temp file first for thread safety")
	//		path.WriteBytesM(blob.Data())
	//		url = imageName
	//		s.CacheMap.Provide(blobId, imageName)
	//		Todo("trim cache periodically; when resource is requested but doesn't exist, some mechanism to have cache read it")
	//	}
	//}
	return url
}
