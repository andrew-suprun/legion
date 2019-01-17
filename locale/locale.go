package locale

import (
	"fmt"

	"github.com/andrew-suprun/legion/es"
)

func Translate(locale, text string, params ...es.Info) string {
	return fmt.Sprintf("translation for %s: %q", locale, text)
}
