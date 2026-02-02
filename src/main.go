/*
  Blink, a powerful source-based package manager. Core of ApertureOS.
	Want to use it for your own project?
	Blink is completely FOSS (Free and Open Source),
	edit, publish, use, contribute to Blink however you prefer.
  Copyright (C) 2025-2026 Aperture OS

  This program is free software: you can redistribute it and/or modify
  it under the terms of the Apache 2.0 License as published by
  the Apache Software Foundation, either version 2.0 of the License, or
  any later version.

  This program is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

  You should have received a copy of the Apache 2.0 License
  along with this program.  If not, see <https://www.apache.org/licenses/LICENSE-2.0>.
*/

package main // main package, entry point

import (
	"context"
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/fang" // For fancy terminal output
	"github.com/spf13/cobra"

	"github.com/Aperture-OS/eyes"
	"github.com/fatih/color"
)

// main function - entry point of the Blink package manager
// sets up Cobra CLI commands and flags
// handles user input and executes corresponding functions
// provides commands for downloading, fetching info, installing packages
// also includes support and version commands
// uses modular functions for package operations to enhance readability and maintainability

func colorScheme(ld lipgloss.LightDarkFunc) fang.ColorScheme {
	return fang.ColorScheme{
		Base: ld(
			lipgloss.Color("#45455e"),
			lipgloss.Color("#d4d8e0"),
		),

		Title: lipgloss.Color("#d4d8e0"),

		Description: ld(
			lipgloss.Color("#6d7592"),
			lipgloss.Color("#9299af"),
		),

		Codeblock: ld(
			lipgloss.Color("#383649"),
			lipgloss.Color("#45455e"),
		),

		Program: lipgloss.Color("#6d7592"),

		DimmedArgument: ld(
			lipgloss.Color("#9299af"),
			lipgloss.Color("#afb4c4"),
		),

		Comment: ld(
			lipgloss.Color("#6d7592"),
			lipgloss.Color("#6d7592"),
		),

		Flag:        lipgloss.Color("#9299af"),
		FlagDefault: lipgloss.Color("#afb4c4"),

		Command: lipgloss.Color("#6d7592"),

		QuotedString: lipgloss.Color("#9299af"),

		Argument: ld(
			lipgloss.Color("#9299af"),
			lipgloss.Color("#afb4c4"),
		),

		Help: ld(
			lipgloss.Color("#9299af"),
			lipgloss.Color("#afb4c4"),
		),

		Dash: ld(
			lipgloss.Color("#d4d8e0"),
			lipgloss.Color("#d4d8e0"),
		),

		ErrorDetails: lipgloss.Color("#f38ba8"),
	}
}

