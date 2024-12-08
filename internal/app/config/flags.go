package config

import "flag"

const defaultURI = ":8080"

var (
	flagURI          string
	flagDBURI        string
	flagMusicInfoURL string
)

func parseFlags() {
	flag.StringVar(&flagURI, "a", defaultURI, "server URI")
	flag.StringVar(&flagDBURI, "d", "", "database URI")
	flag.StringVar(&flagMusicInfoURL, "i", "", "music info URL")
	flag.Parse()
}
