package omnitrail

import (
	"fmt"
	"sort"
)

type PluginInit func() Plugin

var pluginMap = make(map[string]PluginInit)

func RegisterPlugin(name string, initFn PluginInit) {
	pluginMap[name] = initFn
}

func NewTrail(option ...Option) Factory {
	o := &Options{}
	for _, opt := range option {
		opt(o)
	}
	if o.Sha1Enabled == false && o.Sha256Enabled == false {
		o.Sha1Enabled = true
	}
	allowList := []string{}
	plugins := make([]Plugin, 0)
	// Directory plugin depends on the File plugin
	// We assume all other plugins depend on both the File and Directory plugins
	plugins = append(plugins, NewFilePlugin())
	plugins = append(plugins, NewDirectoryPlugin())
	// We load all other plugins here
	for _, pluginInitFunc := range pluginMap {
		plugins = append(plugins, pluginInitFunc())
	}

	fmt.Println(plugins)

	factory := &factoryImpl{
		Options: o,
		Plugins: plugins,
		envelope: &Envelope{
			Header: Header{
				Features: make(map[string]Feature),
			},
			Mapping: make(map[string]*Element),
		},
		AllowList: allowList,
	}

	return factory
}

func FormatADGString(mapping Factory) string {
	res := ""
	sha1adgs := mapping.Sha1ADGs()
	// create a list of all keys sorted in lexical order
	keys := make([]string, 0, len(sha1adgs))
	for k := range sha1adgs {
		keys = append(keys, k)
	}
	// sort the keys
	sort.Strings(keys)

	for _, k := range keys {
		v := sha1adgs[k]
		if v != "" {
			res += fmt.Sprintln(k)
			res += fmt.Sprintln(v)
			res += fmt.Sprintln("--")
		}
	}
	res += fmt.Sprintln("----")

	keys = make([]string, 0, len(sha1adgs))
	sha2adgs := mapping.Sha256ADGs()
	keys = make([]string, 0, len(sha2adgs))
	for k := range sha2adgs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := sha2adgs[k]
		if v != "" {
			res += fmt.Sprintln(k)
			res += fmt.Sprintln(v)
			res += fmt.Sprintln("--")
		}
	}
	return res
}
