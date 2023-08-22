package base

type ErrorHolderStruct struct {
	ErrorList *Array[error]
}

type ErrorHolder = *ErrorHolderStruct

func NewErrorHolder() ErrorHolder {
	t := &ErrorHolderStruct{
		ErrorList: NewArray[error](),
	}
	return t
}

func (h ErrorHolder) Add(e error) ErrorHolder {
	if e != nil {
		h.ErrorList.Add(e)
		Alert("#50<1Added error to holder:", e)
	}
	return h
}

func (h ErrorHolder) First() error {
	var e error
	if !h.ErrorList.IsEmpty() {
		e = h.ErrorList.First()
	}
	return e
}
