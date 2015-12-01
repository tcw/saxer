package tagPath


type TagPath struct {
	Path    []Tag
	PathPos int
}

type Tag struct {
	Name         string
	Attributes   []Attribute
	AttributePos int
}

type Attribute struct {
	Key   string
	Value string
}

func NewTagPath() TagPath{
	return TagPath{make([]Tag, 100),0}
}

func (tp *TagPath) Add(t Tag) {
	tp.Path[tp.PathPos] = t
	tp.PathPos++
}

func NewTag() Tag{
	return Tag{"", make([]Attribute, 100), 0}
}

func (t *Tag) Add(attr Attribute)  {
	t.Attributes[t.AttributePos] = attr
	t.AttributePos++
}

func (t *TagPath) RemoveLast()  {
	t.PathPos--
}