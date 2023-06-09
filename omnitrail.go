package omnitrail

func NewTrail(option ...Option) Factory {
	o := &Options{}
	for _, opt := range option {
		opt(o)
	}
	if o.Sha1Enabled == false && o.Sha256Enabled == false {
		o.Sha1Enabled = true
	}
	plugins := make([]Plugin, 0)
	plugins = append(plugins, NewFilePlugin())
	plugins = append(plugins, NewDirectoryPlugin())
	plugins = append(plugins, NewPosixPlugin())
	return &factoryImpl{
		Options: o,
		Plugins: plugins,
		envelope: &Envelope{
			Header: Header{
				Features: make(map[string]Feature),
			},
			Mapping: make(map[string]*Element),
		},
	}
}
