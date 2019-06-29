package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/t94j0/satellite/config"
	"github.com/t94j0/satellite/path"
	"github.com/t94j0/satellite/server"
)

const ProjectName string = "satellite"

var rootCmd = &cobra.Command{
	Use:   ProjectName,
	Short: "Satellite is an intelligent payload generator",
	Long:  `Satellite is an web payload hosting service which filters requests to ensure the correct target is getting a payload. This can also be a useful service for hosting files that should be only accessed in very specific circumstances.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetLevel(log.DebugLevel)

		cfg, err := config.Config(ProjectName)
		if err != nil {
			log.Fatal(err)
		}

		logLevel := cfg.GetString("log_level")
		logOptions := map[string]log.Level{
			"":      log.DebugLevel,
			"panic": log.PanicLevel,
			"fatal": log.FatalLevel,
			"error": log.ErrorLevel,
			"warn":  log.WarnLevel,
			"info":  log.InfoLevel,
			"debug": log.DebugLevel,
			"trace": log.TraceLevel,
		}
		log.SetLevel(logOptions[logLevel])

		serverRoot := cfg.GetString("server_root")
		log.Debugf("Using config file %s", cfg.ConfigFileUsed())
		log.Debugf("Using server path %s", serverRoot)

		// Parse .info files
		paths, err := path.New(serverRoot)
		if err != nil {
			log.Fatal(err)
		}

		// Warn user about redirection opsec
		notFoundRedirect := cfg.GetString("not_found.redirect")
		notFoundRender := cfg.GetString("not_found.render")
		if notFoundRedirect == "" && notFoundRender == "" {
			log.Warn("Use not_found handlers for opsec")
		}

		log.Debugf("Loaded %d path(s)", paths.Len())

		// Listen for when files in serverRoot change
		go func() {
			if err := createWatcher(serverRoot, "1s", func() error {
				return paths.Reload()
			}); err != nil {
				log.Fatal(err)
			}
		}()

		log.Infof("Listening HTTPS on port %s", cfg.GetString("listen"))

		// Create server and listen
		if err := server.Start(paths, cfg); err != nil {
			log.Fatal(err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
