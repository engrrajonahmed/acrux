# Acrux

1. Project Overview

- Build "acrux" as a Go-based interactive CLI application for managing local and cloud files.
- Use Cobra for application/command structure and Bubble Tea for the interactive terminal interface.
- Use librclone for all rclone-based cloud operations.
- The application must support:
  - Local file and directory browsing.
  - Cloud file and directory browsing through configured rclone remotes.
  - Local-to-local, local-to-cloud, cloud-to-cloud, and cloud-to-local copying.
  - Regular and secure copying.
  - Local and cloud deletion.
  - Account and application settings management.
  - Clean installation and uninstallation.
- Prioritize reliable, functional behavior over unnecessary abstraction or architectural complexity.
- Preserve the existing project structure and organize implementation responsibilities around the existing internal packages and run/exchange entry points.

2. Application Architecture

- Keep the existing separation between:
  - internal/browse
  - internal/chunk
  - internal/cipher
  - internal/codec
  - internal/config
  - internal/file
  - internal/hash
  - internal/manage
  - internal/package
  - internal/status
  - run/exchange
  - run/main
- Use internal/config as the dedicated configuration layer:
  - account.go handles account/rclone remote configuration.
  - preference.go handles application preferences.
- Use Cobra to establish the acrux application entry point and command structure.
- Use Bubble Tea for menus, selections, confirmations, browsing interfaces, and interactive workflows.
- Use librclone as the cloud integration mechanism.
- Keep secure-copy processing modular by using the existing package responsibilities:
  - package for archiving/exporting.
  - codec for compression/decompression.
  - cipher for encryption/decryption.
  - chunk for splitting/joining oversized archives.
  - hash for integrity-related calculations/comparisons.
- Keep local file operations within the appropriate file/browse responsibilities.
- Keep installation and uninstallation within internal/manage.
- Keep configuration data persisted under the user's configuration directory.
- Keep the architecture simple and avoid unnecessary services, abstractions, or additional layers.

3. Installation and Uninstallation

- On startup, determine whether acrux is running from the current user's ~/.local/bin location.
- If acrux is not installed there:
  - Inform the user that acrux can be installed for the current user.
  - Ask for confirmation.
  - If accepted, copy the acrux executable itself into ~/.local/bin.
  - Create ~/.config/.acrux.
  - Create:
    - ~/.config/.acrux/account.conf
    - ~/.config/.acrux/preference.conf
  - Initialize both configuration files appropriately through the configuration layer.
  - Display instructions showing how to run acrux using the acrux command.
- If the user declines installation, do not silently install anything.
- Installation must safely handle an already-existing installation.
- Provide an Uninstall item under Manage.
- Uninstallation must:
  - Exit/terminate the active application workflow cleanly.
  - Remove the installed acrux executable from ~/.local/bin.
  - Remove the acrux configuration directory and its configuration files.
  - Leave the system in a clean state.
- Uninstallation must require confirmation before destructive actions.

4. Configuration and Quick Setup

- Use ~/.config/.acrux as the application's configuration directory.
- Maintain two primary configuration files:
  - account.conf
  - preference.conf
- Use internal/config/account.go for loading, saving, and managing account configuration.
- Use internal/config/preference.go for loading, saving, and managing application preferences.
- preference.conf must contain application settings such as:
  - Starting local directory.
  - Encryption/decryption key.
  - Maximum archive size.
  - Other settings that may be introduced later.
- Default starting directory:
  - Current user's home directory.
- Default maximum archive size:
  - 1 GB.
- When acrux is already installed, initiate Quick Setup as required.
- Quick Setup must allow the user to configure the values managed by preference.go.
- After Quick Setup:
  - If account.conf is empty, ask whether the user wants to configure accounts/rclone remotes.
  - If account.conf already contains accounts, proceed to the account-aware main menu.
- Configuration changes must be performed through the application's interactive interface.
- Reset must be able to clear both configuration files.
- Any operation that changes or clears configuration must require user confirmation.

5. Account Management

- Use internal/config/account.go as the configuration layer for rclone accounts/remotes.
- Store account configuration in account.conf.
- An account should support the information required by the configured rclone remote, including:
  - Account/remote label.
  - Client ID.
  - Secret ID.
  - Root folder ID.
  - Other rclone-related values required by the application.
