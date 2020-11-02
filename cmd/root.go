/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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

	"github.com/jnpr-tjiang/echo-apisvr/pkg/api"
	"github.com/jnpr-tjiang/echo-apisvr/pkg/config"
	"github.com/spf13/cobra"
	"github.com/xeipuuv/gojsonschema"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "echo-apisvr",
	Short: "Demo api server for CSO data model",
	Long:  `This demo server is meant to learn/test echo framework`,
	Run: func(cmd *cobra.Command, args []string) {
		// TestGoJsonSchema()
		// TestGorm()
		api.Run()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
		api.Run()
	}
}

func init() {
	cobra.OnInitialize(func() {
		config.InitConfig(cfgFile)
	})

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.echo-apisvr.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func TestGoJsonSchema() {
	s := gojsonschema.NewReferenceLoader("file:////Users/tjiang/code/playground/echo-apisvr/schemas/device.json")
	d := gojsonschema.NewReferenceLoader("file:////Users/tjiang/code/playground/echo-apisvr/testdata/data.json")
	result, err := gojsonschema.Validate(s, d)
	if err != nil {
		panic(err)
	}
	if result.Valid() {
		fmt.Printf("The document is valid\n")
	} else {
		fmt.Printf("The document is not valid. see errors :\n")
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
	}
	// jsonMap := make(map[string]interface{})
	// err := yaml.Unmarshal([]byte(ymldata), &jsonMap)
	// if err != nil {
	// 	panic(err)
	// }
}
