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


//Construction structs up front and reusing them for performance gain
func NewTagPath() *TagPath {
	tp := &TagPath{make([]Tag, 100), 0}
	for i := 0; i < len(tp.Path); i++ {
		tp.Path[i] = newTag()
		for j := 0; j < len(tp.Path[i].Attributes); j++ {
			tp.Path[i].Attributes[j] = newAttribute()
		}
	}
	return tp
}

func newTag() Tag {
	return Tag{"", make([]Attribute, 100), 0}
}

func newAttribute() Attribute {
	return Attribute{"", ""}
}

func (tp *TagPath) NextTag() *Tag {
	tag := &tp.Path[tp.PathPos]
	tp.PathPos++
	return tag
}

func (tg *Tag) NextAttribute() *Attribute {
	attr := &tg.Attributes[tg.AttributePos]
	tg.AttributePos++
	return attr
}

func (t *Tag) AddAttribute(key string, value string) {
	t.Attributes[t.AttributePos].Key = key
	t.Attributes[t.AttributePos].Value = value
	t.AttributePos++
}

func (t *TagPath) RemoveLast() {
	t.PathPos--
	t.Path[t.PathPos].AttributePos = 0
}