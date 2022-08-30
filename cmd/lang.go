/*
Package cmd
Copyright © 2022 Noah Hsu<i@nn.ci>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/alist-org/alist/v3/internal/bootstrap/data"
	log "github.com/sirupsen/logrus"

	_ "github.com/alist-org/alist/v3/drivers"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/spf13/cobra"
)

type KV[V any] map[string]V

type Drivers KV[KV[interface{}]]

func firstUpper(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func convert(s string) string {
	ss := strings.Split(s, "_")
	ans := strings.Join(ss, " ")
	return firstUpper(ans)
}

func generateDriversJson() {
	drivers := make(Drivers)
	drivers["drivers"] = make(KV[interface{}])
	driverItemsMap := operations.GetDriverInfoMap()
	for k, v := range driverItemsMap {
		drivers["drivers"][k] = k
		items := make(KV[interface{}])
		for i := range v.Additional {
			item := v.Additional[i]
			items[item.Name] = convert(item.Name)
			if item.Help != "" {
				items[fmt.Sprintf("%s-tips", item.Name)] = item.Help
			}
			if item.Type == conf.TypeSelect && len(item.Options) > 0 {
				options := make(KV[string])
				_options := strings.Split(item.Options, ",")
				for _, o := range _options {
					options[o] = convert(o)
				}
				items[fmt.Sprintf("%ss", item.Name)] = options
			}
		}
		drivers[k] = items
	}
	utils.WriteJsonToFile("lang/drivers.json", drivers)
}

func generateSettingsJson() {
	settings := data.InitialSettings()
	settingsLang := make(KV[any])
	for _, setting := range settings {
		settingsLang[setting.Key] = convert(setting.Key)
		if setting.Help != "" {
			settingsLang[fmt.Sprintf("%s-tips", setting.Key)] = setting.Help
		}
		if setting.Type == conf.TypeSelect && len(setting.Options) > 0 {
			options := make(KV[string])
			_options := strings.Split(setting.Options, ",")
			for _, o := range _options {
				options[o] = convert(o)
			}
			settingsLang[fmt.Sprintf("%ss", setting.Key)] = options
		}
	}
	utils.WriteJsonToFile("lang/settings.json", settingsLang)
}

// langCmd represents the lang command
var langCmd = &cobra.Command{
	Use:   "lang",
	Short: "Generate language json file",
	Run: func(cmd *cobra.Command, args []string) {
		err := os.MkdirAll("lang", 0777)
		if err != nil {
			log.Fatal("failed create folder: %s", err.Error())
		}
		generateDriversJson()
		generateSettingsJson()
	},
}

func init() {
	rootCmd.AddCommand(langCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// langCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// langCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
