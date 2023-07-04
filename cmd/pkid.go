package cmd

import (
	"fmt"
	"os"

	"github.com/rawdaGastan/pkid/app"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pkid",
	Short: "A command to start the pkid server",
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile, err := cmd.Flags().GetString("config")
		if err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}

		app, err := app.NewApp(cmd.Context(), configFile)
		if err != nil {
			return fmt.Errorf("failed to create new app: %w", err)
		}

		err = app.Start(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to start app: %w", err)
		}

		return nil

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize()

	rootCmd.Flags().StringP("config", "c", "./config.json", "Enter your configurations path")
}
