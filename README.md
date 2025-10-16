# aer Distribution

This repository produces the public `aer` command-line binaries and GitHub Action.

`aer` (Apex Execution Runtime) lets you validate Apex code, execute tests, and
bring up a lightweight Salesforce-compatible runtime on your workstation. It is
ideal for iterating on Apex logic without deploying to a real org, keeping test
cycles fast and reproducible.

## Install

1. Browse to the **Releases** page of this repository and download the archive
   for your platform:
   - `aer_<platform>.zip` for macOS and Linux
   - `aer_windows_amd64.zip` for Windows
2. Extract the archive and move the `aer` binary somewhere on your `PATH`.
   - macOS/Linux: `chmod +x aer && mv aer /usr/local/bin`
   - Windows (PowerShell): `Move-Item .\aer.exe $env:USERPROFILE\bin`
3. (Optional) Verify the download with the published SHA256 checksums:
   - macOS/Linux: `shasum -a 256 aer_<platform>.zip`
   - Windows: `Get-FileHash .\aer_windows_amd64.zip -Algorithm SHA256`

When a new release is published you only need to replace the binary; your
existing projects continue to work with the updated CLI.

## Quick Start

1. Initialize your project directory with the Apex source you want to run or
   test (for example an `src` folder from an SFDX project).
2. Run `aer compile path/to/src` to validate the Apex classes.
3. Execute your test suite with `aer test path/to/src` (add `-f NamePattern`
   to focus on specific tests).
4. Run individual code paths with `aer run ClassName.methodName path/to/src`.
5. Start a local API-compatible server using `aer server --source path/to/src`
   when you need to exercise HTTP integrations or Metadata API flows.

`aer` reads SObject metadata and mock data alongside your source so tests behave
like they would in Salesforce. You can point multiple commands at the same
schema/database files to share a consistent dataset across runs.

## Working With Salesforce Projects

- **Metadata imports**: Use `aer schema import` to pull metadata from an SFDX
  project or a retrieved Metadata API zip, then run tests against that schema.
- **Test data**: Seed records through `aer run` scripts or by loading CSV/JSON
  files from your project before executing tests.
- **CI usage**: Include this binary in your build pipeline to fail builds on
  Apex regressions without waiting for a scratch org.

## Troubleshooting

- Install errors such as `cannot execute binary file`: confirm you downloaded
  the archive that matches your OS and CPU architecture.
- Permission denied on macOS: add execute permission with
  `chmod +x /path/to/aer`.
- Command not found: ensure the directory where you installed `aer` is listed
  in your `PATH`.
- To report issues with the CLI runtime itself, open a ticket in this
  repository; maintainers will route it to the private source project as
  needed.

