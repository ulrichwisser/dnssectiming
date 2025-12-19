/*
Copyright © 2023 Ulrich Wisser <ulrich@wisser.se>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"os"

	homedir "github.com/mitchellh/go-homedir"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/apex/log"

	_ "github.com/go-sql-driver/mysql"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "dnssectiming [-v] <command>",
	Version: "0.0.1a",
	Short:   "get dnssec timing information",
	Long:    `get dnssec timing information`,
}

func init() {
	cobra.OnInitialize(initConfig)

	// define command line arguments
	rootCmd.PersistentFlags().CountP(VERBOSE, "v", "repeat for more verbose printouts")
	rootCmd.PersistentFlags().StringP(RR, RR_SHORT, RR_DEFAULT, RR_DESCRIPTION)
	rootCmd.PersistentFlags().StringP(TLD, TLD_SHORT, TLD_DEFAULT, TLD_DESCRIPTION)

	// Use flags for viper values
	viper.BindPFlags(rootCmd.Flags())
	viper.BindPFlags(rootCmd.PersistentFlags())
}

func initConfig() {

	// Use flags for viper values
	viper.BindPFlags(pflag.CommandLine)

	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	// Search config in home directory with name ".umpy" (without extension).
	viper.SetConfigName(".dnssect")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(home)
	viper.AddConfigPath(".")

	// read in environment variables that match
	viper.SetEnvPrefix("DNSSECT")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if len(viper.GetString("config")) > 0 {
			log.Fatalf("Error reading config file: '%s' %s", viper.GetString("config"), err.Error())
		}
		log.Info(err.Error())
	} else {
		log.Debugf("Using config file: %s", viper.ConfigFileUsed())
	}

	// init log level
	switch viper.GetInt(VERBOSE) {
	case VERBOSE_QUIET:
		log.SetLevel(log.FatalLevel)
	case VERBOSE_ERROR:
		log.SetLevel(log.ErrorLevel)
	case VERBOSE_WARNING:
		log.SetLevel(log.WarnLevel)
	case VERBOSE_INFO:
		log.SetLevel(log.InfoLevel)
	case VERBOSE_DEBUG:
		log.SetLevel(log.DebugLevel)
	default:
		if viper.GetInt(VERBOSE) > VERBOSE_DEBUG {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.ErrorLevel)
		}
	}
}

func Execute() {
	// Now run the command
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
