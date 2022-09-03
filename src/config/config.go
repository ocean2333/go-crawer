package config

type Config struct {
	SavePath string
}

var config Config

func init() {
	config = Config{
		SavePath: "../zip",
	}
}

func Get() Config {
	return config
}
