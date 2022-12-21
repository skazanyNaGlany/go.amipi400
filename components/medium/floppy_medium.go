package medium

type FloppyMedium struct {
	MediumBase

	fullyCached bool
}

func (fm *FloppyMedium) IsFullyCached() bool {
	return fm.fullyCached
}

func (fm *FloppyMedium) SetFullyCached(fullyCached bool) {
	fm.fullyCached = fullyCached
}
