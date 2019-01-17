package locale

import (
	"fmt"
	"legion/es"
)

func Translate(locale, text string, params ...es.Info) string {
	return fmt.Sprintf("translation for %s: %q", locale, text)
}
