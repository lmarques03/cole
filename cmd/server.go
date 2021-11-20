/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		wg, ctx := errgroup.WithContext(ctx)

		term := make(chan os.Signal, 1)
		signal.Notify(term, os.Interrupt, syscall.SIGTERM)

		srv := createHttpServer(serverPort)

		wg.Go(func() error {
			if err := srv.ListenAndServe(); err != http.ErrServerClosed {
				// logrus.Error("msg", "http server error", err, err)
				return err
			}
			return nil
		})

		select {
		case <-term:
			// logrus.Info("received SIGTERM, exiting gracefully...")
		case <-ctx.Done():
		}

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("server shutdown error ", err)
		}

		cancel()

		if err := wg.Wait(); err != nil {
			// logrus.Error("unhandled error received. Exiting...", err)
			os.Exit(1)
		}

		os.Exit(0)

	},
}

var serverPort = ""

func init() {
	serverCmd.Flags().StringVar(&serverPort, "http.port", ":9754", "listem port for http endpoints")
	rootCmd.AddCommand(serverCmd)

}

func createHttpServer(port string) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/-/health", func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Header().Set("Content-Type", "application/json")

		resp := map[string]string{
			"message": "Healthy",
		}

		jsonResp, err := json.Marshal(resp)

		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}

		rw.Write(jsonResp)
	})
	mux.HandleFunc("/-/ready", func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Header().Set("Content-Type", "application/json")

		resp := map[string]string{
			"message": "Ready",
		}

		jsonResp, err := json.Marshal(resp)

		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}

		rw.Write(jsonResp)
	})

	srv := &http.Server{
		Addr:     port,
		Handler:  mux,
		ErrorLog: &log.Logger{},
	}
	return srv
}
