// Package main provides the entry point for Jurigen - Legal Case Context Builder
//
// @title Jurigen API
// @version 1.0
// @description Legal Case Context Builder API - Build context for legal cases using directed acyclic graphs
// @description
// @description This microservice provides a context builder for legal cases using directed acyclic graphs (DAGs).
// @description Users traverse through question nodes to build comprehensive case context with evidence tracking,
// @description timeline management, and legal assessment capabilities.
//
// @contact.name Legal Tech Support
// @contact.email support@legaltech.example.com
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:8080
// @BasePath /v1
//
// @tag.name DAGs
// @tag.description Operations for managing and retrieving Legal Case DAGs
//
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Bearer token authentication
package main

import (
	"davidterranova/jurigen/backend/cmd"

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
