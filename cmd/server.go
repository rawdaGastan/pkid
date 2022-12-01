/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rawdaGastan/pkid/internal"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pkid",
	Short: "A command to start the pkid server",
	Run: func(cmd *cobra.Command, args []string) {
		logger := zerolog.New(os.Stdout).With().Logger()

		filePath, err := cmd.Flags().GetString("file")
		if err != nil {
			logger.Error().Msg(fmt.Sprint("command failed with error: ", err))
			return
		}

		if filePath == "" {
			logger.Error().Msg("command with error: no file path provided")
			return
		}

		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			logger.Error().Msg(fmt.Sprint("command failed with error: ", err))
			return
		}

		// adding logging mw
		mws := []mux.MiddlewareFunc{}
		loggingMw := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Println(r.RequestURI)
				// Call the next handler, which can be another middleware in the chain, or the final handler.
				next.ServeHTTP(w, r)
			})
		}
		mws = append(mws, loggingMw)

		pkidStore := internal.NewSqliteStore()
		server, err := internal.NewServer(logger, mws, pkidStore, filePath, port)
		if err != nil {
			logger.Error().Msg(fmt.Sprint("new pkid server failed with error: ", err))
			return
		}

		err = server.Start()
		if err != nil {
			logger.Error().Msg(fmt.Sprint("start pkid server failed with error: ", err))
			return
		}

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

	rootCmd.Flags().StringP("file", "f", "", "Enter your file path for DB")
	rootCmd.Flags().IntP("port", "p", 3000, "Enter the port you want the server to run at")

}
