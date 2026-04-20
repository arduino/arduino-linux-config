# Arduino Linux Config CLI

`arduino-linux-config` is a command line tool for Arduino Linux-based boards. It provides a structured interface to configure and manage hardware devices and overlays directly from the terminal, without manually manipulating Device Tree files.

## Quickstart

```sh
# List available carriers and devices
arduino-linux-config carrier list

# Configure a carrier with specific devices
arduino-linux-config carrier enable media-carrier camera1=type1-2lane display=8-dsi-touch-a

# Show current and pending configuration
arduino-linux-config carrier show media-carrier

# Reset a carrier to factory defaults
arduino-linux-config carrier disable media-carrier
```

## Installation

### On the board (via ADB)

```sh
task board:install
```

This will build the Debian package and install it on the connected board via ADB.

### Build locally

```sh
task build
```

The binary will be available at `./build/arduino-linux-config`.

## Commands

- **`list`**: Lists all available carriers and devices for the current hardware.
- **`show <carrier-name>`**: Displays the current and pending (next boot) configuration for a carrier.
- **`enable <carrier-name> [device=option...]`**: Configures a carrier with the specified device options. Factory defaults are applied before the new configuration.
- **`disable <carrier-name>`**: Resets all devices on a carrier to factory defaults.

## Global Flags

- **`--format`**: Output format (`text` or `json`). Default: `text`.
- **`--log-level`**: Log verbosity (`debug`, `info`, `warn`, `error`). Default: `error`.

## Contributions are welcome!

Please read the [Contributor Guide] document, which will show you how to build the source code, run the tests, and
contribute your changes to the project.

:sparkles: Thanks to all our [contributors]! :sparkles:

## Security

If you think you found a vulnerability or other security-related bug in the Arduino CLI, please read our [security
policy] and report the bug to our Security Team 🛡️ Thank you!

e-mail contact: security@arduino.cc

## License

Arduino Linux Config CLI is licensed under the GPL-3.0 license.

You can be released from the requirements of the above license by purchasing a commercial license. Buying such a license
is mandatory if you want to modify or otherwise use the software for commercial activities involving the Arduino
software without disclosing the source code of your own applications. To purchase a commercial license, send an email to
license@arduino.cc

[user documentation]: docs/user-documentation.md
[contributor guide]: docs/CONTRIBUTING.md
