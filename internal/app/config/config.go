package config

import "os"

// Config provides server configuration parameters.
type Config struct {
	URI          string
	DBURI        string
	DBDriver     string
	MusicInfoURL string
}

// Setup calculates server configuration parameters.
func Setup() (*Config, error) {

	var cfg Config
	parseFlags()
	if cfg.URI = os.Getenv("SERVER_URI"); len(cfg.URI) == 0 {
		cfg.URI = flagURI
	}
	if cfg.DBURI = os.Getenv("DB_URI"); len(cfg.DBURI) == 0 {
		cfg.DBURI = flagDBURI
	}
	if cfg.MusicInfoURL = os.Getenv("MUSIC_INFO_URL"); len(cfg.MusicInfoURL) == 0 {
		cfg.MusicInfoURL = flagMusicInfoURL
	}

	cfg.DBDriver = "pgx"
	return &cfg, nil
}