- Provide an interactive Accounts management interface under Manage.
- Allow the user to:
  - View configured accounts.
  - Add/configure accounts.
  - Edit account information.
  - Update account labels and rclone-related configuration.
- Persist account changes through the configuration layer.
- Make configured accounts available to librclone for cloud browsing, copying, and deletion.
- Avoid unnecessarily exposing sensitive account credentials in the interface.
- Require confirmation for account changes where appropriate.

6. Main Menu and Navigation

- After installation, Quick Setup, and account setup as applicable, present the primary acrux menu:
  1. Browse
  2. Copy
  3. Delete
  4. Manage
  5. Exit
- Manage must contain:
  1. Accounts
  2. Settings
  3. Reset
  4. Uninstall
- Use Bubble Tea for interactive menus, navigation, selections, and confirmations.
- Provide a clear way to return to the previous menu where appropriate.
- Exit must terminate the application cleanly.
- Destructive or configuration-changing operations must require explicit confirmation.
- Keep the major workflows separate:
  - Browse is for viewing.
  - Copy is for selecting and transferring.
  - Delete is for selecting and removing.
  - Manage is for configuration and application lifecycle operations.

7. Browse Workflow

- Selecting Browse presents:
  1. Local
  2. Cloud

- Local browsing:
  - Start at the directory configured through internal/config/preference.go and preference.conf.
  - Use the user's home directory as the default starting directory.
  - Allow navigation through directories and files.
  
- Cloud browsing:
  - Present the configured accounts from account.conf through the account configuration layer.
  - After selecting an account, use librclone to browse its directories and files.

- Display available metadata for directories and files where the provider supplies it:
  - Directory size.
  - File size.
  - Created date.
  - Modified date.

- Treat provider-specific metadata as optional because a cloud provider may not supply every field.

- Browse is strictly read-only:
  - Do not provide selection for subsequent operations.
  - Do not provide copy or delete actions from the browser.
  - Do not modify local or cloud content.

- Keep local and cloud browsing behind the existing internal/browse package, with the Bubble Tea interface in run/exchange/browse.go.

8. Copy Workflow

- Selecting Copy presents four transfer modes:
  1. Local to Local
  2. Local to Cloud
  3. Cloud to Cloud
  4. Cloud to Local

- Based on the selected transfer mode:
  - Present the appropriate source location.
  - Allow the user to browse directories and files.
  - Allow selection of one or multiple files and/or directories.
  - Allow the user to select the destination.

- After source and destination selection, present:
  1. Regular Copy
  2. Secure Copy

- Regular Copy:
  - Copy the selected files/directories directly to the selected destination.
  - Use the appropriate local filesystem or librclone operation according to the source and destination.

- Secure Copy:
  - Archive the selected files/directories.
  - Compress the archive.
  - Encrypt the compressed archive using the configured encryption/decryption key.
  - Compare the resulting archive size against the maximum archive size configured in preference.conf.
  - Split the archive when it exceeds the configured maximum size.
  - Copy the resulting archive/chunks to the selected destination.

- The default maximum archive size is 1 GB.
- Read the configured maximum archive size through internal/config/preference.go.
- Use the existing internal packages for the Secure Copy pipeline:
  - internal/package for archiving/exporting.
  - internal/codec for compression.
  - internal/cipher for encryption.
  - internal/chunk for splitting/joining.
  - internal/hash for integrity-related operations where required.
  - internal/file for local file operations.
  - internal/browse for source/destination browsing.
  - librclone for cloud operations.
- Require confirmation before starting the copy operation.
- Provide appropriate status/progress feedback during potentially long-running operations.

9. Delete Workflow

- Selecting Delete presents:
  1. Local
  2. Cloud

- Local:
  - Browse the configured local location.
  - Select one or multiple files and/or directories.

- Cloud:
  - Select an account configured in account.conf.
  - Browse its directories and files through librclone.
  - Select one or multiple files and/or directories.

- Request explicit confirmation after selection and before deletion.
- Delete the selected items from the chosen location.
- Use internal/file for local deletion.
- Use librclone for cloud deletion.
- Keep deletion separate from Browse so Browse remains read-only.
- Provide operation status and clearly report failures.

10. Internal Package Responsibilities

- Preserve the existing package structure and keep responsibilities aligned with the current tree.

- internal/browse
  - Provide local and cloud browsing functionality.
  - Enumerate directories/files and expose available metadata.
  - Support navigation required by the interactive workflows.

