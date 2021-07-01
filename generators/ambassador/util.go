package ambassador

import (
	"encoding/json"
)

type header struct {
	Swagger string `json:"swagger"`
	OpenAPI string `json:"openapi"`
}

// isSwagger tries to decode the spec header
func isSwagger(spec []byte) bool {
	var header header

	// we can ignore the error here
	_ = json.Unmarshal(spec, &header)

	if header.Swagger != "" {
		return true
	}

	return false
}
