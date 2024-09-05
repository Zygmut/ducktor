# Ducktor

Ducktor (pronounced as dʌk.tər) is a Go application designed to manage and monitor the health of various services. It supports multiple health check interfaces, including HTTP/HTTPS, TCP, and more.

## Table of Contents

[Installation](#installation)
[Usage](#usage)
[Configuration]()
[Health Check Interfaces]()
[Logging and Monitoring]()
[Example Services]()
[Contributing]()
[License]()

## Installation

1. Clone the repository:

```bash
git clone https://github.com/yourusername/ducktor.git
cd ducktor
```

2. Use the defined Nix development environment:

```bash
nix-shell
```

3. Build the application
```bash
task build
```

The binary should be in the `/bin` directory

## Usage

Run Ducktor with a configuration file:

```bash
./ducktor -config config.toml
```

## Command-line Arguments

- `config`: Path to the TOML configuration file.

## Configuration

Configuration is managed via a TOML file. [Here](./config.toml.example)'s is an example configuration:

## Health Check Interfaces

Ducktor currently supports the following health check interfaces:

- HTTP/HTTPS: Checks the response status code of an HTTP/HTTPS endpoint.

More interfaces like TCP, GRPC, systemd, ICMP, and custom scripts can be added as needed.

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
