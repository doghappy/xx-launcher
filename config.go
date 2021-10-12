package main

type config struct {
	LauncherUrl string   `yaml:"LauncherUrl"`
	Whitelist   []string `yaml:"Whitelist"`
	Ftp         struct {
		Host     string `yaml:"Host"`
		Port     int    `yaml:"Port"`
		User     string `yaml:"User"`
		Password string `yaml:"Password"`
		Path     string `yaml:"Path"`
	} `yaml:"Ftp"`
	Regions []configRegion `yaml:"Regions"`
}

type configRegion struct {
	RegionId int    `yaml:"RegionId"`
	WorkDir  string `yaml:"WorkDir"`
	Start    string `yaml:"Start"`
	Stop     string `yaml:"Stop"`
}
