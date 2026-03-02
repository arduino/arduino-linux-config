# Arduino Linux Config CLI

`arduino-linux-config` is a command line tool designed for Arduino Linux-based boards. It provides a structured interface to configure, enable, and manage hardware devices and overlays directly from the terminal.

## Quickstart

The project bridges the gap between high-level user requirements and low-level kernel infrastructure. Instead of manually manipulating Device Tree files, users can manage hardware connectivity through a standardized CLI.

# 🛠 CLI Interface

The initial skeleton implements the following core commands and global flags (currently provided as functional stubs):

## Commands

- **`list`**: View all available device configurations.
- **`show`**: Display detailed information for a specific device.
- **`enable`**: Activate a hardware device or overlay.
- **`disable`**: Deactivate a hardware device or overlay.
- **`reset`**: Revert hardware configurations to factory defaults.

## Global Flags

- **`--help`**: Show usage and command syntax.
- **`--verbose`**: Enable detailed process logging.
- **`--debug`**: Display internal execution data for troubleshooting.
- **`--json`**: Output results in JSON format.

## How to contribute

Contributions are welcome!

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
