package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

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
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	if err := newRootCommand(ctx).Execute(); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}

		fmt.Fprintln(os.Stderr, "acrux:", err)
		os.Exit(1)
	}
}

func newRootCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:           appName,
		Short:         "Interactive local and cloud file manager",
		Version:       appVersion,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runApplication(ctx)
		},
	}

	cmd.AddCommand(newVersionCommand())

	return cmd
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:          "version",
		Short:        "Show the acrux version",
		SilenceUsage: true,
		SilenceErrors: true,
		Args:         cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), appVersion)
		},
	}
}

func runApplication(ctx context.Context) error {
	if err := initializeApplication(); err != nil {
		return err
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

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

		case "Exit":
			return nil

		case "":
			return fmt.Errorf("main menu returned without a selection")

		default:
			return fmt.Errorf("unknown main menu selection: %q", selection)
		}
	}
}

func initializeApplication() error {
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}

	executablePath, err = filepath.EvalSymlinks(executablePath)
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}

	runningInstalled, err := manage.IsRunningFromInstalledLocation(executablePath)
	if err != nil {
		return fmt.Errorf("check installation location: %w", err)
	}

	if runningInstalled {
		return validateConfiguration()
	}

	installed, err := manage.IsInstalled()
	if err != nil {
		return fmt.Errorf("check acrux installation: %w", err)
	}

	if installed {
		if err := validateInstalledExecutable(); err != nil {
			return err
		}

		if err := validateConfiguration(); err != nil {
			return fmt.Errorf(
				"existing acrux installation is not usable: %w",
				err,
			)
		}

		return nil
	}

	return runInstallation(executablePath)
}

func validateConfiguration() error {
	preference, err := config.LoadPreference()
	if err != nil {
		return fmt.Errorf("load preferences: %w", err)
	}

	if preference == nil {
		return fmt.Errorf("preference configuration is missing")
	}

	if preference.StartingLocalDirectory == "" {
		return fmt.Errorf("starting local directory is empty")
	}

	localInfo, err := os.Stat(preference.StartingLocalDirectory)
	if err != nil {
		return fmt.Errorf(
			"starting local directory %q is unavailable: %w",
			preference.StartingLocalDirectory,
			err,
		)
	}

	if !localInfo.IsDir() {
		return fmt.Errorf(
			"starting local directory %q is not a directory",
			preference.StartingLocalDirectory,
		)
	}

	if preference.MaxArchiveSize <= 0 {
		return fmt.Errorf("maximum archive size must be greater than zero")
	}

	accounts, err := config.LoadAccounts()
	if err != nil {
		return fmt.Errorf("load accounts: %w", err)
	}

	if accounts == nil {
		return fmt.Errorf("account configuration is missing")
	}

	for index, account := range accounts {
		if account.Label == "" {
			return fmt.Errorf(
				"account %d has an empty label",
				index+1,
			)
		}

		if account.RemoteType == "" {
			return fmt.Errorf(
				"account %q has no remote type",
				account.Label,
			)
		}

		if account.Values == nil {
			account.Values = make(map[string]string)
		}
	}

	return nil
}

func runInstallation(sourceExecutable string) error {
	if sourceExecutable == "" {
		return fmt.Errorf("source executable path is empty")
	}

	installed, err := manage.IsInstalled()
	if err != nil {
		return fmt.Errorf("check acrux installation: %w", err)
	}

	if installed {
		if err := validateInstalledExecutable(); err != nil {
			return fmt.Errorf(
				"existing acrux installation is not usable: %w",
				err,
			)
		}

		if err := validateConfiguration(); err != nil {
			return fmt.Errorf(
				"existing acrux installation has invalid configuration: %w",
				err,
			)
		}

		fmt.Println("acrux is already installed and usable.")
		return nil
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
		if err := validateInstalledExecutable(); err != nil {
			return fmt.Errorf(
				"existing acrux installation is not usable: %w",
				err,
			)
		}

		if err := validateConfiguration(); err != nil {
			return fmt.Errorf(
				"existing acrux installation has invalid configuration: %w",
				err,
			)
		}

		fmt.Println("acrux is already installed and usable.")
		return nil
	}

	if result.ExecutablePath == "" {
		return fmt.Errorf("installation returned an empty executable path")
	}

	if result.ConfigPath == "" {
		return fmt.Errorf("installation returned an empty configuration path")
	}

	if err := validateInstalledExecutable(); err != nil {
		return fmt.Errorf(
			"installation completed but executable validation failed: %w",
			err,
		)
	}

	if err := validateConfiguration(); err != nil {
		return fmt.Errorf(
			"installation completed but configuration validation failed: %w",
			err,
		)
	}

	fmt.Println()
	fmt.Println("acrux installation completed.")
	fmt.Println("Run 'acrux' to start the application.")

	return nil
}

func validateInstalledExecutable() error {
	installedPath, err := manage.InstalledExecutablePath()
	if err != nil {
		return fmt.Errorf("resolve installed executable: %w", err)
	}

	info, err := os.Stat(installedPath)
	if err != nil {
		return fmt.Errorf(
			"installed executable %q is unavailable: %w",
			installedPath,
			err,
		)
	}

	if info.IsDir() {
		return fmt.Errorf(
			"installed executable path %q is a directory",
			installedPath,
		)
	}

	if info.Mode().Perm()&0111 == 0 {
		return fmt.Errorf(
			"installed executable %q is not executable",
			installedPath,
		)
	}

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
			if err := config.ResetAccounts(); err != nil {
				return false, err
			}

			if err := config.ResetPreference(); err != nil {
				return false, err
			}

			return false, nil

		case "Uninstall":
			if err := manage.Uninstall(); err != nil {
				return false, err
			}

			return true, nil

		case "Back":
			return false, nil

		case "":
			return false, fmt.Errorf("manage menu returned without a selection")

		default:
			return false, fmt.Errorf(
				"unknown manage menu selection: %q",
				selection,
			)
		}
	}
}

func runAccounts() error {
	accounts, err := config.LoadAccounts()
	if err != nil {
		return err
	}

	if accounts == nil || len(accounts) == 0 {
		fmt.Println("No accounts are configured.")
		return nil
	}

	fmt.Println("Configured accounts:")

	for _, account := range accounts {
		if account.Label == "" {
			fmt.Println("  [invalid account: missing label]")
			continue
		}

		fmt.Printf("  %s\n", account.Label)
	}

	return nil
}

func runSettings() error {
	preference, err := config.LoadPreference()
	if err != nil {
		return err
	}

	if preference == nil {
		return fmt.Errorf("preference configuration is missing")
	}

	if preference.StartingLocalDirectory == "" {
		return fmt.Errorf("starting local directory is empty")
	}

	if preference.MaxArchiveSize <= 0 {
		return fmt.Errorf("maximum archive size must be greater than zero")
	}

	fmt.Println("Current settings:")
	fmt.Printf(
		"  Starting local directory: %s\n",
		preference.StartingLocalDirectory,
	)
	fmt.Printf(
		"  Maximum archive size: %d bytes\n",
		preference.MaxArchiveSize,
	)

	if preference.EncryptionKey == "" {
		fmt.Println("  Encryption key: not configured")
	} else {
		fmt.Println("  Encryption key: configured")
	}

	return nil
}