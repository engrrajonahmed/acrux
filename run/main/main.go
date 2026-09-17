package main

import (
	"fmt"
	"os"

	"acrux/internal/config"
	"acrux/internal/manage"
	"acrux/run/exchange"

	"github.com/spf13/cobra"
)

const (
	appName    = "acrux"
	appVersion = "0.1.0"
)

func main() {
	if err := newRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "acrux:", err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           appName,
		Short:         "Interactive local and cloud file manager",
		Version:       appVersion,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runApplication()
		},
	}

	cmd.AddCommand(newVersionCommand())

	return cmd
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show the acrux version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), appVersion)
		},
	}
}

func runApplication() error {
	if err := initializeApplication(); err != nil {
		return err
	}

	for {
		selection, err := exchange.MainMenu()
		if err != nil {
			return err
		}

		switch selection {
		case "Browse":
			if err := runBrowse(); err != nil {
				return err
			}

		case "Copy":
			if err := runCopy(); err != nil {
				return err
			}

		case "Delete":
			if err := runDelete(); err != nil {
				return err
			}

		case "Manage":
			shouldExit, err := runManage()
			if err != nil {
				return err
			}

			if shouldExit {
				return nil
			}

		case "Exit", "":
			return nil

		default:
			return fmt.Errorf("unknown main menu selection: %s", selection)
		}
	}
}

func initializeApplication() error {
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}

	runningInstalled, err := manage.IsRunningFromInstalledLocation(executablePath)
	if err != nil {
		return fmt.Errorf("check installation location: %w", err)
	}

	if runningInstalled {
		if _, err := config.LoadPreference(); err != nil {
			if initErr := config.InitializePreference(); initErr != nil {
				return fmt.Errorf("initialize preferences: %w", initErr)
			}
		}

		if _, err := config.LoadAccounts(); err != nil {
			if initErr := config.InitializeAccounts(); initErr != nil {
				return fmt.Errorf("initialize accounts: %w", initErr)
			}
		}

		return nil
	}

	installed, err := manage.IsInstalled()
	if err != nil {
		return fmt.Errorf("check acrux installation: %w", err)
	}

	if installed {
		return nil
	}

	return runInstallation(executablePath)
}

func runInstallation(sourceExecutable string) error {
	if sourceExecutable == "" {
		return fmt.Errorf("source executable path is empty")
	}

	fmt.Println("acrux is not installed for the current user.")
	fmt.Println("Installation is required to continue.")
	fmt.Println()

	result, err := manage.Install(sourceExecutable)
	if err != nil {
		return fmt.Errorf("install acrux: %w", err)
	}

	if result == nil {
		return fmt.Errorf("installation returned no result")
	}

	if result.AlreadyInstalled {
		fmt.Println("acrux is already installed.")
		return nil
	}

	fmt.Println()
	fmt.Println("acrux installation completed.")
	fmt.Println("Run 'acrux' to start the application.")

	return nil
}

func runBrowse() error {
	_, err := exchange.BrowseMenu()
	return err
}

func runCopy() error {
	_, err := exchange.FileMenu("copy")
	return err
}

func runDelete() error {
	_, err := exchange.FileMenu("delete")
	return err
}

func runManage() (bool, error) {
	for {
		selection, err := exchange.ManageMenu()
		if err != nil {
			return false, err
		}

		switch selection {
		case "Accounts":
			if err := runAccounts(); err != nil {
				return false, err
			}

		case "Settings":
			if err := runSettings(); err != nil {
				return false, err
			}

		case "Reset":
			if err := exchange.ResetConfiguration(); err != nil {
				return false, err
			}

		case "Uninstall":
			if err := manage.Uninstall(); err != nil {
				return false, err
			}
			return true, nil

		case "Back", "":
			return false, nil

		default:
			return false, fmt.Errorf("unknown manage menu selection: %s", selection)
		}
	}
}

func runAccounts() error {
	accounts, err := config.LoadAccounts()
	if err != nil {
		return err
	}

	if len(accounts) == 0 {
		fmt.Println("No accounts are configured.")
		return nil
	}

	fmt.Println("Configured accounts:")
	for _, account := range accounts {
		fmt.Printf("  %s\n", account.Label)
	}

	return nil
}

func runSettings() error {
	preference, err := config.LoadPreference()
	if err != nil {
		return err
	}

	fmt.Println("Current settings:")
	fmt.Printf("  Starting local directory: %s\n", preference.StartingLocalDirectory)
	fmt.Printf("  Maximum archive size: %d bytes\n", preference.MaxArchiveSize)

	if preference.EncryptionKey == "" {
		fmt.Println("  Encryption key: not configured")
	} else {
		fmt.Println("  Encryption key: configured")
	}

	return nil
}