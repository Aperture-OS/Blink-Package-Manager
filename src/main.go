/*
  Blink, a powerful source-based package manager. Core of ApertureOS.
	Want to use it for your own project?
	Blink is completely FOSS (Free and Open Source),
	edit, publish, use, contribute to Blink however you prefer.
  Copyright (C) 2025 Aperture OS

  This program is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  This program is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main // main package, entry point

import (
	"fmt"           // For formatted I/O (printing, formatting strings)
	"log"           // For logging info, warnings, and errors
	"os"            // For file and directory operations
	"path/filepath" // For handling file paths in a cross-platform way

	"github.com/spf13/cobra" // Cobra CLI framework (as u might guess lol)
)


//===================================================================//
//							  Functions
//===================================================================//

/****************************************************/
// main function - entry point of the Blink package manager
// sets up Cobra CLI commands and flags
// handles user input and executes corresponding functions
// provides commands for downloading, fetching info, installing packages
// also includes support and version commands
// uses modular functions for package operations to enhance readability and maintainability
/****************************************************/

// these comments on the main func should prob be added but theyre so boring so imma skip them
// if anyone wanna add them feel free ;)
func main() {

	log.SetFlags(log.Ltime)         // Log only time, no date as its useless
	log.SetPrefix("[Blink Debug] ") // Log prefix for debug messages (eg. [Blink Debug] INFO: { ... } )

	// Flags for CLI commands
	var force bool  // Force re-download or reinstall
	var path string // Custom cache path

	// Root Cobra command
	rootCmd := &cobra.Command{
		Use:   "blink",
		Short: "Blink - lightweight, source-based package manager for Aperture OS",
		Long:  "Blink - lightweight, fast, source-based package manager for Aperture OS and Unix-like systems.",
	}

	//  blink get <pkg>
	getCmd := &cobra.Command{
		Use:     "get <pkg>",
		Short:   "Download a package recipe (JSON file)",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"d", "download", "g", "dl"},
		Run: func(cmd *cobra.Command, args []string) {
			pkgName := args[0]
			if path == "" {
				path = filepath.Join(defaultCachePath, "recipes")
			}
			if err := getpkg(pkgName, path); err != nil {
				log.Fatalf("Error fetching package: %v", err)
			}
		},
	}
	getCmd.Flags().BoolVarP(&force, "force", "f", false, "Force re-download")
	getCmd.Flags().StringVarP(&path, "path", "p", defaultCachePath, "Specify cache directory")

	//  blink info <pkg>
	infoCmd := &cobra.Command{
		Use:     "info <pkg>",
		Short:   "Fetch & display package information",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"information", "pkginfo", "details", "fetch"},
		Run: func(cmd *cobra.Command, args []string) {
			pkgName := args[0]
			if path == "" {
				path = filepath.Join(defaultCachePath, "recipes")
			}
			if _, err := fetchpkg(path, force, pkgName); err != nil {
				log.Fatalf("Error reading package info: %v", err)
			}
		},
	}
	infoCmd.Flags().BoolVarP(&force, "force", "f", false, "Force re-download")
	infoCmd.Flags().StringVarP(&path, "path", "p", defaultCachePath, "Specify cache directory")

	//  blink install <pkg>
	installCmd := &cobra.Command{
		Use:     "install <pkg>",
		Short:   "Download and install a package",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"i", "add", "inst"},
		Run: func(cmd *cobra.Command, args []string) {

			requireRoot() // ensure running as root

			pkgName := args[0]
			if path == "" {
				path = filepath.Join(defaultCachePath, "recipes")
			}
			if err := install(pkgName, force, path); err != nil {
				log.Fatalf("Error installing package: %v", err)
			}
		},
	}
	installCmd.Flags().BoolVarP(&force, "force", "f", false, "Force reinstall")
	installCmd.Flags().StringVarP(&path, "path", "p", defaultCachePath, "Specify cache directory")

	//  blink support
	supportCmd := &cobra.Command{
		Use:     "support",
		Aliases: []string{"issue", "bug", "contact", "discord", "s", "-s", "--support", "--bug"},
		Short:   "Show support information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s", supportPage)
		},
	}

	cleanCmd := &cobra.Command{
		Use:     "clean",
		Aliases: []string{"cleanup", "clear", "c", "-c", "--clean", "--cleanup"},
		Short:   "Clean cache info.",
		Run: func(cmd *cobra.Command, args []string) {

			requireRoot() // ensure running as root

			clean()
		},
	}

	//  blink version
	versionCmd := &cobra.Command{
		Use:     "version",
		Aliases: []string{"v", "ver", "--version", "-v"},
		Short:   "Show Blink version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s", versionPage)
		},
	}

	// Add commands to root
	rootCmd.AddCommand(getCmd, infoCmd, installCmd, supportCmd, versionCmd, cleanCmd)

	// Disable default Cobra completion
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	//  Shell completion command
	completionCmd := &cobra.Command{
		Use:       "completion [bash|zsh|fish]",
		Short:     "Generate shell completion scripts",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish"},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				return rootCmd.GenFishCompletion(os.Stdout, true)
			default:
				return cmd.Help()
			}
		},
	}
	rootCmd.AddCommand(completionCmd)

	// Print welcome message
	fmt.Printf("Blink Package Manager Version: %s\n", Version)
	fmt.Printf("© Copyright 2025-%d Aperture OS. All rights reserved.\n", currentYear)

	// Execute root command
	if err := rootCmd.Execute(); err != nil {
		log.Printf("FATAL: Command Line Interface failed to run. (Is there any syntax error(s)?)\nERR: %v ", err)
		os.Exit(1)
	}
}

/********************************************************************************************************/

// if ur reading this pls contribute to the repository if its out :sob:
