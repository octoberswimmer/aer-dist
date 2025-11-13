[![OSS Apex Tests](https://github.com/octoberswimmer/aer-dist/actions/workflows/oss-tests.yml/badge.svg)](https://github.com/octoberswimmer/aer-dist/actions/workflows/oss-tests.yml)

# aer

`aer` (Apex Execution Runtime) lets you validate Apex code, execute tests, and
bring up a lightweight Salesforce-compatible runtime on your workstation. It is
ideal for iterating on Apex logic without deploying to a real org, keeping test
cycles fast and reproducible.

![Demo](demo.gif)

## Install

1. Browse to the **Releases** page of this repository and download the archive
   for your platform:
   - `aer_<platform>.zip` for macOS and Linux
   - `aer_windows_amd64.zip` for Windows
2. Extract the archive and move the `aer` binary somewhere on your `PATH`.
   - macOS/Linux: `mv aer /usr/local/bin`
   - Windows (PowerShell): `Move-Item .\aer.exe $env:USERPROFILE\bin`
3. (Optional) Verify the download with the published SHA256 checksums:
   - macOS/Linux: `shasum -a 256 aer_<platform>.zip`
   - Windows: `Get-FileHash .\aer_windows_amd64.zip -Algorithm SHA256`

## GitHub Action

To run `aer test` in your GitHub Actions pipeline, add a workflow like:

```yaml
name: Apex Tests

on:
  push:
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run Apex Tests
        uses: octoberswimmer/aer-dist@main
        with:
          source: sfdx
```

Adjust `with.source` for your project's Apex root, and pin the `uses:` clause to the latest released tag (for example `@v0.1.0`).

Set a license key for production use (running more than 100 tests).

```yaml
      - name: Run Apex Tests
        uses: octoberswimmer/aer-dist@main
        with:
          source: sfdx
        env:
          AER_LICENSE_KEY: ${{ secrets.AER_LICENSE_KEY }}
```


## Quick Start

1. Initialize your project directory with the Apex source you want to run or
   test (for example an `sfdx` folder from an SFDX project).
2. Execute your test suite with `aer test sfdx` (add `-f NamePattern`
   to focus on specific test classes).
3. Run individual code paths with `aer run ClassName.methodName --path sfdx`.
4. Use the interactive debugger to step through code, inspect variables, and
   troubleshoot issues within VS Code with `aer test --debug` or `aer exec
   --debug`.

`aer` reads SObject metadata alongside your source so tests behave like they
would in Salesforce.

**Learn more:**
- [Getting Started Guide](https://www.octoberswimmer.com/tools/aer/getting-started/)
- [Interactive Debugging with VS Code](https://www.octoberswimmer.com/tools/aer/docs/interactive-debugging/)
- [Interactive Debugging with IntelliJ](https://www.octoberswimmer.com/tools/aer/docs/intellij-debugging/)
- [Documentation](https://www.octoberswimmer.com/tools/aer/docs/)
- [Subscribe](https://www.octoberswimmer.com/tools/aer/subscribe/)

## Working With Salesforce Projects

- **Metadata imports**: Use `aer schema import` to pull metadata from a project
  repo, then run tests against that schema.
- **CI usage**: Add this repo's GitHub action to fail builds on
  Apex regressions without waiting for a scratch org or sandbox deploy.

## Troubleshooting

- Install errors such as `cannot execute binary file`: confirm you downloaded
  the archive that matches your OS and CPU architecture.
- Command not found: ensure the directory where you installed `aer` is listed
  in your `PATH`.
- To report issues with the CLI runtime itself, open a ticket in this
  repository; maintainers will route it to the private source project as
  needed.
