package main

import (
	"github.com/harpoon/hpn/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hpn",
	Short: "Container image management CLI (pull, save, load, push, rmi, ls, list-images)",
	Long:  `Harpoon (hpn) - Manage container images with multiple runtimes. Supports pull, save, load, push, rmi, ls, and listing images from Helm charts.`,
	Version:      version.GetFullVersion(),
	SilenceUsage: true,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		version.PrintDetailedVersion()
	},
}

var platform string

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Config file (default $HOME/.hpn/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&runtimeName, "runtime", "", "Container runtime (docker|podman|nerdctl|skopeo)")
	rootCmd.PersistentFlags().BoolVar(&autoFallback, "auto-fallback", false, "Auto fallback to available runtime")
	rootCmd.PersistentFlags().StringVar(&platform, "platform", "", "Target platform (e.g. linux/amd64, all)")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(listImagesCmd)
}
