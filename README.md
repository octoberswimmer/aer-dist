[![OSS Apex Tests](https://github.com/octoberswimmer/aer-dist/actions/workflows/oss-tests.yml/badge.svg)](https://github.com/octoberswimmer/aer-dist/actions/workflows/oss-tests.yml)

# aer

**Run Apex and Apex unit tests locally — no org, no deploy.**

`aer` (Apex Execution Runtime) runs your Apex code and tests on your own machine.
There's no scratch org to spin up, no sandbox to deploy to, and no API calls to an
org. It loads your Apex and the metadata next to it — classes, triggers, flows,
objects, and more — so SOQL, DML, and test data behave the way they do in
Salesforce.

Because everything runs locally, your feedback loop is the speed of a local binary
instead of a deploy-and-poll cycle against an org: run a focused test class in about
the time it takes to save a file.

**Use `aer` to:**
- Run Apex unit tests locally, with code coverage, from the CLI or in CI.
- Execute anonymous Apex against your local metadata.
- Step through Apex in an interactive debugger (VS Code or IntelliJ) — set
  breakpoints, inspect variables, and walk logic line by line.

![Demo](demo.gif)

## What aer supports

`aer` aims to behave like the Salesforce Apex runtime. It runs
against an embedded database seeded from your metadata, so the parts of Apex that
normally require an org work locally:

- **SObjects & database** — SOQL (WHERE, ORDER BY, LIMIT, relationship queries),
  DML (insert/update/delete/undelete), the `Database` methods, record types,
  picklist dependencies, field sets, and `Schema`/describe information.
- **Triggers, validation rules & flows** — Apex triggers (before/after
  insert/update/delete/undelete), validation rules enforced on DML (raising
  `FIELD_CUSTOM_VALIDATION_EXCEPTION`), and record-triggered flows and Flow
  interviews all run alongside your Apex.
- **Governor limits** — SOQL, DML, CPU time, heap, callouts, and more are tracked
  and enforced, raising the same `LimitException` ("Too many SOQL queries", "Too
  many DML statements") you'd hit in an org.
- **Standard library** — Collections, String/Math/Date, JSON, Crypto, Pattern,
  HTTP callouts with `HttpCalloutMock`, `Messaging` email, DataWeave script
  execution, and many platform namespaces (System, Database, Schema, ConnectApi,
  EventBus, and more).
- **Testing framework** — `@IsTest`, `Test.startTest()`/`stopTest()`, mock
  callouts, test-data isolation, and code coverage (Cobertura or JSON).
- **Type checking** — running tests parse- and type-checks your code, catching
  errors before you deploy.

## Try it in your browser

Run `aer` without installing anything: the [VS Code Web Demo](https://www.octoberswimmer.com/aer-demo/)
launches a preconfigured editor with sample Apex source so you can execute tests
and step through code in the interactive debugger from your browser.

## Install

On macOS or Linux with [Homebrew](https://brew.sh):

```sh
brew install octoberswimmer/tap/aer
```

You can also install `aer` as a [Salesforce CLI plugin](https://www.npmjs.com/package/@octoberswimmer/aer-sf-plugin):

```sh
sf plugins install @octoberswimmer/aer-sf-plugin
```

Or install the [VS Code extension](https://marketplace.visualstudio.com/items?itemName=OctoberSwimmer.aer-dap-client),
which installs `aer` for you and integrates the interactive debugger.

Otherwise:

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
      - uses: actions/checkout@v6

      - name: Run Apex Tests
        uses: octoberswimmer/aer-dist@main
        with:
          source: force-app
```

Adjust `with.source` for your project's Apex root, and pin the `uses:` clause to the latest released tag (for example `@v1.0.0`).

Optional inputs:
- `flags` lets you append additional CLI arguments (for example `--skip SomeTest`).
- `default-namespace` mirrors the `--default-namespace` flag to run tests against as if the code is within a package's namespace.

Set a license key for production use (running more than 100 tests).

```yaml
      - name: Run Apex Tests
        uses: octoberswimmer/aer-dist@main
        with:
          source: force-app
        env:
          AER_LICENSE_KEY: ${{ secrets.AER_LICENSE_KEY }}
```

## Quick Start

1. Initialize your project directory with the Apex source you want to run or
   test (for example an `force-app` folder from an SFDX project).
2. Execute your test suite with `aer test force-app` (add `-f NamePattern`
   to focus on specific test classes).
3. Run individual code paths with `aer exec "ClassName.methodName();" --path force-app`.
4. Use the interactive debugger to step through code, inspect variables, and
   troubleshoot issues within VS Code with `aer test --debug` or `aer exec
   --debug`.

**Learn more:**
- [Getting Started Guide](https://www.octoberswimmer.com/tools/aer/getting-started/)
- [Interactive Debugging with VS Code](https://www.octoberswimmer.com/tools/aer/docs/interactive-debugging/)
- [Interactive Debugging with IntelliJ](https://www.octoberswimmer.com/tools/aer/docs/intellij-debugging/)
- [Documentation](https://www.octoberswimmer.com/tools/aer/docs/)
- [Subscribe](https://www.octoberswimmer.com/tools/aer/subscribe/)

## Troubleshooting

- Install errors such as `cannot execute binary file`: confirm you downloaded
  the archive that matches your OS and CPU architecture.
- Command not found: ensure the directory where you installed `aer` is listed
  in your `PATH`.
- To report a bug or request a feature, open an issue in this repository.
