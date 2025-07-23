# Requirements Document

## Introduction

This document outlines the requirements for rewriting the Harpoon container image management tool from shell script to Go language, providing a modern `hpn` CLI binary with enhanced functionality. The new Go implementation will maintain all existing features while adding improved performance, better error handling, and enhanced user experience for cloud-native container image management in Kubernetes environments.

## Requirements

### Requirement 1

**User Story:** As a DevOps engineer, I want a native Go binary tool that provides all current shell script functionality, so that I can manage container images with better performance and reliability.

#### Acceptance Criteria

1. WHEN the user runs `hpn --version` THEN the system SHALL display the current version information
2. WHEN the user runs `hpn --help` THEN the system SHALL display comprehensive usage information with all available commands and options
3. WHEN the user executes any command THEN the system SHALL provide colored log output with timestamp information
4. WHEN the user specifies a log file THEN the system SHALL write all operations to the specified log file
5. IF the system encounters an error THEN the system SHALL provide clear error messages with suggested solutions

### Requirement 2

**User Story:** As a container platform administrator, I want to pull container images from external registries with proxy support, so that I can prepare images for offline or restricted environments.

#### Acceptance Criteria

1. WHEN the user runs `hpn pull -f <image-list>` THEN the system SHALL pull all images specified in the file
2. WHEN HTTP/HTTPS proxy environment variables are set THEN the system SHALL use the proxy for image pulling
3. WHEN the user specifies custom proxy settings THEN the system SHALL use the provided proxy configuration
4. IF an image pull fails THEN the system SHALL retry with configurable retry attempts and continue with remaining images
5. WHEN pulling is complete THEN the system SHALL display a summary of successful and failed pulls

### Requirement 3

**User Story:** As a system administrator, I want to save container images to tar files with multiple organization modes, so that I can efficiently manage and transfer images for different deployment scenarios.

#### Acceptance Criteria

1. WHEN the user runs `hpn save -f <image-list> --save-mode 1` THEN the system SHALL save tar files to the current directory
2. WHEN the user runs `hpn save -f <image-list> --save-mode 2` THEN the system SHALL save tar files to ./images/ directory
3. WHEN the user runs `hpn save -f <image-list> --save-mode 3` THEN the system SHALL save tar files organized by project in ./images/<project>/ directories
4. WHEN saving images THEN the system SHALL generate tar filenames using the format: <registry>_<project>_<image>_<tag>.tar
5. IF insufficient disk space exists THEN the system SHALL display clear error messages with space requirements

### Requirement 4

**User Story:** As a deployment engineer, I want to load container images from tar files with flexible loading modes, so that I can restore images in different environments efficiently.

#### Acceptance Criteria

1. WHEN the user runs `hpn load --load-mode 1` THEN the system SHALL load all *.tar files from the current directory
2. WHEN the user runs `hpn load --load-mode 2` THEN the system SHALL load all *.tar files from ./images/ directory
3. WHEN the user runs `hpn load --load-mode 3` THEN the system SHALL recursively load *.tar files from ./images/*/ subdirectories
4. WHEN loading images THEN the system SHALL display progress information for each loaded image
5. IF a tar file is corrupted THEN the system SHALL skip the file and continue with remaining files

### Requirement 5

**User Story:** As a registry administrator, I want to push container images to private registries with multiple tagging strategies, so that I can organize images according to different enterprise policies.

#### Acceptance Criteria

1. WHEN the user runs `hpn push -f <image-list> --push-mode 1` THEN the system SHALL push images as registry/image:tag
2. WHEN the user runs `hpn push -f <image-list> --push-mode 2 -p <project>` THEN the system SHALL push images as registry/project/image:tag
3. WHEN the user runs `hpn push -f <image-list> --push-mode 3` THEN the system SHALL push images preserving original project paths
4. WHEN pushing to private registries THEN the system SHALL support authentication via Docker config or environment variables
5. IF push authentication fails THEN the system SHALL provide clear error messages with authentication guidance

### Requirement 6

**User Story:** As a container platform user, I want the tool to automatically detect and work with different container runtimes, so that I can use it in various environments without manual configuration.

#### Acceptance Criteria

1. WHEN the system starts THEN the system SHALL automatically detect available container runtime in priority order: Docker, Podman, Nerdctl
2. WHEN using Nerdctl THEN the system SHALL automatically add --insecure-registry parameters for private registries
3. WHEN no container runtime is found THEN the system SHALL display clear error message with installation guidance
4. WHEN multiple runtimes are available THEN the system SHALL allow user to specify preferred runtime via configuration or flag
5. IF container runtime commands fail THEN the system SHALL provide runtime-specific troubleshooting suggestions

### Requirement 7

**User Story:** As a DevOps engineer, I want comprehensive configuration management, so that I can customize the tool behavior for different environments and use cases.

#### Acceptance Criteria

1. WHEN the user creates a config file THEN the system SHALL support YAML and JSON configuration formats
2. WHEN configuration is provided THEN the system SHALL support setting default registry, project, proxy settings, and operation modes
3. WHEN environment variables are set THEN the system SHALL override config file settings with environment variables
4. WHEN command line flags are provided THEN the system SHALL override both config file and environment variable settings
5. IF configuration is invalid THEN the system SHALL display validation errors with specific field information

### Requirement 8

**User Story:** As a system administrator, I want parallel processing capabilities, so that I can efficiently handle large numbers of container images.

#### Acceptance Criteria

1. WHEN processing multiple images THEN the system SHALL support configurable concurrent operations
2. WHEN the user specifies --parallel <number> THEN the system SHALL limit concurrent operations to the specified number
3. WHEN parallel processing is enabled THEN the system SHALL display progress information for all concurrent operations
4. IF system resources are limited THEN the system SHALL automatically adjust concurrency to prevent system overload
5. WHEN operations complete THEN the system SHALL provide summary statistics including timing information

### Requirement 9

**User Story:** As a security-conscious administrator, I want image verification and validation capabilities, so that I can ensure image integrity and security compliance.

#### Acceptance Criteria

1. WHEN the user enables verification THEN the system SHALL validate image checksums before and after operations
2. WHEN processing images THEN the system SHALL support image signature verification if signatures are available
3. WHEN the user specifies size limits THEN the system SHALL validate image sizes against configured limits
4. IF image validation fails THEN the system SHALL skip the image and log detailed failure information
5. WHEN verification is complete THEN the system SHALL provide a security compliance report

### Requirement 10

**User Story:** As a CI/CD pipeline developer, I want comprehensive JSON output and exit codes, so that I can integrate the tool seamlessly into automated workflows.

#### Acceptance Criteria

1. WHEN the user specifies --output json THEN the system SHALL provide structured JSON output for all operations
2. WHEN operations complete THEN the system SHALL return appropriate exit codes (0 for success, non-zero for failures)
3. WHEN the user enables --quiet mode THEN the system SHALL suppress non-essential output while maintaining error reporting
4. WHEN processing in batch mode THEN the system SHALL provide machine-readable progress and status information
5. IF integration fails THEN the system SHALL provide detailed error information suitable for automated error handling