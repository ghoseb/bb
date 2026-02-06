package cmdutil

import (
	"encoding/json"
	"io"
)

// WriteJSON writes v as indented JSON to w
func WriteJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
