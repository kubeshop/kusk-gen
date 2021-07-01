package ambassador

import (
	"encoding/json"
)

type header struct {
	Swagger string `json:"swagger"`
	OpenAPI string `json:"openapi"` // we might need that to distinguish 3.1.x vs 3.0.x
}

// isSwagger tries to decode the spec header
func isSwagger(spec []byte) bool {
	var header header

	// we can ignore the error here
	_ = json.Unmarshal(spec, &header)

	return header.Swagger != ""
}
