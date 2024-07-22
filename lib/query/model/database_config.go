package model

type Config struct {
	Uri     string
	MinSize int
	MaxSize int
	Ssl     bool
}

type DatabaseConfiguration struct {
	Client string
	Config Config
}
