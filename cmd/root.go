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

	"github.com/spf13/cobra"
	"gorm.io/datatypes"

	"github.com/jnpr-tjiang/echo-apisvr/internal/config"
	"github.com/jnpr-tjiang/echo-apisvr/internal/database"
	"github.com/jnpr-tjiang/echo-apisvr/internal/models"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "echo-apisvr",
	Short: "Demo api server for CSO data model",
	Long:  `This demo server is meant to learn/test echo framework`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.Init()
		if err != nil {
			panic(fmt.Sprintf("Failed to initialize the database: %v", err))
		}

		domain := models.Domain{
			Base: models.BaseModel{
				Name:    "default",
				Payload: datatypes.JSON([]byte(`{"display_name": "default", "system": {"serial": "SN1234", "mac":"ab:34:12:f3"}}`)),
			},
		}
		db.Create(&domain)

		project := models.Project{
			Base: models.BaseModel{
				Name:     "juniper",
				ParentID: domain.Base.ID,
			},
		}
		db.Create(&project)

		id := domain.Base.ID
		domain = models.Domain{}
		db.First(&domain, id)
		fmt.Printf("domain[%s]", domain.Base.Name)
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