func main() {

	eyes.SetLoggerConfiguration(eyes.LoggerConfiguration{
		DisplayName:      "BLINK",
		PrefixTemplate:   "[{display_name}] {timestamp} {log_level}: ",
		TimestampFormat:  "15:04:05",
		InfoTextColor:    color.New(color.FgHiBlue),
		WarnTextColor:    color.New(color.FgHiYellow),
		SuccessTextColor: color.New(color.FgHiGreen),
		FatalTextColor:   color.New(color.BgRed, color.Bold, color.FgWhite),
	})

	// Flags for CLI commands
	var force bool  // Force re-download or reinstall
	var path string // Custom cache path
	var root = DefaultRoot

	//  Root command
	rootCmd := &cobra.Command{
		Use:   "blink",
		Short: fmt.Sprintf("Blink - lightweight, source-based package manager for %s", DistroName),
		Long:  fmt.Sprintf("Blink - lightweight, fast, source-based package manager for %s and Linux systems.", DistroName),
	}

	//  blink get <pkg>
	getCmd := &cobra.Command{
		Use:     "get <pkg>",
		Short:   "Download a package recipe (JSON file)",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"d", "download", "g", "dl"},
		Run: func(cmd *cobra.Command, args []string) {

			requireRoot() // ensure running as root

			if err := ApplyRoot(root); err != nil {
				eyes.Fatalf("Invalid root: %v", err)
			}
			if err := EnsureConfig(); err != nil {
				eyes.Fatalf("Failed to ensure config: %v", err)
			}

			_, err := LoadConfig()
			if err != nil {
				eyes.Fatalf("Failed to load repositories: %v", err)
			}

			if path == "" {
				path = RecipeDirPath
			}

			for _, pkgName := range args {
				if err := getpkg(pkgName, path); err != nil {
					eyes.Errorf("Failed to fetch %s: %v", pkgName, err)
					return
				}
			}

		},
	}

	//  blink search <pkg>
	infoCmd := &cobra.Command{
		Use:     "search <pkg>",
		Short:   "Fetch & display package information",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"information", "pkginfo", "details", "fetch", "info", "f", "searchfor"},
		Run: func(cmd *cobra.Command, args []string) {

			requireRoot() // ensure running as root

			if err := ApplyRoot(root); err != nil {
				eyes.Fatalf("Invalid root: %v", err)
			}
			if err := EnsureConfig(); err != nil {
				eyes.Fatalf("Failed to ensure config: %v", err)
			}

			_, err := LoadConfig()
			if err != nil {
				eyes.Fatalf("Failed to load repositories: %v", err)
			}

			if path == "" {
				path = RecipeDirPath
			}

			for _, pkgName := range args {
				if _, err := fetchpkg(path, force, pkgName, false); err != nil {
					eyes.Errorf("Failed to fetch info for %s: %v", pkgName, err)
					return
				}
			}

		},
	}

	//  blink install <pkg>
	installCmd := &cobra.Command{
		Use:     "install <pkg>",
		Short:   "Download and install a package",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"i", "add", "inst"},
		Run: func(cmd *cobra.Command, args []string) {

			requireRoot() // ensure running as root

			if err := ApplyRoot(root); err != nil {
				eyes.Fatalf("Invalid root: %v", err)
			}
			if err := EnsureConfig(); err != nil {
				eyes.Fatalf("Failed to ensure config: %v", err)
			}

			_, err := LoadConfig()
			if err != nil {
				eyes.Fatalf("Failed to load repositories: %v", err)
			}

			if path == "" {
				path = RecipeDirPath
			}

			for _, pkgName := range args {
				eyes.Infof("Processing package: %s", pkgName)

				if err := install(pkgName, force, path); err != nil {
					eyes.Errorf("Failed to install %s: %v", pkgName, err)
					return
				}
			}

		},
	}

	//  blink uninstall <pkg>
	uninstallCmd := &cobra.Command{
		Use:     "uninstall <pkg>",
		Short:   "Download and install a package",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"remove", "u", "uninst"},
		Run: func(cmd *cobra.Command, args []string) {

			requireRoot() // ensure running as root

			if err := ApplyRoot(root); err != nil {
				eyes.Fatalf("Invalid root: %v", err)
			}
			if err := EnsureConfig(); err != nil {
				eyes.Fatalf("Failed to ensure config: %v", err)
			}

			_, err := LoadConfig()
			if err != nil {
				eyes.Fatalf("Failed to load repositories: %v", err)
			}

			if path == "" {
				path = RecipeDirPath
			}

			for _, pkgName := range args {
				eyes.Infof("Processing package: %s", pkgName)

				if err := uninstall(pkgName, force, path); err != nil {
					eyes.Errorf("Failed to uninstall %s: %v", pkgName, err)
					return
				}
			}

		},
	}

	// Sync command for syncing the package repository
	syncCmd := &cobra.Command{
		Use:     "sync",
		Short:   "Syncs the package repository to the latest version.",
		Args:    cobra.NoArgs,
		Aliases: []string{"s", "--sync", "repo", "reposync"},
		Run: func(cmd *cobra.Command, args []string) {

			requireRoot() // ensure running as root

			if err := ApplyRoot(root); err != nil {
				eyes.Fatalf("Invalid root: %v", err)
			}
			if err := EnsureConfig(); err != nil {
				eyes.Fatalf("Failed to ensure config: %v", err)
			}

			_, err := LoadConfig()
			if err != nil {
				eyes.Fatalf("Failed to load repositories: %v", err)
			}

			if err := ensureRepoOnce(force); err != nil {
				eyes.Fatalf("Failed to sync repositories: %v", err)
			}

		},
	}

	// Update command for updating installed packages
	updateCmd := &cobra.Command{
		Use:     "update",
		Short:   "Update installed packages",
		Aliases: []string{"upgrade", "up"},
		Run: func(cmd *cobra.Command, args []string) {
			requireRoot()

			if err := ApplyRoot(root); err != nil {
				eyes.Fatalf("Invalid root: %v", err)
			}
			if err := EnsureConfig(); err != nil {
				eyes.Fatalf("Failed to ensure config: %v", err)
			}

			_, err := LoadConfig()
			if err != nil {
				eyes.Fatalf("Failed to load repositories: %v", err)
			}

			if path == "" {
				path = RecipeDirPath
			}

			if err := updateAll(path); err != nil {
				eyes.Fatalf("Update failed: %v", err)
			}
		},
	}

	// Support command for displaying support information
	supportCmd := &cobra.Command{
		Use:     "support",
		Aliases: []string{"issue", "bug", "contact", "discord", "--support", "--bug"},
		Short:   "Show support information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s", SupportInformationSnippet)
		},
	}

	// Clean command for cleaning the data folders
	// containing recipes, build directories, etc
	cleanCmd := &cobra.Command{
		Use:     "clean",
		Aliases: []string{"cleanup", "clear", "c", "-c", "--clean", "--cleanup"},
		Short:   "Clean cache info.",
		Run: func(cmd *cobra.Command, args []string) {

			requireRoot() // ensure running as root

			clean()
		},
	}

	// version command for displaying the current version
	// of Blink. Not using fang for this one because its
	// better like this.
	versionCmd := &cobra.Command{
		Use:     "version",
		Aliases: []string{"v", "ver", "--version", "-v"},
		Short:   "Show Blink version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s", VersionInformationSnippet)
		},
	}

	// Disable default Cobra completion
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Command for generating shell completion scripts
	// for bash, zsh, and fish
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

	// Add flags to commands

	getCmd.Flags().BoolVarP(&force, "force", "f", false, "Force re-download")
	getCmd.Flags().StringVarP(&path, "path", "p", "", "Specify recipes directory")
	getCmd.Flags().StringVarP(&root, "root", "r", DefaultRoot, "Specify root directory")
	infoCmd.Flags().BoolVarP(&force, "force", "f", false, "Force re-download")
	infoCmd.Flags().StringVarP(&path, "path", "p", "", "Specify recipes directory")
	infoCmd.Flags().StringVarP(&root, "root", "r", DefaultRoot, "Specify root directory")
	installCmd.Flags().BoolVarP(&force, "force", "f", false, "Force reinstall")
	installCmd.Flags().StringVarP(&path, "path", "p", "", "Specify recipes directory")
	installCmd.Flags().StringVarP(&root, "root", "r", DefaultRoot, "Specify root directory")
	uninstallCmd.Flags().BoolVarP(&force, "force", "f", false, "Force uninstall")
	uninstallCmd.Flags().StringVarP(&path, "path", "p", "", "Specify recipes directory")
	uninstallCmd.Flags().StringVarP(&root, "root", "r", DefaultRoot, "Specify root directory")
	syncCmd.Flags().BoolVarP(&force, "force", "f", false, "Force re-sync")
	syncCmd.Flags().StringVarP(&root, "root", "r", DefaultRoot, "Specify root directory")
	updateCmd.Flags().StringVarP(&path, "path", "p", "", "Specify recipes directory")
	updateCmd.Flags().StringVarP(&root, "root", "r", DefaultRoot, "Specify root directory")
	cleanCmd.Flags().StringVarP(&root, "root", "r", DefaultRoot, "Specify root directory")

	// Add commands to cobra cli root command
	rootCmd.AddCommand(getCmd, infoCmd, installCmd, supportCmd, versionCmd, cleanCmd, completionCmd, syncCmd, uninstallCmd, updateCmd)

	// Print welcome message
	fmt.Printf("Blink Package Manager Version: %s\n", CurrentBlinkVersion)
	fmt.Printf("© Copyright 2025-%d Aperture OS. All rights reserved.\n", CurrentYear)

	// Execute root command
	if err := fang.Execute(context.Background(), rootCmd, fang.WithoutVersion(), fang.WithColorSchemeFunc(colorScheme)); err != nil {
		eyes.Fatalf("Command Line Interface failed to run. (Is there any syntax error(s)?)\nERR: %v ", err)
	}
}

// if ur reading this pls contribute to the package repository :sob:
