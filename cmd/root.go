/*
Copyright © 2019 Simon Fuhrer

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
	"os"

	"github.com/k0kubun/pp"
	"github.com/spf13/cobra"
	"go.etcd.io/etcd/pkg/transport"
)

// GlobalFlags are flags that defined globally
// and are inherited to all sub-commands.
type GlobalFlags struct {
	Endpoints []string
	TLS       transport.TLSInfo
	User      string
	Password  string
	Debug     bool
	DryRun    bool
}

const (
	cliName = "etcd-manipulator"
	version = "v0.0.1"
)

var (
	rootCmd = &cobra.Command{
		Use:   cliName,
		Short: "A simple command line client for etcd3 to manipulate data (dangerous).",
		Long:  "!!!!Dangerous!!! use at your own risk",
	}
	listcmd = &cobra.Command{
		Use:   "list",
		Short: "list all pvs",
		Run:   listCommandFunc,
	}
	versioncmd = &cobra.Command{
		Use:   "version",
		Short: "Prints the version of the cli tool",
		Run:   versionCommandFunc,
	}
)

var globalFlags = GlobalFlags{}

func listCommandFunc(cmd *cobra.Command, args []string) {
	pp.Println(globalFlags)
}

func versionCommandFunc(cmd *cobra.Command, args []string) {
	fmt.Println(fmt.Sprintf("%s version:", cliName), version)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.EnablePrefixMatching = true
	rootCmd.PersistentFlags().BoolVarP(&globalFlags.Debug, "debug", "d", false, "enable client-side debug logging")
	rootCmd.PersistentFlags().StringSliceVarP(&globalFlags.Endpoints, "endpoints", "e", []string{"127.0.0.1:2379"}, "gRPC endpoints")
	rootCmd.PersistentFlags().StringVar(&globalFlags.TLS.CertFile, "cert", "", "identify secure client using this TLS certificate file")
	rootCmd.PersistentFlags().StringVar(&globalFlags.TLS.KeyFile, "key", "", "identify secure client using this TLS key file")
	rootCmd.PersistentFlags().StringVar(&globalFlags.TLS.TrustedCAFile, "cacert", "", "verify certificates of TLS-enabled secure servers using this CA bundle")
	rootCmd.PersistentFlags().StringVarP(&globalFlags.User, "user", "u", "", "username[:password] for authentication (prompt if password is not supplied)")
	rootCmd.PersistentFlags().StringVarP(&globalFlags.Password, "password", "p", "", "password for authentication (if this option is used, --user option shouldn't include password)")

	listcmd.Flags().BoolVarP(&globalFlags.DryRun, "dry-run", "t", false, "dry-run")

	rootCmd.AddCommand(listcmd)
	rootCmd.AddCommand(versioncmd)

}
