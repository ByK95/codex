package enum

type Enum struct {
	strToIdx map[string]int
	idxToStr []string
}

func NewEnum() *Enum {
	return &Enum{
		strToIdx: make(map[string]int),
		idxToStr: make([]string, 0),
	}
}

func (e *Enum) Add(value string) int {
	if idx, ok := e.strToIdx[value]; ok {
		return idx
	}
	idx := len(e.idxToStr)
	e.strToIdx[value] = idx
	e.idxToStr = append(e.idxToStr, value)
	return idx
}

func (e *Enum) GetIndex(value string) (int, bool) {
	idx, ok := e.strToIdx[value]
	return idx, ok
}

func (e *Enum) GetValue(index int) (string, bool) {
	if index < 0 || index >= len(e.idxToStr) {
		return "", false
	}
	return e.idxToStr[index], true
}
