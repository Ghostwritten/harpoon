package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
	"github.com/spf13/cobra"
	containerruntime "github.com/harpoon/hpn/internal/runtime"
)

var (
	loginRegistry     string
	loginUsername     string
	loginPassword     string
	loginPasswordStdin bool
	loginInsecure     bool
)

var loginCmd = &cobra.Command{
	Use:   "login [REGISTRY]",
	Short: "Log in to a container registry",
	Long:  `Log in to a container registry using the selected runtime.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  executeLogin,
}

func init() {
	loginCmd.Flags().StringVarP(&loginUsername, "username", "u", "", "Registry username")
	loginCmd.Flags().StringVarP(&loginPassword, "password", "p", "", "Registry password (not recommended, use --password-stdin or interactive input)")
	loginCmd.Flags().BoolVar(&loginPasswordStdin, "password-stdin", false, "Read password from stdin")
	loginCmd.Flags().BoolVar(&loginInsecure, "insecure", false, "Allow connections to registries without TLS or with self-signed certificates")
}

func executeLogin(cmd *cobra.Command, args []string) error {
	// Get registry from args or flag
	if len(args) > 0 {
		loginRegistry = args[0]
	}

	if loginRegistry == "" {
		return fmt.Errorf("registry is required")
	}

	// Select container runtime
	selectedRuntime, err := selectContainerRuntime()
	if err != nil {
		return fmt.Errorf("container runtime selection failed: %v", err)
	}

	fmt.Printf("Logging in to %s using %s...\n", loginRegistry, selectedRuntime.Name())

	// Check for environment variables
	if loginUsername == "" {
		loginUsername = os.Getenv("REGISTRY_USERNAME")
	}
	if loginPassword == "" && !loginPasswordStdin {
		loginPassword = os.Getenv("REGISTRY_PASSWORD")
	}

	// Get username
	username := loginUsername
	if username == "" {
		username, err = promptUsername()
		if err != nil {
			return err
		}
	}

	// Get password
	password := loginPassword
	if loginPasswordStdin {
		// Read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			password = scanner.Text()
		}
	} else if password == "" {
		// Interactive input
		password, err = promptPassword()
		if err != nil {
			return err
		}
	} else {
		// Password provided via -p flag, show warning
		fmt.Printf("Warning: Using --password flag is insecure. " +
			"Password may appear in shell history. " +
			"Consider using --password-stdin or interactive input.\n")
	}

	// Prepare login options
	loginOptions := containerruntime.LoginOptions{
		Username:      username,
		Password:      password,
		PasswordStdin: loginPasswordStdin,
		Insecure:      loginInsecure,
		Debug:         IsDebug(),
	}

	// Execute login
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := selectedRuntime.Login(ctx, loginRegistry, loginOptions); err != nil {
		return fmt.Errorf("login failed: %v", err)
	}

	fmt.Printf("✅ Successfully logged in to %s\n", loginRegistry)
	return nil
}

// promptUsername prompts for username interactively
func promptUsername() (string, error) {
	fmt.Print("Username: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return "", fmt.Errorf("failed to read username")
	}
	return strings.TrimSpace(scanner.Text()), nil
}

// promptPassword prompts for password interactively (hidden input)
func promptPassword() (string, error) {
	fmt.Print("Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %v", err)
	}
	fmt.Println() // New line after password input
	return string(bytePassword), nil
}
