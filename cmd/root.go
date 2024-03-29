/*
 * Licensed to Apache Software Foundation (ASF) under one or more contributor
 * license agreements. See the NOTICE file distributed with
 * this work for additional information regarding copyright
 * ownership. Apache Software Foundation (ASF) licenses this file to you under
 * the Apache License, Version 2.0 (the "License"); you may
 * not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/apache/skywalking-kubernetes-event-exporter/assets"
	"github.com/apache/skywalking-kubernetes-event-exporter/configs"
	"github.com/apache/skywalking-kubernetes-event-exporter/internal/pkg/logger"
)

var (
	verbosity  string
	configFile string
)

var rootCmd = &cobra.Command{
	Use:           "skywalking-kubernetes-event-exporter",
	Long:          "Export Kubernetes events to Apache SkyWalking",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		level, err := logrus.ParseLevel(verbosity)
		if err != nil {
			return err
		}
		logger.Log.SetLevel(level)

		content := assets.DefaultConfig
		if configFile != "" {
			c, err := os.ReadFile(configFile)
			if err != nil {
				return err
			}
			content = c
		}
		if err := configs.ParseConfig(content); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&verbosity, "verbosity", "v", logrus.InfoLevel.String(), "log level (debug, info, warn, error, fatal, panic")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "the config file")
}
