# omnitrail-go

## Overview

Omnitrail-go is a Go library designed to manage and track file and directory metadata, including permissions, ownership, and cryptographic hashes. It supports various plugins to handle different types of metadata and hashing algorithms.

## Features

- **File Plugin**: Computes SHA1, SHA256, and Gitoid hashes for files.
- **Directory Plugin**: Manages directory structures and computes Gitoid hashes for directories.
- **Posix Plugin**: Tracks POSIX file permissions, ownership, and size.

## Installation

To install the library, use the following command:

```sh
go get github.com/yourusername/omnitrail-go
```

## Usage

### Creating a New Trail

To create a new trail, use the `NewTrail` function:

```go
import "github.com/yourusername/omnitrail-go"

trail := omnitrail.NewTrail()
```

### Adding Files and Directories

To add files and directories to the trail, use the `Add` method:

```go
err := trail.Add("/path/to/file_or_directory")
if err != nil {
    log.Fatalf("Failed to add path: %v", err)
}
```

### Generating ADG Strings

To generate ADG strings, use the `FormatADGString` function:

```go
adgString := omnitrail.FormatADGString(trail)
fmt.Println(adgString)
```

## Testing

To run the tests, use the following command:

```sh
go test ./...
```

## License

This project is licensed under the ApacheV2 License. See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any changes.

## Acknowledgements

Special thanks to all contributors and the open-source community for their support.
