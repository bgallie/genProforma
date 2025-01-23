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
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	dbug "runtime/debug"
	"time"

	"github.com/spf13/cobra"
)

var (
	GitCommit  string = "not set"
	GitBranch  string = "not set"
	GitState   string = "not set"
	GitSummary string = "not set"
	GitDate    string = "not set"
	BuildDate  string = "not set"
	Version    string = ""
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long:  `Display version and detailed build information for genProforma.`,
	Run: func(cmd *cobra.Command, args []string) {
		if Version == "" {
			Version = "(devel)"
		}
		fmt.Println("    Version:", Version)
		if len(GitDate) > 1 {
			fmt.Println("Commit Date:", GitDate)
		}
		if len(GitCommit) > 1 {
			fmt.Println("     Commit:", GitCommit)
		}
		fmt.Println("      State:", GitState)
		if GitSummary != "not set" {
			fmt.Println("    Summary:", GitSummary)
		}
		if BuildDate != "not set" {
			fmt.Println(" Build Date:", BuildDate)
		}
	},
}

func getBuildSettings(settings []dbug.BuildSetting, key string) string {
	for _, v := range settings {
		if v.Key == key {
			return v.Value
		}
	}
	return ""
}

func init() {
	// Extract version information from the stored build information.
	bi, ok := dbug.ReadBuildInfo()
	if ok {
		if Version == "" {
			Version = bi.Main.Version
		}
		rootCmd.Version = Version
		GitDate = getBuildSettings(bi.Settings, "vcs.time")
		GitCommit = getBuildSettings(bi.Settings, "vcs.revision")
		if len(GitCommit) > 1 {
			GitSummary = fmt.Sprintf("%s-1-%s", Version, GitCommit[0:7])
		}
		GitState = "clean"
		if getBuildSettings(bi.Settings, "vcs.modified") == "true" {
			GitState = "dirty"
		}
	}

	// Get the build date (as the modified date of the executable) if the build date
	// is not set.
	if BuildDate == "not set" {
		fpath, err := os.Executable()
		cobra.CheckErr(err)
		fpath, err = filepath.EvalSymlinks(fpath)
		cobra.CheckErr(err)
		fsys := os.DirFS(filepath.Dir(fpath))
		fInfo, err := fs.Stat(fsys, filepath.Base(fpath))
		cobra.CheckErr(err)
		BuildDate = fInfo.ModTime().UTC().Format(time.RFC3339)
	}
	rootCmd.AddCommand(versionCmd)
}
