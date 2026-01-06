package config

type ConfigProvider interface {
	Get() (Config, error)
}

type ConfigProviderObject struct {
	config *Config
	loaded bool
	err    error
}

func (o *ConfigProviderObject) Get() (Config, error) {
	if o.loaded {
		return *o.config, o.err
	}
	o.config, o.err = LoadConfig()
	o.loaded = true
	return *o.config, o.err
}

func NewConfigProvider() *ConfigProviderObject {
	return &ConfigProviderObject{
		config: nil,
		loaded: false,
		err:    nil,
	}
}
