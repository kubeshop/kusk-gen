package ambassador

import (
	"fmt"
	"strings"

	"github.com/kubeshop/kusk-gen/options"
)

type corsTemplateData struct {
	Origins        string
	Methods        string
	Headers        string
	ExposedHeaders string

	Credentials bool
	MaxAge      string
}

func newCorsTemplateData(corsOpts *options.CORSOptions) corsTemplateData {
	res := corsTemplateData{
		Origins:        strings.Join(corsOpts.Origins, ","),
		Methods:        strings.Join(corsOpts.Methods, ","),
		Headers:        strings.Join(corsOpts.Headers, ","),
		ExposedHeaders: strings.Join(corsOpts.ExposeHeaders, ","),
		MaxAge:         fmt.Sprint(corsOpts.MaxAge),
	}

	if corsOpts.Credentials != nil {
		res.Credentials = *corsOpts.Credentials
	}

	return res
}
