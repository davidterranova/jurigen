package main

import (
	"davidterranova/jurigen/cmd"

	"github.com/rs/zerolog/log"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.
			Fatal().
			Err(err).
			Msg("failed to start contacts")
	}
}
