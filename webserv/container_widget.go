package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// A concrete Widget that can contain others
type ContainerWidgetObj struct {
	BaseWidgetObj
	children []Widget
}

type ContainerWidget = *ContainerWidgetObj

func NewContainerWidget(id string) ContainerWidget {
	w := ContainerWidgetObj{}
	w.BaseId = id
	return &w
}

func (w ContainerWidget) String() string {
	return "<" + w.BaseId + " ContainerWidget>"
}

func (w ContainerWidget) Children() []Widget {
	return w.children
}

func (w ContainerWidget) ClearChildren() {
	w.children = nil
}

func (w ContainerWidget) AddChild(c Widget, manager WidgetManager) {
	cols := manager.pendingChildColumns
	if cols == 0 {
		BadState("no pending columns for widget:", c.Id())
	}
	c.SetColumns(cols)
	w.children = append(w.children, c)
	pr := PrIf(false)
	pr(VERT_SP, "ContainerWidget", w.BaseId, "adding child", c.Id(), "to container", w.BaseId, "columns:", w.Columns())
}

func (w ContainerWidget) RenderTo(s Session, m MarkupBuilder) {
	// It is the job of the widget that *contains* us to set the columns that we
	// are to occupy, not ours.
	m.Comments(`ContainerWidget`, w.IdSummary())
	m.OpenTag(`div id='` + w.BaseId + `'`)
	if len(w.children) != 0 {
		m.OpenTag(`div class='row'`)
		for _, child := range w.children {
			str := `div class="col-sm-` + IntToString(child.Columns()) + `"`
			if WidgetDebugRenderingFlag {
				str += ` style="background-color:` + DebugColorForString(child.Id()) + `;`
				str += `border-style:double;`
				str += `"`
			}
			m.Comments(`child`).OpenTag(str)
			{
				verify := m.VerifyBegin()
				RenderWidget(child, s, m)
				m.VerifyEnd(verify, child)
			}
			m.CloseTag()
		}
		m.CloseTag().Br()
	}
	m.CloseTag() // ContainerWidget
}
