package valueobject

import (
	"fmt"

	"github.com/samber/lo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Title string

var _ fmt.Stringer = Title("")

// NewTitle allocates 'v' as a [Title].
func NewTitle(v string, langTag language.Tag) Title {
	langTag = lo.CoalesceOrEmpty(langTag, language.English)
	titleBuilder := cases.Title(langTag)
	return Title(titleBuilder.String(v))
}

func (t Title) String() string {
	return string(t)
}