- internal/chunk
  - Split oversized Secure Copy archives.
  - Join chunks when required by the corresponding workflow.

- internal/cipher
  - Encrypt and decrypt Secure Copy data using the configured key.

- internal/codec
  - Compress data for Secure Copy.
  - Decompress data when required.

- internal/config
  - Provide the application's configuration persistence layer.
  - account.go:
    - Load, save, and manage accounts/rclone remote configuration.
  - preference.go:
    - Load, save, and manage application preferences.

- internal/file
  - Perform local file operations such as copying and deleting.

- internal/hash
  - Calculate and compare hashes where required for integrity verification.

- internal/manage
  - Handle application installation and uninstallation.

- internal/package
  - Create archives from selected files/directories.
  - Handle archive/export operations for Secure Copy.

- internal/status
  - Provide local and cloud status information required by application workflows.

- run/exchange
  - Contain the Bubble Tea interactive workflows.
  - Coordinate user selections with the internal operational packages.
  - Keep the existing workflow-specific files for browse, copy, delete, manage, package, codec, cipher, chunk, and hash interactions.

- run/main
  - Serve as the executable entry point.
  - Initialize the Cobra application and launch acrux.
  
11. Error Handling and Confirmation

- Require explicit user confirmation before all destructive or modifying operations, including:
  - Installation.
  - Copy.
  - Delete.
  - Account changes.
  - Settings changes.
  - Reset.
  - Uninstall.
- Validate required configuration before starting an operation.
- Handle missing or invalid preference.conf values without crashing the application.
- Handle missing, malformed, or unusable account.conf entries gracefully.
- Handle unavailable local files/directories and inaccessible cloud remotes with clear error messages.
- Handle librclone and cloud-provider errors without reporting the operation as successful.
- Handle failures during individual Secure Copy stages:
  - Archive.
  - Compression.
  - Encryption.
  - Chunking.
  - Destination copy.
- Avoid leaving incomplete or misleading output when an operation fails.
- Provide clear status/progress information for operations that may take significant time.
- Ensure Exit and Uninstall terminate the application cleanly.
- Keep error handling practical and focused on reliable user-facing behavior rather than introducing an elaborate error framework.

12. Implementation Sequence

- Establish the Cobra application entry point and Bubble Tea application flow.
- Implement configuration handling in internal/config:
  - Account loading/saving.
  - Preference loading/saving.
  - Default values.
- Implement installation detection and installation workflow.
- Implement creation of the .acrux configuration directory and configuration files.
- Implement Quick Setup and settings management.
- Implement rclone account configuration and account management.
- Implement the main menu and Manage submenu.
- Implement local and cloud browsing.
- Implement source selection and destination selection for Copy.
- Implement Regular Copy across all four transfer directions.
- Implement the Secure Copy pipeline using the existing archive, compression, encryption, chunking, and hash packages.
- Implement local and cloud Delete.
- Add confirmations, validation, error handling, and operation status throughout the workflows.
- Implement Reset and clean Uninstall.
- Test individual workflows, followed by complete end-to-end workflows.

13. Completion Criteria

- acrux builds and launches successfully as a Go CLI application.
- Cobra provides the application/command structure and Bubble Tea provides the interactive interface.
- A non-installed launch can install acrux for the current user.
- Installation creates:
  - ~/.local/bin/acrux
  - ~/.config/.acrux/account.conf
  - ~/.config/.acrux/preference.conf
- internal/config correctly manages accounts and application preferences.
- Quick Setup correctly initializes and updates settings.
- Accounts can be configured and used as rclone remotes.
- Local and cloud browsing work without modifying content.
- Copy works for:
  - Local to Local.
  - Local to Cloud.
  - Cloud to Cloud.
  - Cloud to Local.
- Regular Copy transfers selected content directly.
- Secure Copy correctly archives, compresses, encrypts, splits oversized archives, and copies the resulting data.
- The Secure Copy maximum archive size is configurable, with 1 GB as the default.
- Delete works for both local and cloud locations.
- Manage provides:
  - Accounts.
  - Settings.
  - Reset.
  - Uninstall.
- Reset clears account.conf and preference.conf after confirmation.
- Required destructive and modifying operations request confirmation.
- Uninstall removes the installed executable and .acrux configuration directory cleanly.
- Errors are reported clearly and failed operations are not falsely reported as successful.
- The implementation remains aligned with the updated project tree without unnecessary architectural expansion.
