package main

type config struct {
	LauncherUrl string         `yaml:"LauncherUrl"`
	Archive     string         `yaml:"Archive"`
	Whitelist   []string       `yaml:"Whitelist"`
	Ftp         configFtp      `yaml:"Ftp"`
	Regions     []configRegion `yaml:"Regions"`
}

type configFtp struct {
	Host     string `yaml:"Host"`
	Port     int    `yaml:"Port"`
	User     string `yaml:"User"`
	Password string `yaml:"Password"`
	Path     string `yaml:"Path"`
	Timeout  int    `yaml:"Timeout"`
}

type configRegion struct {
	RegionId int    `yaml:"RegionId"`
	WorkDir  string `yaml:"WorkDir"`
	Start    string `yaml:"Start"`
	Stop     string `yaml:"Stop"`
	Bat      string `yaml:"Bat"`
}

func (c *config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawConfig config
	raw := rawConfig{
		LauncherUrl: "127.0.0.1:9599",
		Archive:     "servers",
		Ftp: configFtp{
			Timeout: 10000,
		},
	}
	if err := unmarshal(&raw); err != nil {
		return err
	}
	*c = config(raw)
	return nil
}
