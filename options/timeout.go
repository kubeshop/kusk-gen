package options

type TimeoutOptions struct {
	// RequestTimeout is total request timeout
	RequestTimeout uint32 `yaml:"request_timeout,omitempty" json:"request_timeout,omitempty"`
	// IdleTimeout is timeout for idle connection
	IdleTimeout uint32 `yaml:"idle_timeout,omitempty" json:"idle_timeout,omitempty"`
}

func (o *TimeoutOptions) Validate() error {
	return nil
}
