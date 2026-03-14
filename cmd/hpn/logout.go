package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout [REGISTRY]",
	Short: "Log out from a container registry",
	Long: `Log out from a container registry using the selected runtime.
Credentials stored by the runtime (e.g. ~/.docker/config.json) are removed.`,
	Args: cobra.MaximumNArgs(1),
	RunE: executeLogout,
}

func executeLogout(cmd *cobra.Command, args []string) error {
	var registry string
	if len(args) > 0 {
		registry = args[0]
	}
	if registry == "" {
		return usageErrorf("registry is required")
	}

	selectedRuntime, err := selectContainerRuntime()
	if err != nil {
		return fmt.Errorf("container runtime selection failed: %v", err)
	}

	fmt.Printf("Logging out from %s using %s...\n", registry, selectedRuntime.Name())

	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
	defer cancel()

	if err := selectedRuntime.Logout(ctx, registry); err != nil {
		return fmt.Errorf("logout failed: %v", err)
	}

	fmt.Printf("✅ Successfully logged out from %s\n", registry)
	return nil
}
