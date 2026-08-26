# Changelog

## v1.4 (in development)

### New capabilities

- **PostgreSQL storage backend.** The `exec`, `test`, and `server` commands
  accept `--db postgres://…` (or `postgresql://…`) to run against a PostgreSQL
  database instead of the embedded SQLite store, and `--db postgresqltest://`
  starts a private, disposable PostgreSQL server for the run (a local
  PostgreSQL installation is required). The schema is created with identity
  columns and PL/pgSQL triggers, the formula and SOQL function surface
  (`DATEVALUE_TZ`, `ADDMONTHS`, format and distance functions, Id prefix
  lookups) is installed at bootstrap, and text comparison uses a collation
  matching sfapex.  On the server, each pooled VM opens its own connections and
  transaction so concurrent requests run in parallel instead of serializing on
  one storage handle, and `SELECT … FOR UPDATE` takes real row locks scoped to
  the queried object's table, bounded by a lock timeout that surfaces as
  a `QueryException` with `UNABLE_TO_LOCK_ROW`, matching sfapex.
- **Server landing page redesigned as an endpoint browser.** The authenticated
  landing page is now a self-contained dark/light single-page endpoint browser. A
  sidebar groups the development tools and API surface into categories with live
  filtered counts, a search box (focusable with `/`, clearable with Escape)
  filters endpoints by path or description, each path can be copied to the
  clipboard, and a theme toggle persists across reloads. Endpoint data is built
  server-side from the live routes, including dynamic Apex REST discovery. The
  status indicator polls a cheap, auth-free endpoint and flips from a green
  'Running' dot to a red 'Stopped' dot when the server stops responding, pausing
  while the tab is hidden. The default Salesforce API version the server exposes
  is raised from 60.0 to 67.0 across the `--api-version` default, the core
  fallback, and the SOAP and Bulk v2 fallbacks.
- **`.aerrc` config files for default flags.** Default flags can be set in
  `.aerrc` files so they need not be repeated on every invocation. Files are
  consulted in increasing precedence: `$XDG_CONFIG_HOME/aer/aerrc`, `~/.aerrc`,
  then `./.aerrc`. Flags are injected after the resolved subcommand and before
  the user's own arguments, so an explicit command-line flag wins; a flag not
  defined on the invoked command is dropped, so one file can hold defaults for
  several commands. Blank lines, comments, and lines not starting with a dash are
  ignored, environment variables are expanded, and `AER_NO_RC` skips all config
  files. A new global `--help-aerrc` flag explains how the config file works.
- **`--allow-email` delivers `Messaging.sendEmail` over real SMTP.** By default
  `Messaging.sendEmail` is mocked — addresses and bodies are validated the way
  sfapex does, but no message leaves the process. A new `--allow-email` flag
  on `exec` and `server` delivers real email over SMTP, configured by
  `AER_SMTP_HOST`, `AER_SMTP_PORT`, `AER_SMTP_USER`, and `AER_SMTP_PASSWORD`
  (defaulting to a local MTA on `localhost:25` with no authentication).
  Recipients come from the To/Cc/Bcc lists, record-Id and `targetObjectId`
  recipients resolve to their Email field, and the From is resolved from the
  org-wide email address or the running user. The outcome is logged to stderr
  either way, and Apex still observes a successful `SendEmailResult` when email
  is discarded because the flag is not set.
- **VS Code extension derives source paths, namespace, and replacements from
  `sfdx-project.json`.** A new `aer.useSfdxProject` setting (on by default)
  derives source paths from `packageDirectories`, the default namespace from
  `namespace`, and applies `replacements` by staging source into a temp
  directory before running, mirroring the aer sf plugin. Explicit
  `aer.sourcePaths` and `aer.defaultNamespace` settings still take precedence.
  This is wired through the language server, test runner, watch mode, and
  debugger; staging maps failure locations back to the real files and watch mode
  mirrors edits into the staging directory.
- **VS Code extension honors `unpackagedMetadata` and `apexTestAccess` from
  `sfdx-project.json`.** When `aer.useSfdxProject` is enabled, each package
  directory's `unpackagedMetadata.path` is loaded alongside the packaged source
  and staged with it when replacements apply, so tests compile and run against
  that metadata across the language server, test runner, watch mode, and
  debugger. `apexTestAccess.permissionSets` are passed to aer as
  `--assign-perms`, combined with the `aer.assignPermissionSets` setting.
- **VS Code Test Explorer shows Apex code coverage.** A new "Run Apex Tests
  with Coverage" profile in the VS Code extension runs the selected tests with
  `aer test --coverage` and reports covered and uncovered lines to VS Code, so
  they appear in the editor gutter and the Test Coverage view. Unit and
  integration selections merge, with a line counted as covered when either run
  executed it. The coverage JSON now lists each class's covered lines in a
  `coveredLines` field alongside the uncovered lines. The extension's minimum
  VS Code version rises to 1.88.0, where the test coverage API was finalized.
- **`SIGUSR1` prints live test-run status.** Sending `SIGUSR1` to a running
  `aer test` process (`kill -USR1 <pid>`) prints a one-line status summary to
  stderr without waiting for the periodic progress output. During setup it
  reports the current phase (schema and package loading, VM pool
  initialization) and elapsed time; once execution starts it reports test
  counts.
- **Server deployments accept every metadata type aer loads from source.** The
  deploy endpoints previously handled only Apex classes, single-file custom
  objects, and `CustomField`/`CustomObject` destructive changes. Deployment
  zips are now extracted and imported with the same metadata walker used at
  startup, so objects and fields in metadata and source formats, triggers,
  flows, flexipages, workflow rules, permission sets, profiles, custom metadata
  records, queues, labels, value sets, static resources, Visualforce pages,
  email templates, reports, and dashboards can be deployed at runtime. The
  pipeline compiles Apex against the merged schema, validates `package.xml`
  members the way the Salesforce Metadata API does (missing members fail per
  component; unmodeled types become warning components), migrates storage
  tables for the deployed objects, and commits only on success.
  Deployments queue in `Pending` status, apply serially, run as the
  authenticated deploying user, and roll back on failure; `checkOnly` validates
  without applying.
- **Revenue Cloud (Agentforce Revenue Management) support.** The full Agentforce
  Revenue Management SObject suite (general ledger and accounting-period
  objects, transactions and journals, billing, payments, usage management,
  contracts and their documents, the product catalog, and quote, order,
  rate-card, and fulfillment lines) is modeled behind the `RevenueCloud`
  feature. Enabling the feature also enables the B2B Commerce objects these
  records reference. sfapex's field derivation and status validation are
  reproduced: `GeneralLedgerAccount.Name` and `AccountingPeriod.Name` recompute
  when their source fields change on update, `Order.Status` is restricted to
  `Draft` and `Activated` with `Order.StatusCode` derived from it,
  `FulfillmentOrder.StatusCategory` derives from `Status`, and an
  `AccountingPeriod.EndDate` referenced by a `LegalEntyAccountingPeriod`
  becomes read-only.
- **Server dev-tool pages share the landing page design.** The preview
  (`/dev/lwc`, `/dev/uibundle`, `/dev/visualforce`), monitoring (`/dev/events`,
  `/dev/jobs`), and inspect (`/dev/explorer`, `/dev/mocks`) pages now use the
  same dark-default, light-override design as the redesigned landing page, with
  a persistent theme toggle whose choice follows the user across pages.
- **UI Bundle GA support.** The Data SDK shim resolves
  `@salesforce/platform-sdk` (the GA rename of `@salesforce/sdk-data`) and
  exposes the GA `graphql.query` and `graphql.mutate` API while staying callable
  in the older beta form. Bundles are served at the GA application URL
  `/app/c__<bundleName>` alongside the `/dev/uibundle` preview, authenticated
  by the session cookie. Source-tree walkers now skip `node_modules` and hidden
  directories.
- **Analytics REST API, and a much broader report engine behind it.** The
  Analytics REST subset now covers `analytics/reports`,
  `analytics/reports/<id>` with `factMap`-shaped results,
  `analytics/reports/<id>/describe`, `analytics/dashboards`, and
  `analytics/dashboards/<idOrDevName>`, which returns dashboard metadata plus
  per-component report results, memoizing runs so components sharing a source
  report execute it once. Dashboard components run in the dashboard's running
  user context: a `SpecifiedUser` dashboard resolves its stored running user
  (failing when that user does not exist) and `LoggedInUser`/`MyTeamUser`
  dashboards resolve the session user, so user-scoped report filters reflect
  the right user. The report engine gained row-level formulas
  (`customDetailFormulas`), joined (MultiBlock) reports, bucket-field criteria,
  criteria on joined-child columns with boolean filter logic decomposed across
  the join, Record Type filters compared by developer name, lookup filters
  compared by the related record's name, filter literals typed from the
  filtered field's schema type rather than guessed from the value, the full
  `UserDateInterval` timeframe vocabulary, activity reports over the Task/Event
  union including their polymorphic Who and What columns, the campaign-member
  and "Activities with <Object>" report families, custom report types with a
  custom child or an Activities join, multi-hop relationship-terminating
  columns, and the legacy report column token vocabulary (`CUST_`, `FK_`,
  `ADDRESS1_`, `OWNER_ROLE`, and the rest). Fields the running user cannot read
  are dropped from a report instead of failing the query, matching Salesforce,
  and `contains`/`does not contain` filters match per comma-separated entry
  with blank values handled the way Salesforce handles them.
- **Per-test Apex code coverage.** Salesforce records which test method covered
  which lines of each class; aer now reports the same. `aer test --json
  --coverage-per-test` emits a `perClassCoverage` field listing the covered and
  uncovered lines of every class a test touched, and the server serves
  `ApexCodeCoverage`, `ApexCodeCoverageAggregate`, and `ApexOrgWideCoverage`,
  populated for asynchronous runs and for synchronous `runTests` unless the
  request sets `skipCodeCoverage`.
- **`--locale` sets the org's default locale.** The `exec`, `test`, and `server`
  commands accept `--locale`, mirroring `--timezone`. It seeds the default
  user's `LocaleSidKey` and becomes the fallback everywhere `en_US` was
  hardcoded: `UserInfo.getLocale()`, `Integer.format()`, `runAs` users, the
  `Organization.DefaultLocaleSidKey` record, and flow interviews.
- **`--currency` sets the org's default currency.** The `exec`, `test`, and
  `server` commands accept `--currency`, validated against the ISO 4217 codes.
  `UserInfo.getDefaultCurrency()` resolves the running user's
  `CurrencyIsoCode`, then the corporate `CurrencyType`, then the user's
  `DefaultCurrencyIsoCode`, then this value. With `MultiCurrency` enabled the
  configured currency becomes the corporate `CurrencyType` at a conversion rate
  of 1.0, and the seeded `DatedConversionRate` rows match.
- **`--assign-psg` assigns permission set groups to the default user.** `aer
  test` and `aer server` accept a repeatable `--assign-psg`, the permission set
  group counterpart of `--assign-perms`. Each value is a `DeveloperName` or
  `NamespacePrefix__DeveloperName`.
- **`--bootstrap-db` merges rather than replaces.** Copying bootstrap data no
  longer deletes the target's rows first. A bootstrap row whose Id already
  exists is ignored, rows the target holds beyond the bootstrap data are left
  alone, and re-running the copy is a no-op. The old wipe deleted the seeded
  `admin@aer.local` and Automated Process users on every restart with `--db
  --bootstrap-db`, which then failed startup with
  `INSUFFICIENT_ACCESS_ON_CROSS_REFERENCE_ENTITY`. A table with no `Id` column
  shared with the target is skipped with a warning, since re-running the copy
  would duplicate its rows.
- **A `.pkg` file given as a source path loads as an unmanaged package.**
  Passing one directly (`aer test force-app fflib.pkg`), or leaving one inside
  a source directory, no longer requires `--package`/`--package-dir`. Its
  contents load exactly like source code, modeling an unmanaged package whose
  components belong to the org: they take the path's `@ns` suffix or the
  default namespace rather than the package file's own name, and test classes
  inside the package are discovered and run like source tests. This works the
  same way for `aer test`, `aer exec`, and `aer server`.
- **The server supports more Tooling API endpoints.**
  Tooling queries return `ApexClass`, `ApexSettings`, and
  `MetadataComponentDependency`, whose rows come from the dependency graph with
  class members rolled up to their classes and entity types mapped to
  Salesforce metadata component types. `MetadataContainer`, `ApexClassMember`,
  and `ContainerAsyncRequest` implement the container save path: a deploy
  applies every member body atomically, reports parse failures as component
  messages, restores the previous code when the request fails or is check-only,
  and swaps the deployed class into the running program so the new body is what
  executes. `runTestsSynchronous` runs a test class inline, recording coverage
  for the follow-up `ApexCodeCoverage` queries.
- **Test suites run from `.testSuite-meta.xml` metadata.** Test suite metadata
  in `testSuites/*.testSuite-meta.xml` loads into queryable `ApexTestSuite` and
  `TestSuiteMembership` records, and tests can be selected by suite: `aer test
  --suite <name>` runs the classes a suite names, and the server's
  `runTestsAsynchronous` accepts `suiteNames` and `suiteids` in both the
  comma-separated and array forms, so `sf apex run test --suite-names` works
  against a running server.
- **`--seed-session-token` registers a pre-shared session token.** `aer server`
  accepts `--seed-session-token` (or `AER_SEED_SESSION_TOKEN`) to register a
  token as an authenticated session for the default admin user at startup, so a
  client that already holds it can log in without an interactive flow, e.g.  `sf org
  login access-token`.  The token must be in session-id form
  (`00D000000000000!` followed by characters matching `[A-Za-z0-9_.]`).
- **REST and Bulk writes run the full DML pipeline.** A record written through
  the REST or Bulk endpoints now fires the same triggers, flows, workflow rules,
  and validation rules an Apex insert would, and each row reports its own
  errors. Retired automation stays retired: an inactive trigger neither runs nor
  appears as an `ApexTrigger` record, and only a flow's active version runs;
  `Draft`, `Obsolete`, and `InvalidDraft` versions load as metadata and are
  never executed.
- **Record-triggered flows convert Transform elements.** A Map builds one output
  per source member, binding `Source[$EachItem]` to the member being mapped,
  both in direct references and inside formulas; `Count` and `Sum` reduce a
  collection named through an `aggregationValues` input parameter to a scalar,
  and reducing an empty collection gives zero.

### Fixes and performance

- **The async job worker only wakes when async work was enqueued.** Every
  `executeAnonymous` request signaled the async job worker, which ran three
  full `AsyncApexJob` table scans per wake even for workloads that never
  enqueue async work, keeping database connections busy under high
  throughput. The worker is now signaled only when the execution actually
  persisted a deferred async job, and each wake performs a single scan with
  the queued-status filtering done in SQL.
- **Concurrent server requests no longer share a storage transaction.** Pooled VMs
  share one storage instance on the SQLite backends, and the storage executor
  attached every operation to whichever transaction happened to be open on it.
  Code that issued an operation without owning a transaction used a concurrent
  request's transaction, so its work was committed or rolled back with that
  request and failed once the owner finished first. Under concurrent Apex REST
  and `executeAnonymous` load this produced 500s, lost writes, and
  "transaction has already been committed or rolled back" errors. There are now
  two explicit transaction boundaries: Apex execution owns its transaction per
  VM with savepoint nesting, and direct-storage units of work (REST row
  handlers, LDS record operations, bulk and Bulk v2 processors, the async job
  worker, SOAP undelete, login auditing, storage sync) run inside a single
  `RunInTransaction` helper that commits on success and rolls back on error.
- **Apex REST requests no longer leak pooled VMs.** The Apex REST handler
  checked a VM out of the pool for every request but never returned it, so each
  leaked VM permanently consumed a pool slot. Once the pool reached its maximum
  the server deadlocked and Apex REST requests, deployments, and OAuth
  token requests blocked forever; the SOAP undelete and `runTests` handlers
  leaked the same way. Those handlers now release their VM when the request
  finishes. The handler also resolved the session user while holding a VM,
  which required a second pool VM for the query; at the pool maximum, every VM
  was held by a request waiting for its second VM. The user is now resolved
  before checkout.
- **Code deploys no longer panic against in-flight execution.** Pooled VMs
  share the base VM's code registry by pointer, so a metadata deploy that
  reloaded the registry in place changed the class ASTs every pooled VM was
  executing, and a request VM running during the swap panicked when it
  evaluated a newly deployed class absent from its resolved binding. The
  storage phase of a deploy already quiesced request work, but the code-swap
  phase did not; it now holds the request-VM gate across applying and reverting
  Apex artifacts and destructive changes, so new requests queue and load the
  new program on their next acquisition. A reloaded VM's program version is
  also recorded in the same critical section as the assets it loaded, so
  a deploy landing mid-load cannot leave a VM marked current while holding the
  older program.
- **Code-only deploys no longer force a full schema reload.** Every deployment
  that introduced changes rebuilt the base VM and bumped the schema version,
  forcing every pooled VM into an expensive schema reload on its next
  acquisition. Deploying an Apex class was misclassified as a schema change and
  repeated deploys stalled every subsequent request behind a schema reload. The
  base-VM rebuild and version bump are now gated on an actual schema change.
- **Warm runs restore a cached workspace image.** The parsed, canonicalized,
  and type-checked program is snapshotted into a workspace image keyed by
  source identity, so a rerun over unchanged sources skips parsing,
  canonicalization, semantic analysis, flow and external-service class
  generation, AST line info, and type-check validation. The symbol graph and
  binding result are stored alongside it, so a warm run also skips symbol
  resolution.
- **Package merges no longer re-migrate the whole database.** Merging each
  package schema into a VM ran a full storage migration over the entire merged
  schema, re-inspecting every table's columns and re-applying every generated
  index and trigger. Migration is now scoped to the tables a merge actually
  touches.
- **Persistent databases restart without re-migrating and reseeding.** A server
  started against a `--db` database re-migrated every table and reseeded all
  metadata on every restart, even when nothing had changed. A startup
  fingerprint over the schema, package schemas, and startup options is now
  recorded after a full startup; a later start whose fingerprint matches
  attaches storage without migration, skips metadata seeding, and reconstructs
  its in-memory state from the existing rows. Any input change produces a
  different fingerprint and falls back to the full path.
- **A persistent database survives an aer upgrade.** The synthetic
  `admin@aer.local` user is recreated at startup and its id advances whenever
  metadata is loaded, so after an upgrade or a schema-cache version change the
  id aer computed no longer matched the one the database held, and startup
  aborted while loading standard metadata so upgrading aer required rebuilding
  the database. A persisted default user or profile id that resolves to no
  record is now dropped and re-resolved against the actual database.
  Separately, a `Profile` copied in by `--bootstrap-db` whose `CreatedById`
  references a user that was not copied no longer aborts startup with a foreign
  key error.
- **Fields named after reserved SQL keywords are stored and queried
  consistently.** A field such as `IdpEventLog.Timestamp` is stored in an
  underscore-prefixed column, but the storage layer and the SOQL-to-SQL
  generator kept two disagreeing lists of reserved words, so the column was
  written as `_Timestamp` and queried as `Timestamp`. Both now share one
  reserved set and one column-to-field mapping.
- **Server REST writes require an authenticated session and apply the profile's
  default record type.** REST sObject create, update, and delete now resolve
  the request's session token to a real user and run in that user's context,
  rejecting the request with 401 when no valid user can be resolved; they
  previously ran with no user context, so records were created with no default
  record type and profile-driven access checks resolved against nobody. Create
  assigns the profile's default record type when one is not supplied, matching
  Salesforce. The LWC and UI bundle preview routes reject stale and expired
  sessions consistently rather than serving a page whose data operations would
  then fail. Session tokens issued against a persistent `--db` database are
  themselves persisted, so a login survives a server restart.
- **Custom rich text fields describe as rich text.** The source importer mapped
  custom `Html` fields to plain `String` fields; they are now imported as
  `textarea` with `htmlFormatted` set, which also drives
  `extraTypeInfo = richtextarea`, matching how Salesforce describes them.
- **`Owner.Type` no longer reports null for user-owned records.** The SQL
  generator's multi-target lookup branch shadowed the `Owner.Type`
  User/Queue discriminator, joining `Group`'s own `Type` column and yielding
  null wherever the owner was a user. `Owner.Type` now resolves through the id
  prefix.
- **Two LWC compiler fixes.** A class field whose initializer contained a
  template literal with a `${...}` interpolation had its removal range run past
  the end of the enclosing class, deleting later methods and producing invalid
  JavaScript. A base class exported by name rather than as the default export
  was treated as a utility module, so its `@track`/`@api`/`@wire` decorators
  were never stripped and every component extending it broke. The
  `getObjectInfo` wire adapter now also returns `recordTypeInfos` and the
  default record type id, so component code iterating them no longer throws.
- **The LWC preview no longer loads its stylesheet from the network.** Every
  preview fetched the Salesforce Lightning Design System stylesheet from
  unpkg.com, a live network dependency; the same version is now vendored and
  served by the dev server.
- **Repeated class deploys keep dependent classes working.** A class deploy or
  watch reload re-resolves only the changed classes and their dependents, but
  the re-resolved classes recorded no dependency edges of their own, so after
  the first deploy aer no longer knew what depended on the changed class. A
  second deploy then left the dependents bound to deleted symbols, and a
  chained access through a static property whose name shadows an inner
  interface threw a `NullPointerException` at runtime.
- **An Apex class keeps one Id across the Tooling API, the Data API, and Apex.**
  A program load now updates the existing `ApexClass` rows instead of deleting
  and reinserting them, and the server takes each class's id from its row rather
  than inserting a second row to obtain one. Id assignment is also serialized,
  since concurrent callers each assigned a different id to one class and
  registered it under every one of them, repeating the class in later query
  results.
- **A Tooling query returns the fields it selected and nothing else.** Selecting
  `Id` from `ApexClass` answered with every class body, including the subfields
  of a relationship. A record's `attributes.url` carries its id even when the
  query does not select `Id`, and a row with no id of its own, such as an
  aggregate, carries no url.
- **Session tokens are accepted in the classic `OAuth <sessionId>` scheme.** The
  CometD client behind `sf apex get test` sends that scheme on every streaming
  request and was rejected with `invalid_session` even though its REST calls
  authenticated; `Bearer` is still accepted. Request paths with repeated slashes
  are also collapsed before dispatch instead of answering with a 307 redirect,
  which clients joining an instance URL that ends in a slash to an absolute path
  could not follow. Encoded slashes (`%2F`) are preserved.
- **The seeded Organization is named "aer Organization".** It was "Test
  Organization", which `UserInfo.getOrganizationName`, `$Organization.Name`, a
  query against the `Organization` singleton, and
  `ConnectApi.Organization.getSettings` all reported. The server also sets a
  `Server` response header of `aer/<version>` on every response, including 404s
  and unauthenticated routes.
- **A record-triggered flow saves once per element, not once per record.** The
  Flow runtime advances every interview of a trigger batch element by element,
  so a record DML element saves the records built by all of them in one DML. aer
  ran each interview to completion instead, so a Create Records element inserted
  once per triggering record, firing the created object's triggers again for
  each one. Any flow that saves records now runs as a generated handler class
  whose interviews suspend at a DML element and resume after the save, so one
  DML covers the batch, with save batches deduplicated by Id. An update of the
  triggering record still folds into the save already in progress as one
  batched follow-up update per trigger batch.
- **A flow formula converts the functions and references it names, or the
  conversion fails.** `REGEX(text, pattern)` converts to `Pattern.matches`,
  matching the text end to end, with an unset text matched as the empty string
  and an unset pattern never matching. A function aer does not model no longer
  falls through to a bare identifier or emits the formula's own source text as
  an Apex literal.  A formula's expression is also converted to the data type
  the formula declares, so a formula declared `Date` whose expression adds
  a day to a datetime field no longer feeds a `Datetime` to every `Date`
  target.
- **A `{!$Name}` merge field naming no global variable renders as literal
  text.** Such a reference resolves only against Flow's global variables, so a
  text template holding `{!$link_formula}` renders those characters while
  `{!link_formula}` renders the formula's value. aer emitted the reference as an
  invalid identifier, and the generated class failed to compile with
  "Variable does not exist: $Name". A text template's value is also built as a
  real expression rather than a literal holding generated Apex source, which had
  assigned the source text to the field the template fed.
- **A flow value assigned to a text field is rendered in the running user's
  locale.** A Boolean renders as `true` or `false`, a Number and a Datetime
  through their locale-aware `format()`, and a Date in the locale's long date
  form, for every locale and language a user can be assigned. A formula that
  offsets a Date by a Number converts to `addDays` over the whole part, which
  is what Flow does with the fraction, and assignments to formula and roll-up
  fields are dropped, since Flow's runtime ignores them while Apex refuses to
  compile one. A flow loaded into a namespace now resolves field types
  namespace-first, which had left a namespaced flow with no field types and
  therefore no conversions at all.
- **A flow variable's default is applied when the interview starts.** A declared
  default became a field initializer on the generated handler class, so a
  default reading the triggering record ran before the handler was given one and
  the flow failed with a null dereference. Defaults are now assigned at the top
  of the interview, so each interview of a trigger batch computes its own from
  its own record.
- **A flow polymorphic qualifier converts and types the field behind it.** A
  formula reference like `Owner:User.Alias` used to reach the generated Apex
  verbatim, which is not Apex syntax; it now becomes the same guarded read an
  element reference builds, reading the field only when the related record is of
  that type and giving null otherwise. The qualifier also names the branch's
  type, so `Owner:User.CreatedDate` is a `Datetime` and converts on assignment
  to a `Date` target, while an Id reads as an Id.
- **A flow implies the optional features of the objects it names.** Flows are
  converted before the run applies optional features, so a text formula function
  over an Id-typed lookup on a feature-gated object generated Apex calling a
  `String` method on an `Id`. The objects a flow names in its metadata now imply
  their feature the way a static Apex reference does, in `aer test`, `aer exec`,
  and `aer server` alike, and a warm run's cached flows carry the objects they
  name.
- **A polymorphic relationship in SOQL selects only `Name` entity fields.** A
  field selected through `What`, `Who`, or `Owner` used to resolve against the
  relationship's first `referenceTo`, so `What.Amount` returned null for every
  task instead of being rejected with "No such column 'Amount' on entity
  'Name'". Reading a branch's own fields still takes a `TYPEOF`, whose projected
  numeric fields no longer panic on their packed decimal representation.
- **`Pattern.matches` and `Matcher.matches` anchor the pattern.** Both found the
  leftmost match and then checked whether it covered the whole input, so an
  alternation whose shorter alternative matched first reported no match: the
  regex `[a-zA-Z0-9]{15}|[a-zA-Z0-9]{18}` matched only the first fifteen
  characters of an eighteen-character Id. The pattern is now anchored the way
  Java anchors it for `matches()`, leaving capture group numbering unchanged,
  and `Matcher.matches()` matches the current region rather than the whole
  input.
- **A `static final String` initialized with a literal is a compile-time
  constant.** sfapex inlines it into its use sites, so a static initializer
  declared above such a field still reads its value and the field's own
  declaration is not an executable location for coverage. aer evaluated static
  fields in strict declaration order, so a class whose constructor read a
  constant declared below it saw null.
- **Apex comparison and equality operators share one precedence level, evaluated
  left to right.** aer's parser bound the relational operators tighter, as Java
  does, so `(a < 0) != (b < 0)` re-parsed as `((a < 0) != b) < 0`. A bare
  `instanceof` operand of a comparison is now rejected, since `instanceof` is
  non-associative with those operators, and an ordering comparison of Booleans
  is rejected with the message sfapex uses for it.
- **Approval process and workflow criteria match on a record's record type.** A
  criteria item on the `RecordType` pseudo-field named no field of the object,
  so it read as null and every process gated on a record type was skipped,
  failing submission with `NO_APPLICABLE_PROCESS`. The comparison now
  substitutes the record's record type `Name`.
- **A nested formula field is inlined at every reference, not just the first.**
  The SQL generator marked a referenced formula field visited for the rest of
  the enclosing expression, so every reference after the first was treated as a
  circular reference and emitted as a bare column. A formula field has no stored
  value, so those references evaluated to null: a formula like
  `IF(NOT(ISBLANK(a)), b, IF(c, b * 1.15, b))` over a formula field `b` produced
  null for rows taking either `ELSE` branch, `SUM()` over them returned null,
  and a `WHERE` filter on the formula matched nothing.
- **An `Iterator<T>` parameter accepts `List<T>.iterator()`.** Return type
  inference had an iterator case for `Set<T>` but none for `List<T>`, so the
  argument reached overload resolution typed as a bare `Iterator`, which matches
  no `Iterator<T>` parameter: constructors raised `IllegalArgumentException` and
  a method overload silently selected a different candidate.
- **`Task.AccountId` and `Event.AccountId` derive from a contact `WhoId`.** An
  activity linked only to a contact came back with a null `AccountId`, so rollup
  code grouping activities by account silently counted nothing. When the
  `WhatId` does not own the derivation, the `WhoId` contact's account
  supplies it, and a Lead `WhoId` supplies nothing. Updating an activity no
  longer clears a correct `AccountId` at every `WhatId` dead end, and moving a
  contact to another account takes its activities with it.
- **State and country validation applies only where the repo has the picklists.**
  A `*StateCode` / `*CountryCode` component field exists only when State and
  Country/Territory Picklists are enabled; aer carried them on every address
  compound unconditionally, so describes reported fields the repo does not have
  and every insert was validated against picklists the repo had never enabled.
  With the picklists on, a compound `*Country` / `*State` text field holds the
  picklist entry's integration value and nothing else, so a label or an ISO code
  is rejected as an invalid country or state rather than resolving and then
  failing with a mismatched code. Turning the picklists on later restores the
  component fields verbatim.
- **A new record takes the org's default country in every address compound.**
  Address Settings' default country is stamped on each address compound of a new
  record whether or not the record carries an address, so a state supplied
  without a country resolves within that default instead of raising the
  country-required error, and a state belonging to another country is rejected
  as an invalid state. A person account takes the default in its person
  addresses; a business account leaves those blank.
- **The server's test bookkeeping no longer re-migrates the database.**
- **A database whose table name differs only in case is repaired on open.**
  SQLite compares identifiers case-insensitively, so a table stored under an
  object's older spelling satisfied `CREATE TABLE IF NOT EXISTS` for the
  canonical one; migration reported success, the startup fingerprint was
  written, and the next startup matched it, skipped migration, and failed
  validation with "missing table QueueSobject".  The migrator now renames such
  a table to its canonical spelling, preserving rows, indexes, and triggers,
  before generating DDL and before validating a database that skipped
  migration.
- **Legacy SQLite feed tables are replaced with views during migration.** A
  database created by an older aer version holds entity feeds, `ContentNote`,
  and picklist masters as real tables, where the current generator emits views
  over their backing tables. On SQLite the conversion silently did nothing, and
  the following write-through trigger failed with "cannot create INSTEAD OF
  trigger on table: AccountFeed". The legacy table is now dropped before the
  view is created, matching the PostgreSQL path.
- **A failed insert names the field and value the database rejected** instead of
  reporting the bind parameter's position. Bootstrap data also merges users on
  their username as well as their Id, so a bootstrap file captured from an org
  no longer leaves two users sharing a username once the target has assigned its
  own admin a different Id.
- **The web extension's wasm binary is 20% smaller**   The API server, the
  LWC/Aura preview bundler, package authoring, and the VS Code launcher were
  all reachable from the wasm build and retained; they are excluded now.

## v1.2.37 — 2026-08-26

- **Non-reparentable master-detail fields are read-only once a record has an
  Id.** Assigning such a field on an SObject that already carries an Id, even
  a fake Id stamped on an in-memory record, now throws `System.SObjectException`
  "Field is not writeable" at assignment time, matching sfapex. Previously
  the write was silently allowed. Named field arguments in the SObject
  constructor remain permitted.
- **`JSON.deserialize` into SObjects resolves columns by exact API name and
  raises sfapex's unknown-column errors.** Scalar fields, child relationships,
  and parent `__r` keys now match only their fully qualified API names; a bare
  key in a namespaced org no longer binds to the namespaced field. Unknown
  columns on the top-level record are tolerated and dropped (their nested
  known columns still apply to the record, and a non-empty array anywhere
  inside throws "No field name specified on column for sobject of type
  <T>"); a child-relationship row rejects any unknown column with "No such
  column '<key>' on sobject of type <RowType>"; a parent-relationship value
  rejects one with "Cannot deserialize instance of ... or request may be
  missing a required field". A child-relationship value must be a
  `QueryResult` object ("QueryResult must start with '{'"), a nested
  `attributes` key that duplicates the record's own throws "Duplicate column
  specified for sobject at location: [Line: L, Column: C]", and a parent
  value deserializes as its `attributes.type` so polymorphic lookups keep
  their `Type` and `Name` fields.
- **`JSON.serialize` no longer emits dropped or internal columns.** Unknown
  columns dropped during deserialization stay invisible when the record is
  serialized again, and deep parent-path selects (`Account.Owner.Name`) no
  longer leave dotted storage columns in the output; only the nested
  relationship records are serialized.
- **`SObject.putSObject` validates the target relationship.** Passing a field
  that is not a reference field throws `System.SObjectException` — "<Type>.<Field>
  is not a relationship" for the token form, "Invalid relationship <name> for
  <Type>" for the string form, and `put` of an SObject value into a String
  field throws "Illegal assignment from <Type> to String".
- **Custom metadata records serialize with only qualified field names.** A
  packaged `__mdt` record no longer serializes each namespaced field twice
  (`ns__Active__c` and `Active__c`), so `JSON.serialize` followed by
  `JSON.deserializeStrict` into the same type round-trips from package code.

## v1.2.36 — 2026-08-18

- **A queried row returns only the branch it takes through a `TYPEOF` clause.**
  The relationship was typed by whichever `WHEN` clause a projected column
  belonged to, so a `ContentDocumentLink.LinkedEntityId` pointing at one object
  came back as an empty instance of an unrelated one, and every branch's fields
  were marked queried on every row. A row whose type matches a `WHEN` clause now
  comes back typed as that object with that clause's fields; a row matching no
  clause comes back typed `Name` when the query has an `ELSE` branch, and null
  with the lookup Id field still populated when it does not. The Id always comes
  back whether or not the branch selected it, and reading a field that a
  different branch projected throws `SObjectException`, even when that field
  exists on the row's own object. The `ELSE` branch also no longer builds a
  `COALESCE` over every object the polymorphic field can reference, which
  exceeded SQLite's argument limit and failed queries on a relationship as wide
  as `ContentDocumentLink.LinkedEntity` outright.
- **`Approval.process` accepts its `allOrNone` argument.** With
  `allOrNone=false` a failing request rolls back only its own work and returns a
  failed `ProcessResult` carrying the failure as a `Database.Error`, with null
  `entityId`, `instanceId`, and `instanceStatus` and an empty `newWorkitemIds`,
  while the remaining requests commit. With `allOrNone=true`, and with the
  one-argument form, a failure throws the row-naming `DmlException` and rolls
  back every request in the batch, including ones that had already succeeded; a
  null `allOrNone` throws `NullPointerException`. Per-request failures now carry
  a real status code, `NO_APPLICABLE_PROCESS`, `REQUIRED_FIELD_MISSING`,
  or `MANAGER_NOT_DEFINED`, instead of a formatted string, a request naming a
  nonexistent workitem fails with `INVALID_CROSS_REFERENCE_KEY` rather than
  returning a fabricated success, and a successful `ProcessResult` reports
  `getErrors()` as null rather than an empty list.
- **Approval processes evaluate their entry criteria.** Submitting a record
  without naming a process evaluates each active process's criteria in process
  order and enters the first that accepts the record; an explicitly named
  process must also accept it, and when an object has approval processes but
  none accept, the submission fails with `NO_APPLICABLE_PROCESS`. Criteria
  items, boolean filters, and criteria formulas are read from the approval
  process metadata and evaluated with the workflow criteria machinery, and are
  qualified with the package namespace when they ship in a package.
- **Approval `ActorId` and `TargetObjectId` fields are name-pointing polymorphic
  lookups.** `ProcessInstance.TargetObjectId`,
  `ProcessInstanceHistory.TargetObjectId`, and the `ActorId` and
  `OriginalActorId` fields on `ProcessInstanceStep`, `ProcessInstanceWorkitem`,
  and `ProcessInstanceHistory` now resolve their `Type` pseudo-field to the
  object-type discriminator and their `Name` pseudo-field through the `Name`
  object, so `ActorId` on a record whose owner is a queue or a user projects the
  right name. Plain multi-target lookups such as `EmailTemplate.Folder.Type`
  keep resolving real referent fields. `Name` rows carry AutoNumber names, and
  the `Name` object follows sfapex's test data isolation: rows that existed
  before a test project null names, rows inserted during the test read back, and
  the transaction rollback restores the baseline.
- **A queue and a public group can share a `DeveloperName`.**
  `Group.DeveloperName` is unique per `Type`, not globally, but the metadata
  loader upserted `Group` records by developer name alone, so a public group
  loaded after a same-named queue overwrote the queue's row.  `aer test` also
  accepts `.queue`/`.queue-meta.xml` files passed explicitly, which it
  previously ignored.
- **Dangling lookup Ids report `INVALID_CROSS_REFERENCE_KEY` on the fields
  sfapex reports it for.** Writing a well-formed but nonexistent Id into a
  reference field raised `INSUFFICIENT_ACCESS_ON_CROSS_REFERENCE_ENTITY` for
  every field except `OwnerId`, so reassigning an approval work item to a
  deleted user (`ProcessInstanceWorkitem.ActorId`) reported the wrong status
  code and message.
- **Flow formula text functions accept non-text operands.** `LEFT`, `RIGHT`,
  `MID`, `LEN`, `UPPER`, `LOWER`, `TRIM`, `LPAD`, `RPAD`, `CONTAINS`,
  `SUBSTITUTE`, and `FIND` applied to an Id-typed reference generated a `String`
  method call on the Id, which failed to compile with "Method does not exist:
  left on builtin type Id". The operand is now rendered as text when its type
  is not `String`.
- **Three constructs Salesforce rejects at compile time are now rejected.** A
  bare expression that computes a value and does nothing with it fails with
  "Expression cannot be a statement".  Previously only unary operators were
  checked, so an identifier, literal, field access, comparison, ternary, cast,
  index, or `instanceof` passed. The operand of `++` or `--` must be assignable
  wherever the operator appears, so `switch on ---p` fails with "Expression
  cannot be assigned". A bare identifier on the left of an assignment that names
  nothing in scope fails with "Variable does not exist". Switch statements are
  covered too: the subject, every `when` value, and every `when` body are now
  checked like any other expression or statement, with a `when Type var` name
  bound over its case body.
- **Loading a source directory is faster.** Each source tree was traversed
  multiple times.  Each tree is now captured once and every traversal answered
  from that capture.
- **Test runs spend less time on permissions, formulas, and name lookups.**

## v1.2.35 — 2026-08-17

- **`ADDMONTHS` works in formulas evaluated by Apex.** Only the SOQL translation
  implemented it, so `recalculateFormulas` and dynamic `Formula.builder`
  evaluation always yielded null. It now returns the shifted date with month-end
  clamping, matching the SOQL behavior, including leap-day clamping, year
  rollover, negative month counts, and null dates.
- **Triggers at API version 67.0 or later default to user context.** The default
  was taken from the current class's API version so a 67.0 trigger kept the
  system-mode default for SOQL, DML, and `Database` methods. The executing
  trigger's own API version now decides; `WITH SYSTEM_MODE` and `as system` opt
  out, and a class called from the trigger keeps its own default.
- **`USER_MODE` queries enforce field-level security on `WHERE`-clause fields,
  while `WITH SECURITY_ENFORCED` does not.** A user-mode query filtering on a
  field the running user cannot read now fails with the inaccessible-field
  `QueryException`, whose message gains the "If you are attempting to use a
  custom field" hint and echoes the field as authored, bare or
  namespace-qualified. `WITH SECURITY_ENFORCED` applies only to the `SELECT` and
  `FROM` clauses, so a query selecting accessible fields while filtering on an
  inaccessible one succeeds with the filter applied.
- **Workflow field updates bypass custom validation rules.** A field update may
  write a value a direct DML update would be rejected for, and the save of the
  re-fired update triggers' changes does not re-run the rules either. Field
  updates ran through the full update pipeline, so a rule matching `ISCHANGED` on
  the updated field failed the triggering DML.
- **Workflow rules with `triggerType` `onCreateOnly` fire.** They previously
  never executed. Such a rule now runs its actions on insert only; updates never
  evaluate it, not even an edit that moves the record into the criteria.
- **`Datetime.format` renders full month and day names for pattern runs of four
  or more letters.** Month (`M`, `L`) and day-of-week (`E`) are
  `SimpleDateFormat` text fields, where any run of four or more letters gives the
  full name; the converter topped out at four letters, so `MMMMM` produced
  "November11" and `EEEEE` produced "SaturdaySat".
- **`Datetime.format` supports the general time-zone pattern `z`.** It renders
  the short zone name (`PDT`), `GMT` for UTC, and a GMT offset for a zone with no
  name.
- **A subscriber permission set or profile can grant a custom permission shipped
  in a managed package.** Schema import failed with "no CustomPermission named
  `pkg__Advanced_Access` found" before any package schema was merged, so the
  reference could never be satisfied by `--package`, `--package-dir`, or a
  `path@ns` source dir. Such references are now resolved after the last package
  merge, and the same error with its package hint is raised only for whatever
  remains unresolved. A namespace-qualified name resolves against the owning
  namespace's record, and a bare name still never binds to a packaged permission.
- **`aer exec` handles subscriber source alongside `path@ns` sources.** It built
  one schema from every directory and then applied the single namespace to all of
  it, so subscriber objects like `Entry__c` became `pkg__Entry__c` and failed
  type checking; with two or more namespaces no namespace was applied at all and
  package objects were unresolvable. `exec` now mirrors `test`: the base schema
  comes from unnamespaced dirs only, and each namespaced path group is built with
  its own namespace and merged as a package.
- **`aer license show --key` displays a public-only license without
  `GITHUB_TOKEN`.** The command ran full validation, including GitHub org
  membership verification. It now decodes the key and checks its format,
  expiration, and signature only; `license show` for an installed license still
  runs full validation.

## v1.2.34 — 2026-08-16

- **Duplicate rule matching implements the `CompanyName`, `Street`, and `City`
  methods.** Rule items using them fell through to exact comparison, so the
  standard Account matching rule never detected the fuzzy duplicates Salesforce
  detects.
- **A rollup over similarly named parent records no longer fails with
  `DUPLICATES_DETECTED`.** The rollup engine recalculates parents in batch but
  saved them outside a duplicate-rule save scope, so the parents were compared
  against each other. Records saved in the same call are never compared against
  each other, so a bulk rollup over many similarly named parents now saves
  cleanly.
- **A successful partial-save retry keeps its triggers' side effects.** When a
  partial save (`allOrNone` false) has after triggers fail some rows, the
  successful subset is retried with triggers re-fired. The batch-update path
  then ran its failed-row restore anyway, rolling back to a savepoint taken
  just before the retry and re-saving only the row fields, so an after trigger
  that wrote to another record on behalf of a surviving row silently lost that
  write whenever any other row in the call failed. The restore now runs only
  when the final attempt still had failures.
- **`Schema.getGlobalDescribe()` resolves keys of any casing.** The returned map
  stored lowercase keys plus an original-case alias, so `get` and `containsKey`
  with a mixed-case key matching neither spelling returned null. `keySet()`
  still returns lowercase names, and a `putAll` copy into a regular map remains
  case-sensitive with lowercase keys.

## v1.2.33 — 2026-08-15

- **Updating a `List<SObject>` enforces the same sharing write check as
  updating one record.** The single-record `update` path denies the write when
  class sharing is enforced and the running user has no edit access to the
  stored record, but the list path validated CRUD/FLS, duplicate Ids, and
  deleted entities and then wrote every row, so a `with sharing` class silently
  modified records the same update would have refused one at a time. The check
  now runs per row, and the failing row index is carried in the `DmlException`
  message.
- **A unique field constraint is evaluated against the final state of the whole
  DML statement.** One record can now take over a value another record in the
  same statement releases, and two records can swap values outright, matching
  sfapex.  A keyword DML exception's `getDmlIndex()` now reports the row that
  actually failed instead of echoing its argument.
- **A custom object share's `AccessLevel` is validated as a restricted
  picklist.** An invalid value now fails with
  `INVALID_OR_NULL_FOR_RESTRICTED_PICKLIST` on `[AccessLevel]` on the
  keyword list-insert path as well, reporting the failing row index.
- **A failed all-or-none list insert leaves `Id` unset.** When a later step,
  an after-insert trigger, workflow, or after-save flow, failed and rolled
  back, the Ids were still copied onto the caller's records, so re-inserting
  the same list threw "cannot specify Id in an insert call". `Id` and
  `PersonContactId` are now cleared on rollback and the retry succeeds.
- **Duplicate rule name matching applies its fuzzy algorithms.** The
  `FirstName` matching method matches nickname variants (Will/William,
  Liz/Elizabeth), initials against full names (J/John), and misspellings within
  its 85% threshold; `LastName` matches misspellings within its 90% threshold.
  A shared phonetic encoding alone is not a match. `Datacloud.MatchRecord` now
  reports one `FieldDiff` per matching rule item with the values `Same`,
  `Different`, and `Null`, using the relationship name for a parent item like
  `Account.Name`, and `getMatchConfidence()` returns 100.0 for every match.
- **Rollup summary fields aggregate to their real values.** `SUM`, `AVG`,
  `MIN`, and `MAX` over a rollup summary field silently returned zero, because
  the aggregate path treated the field as a query-time formula and read its
  stored column raw.
- **`Decimal.toString()` renders the value's scale.** A `Decimal` with a scale
  of two printed as `34` where sfapex prints `34.00`; it now agrees with
  `String.valueOf()` and with string concatenation. `Decimal.precision()`
  likewise counts the digits of the unscaled value, so `34.00` has a precision
  of four and `0.05` a precision of one.
- **`Decimal.format()` respects the running user's locale.** It hardcoded a
  comma group separator and a period decimal point; it now uses each of the
  locales' group and decimal separators, negative prefix, digit
  script, and grouping style (thousands or the Indian lakh/crore pattern).
  `Long.format()` and the `Double` overload share the same behavior.
- **Flows read auto-stored subflow outputs.** A subflow element with
  `storeOutputAutomatically` exposes the child flow's output variables as
  `ElementName.variable`, but only explicit output assignments were handled, so
  any converted flow reading one failed at runtime with "identifier not found
  in binding". A reference on a path where the subflow never ran resolves to
  null, matching Flow semantics.
- **Flow datetime literals evaluate correctly.** A `dateTimeValue` literal was
  emitted verbatim into `Datetime.valueOf`, which does not accept the ISO 8601
  form flow metadata stores, so an entry criterion like `CreatedDate >=
  <datetime>` threw "Invalid date/time" on every record save on the object.
- **Flow constants resolve on every reference path.** A flow's `<constants>`
  were parsed but never seeded into the interview, so any reference raised
  "Unknown flow variable reference". They are now seeded before variables, so a
  variable's default value can reference one, and cloned interviews carry them.
- **Record-triggered flows no longer expose relationship data on `$Record`.**
  Flows received the caller's record image verbatim, where Apex triggers get
  theirs stripped, so when the record came from a query with a partial parent
  projection a spanning reference threw `SObjectException`, "SObject row was
  retrieved via SOQL without querying the requested field", on the second hop.

## v1.2.32 — 2026-08-14

- **Flow global variable lookups no longer consume SOQL queries.** A flow
  reference to a global variable field with no `UserInfo` accessor
  (`$User.Custom_Field__c`, `$Profile.Description`, `$UserRole.DeveloperName`,
  `$Organization.City`) issued a query that counted against the governor limits
  at every evaluation so a trigger batch could exceed 100 queries on globals
  alone. The Flow runtime resolves these from the running user's context, so
  these lookups now run without counting queries or query rows.
- **Get Records is bulkified in handler-class flows.** A flow generated through
  the handler path (converging branches) ran its Get Records element once per
  record, consuming one query per record where the Flow runtime bulkifies the
  lookup across a trigger batch. A before-save handler Get whose filters AND
  together with exactly one per-record condition now runs a single combined
  query for the whole batch.
- **Process Builder record updates apply on converging-branch processes.** The
  generated handler discarded the updates without issuing DML, so every record
  update in such a process silently did nothing. Updates from filter-criteria
  action groups now run in place, in group order, so two action groups updating
  the same record no longer collide on a duplicate Id.
- **Process Builder update filters comparing old record values work on
  insert.** A filter referencing the prior value threw a
  `NullPointerException` on insert, where the old map is null; the comparison
  is now made against a null value, matching the Flow runtime.
- **Flow record updates match Salesforce write-set semantics.** A record passed
  into `Flow.Interview` from Apex writes back its populated fields while
  a `Trigger.new` record passed as `$Record` into a subflow writes only the
  fields the flow assigns. SObject values entering an interview are copied, so
  assignments inside the interview no longer mutate the caller's instance.
  A record update whose input names a single-record variable now updates that
  record instead of silently doing nothing.
- **Before-save flows whose record lookup filters on a formula no longer fail
  with "undefined bind variable."** The generated query referenced the formula
  by name with nothing defining it.
- **Flow apex action inputs convert Number resources to `Integer`.** Passing a
  flow Number resource into an `@InvocableVariable` declared `Integer` failed
  the generated class's typecheck with "Illegal assignment from Decimal to
  Integer" and aborted the run before the flow executed. The conversion the
  Flow runtime performs implicitly is now applied to action inputs as well as
  to field assignments, truncating a fractional value toward zero.
- **Flow formulas that subtract two dates generate valid Apex.** A formula
  subtracting two `Date` or two `Datetime` values produced Apex that sfapex
  rejects at compile time.
- **A flow formula's spanning field references are queried by after-commit
  elements.** An async-after-commit path selected only the fields other
  elements referenced, so a formula reading any other field threw "SObject row
  was retrieved via SOQL without querying the requested field" at runtime.
- **`SELECT COUNT()` always yields an `Integer`.** Assigning the result of an
  argument-less `COUNT()` query to a static field of an enclosing class from an
  inner class stored the raw `AggregateResult` instead of the count, because
  the unwrapping depended on being able to determine the target's declared
  type. `COUNT()` is typed `Integer` in Apex regardless of the assignment
  target, so it is now always unwrapped.
- **Rollup summaries with more than one record-type filter evaluate against
  every filter.** When an object carried multiple `RecordTypeId`-filtered
  rollups, all but one kept the authored record type label rather than its Id,
  and which one survived varied between runs.
- **`--json` test output is no longer lost after a workspace-expansion retry.**
  Every test result written after startup failed with "write |1: file already
  closed" and the JSON output disappeared.

## v1.2.31 — 2026-08-14

- **Duplicate rules are enforced during insert and update DML.** Each record's
  active duplicate rules run in the save order, after validation rules and
  before the write. A rule whose action is Block or whose operations include
  Alert fails the record with `DUPLICATES_DETECTED`, using the rule's alert
  text as the error message; report-only rules never fail the save. Partial
  saves surface the failure as a `Database.DuplicateError` whose
  `getDuplicateResult()` carries the matched records.
- **A new `Division` feature models Salesforce divisions.** `--feature
  Division` adds the `Division` object and the restricted `Division` picklist
  (label "Division ID") to the standard objects that carry it in a
  divisions-enabled org, plus `DefaultDivision` on `User` and `Group`.
  Changing a record's division or reparenting it transfers its related
  records. Custom objects opt in through the CustomObject `enableDivisions`
  flag, and master-detail details of a division-carrying object get a
  read-only field automatically.
- **Every active validation rule evaluates on every update.** A rule was
  skipped during update DML unless one of the changed field names appeared in
  its formula text, and all rules were skipped when no field values changed.
  sfapex evaluates every active rule on every update, including no-op
  updates, so a rule referencing only the parent lookup and a cross-object
  formula field now fires when an unreferenced child field is edited.
- **Formula fields that reference other formula fields evaluate in dependency
  order.** Evaluation order was previously undefined, so a dependent formula
  could be computed before its dependencies and store a null that was never
  repaired, surfacing as an intermittent `NullPointerException` when Apex
  divided by a formula field after `Formula.recalculateFormulas`.
- **`Formula.recalculateFormulas` matches sfapex semantics.** Formulas
  recompute from in-memory values, with missing dependencies fetched from the
  database by Id; loaded and mutated values win over stored ones, and blank
  handling is honored. In a multi-currency org, a currency-referencing formula
  whose currency-typed dependencies hold a non-null value fails with "isoCode
  of currency data should never be null" when the record supplies no currency
  context, nulling the field and reporting a `FormulaRecalcFieldError`; a
  non-null in-memory `CurrencyIsoCode` or a database fetch supplies the
  context. Single-currency orgs always recalculate currency formulas.
- **Percent formula fields convert between display and fraction forms
  consistently.** A formula returning a bare field reference, such as
  `IF(ISNULL(Pct__c), 0, Pct__c)`, skipped the display-form conversion, so
  the same field carried the fraction or the display form depending on the
  formula's shape, and a validation rule comparing a Percent formula total to
  `1.00` fired while the total displayed 100%. Recomputed values now get the
  same conversions as ordinary reads: Apex sees the display form and formulas
  read the fraction form.
- **Blank date fields stay null in "treat blanks as zeroes" formulas.** The
  zero substitution covers only number, currency, and percent fields, so a
  Number formula like `End_Date__c - Start_Date__c` with a blank operand now
  yields null instead of a date value, which had failed with an invalid
  numeric value error when saved into a Number field from a trigger.
- **A checkbox reached through a null lookup reads false in queried
  formulas.** sfapex applies blank substitution at the leaf field
  references of a queried spanning formula, so a checkbox behind a null
  relationship is false, never blank, and `NOT(Parent__r.Checkbox__c)` is
  true when the lookup is empty; aer evaluated the whole reference to false.
  Trigger records and validation rules keep the different, also
  reference-verified, behavior: there the whole cross-object reference is
  null, so the `NOT(...)` stays false.
- **SOQL `LIKE` honors backslash escapes.** A backslash escapes only the `_`
  and `%` wildcards; before any other character it pairs with that character
  as two literals. `'Alpha\_%'` matches a literal underscore and `'Alpha\\_%'`
  matches two literal backslashes followed by the `_` wildcard. Previously the
  escapes were ignored entirely, so a pattern like `'Alpha\_%'` matched
  nothing.
- **Updating a clone of a queried row writes back every populated field.** aer
  limited a clone's update payload to explicitly assigned fields. sfapex
  derives the write set from the record's populated fields, and `clone()`
  copies all of them, so the update writes every populated field back,
  overwriting whatever another update committed in the meantime, and
  `Trigger.new` exposes those stale values to trigger guards. Clones keep
  their query provenance: accessing an unqueried field still throws.
- **User-mode DML enforces edit field permissions across the whole update
  payload.** An update only checked that the user could read each field it
  wrote, so it succeeded without edit access; it now checks every field the
  record carries, not only the fields assigned since the query. Standard
  profiles' implicit grants on standard-object fields now extend to writes
  when the user has object create or edit access, so describe no longer
  reports an object as updateable while its own standard fields are not.
- **User-mode DML accepts system-generated and upserted immutable fields.**
  An AutoNumber field carried in a queried payload no longer fails the edit
  check, and AutoNumber fields now describe as neither createable nor
  updateable. A non-reparentable master-detail field is rejected only by the
  update operation: an upsert accepts it on its update leg as well as its
  insert leg, matching sfapex.
- **Restricted picklist DML errors include the field label.** `getMessage()`
  reads "INVALID_OR_NULL_FOR_RESTRICTED_PICKLIST, Status: bad value for
  restricted picklist field: Bogus Value: [Status__c]" and `getDmlMessage()` /
  `Database.Error.getMessage()` read "Status: bad value for restricted
  picklist field: Bogus Value", matching sfapex; the label prefix was
  previously missing from both.
- **`Task.CompletedDateTime` is maintained at save time.** The field was
  previously never populated.
- **Manual shares granting a user no more access than the org-wide default are
  rejected.** Salesforce refuses such rows with `FIELD_INTEGRITY_EXCEPTION`
  ("trivial share level") on `AccessLevel`: with a ReadWrite sharing model
  both Edit and Read user shares are trivial, with Read only Read is, and
  Private accepts both.
- **Before-save flow record lookups are bulkified per trigger batch.** A Get
  Records element whose filters bind per-record values ran its query once per
  record, so a 200-record batch exceeded the 100-query governor limit.
  Repeated bind values within a batch now reuse the first lookup's result;
  before-save flows cannot perform DML, so the reuse cannot observe stale
  data. After-save flows keep per-record queries.

## v1.2.30 — 2026-08-12

- **Workflow rules run at correct stage of save order, on every DML path.**
  Previously only the `insert` and `update` keyword statements evaluated
  workflow rules at all, and they ran before the database write, so after
  triggers saw field-update values early and `Database.insert`/`update`,
  `upsert`, and lead conversion skipped workflow entirely. Rules now evaluate
  after the after triggers in all of those paths, and when a field update
  actually changes the record the before-update and after-update triggers fire
  one more time, during inserts as well, as on sfapex. Field-update and
  criteria formulas hydrate spanning parents from the lookup id rather than
  reusing a query-projected relationship image, which had resolved unselected
  parent fields to blank, and `ISNULL` over a text field is always false,
  matching the formula documentation.
- **A comma-separated workflow criteria value is a list of alternatives for
  every operation.** The list was split for `equals` and `notEqual` but
  compared as one joined string for `contains`, `notContain`, and
  `startsWith`, so a rule like `startsWith "Alpha, Beta"` never matched a
  record starting with "Alpha - ". Every operation now applies its semantics
  per alternative.
- **Rollup summary filters match each comma-separated value on its own.** The
  joined list was matched as a single pattern, with prefix or substring
  semantics applied only to the last element, so a `startsWith` filter like
  "Alpha, Beta, Gamma" gave prefix semantics to `Gamma` alone and required
  exact matches for the rest.  A child whose type began with "Alpha - " never
  entered the rollup. `contains` had the same defect with its leading
  wildcard.
- **Multi-step approval processes run from `approvalProcesses` metadata.**
  Submitting a record whose object has an active approval process assigns the
  first step's approver: a related user field is read from the target record, a
  user approver resolves by username, and a queue approver by the queue's
  developer name. An empty related-user approver field fails the submission
  with `MANAGER_NOT_DEFINED` naming the field's label, and `nextApproverIds`
  does not substitute for it; it only satisfies manually-chosen approvers.
  Approving a non-final step records a `ProcessInstanceStep`, replaces the
  acted-on workitem with the next step's, and keeps the instance `Pending`;
  only the final step's approval completes it. Validation errors carry
  sfapex's "Process failed. First exception on row N" wording. Objects
  without a definition keep the previous adhoc single-step behavior.
  `ProcessInstance.TargetObjectId` and `ProcessInstanceHistory.TargetObjectId`
  now reference every custom object, each custom object gains the
  `ProcessInstances` and `ProcessSteps` child relationships, and `ActorId` /
  `OriginalActorId` reference `Group` as well as `User`, since queues can be
  approval actors.
- **Approval process field update actions run.** Submitting, approving,
  rejecting, or recalling a record left it untouched, so a process whose
  actions set a status field never set it. A step's `approvalActions` and
  `rejectionActions` and the process's `initialSubmissionActions`,
  `finalApprovalActions`, `finalRejectionActions`, and `recallActions` are now
  resolved against the object's workflow field update definitions, including
  definitions no workflow rule references, which are now retained.
- **Record-triggered flows run in trigger order.** Registration order was
  previously undefined, so a flow whose entry criteria test a field an earlier
  ordered flow assigns could run first and silently skip itself.
- **Flow formula functions match the formula engine's blank-in, blank-out
  rule.** `DATEVALUE`, `TIMEVALUE`, `YEAR`, `MONTH`, `DAY`, `HOUR`, `MINUTE`,
  `SECOND`, `MILLISECOND`, `ADDMONTHS`, and `DATE` null-guard their operands,
  as do `LEFT`, `RIGHT`, `MID`, `UPPER`, `LOWER`, and `SUBSTITUTE`; `LEN` of
  blank text is 0. A Process Builder formula assigning `LEFT` of a null long
  text area, or an entry filter subtracting `DATEVALUE` of a null relationship
  field, threw a `NullPointerException`. `MID`'s start position is now 1-based
  and clamped to the first character, like `FIND`. `ROUND(num, digits)` now
  converts.  A flow formula using it previously aborted the run with a
  symbol-resolution error.
- **A flow's entry criteria formula is checked against prior values.** The
  "record changed to meet criteria" gate applied only to filter condition
  lists, not filter formulas, so a record that already met the criteria
  re-entered the flow and a flow's own update of its triggering record could
  re-trigger it. Cross-object paths on the old record hydrate through its
  lookup values, so they see the prior parent when the lookup itself changed.
- **Fields read only through a Loop element are queried.** A Get Records
  element's generated SOQL collected only fields referenced through the
  element's own name, so a flow that loops over the lookup's collection threw
  "SObject row was retrieved via SOQL without querying the requested field".
- **An Apex action's automatically-stored output is readable downstream.** With
  `storeOutputAutomatically`, references of the form `ElementName.field`
  resolved as static field reads and failed with "undefined static field",
  because the invocable's return value was discarded. The returned row is now
  captured per record and exposed to downstream elements.
- **The Add assignment operator on a scalar variable null-guards.** A bare
  compound assignment threw a `NullPointerException` when the value was null,
  e.g. a blank-as-blank formula field read off a queried record; a null value
  now leaves the target unchanged and a null target takes the value.
- **A flow that assigns a decimal value to an Integer field converts.**
  Assigning a Number formula, Number variable, or Decimal field reference to a
  field like `Account.NumberOfEmployees` produced Apex that failed to type
  check with "Illegal assignment from Decimal to Integer". The value is now
  coerced, truncating fractional values toward zero as sfapex does.
- **Deleting your own record no longer fails with
  `INSUFFICIENT_ACCESS_OR_READONLY`.** The `delete` keyword enforced
  object-level CRUD on every statement, but bare DML below API 67 runs in
  system mode, where object permissions are not enforced and deletes are
  governed by record-level sharing. The check now only rejects objects that
  never support delete.
- **`aer test` shows what it is doing during a long startup.** Startups that
  took a long time on large workspaces printed nothing. A heartbeat now reports
  the current phase on stderr, with done/total counts for parsed source files,
  type-checked classes, resolved classes, and loaded classes. It prints only
  when nothing else has, so verbose runs and warnings are never interleaved
  with it, and is skipped under `--quiet` and `--debug`.
- **Startup and permission checks are substantially faster.** A warm start over
  a large org schema spends less time preparing the schema.
  Field and object permission checks no longer issue a `FieldPermissions` or
  `ObjectPermissions` query on every `isAccessible`/`isUpdateable` call,
  cutting a trigger-heavy test class run times.
- **ConnectApi topics read and write storage.** `createTopic`, `getTopics`,
  `deleteTopic`, `assignTopic`, `assignTopicByName`, `unassignTopic`,
  `mergeTopics`, and `reassignTopicsByName` returned values but touched no
  storage, and the mutations reported a missing resource. They now work against
  `Topic` and `TopicAssignment` rows: assigning stores an assignment naming the
  record, its type and its key prefix; assigning the same topic twice stores
  one; the by-name form creates the topic when none of that name exists,
  comparing without regard to case; and `getTopics` lists the stored topics or
  a record's topics. Merging assigns the target wherever a merged topic is
  assigned, and reassigning by name replaces the topics a record carries.
- **Following and subscriptions are backed by `EntitySubscription`.**
  `ConnectApi.ChatterUsers.follow`, `getFollowings`, and `getFollowers`,
  `ConnectApi.Chatter.getSubscription`, `deleteSubscription`, and
  `getFollowers`, and `ConnectApi.ChatterGroups.follow` and `getFollowings` now
  read and write real rows. Following the same record twice reports the
  subscription already there rather than storing a second, a subscription is
  read and deleted by id, and paging overloads slice the result while the total
  keeps counting them all. A user who asks to follow themselves is refused with
  "users cannot subscribe to self". `Chatter.getFollowers` had returned an empty
  page for every record. `EntitySubscription.SubscriberId` now also refers to
  `CollaborationGroup`, which had been refused with
  `FIELD_INTEGRITY_EXCEPTION`.
- **`ConnectApi.UserProfiles.getUserProfile` reads the stored user** rather
  than reporting a missing resource.
- **Profile photos and banner photos are stored for users and groups.**
  `getPhoto`, `setPhoto`, `deletePhoto`, and the banner equivalents on
  `ConnectApi.UserProfiles` and `ConnectApi.ChatterGroups` read and write the
  record's photo columns, so a photo survives a restart against a persistent
  database. The `…WithAttributes` overloads now perform the write from the file
  the input names, and the upload overloads set the photo from content the call
  carries; `ConnectApi.BinaryInput` was registered with no properties, so
  constructing one did not compile at all. A crop region on the input is
  honored, and a refusal names the offending attribute. The file's content is
  decoded rather than sniffed, so a PNG truncated to its signature is refused;
  content carrying no image signature is reported as not a valid image, and
  content that announces a format and then fails to decode is told which
  formats are accepted.
- **Chatter group invitations are stored.**
  `ConnectApi.ChatterGroups.inviteUsers` stores a `CollaborationInvitation` for
  each invited address, against the group, with the running user as the inviter
  and a `Sent` status.
- **Creating an order summary builds the whole graph.** Only the root row was
  created, so the order product and delivery group summaries that hang off it
  did not exist and any query for them came back empty. Each delivery group now
  becomes a delivery group summary and each order product an order product
  summary under it, with inherited fields read from the schema rather than
  listed by hand. An order whose products have no total line amount is refused
  before anything is written, naming each offending product. Behind that, an
  order that has left `Draft` no longer accepts or releases order products and
  refuses to be deleted, and an order product takes its product from its price
  book entry when the insert does not name one.
- **`CartExtension.CartTestUtil.getCart` hydrates delivery groups and cart
  items from their records.** The wrappers were built from an Id and a handful
  of columns, leaving every other getter null, so a calculator validating the
  delivery address or an existing delivery method took its error branch and
  added a `CartValidationOutput` that sfapex does not produce. A delivery
  group now carries its address, recipient, shipping instructions, gift
  details, delivery method, totals, audit fields, `CartDeliveryGroupMethod`
  records, and selected method; a cart item carries its prices, line totals,
  adjustment amounts, weight, lookups, audit fields, `CartTax` records, and
  `CartItemPriceAdjustment` records. A child item nests under its parent rather
  than appearing as a top-level item. `CartItem.TotalWeight`,
  `TotalLineNetAmount`, and `TotalLineGrossAmount` are maintained by the
  platform and are now non-createable and non-updateable.
- **A static product bundle's components traverse back from the parent
  `Product2`.** The parent-to-components subquery failed with SQL "no such
  column" because `Product2`'s child relationship pointed at a `ProductId`
  field that does not exist on `ProductRelatedComponent`, and the names were
  cross-wired: `ParentProductId` is reached through
  `ChildProductRelatedComponents` and `ChildProductId` through
  `ParentProductRelatedComponents`.
- **Metadata file names are percent-decoded.** The Metadata API percent-encodes
  special characters when it writes a full name as a file name, so a profile
  named "My (Custom) Profile" ships as
  `My %28Custom%29 Profile.profile-meta.xml`. The encoded string was stored as
  the record name, so a SOQL query for the real name returned nothing.
- **Public groups load from source metadata.** `groups/*.group-meta.xml` was
  never parsed, so queries against `Group` returned no `Regular`-type rows.
  `Name`, `DeveloperName`, and `DoesIncludeBosses` are read, with
  `DoesIncludeBosses` defaulting to true when omitted, and explicit `.group`
  file arguments work with `aer test`.
- **A product attribute set's items load.**
- **A user-supplied `TaskStatus` standard value set drives `Task.IsClosed`.**
  It updated only the `Task.Status` picklist, so a task with a custom closed
  status like `Canceled` kept `IsClosed` false and appeared in `WHERE IsClosed
  = false` queries.
- **A field named after a SQL keyword can be read and written.** Fields such as
  `Payment.Date` and `PartyConsent.Action` could not be written at all, and
  reading one failed with a raw SQL error, including from a subquery.
  Separately, `AssessmentQuestionResponse`,
  `AssessmentQuestion`, `AssessmentQuestionVersion`, `AssessmentQstnVerChoice2`,
  and `DisclosureDefinitionVersion` join the Health Cloud feature.
- **`Task` has a `RecordTypeId` field, and `RecordTypeId` is hidden on objects
  with no record types.** `Task` carried no `RecordTypeId` while `Account`,
  `Contact`, `Lead`, `Event`, and `Campaign` all did, so a package shipping a
  formula field on `Task` that dereferences `RecordType.Name` failed to load
  with 'relationship "RecordType" not found on Task'. sfapex hides
  `RecordTypeId` entirely on an object with no record types, but a query for it
  died with a raw SQL error instead of a `QueryException`, and `get` returned
  null while `put` silently succeeded.
- **An invalid validation rule is rejected at load.** A rule whose formula or
  error display field references something that does not exist is now rejected
  the way an unresolvable lookup filter reference is, or dropped with a report
  under `--skip-errors`. An unresolvable reference no longer reads as null,
  which had let `ISBLANK` of a missing field evaluate true and fire the rule
  against records it was never meant to match, and a non-boolean rule result is
  an error rather than a firing rule. References inside function arguments are
  namespaced and checked like any other, having previously been skipped.
- **`$RecordType` resolves in validation rule formulas.** The global never
  resolved, so a rule comparing `$RecordType.DeveloperName` silently skipped
  its record-type branch. It now reads the record's `RecordTypeId` and resolves
  the remaining path against the `RecordType` row; a record with no record type
  resolves to the synthetic Master record type. A validation rule
  `DmlException` also records the failing row index, so `getDmlIndex` reports
  the failing record's position in a bulk statement instead of echoing its
  argument.
- **A trigger `addError` `DmlException` reports the failing record's row.** The
  message read "First exception on row 1" for a failure on row 0 of a
  two-record insert, and `getDmlIndex` reported 0 even when the second record
  was the one that failed. Both now carry the failing record's position, for
  single-record and list `insert` and `update`.
- **An untranslatable formula construct fails conversion instead of emitting
  invalid SQL.** A `CASE` formula whose value or result expression had no SQL
  translation silently skipped that `WHEN` pair, emitting
  `CASE <operand> ELSE NULL END`, which SQLite rejects with a syntax error — so
  every query selecting the object failed. The same silent drop existed for
  concatenation operands, `{!field}` merge-field syntax, and the
  `fieldReference.VALUE` grammar. Such a formula now reports the field that
  could not be translated.
- **`Math.min` and `Math.max` return `Integer` and `Long`.** Non-`Decimal`
  arguments were coerced to a float, so the `(Integer, Integer)` and
  `(Long, Long)` overloads returned `Double`, breaking callers that check the
  runtime type, such as `Database.Cursor.fetch`. Two `Integer`s return
  `Integer`, `Integer`/`Long` pairs widen to `Long`, `Decimal` pairs return the
  picked operand preserving its scale.
- **`System.assertEquals` and `System.assertNotEquals` reject a null message.**
- **A builtin instance method's return type is known at compile time.**
  Comparing a `String` to an enum-typed `CartExtension` getter result fails to
  compile on sfapex but only failed at runtime in aer. Builtin instance
  methods beyond a hard-coded set had no known return type until they ran.
- **`ClassName.Member` resolves a nested type over a same-named instance
  field.** When a public instance field or property shared its name with an
  inner enum or class, differing only by case, the member access bound to the
  instance field and expressions like `Outer.InnerEnum.CONSTANT` threw a
  `NullPointerException`. Static fields keep precedence over nested types.
- **A `COUNT()` result assigned to another class's static field unwraps.**
  `Other.total = [SELECT COUNT() …]` stored the raw `AggregateResult` instead
  of the numeric count, unlike the same assignment to a local variable or an
  instance field.
- **A large `Blob` body inserts as `ContentVersion`.** `Blob.valueOf` stores a
  large concatenation result unflattened, which the content-size calculation
  did not recognize, so the insert failed with a spurious
  `REQUIRED_FIELD_MISSING` on `VersionData` whenever the body was 16KB or
  larger. An empty body still fails with `REQUIRED_FIELD_MISSING`, matching
  sfapex.
- **`System.enqueueJob` from trigger code starts a fresh queueable chain.** An
  enqueue from a trigger fired by DML inside an executing queueable is not
  chained onto that queueable: sfapex starts a fresh chain at depth 1, so
  the enqueue is exempt from the test-context chaining `AsyncException` and may
  configure its own explicit depth. It still counts against the
  one-queueable-job-per-queueable-transaction `LimitException`. `AsyncInfo`'s
  queueable-only methods throw "not allowed outside a Queueable of Finalizer
  execution" in trigger code even when the trigger was fired from a running
  queueable.
- **A managed package's `@TestSetup` queries are charged to the package.** The
  CLI class runner ran `@TestSetup` without a limits context, and the fallback
  counters have no namespace segregation, so queries a package ran during setup
  merged into the subscriber baseline every test method inherits.  A package
  seeding twenty queries in setup could fail a subscriber test with "Too many
  SOQL queries: 101".
- **`DataSource.Table.get(name, nameColumn, columns)` defaults its labels.**
  The three-argument overload left `labelSingular`, `labelPlural`, and
  `description` unset; the documented behavior is that all three take the table
  name.
- **`aer package mock` captures protected custom settings.** A protected custom
  setting is hidden from every subscriber-facing metadata surface, but a
  package's global class stubs can still reference it in method signatures, so
  `aer package create` failed on such a mock with unknown-type errors. Missing
  signature types are now retrieved through the Tooling API and confirmed
  protected, and the setting's protected visibility is preserved when the
  package is built, loaded, and unpacked.
- **`aer package mock` fails instead of producing a package that cannot be
  loaded.** It now names the fields the package references but does not contain,
  the sign that the org user running it cannot read them. It also keeps only
  the child relationships reached through the package's own fields, since a
  subscriber lookup added to a packaged object surfaced under an unprefixed
  name that clobbered the package's own relationship. Separately, a saved
  package no longer carries namespaced lookup filter references that belong to
  the merged schema rather than the package.

## v1.2.29 — 2026-08-02

- **ConnectApi Chatter feeds are backed by storage.**
  `ConnectApi.ChatterFeeds.postFeedElement` threw "Resource not found." for
  every call and `getFeedElementsFromFeed` always returned an empty page.
  Posts, comments, and likes now read and write `FeedItem`, `FeedComment`, and
  `FeedLike` rows, so a post is visible through SOQL and through the subject's
  entity feed, and `CommentCount` and `LikeCount` are maintained as sfapex
  maintains them. Posting runs the DML pipeline, so before-save flows and
  before/after triggers see it. `updateFeedElement`, `updateComment`,
  `deleteFeedElement`, and `deleteComment` complete the edit and delete path,
  and deleting a feed element takes its comments and likes with it.
  `FeedItem`'s `FeedComments` and `FeedLikes` child relationships were not
  marked cascade-delete, so a plain DML delete of a post left its comments
  behind. Inside a test, `getFeedElementsFromFeed` and `searchGroups` are still
  answered only from a response registered with the matching `setTest…` method.
- **The feed element representation matches sfapex.** A plain text post
  carries twelve capabilities; ten were missing (`associatedActions`,
  `bookmarks`, `close`, `edit`, `interactions`, `mute`, `readBy`, `status`,
  `topics`, `upDownVote`). A post now defaults to `Visibility=InternalUsers`
  with a header reading "<user> to <org> Only", `postFeedElement`
  honors the input's visibility, `relativeCreatedDate` reads "Just now" on a
  fresh post, and `FeedItem.Status` is `Published`. `createdDate` and
  `modifiedDate` had been null on every feed element. `getBuildVersion` returns
  the API version rather than 1.0, and a ConnectApi object's `toString` renders
  nested values through the Apex formatter instead of producing
  `ConnectApi.X:[…]` with a stray colon.
- **ConnectApi Chatter groups are backed by storage.** `createGroup`,
  `getGroup`, `updateGroup`, `deleteGroup`, `getGroups`, and the batch readers
  work against `CollaborationGroup`, and the membership methods against
  `CollaborationGroupMember`, with `MemberCount` maintained in storage. Group
  records, announcements, membership requests, and group photos are stored too,
  and `getMyChatterSettings` / `updateMyChatterSettings` read and write the
  membership's `NotificationFrequency`. Group names are unique across the org,
  so a plain insert or rename of a `CollaborationGroup` with a duplicate name
  is refused with `DUPLICATE_VALUE` on `Name`. Deleting a group takes its
  memberships and membership requests with it. Throughout, an id belonging to
  another table is rejected as an illegal parameter value while a well-formed
  id with no stored row is reported as not found.
- **Deleting a parent record no longer fails with "FOREIGN KEY constraint
  failed".** Cascade children were deleted after the parent row, so a child
  whose foreign key is not declared `ON DELETE CASCADE` failed the constraint
  before the cascade could run. Children are deleted first now.
- **Catching a ConnectApi exception as `Exception` and calling `getTypeName()`
  no longer panics the VM** with "marked as IsNative but no implementation
  registered". Native implementations resolve through the declaring type's
  extends chain, which covers every exception subclass in every namespace.
- **Flow formulas match the formula engine's null semantics.** Flow formulas
  never throw on null values, and the generated Apex now agrees: arithmetic
  evaluates to null when an operand is null, field references through a
  single-record lookup output get safe navigation, `TEXT(null)` converts to `''`
  so downstream string methods no longer throw, and a blank text value assigned
  to a lookup field stores null rather than raising `StringException`.
- **More flow elements and formula functions convert.** An element or formula
  function the converter did not know left a call to a never-generated method or
  a reference to a never-declared variable, and the run aborted with a type
  error. Custom Error, Collection Filter, Collection Sort, subflow elements, the
  Submit for Approval core action, and Process Builder Wait elements are now
  converted, along with the `TRIM`, `CONTAINS`, and `FIND` functions and `+` on
  text operands. Update Records and Delete Records accept record collection
  variables and query results, not only a single SObject. Action fault
  connectors are honored: a failed action routes to its fault connector with
  `$Flow.FaultMessage` populated.
- **`$Profile`, `$Organization`, and `$Flow.InterviewGuid` resolve in flows.** A
  reference to any of them fell through to a bare identifier or a null literal,
  and the generated handler failed to compile. `$Profile` and `$Organization`
  now resolve like `$UserRole`, and `$Flow.InterviewGuid` is seeded once per
  interview. A global variable's Id renders in its 15-character form where the
  value becomes text, matching Flow, while the reference itself keeps all 18 so
  it stays assignable to a lookup field.
- **A flow reference naming a polymorphic branch reads only that branch.**
  `Owner:Group.Name` reads the owner's name when the owner is a Group and null
  when it is a User; the converter dropped the type specifier, so both branches
  read the same value. The specifier is now a guard on the related record's
  actual type.
- **A Create Records element's output is the new record's Id.** The converter
  named the created SObject after the element, so a flow assigning
  `{!Create_X}` to a lookup field generated an illegal assignment and the
  generated class failed to compile. Behind that, the trigger optimizer batched
  in-loop DML after the loop, leaving the inserted record's Id null for the rest
  of the iteration; an insert now stays inline when a later statement in the
  loop reads its target.
- **Two Get Records elements over the same object no longer collide.** Both
  produced the same local, list temporary, and per-batch cache, which the
  generated class rejects as duplicate declarations; the local takes the
  element's name now. An element two scheduled paths converge on was also
  emitted once per path, duplicating its declarations.
- **Methods can be declared in a trigger body.** sfapex compiles a trigger
  body like an anonymous block — methods declared alongside statements are
  static whether or not the keyword is written, and the body's top-level
  variables are shared with them. aer rejected such triggers with "unsupported
  trigger member declaration", and because the parse produced errors the files
  were re-parsed on every run.
- **The `Id` field is accepted as an upsert external ID.** `Id` is a standard
  indexed field on every object, but the type checker rejected it, and the
  `upsert` statement form failed at runtime with `MISSING_ARGUMENT` because the
  reference reached the VM as authored (`Account.Id`). The `Object.Fields.Field`
  spelling and the two-argument overload are validated now as well, and the
  `MISSING_ARGUMENT` message for an empty external ID value reads
  "<Field> not specified" with no field names, matching sfapex.
- **`System.Location.latitude` and `.longitude` can be read as fields.** Only
  `getLatitude()` and `getLongitude()` worked; reading `loc.latitude` raised
  "invalid field access target".
- **A double implicitly converted to `Decimal` keeps its rendering.** A JSON
  numeric deserialized into a Number or Currency field is a runtime Double, and
  Salesforce keeps the double's rendering when it becomes a `Decimal`, so a
  whole number keeps its trailing `.0`. aer dropped it everywhere except the
  explicit `(Decimal)` cast.
- **A `List` custom setting's `Name` is required.** `isNillable()` reported it
  as nillable while DML threw `REQUIRED_FIELD_MISSING`, so describe-driven test
  factories never populated `Name` and their inserts failed. A Hierarchy
  setting's `Name` stays nillable, since its records are keyed by
  `SetupOwnerId`.
- **Person Account behavior is derived from the loaded schema.** Inserting a
  person account through `aer exec` failed with `REQUIRED_FIELD_MISSING` on
  `Name`: only the `test` command enabled the runtime flag, so under `exec` and
  the server no person account record type was assigned and `Account.Name` was
  never derived from the name fields. The schema is now the signal, covering
  `test`, `exec`, `server`, and the pooled VM restore.
- **Ordering by a standard picklist follows the value set rather than the
  alphabet**, as `MIN` and `MAX` already did, with values outside the value set
  breaking the tie on the value itself.
- **Health Cloud, Survey, Omni-Channel, and Workplace Command Center objects are
  modeled.** `--feature HealthCloud` gains the care plan family (`CarePlan`,
  `CarePlanActivity`, `CarePlanDetail`, `CarePlanTemplate` and friends) and the
  assessment objects (`Assessment`, `GoalAssignmentDetail`,
  `ExternalAssessmentDefinition`, `OmniProcess`); `SurveySettings.enableSurvey`
  gains the full Survey family from `Survey` and `SurveyVersion` through
  `SurveyQuestionResponse` and `SurveyEngagementContext`; Omni-Channel gains
  `UserServicePresence`; and a new `WorkplaceCommandCenter` feature models
  `LocationTrustMeasure`.
- **Settings-gated objects auto-enable from a static Apex reference.** Code
  referencing `Territory2`, `AccountTeamMember`, `OpportunityTeamMember`, a
  Survey object, or a Health Cloud object failed type checking with an
  unresolved reference unless the matching settings file was found. A static
  reference now turns the setting on, including through a field token such as
  `CarePlan.CaseId.getDescribe()` where the object never appears in a type
  position.
- **An unrecognized `--feature` value is an error.** `--feature InvalidFeature`
  was silently ignored, and the only symptom was a confusing "Invalid type" from
  the type checker. The value is now rejected at flag parse time with the list
  of valid features, across `exec`, `test`, `server`, and the schema inspect
  commands, and an accepted value is canonicalized so `--feature healthcloud`
  works.
- **Packages carry Visualforce pages and custom tabs.** Subscriber source
  referencing a managed package's page (`Page.ns__Foo`) or a tab in a profile or
  permission set looked like a missing org dependency, because `package mock`
  captured neither. `package mock` now queries `ApexPage` and lists `CustomTab`,
  `package list` gains Visualforce Pages and Custom Tabs sections, and
  `package unpack` writes `pages/<Name>.page` and
  `tabs/<Name>.tab-meta.xml`. A tab whose name is not an SObject API name (a
  Visualforce, Lightning component, or web tab) now takes the package
  namespace.

## v1.2.28 — 2026-07-31

- **User-mode DML no longer rejects records queried with a parent
  relationship.** Updating a record in user mode failed with
  `NoAccessException: No Access to field: Child__c.Parent__r` whenever the
  record had been queried with a relationship field, even when the lookup field
  and both objects were explicitly permitted. A relationship key holds the
  related record the query populated rather than a value the DML writes, and no
  `FieldPermissions` row can grant access to a `__r` pseudo field, so the check
  now skips it; field-level security still applies to the lookup field itself.
- **`SObject.put()` marks only related records as relationship fields.** Every
  field written through `put()` was recorded as a relationship container,
  whatever the value, so a record whose fields were all set that way presented
  an empty write set: `Security.stripInaccessible` removed nothing and
  `Database.update(record, AccessLevel.USER_MODE)` raised nothing, and fields
  the running user has no field-level access to were written. Only an `SObject`
  value is marked now.
- **Permission sets from source and packages report `IsCustom = true`.**
  `PermissionSet.IsCustom` was always false, so a selector using the common
  `WHERE IsCustom = true` filter skipped every permission set deployed from
  source or shipped in a package, and code assigning permission sets through
  such a selector failed. `IsCustom` is now derived at insert: a permission set
  is custom unless the builtin schema seeds it or it belongs to a profile that
  is not custom. A permission set the platform ships also reports
  `Type = Standard` rather than `Regular`.
- **`PermissionSet` subqueries resolve.** `PermissionSet` carried no child
  relationships, so a subquery such as `SELECT (SELECT SobjectType FROM
  ObjectPerms) FROM PermissionSet` failed to resolve the relationship name.
  Seven relationships are now declared: `ObjectPerms`, `FieldPerms`,
  `SetupEntityAccessItems`, `Assignments`, `PermissionSetGroupComponents`,
  `SessionActivations`, and `PermissionSetTabSetting`.
- **`UserInfo.isCurrentUserLicensed` reports loaded packages as licensed.** It
  returned false for every namespace, so a package loaded with `--package`
  looked unlicensed to both the subscriber and the package's own code. It now
  returns true for the namespace of any loaded package and for the VM's default
  namespace, which models a packaging org licensing its own namespace to the
  developer. A namespace with no package and no schema behind it still raises
  `System.TypeException`.
- **The default System Administrator profile grants `PermissionsInstallPackaging`
  and `PermissionsModifyMetadata`.** Both were false on the profile-owned
  permission set, so `UserPermissionAccess` reported them as false too. This
  also resolves an inconsistency in the permission dependency closure:
  `PermissionsAuthorApex` requires `PermissionsModifyMetadata`, which the admin
  profile granted without it.
- **Entity feeds are modeled as projections of `FeedItem`.** An entity feed such
  as `AccountFeed` or `MyObject__Feed` was synthesized as a standalone object
  backed by a table nothing ever wrote to, so a feed query always came back
  empty and a subquery over `FeedComments` failed to compile. Each feed is now
  built from the `FeedItem` field definitions with `ParentId` narrowed to its
  parent type, gaining the `IsRichText`, `RelatedRecordId`, `InsertedById`, and
  `BestCommentId` fields and the `FeedAttachments`, `FeedComments`, `FeedLikes`,
  `FeedSignals`, and `FeedTrackedChanges` child relationships, and posted
  `FeedItem` records are visible through it. Every custom object is listed in
  `FeedItem.ParentId`, so posting to a custom record no longer fails with
  `INSUFFICIENT_ACCESS_ON_CROSS_REFERENCE_ENTITY`. `FeedPostId` is now hidden
  after API 21.0 for describe as well as inline and dynamic SOQL, and a bare
  `__Feed` name in a `FROM` clause resolves against the executing namespace.
- **`--skip-errors` drops metadata with missing dependencies.** The flag covered
  parse and type check errors only, so a workspace whose metadata referenced
  something that was not loaded aborted the run outright. Four schema-assembly
  errors now drop the offending metadata instead: a custom field whose
  `referenceTo` names an absent sObject, two lookups on one parent deriving the
  same child relationship name, a field bound to an absent global or standard
  value set, and a permission set group naming an absent permission set.
  Removals cascade: formula fields (including those reaching through a removed
  lookup), rollup summaries, compound components, validation rules, field set
  members, lookup filters, dependent picklist controllers, and field permissions
  go with what they depended on, and everything removed is reported in one
  warning. All four errors still fail the run without the flag.
- **Apex custom adapters for external data sources are supported.** An
  `ExternalDataSource` whose `Type` names an Apex class extending
  `DataSource.Provider` was rejected for having no URL or endpoint; a custom
  adapter supplies its own connection and needs an endpoint only when the
  provider declares `DataSource.Capability.REQUIRE_ENDPOINT`. The overridable
  `DataSource.Provider` and `DataSource.Connection` methods
  (`getCapabilities`, `getConnection`, `query`, `search`, `sync`, `upsertRows`,
  `deleteRows`) are also marked virtual, so overriding them no longer fails with
  "cannot override non-virtual parent method". When a namespace is applied, the
  adapter class reference is qualified with it (`ns.ProviderClass`), which is
  how the platform stores it.
- **Datetimes are stored at whole-second precision, and `SystemModstamp`
  advances on update.** sfapex keeps datetimes only to the whole second,
  including `CreatedDate`, `LastModifiedDate`, and `SystemModstamp`, so rows
  stamped within one second tie under `ORDER BY` and the tie stays in insert
  order. Storage wrote sub-second precision, which broke ordering two ways: a
  row carrying a fraction sorted after its same-second siblings, and querying a
  record and updating it rewrote the field with the shorter value, moving that
  record under a later `ORDER BY`. `SystemModstamp` is also included in the
  `UPDATE` statement now.  Nothing can assign it, so it kept the value it
  received at insert and `ORDER BY SystemModstamp` could not find the most
  recently modified record.
- **Formula fields reject every function sfapex forbids.** Field validation
  rejected `REGEX` alone. `HTMLENCODE`, `PRIORVALUE`, `ISCHANGED`, `ISNEW`, and
  `IMAGEPROXYURL` are equally unusable in a formula field, and are now rejected.
- **`CURRENCYRATE`, `GETSESSIONID`, and `IMAGE` are implemented.** The first two
  failed with an unsupported-function error, and `IMAGE` returned the bare URL
  rather than markup.
- **More formula functions can be filtered and sorted on.** A formula field
  using a function with no SQL translation could not be used in a `WHERE` or
  `ORDER BY` clause.  Translations were added for `TRUNC`, `MOD`, `PI`, `SQRT`,
  `EXP`, `LN`, `LOG`, `SIN`, `COS`, `TAN`, `ASIN`, `ACOS`, `ATAN`, `ATAN2`,
  `INITCAP`, `LPAD`, `RPAD`, `ASCII`, `CHR`, `DAYOFYEAR`, `ISOWEEK`, `ISOYEAR`,
  `DATETIMEVALUE`, `UNIXTIMESTAMP`, `FROMUNIXTIME`, `PICKLISTCOUNT`,
  `FORMATDURATION`, `TIMEVALUE`, `TIMENOW`, `SECOND`, `MILLISECOND`, and
  `DISTANCE` over a Location field.
- **`DATE()` returns null for an out-of-range or null component.**
  `DATE(2024, 13, 1)`, `DATE(2024, 0, 1)`, `DATE(2024, 2, 30)`, and
  `DATE(2024, 1, 0)` are all null rather than rolling over into January 2025,
  December 2023, March 1, and December 31. In generated SQL, a null year, month,
  or day also produced the string `0000-00-00` instead of a null date, so a
  query for a date formula over blank inputs both returned a non-null value and
  failed to match a null filter.
- **`BlankAsZero` zeroes only blank field references.** The setting coerced
  every null it encountered to zero, including nulls returned by functions.
  `YEAR()`, `MONTH()`, and `DAY()` of a blank date correctly return null, so
  `DATE(YEAR(d) + 3, MONTH(d), DAY(d))` over blank inputs built year 3 with
  month 0 and day 0, which rolls back to a date outside the storable range and
  failed the insert with `FIELD_INTEGRITY_EXCEPTION`. A blank numeric field
  still reads as zero; a null returned by a function now propagates.
- **`MOD` by zero returns the dividend.** `MOD(7, 0)` is 7, not null.
- **`INITCAP` capitalizes after any non-alphanumeric character.** It split on
  whitespace and rejoined with a single space, so runs of spaces collapsed and
  `a  b` capitalized to `A b`. Separators are copied through unchanged now, so
  `o'neill-smith` becomes `O'Neill-Smith` while a letter after a digit is left
  alone and `3rd` stays `3rd`. `LPAD` and `RPAD` with a width of zero or less
  yield an empty string instead of failing.
- **Numeric formula fields are rounded to the field's scale in queries.** A
  formula field's value is stored at its scale, so a translated expression
  carrying full precision disagreed with the runtime as soon as a result had
  more digits than the field holds — `LN(100)` read as 4.605170186 against a
  stored 4.6052. Rounding previously applied to Currency only.
- **`TIMENOW()` and time components are available in formulas and flows.**
  `TIMENOW()` returns the current GMT time of day, honoring the mock clock like
  `NOW()`. `HOUR`, `MINUTE`, `SECOND`, and `MILLISECOND` gained runtime
  implementations extracting UTC components, and all of them, along with
  `TIMEVALUE`, are converted by the flow converter.
- **`??` promotes its operands to the wider numeric type.** The result type was
  taken from the left operand alone, so `(nullDecimal ?? 0).longValue()` failed
  at runtime with "method not found: Integer.longValue" and
  `(someInteger ?? 0.0).longValue()` was rejected at compile time. Both sides
  now widen along Apex's implicit numeric chain (Integer to Long to
  Double/Decimal), including through chained coalescing.
- **A blank picklist field argument is typed as `String` in overload
  resolution.** A null value carries no runtime type, so an SObject field
  argument reading as null takes its type from the schema.  Picklist and
  multi-select picklist had no Apex equivalent, so the argument was typed
  `picklist`, matched no parameter, and the call failed with no matching
  constructor found. Both map to `String` now.
- **`SObject.Id` is typed, and unconvertible primitive arguments are rejected.**
  Field access on a value typed as the generic `SObject` base type produced no
  type, so an overload set taking `List<Id>` and `Id` bound the wrong overload
  and its `List<SObject>` return value then failed to assign. Separately, a call
  to a user-defined method was judged only on arity, so an argument whose type
  can never convert to the parameter type was accepted; a `String`, `Boolean`,
  or `Id` argument for a `Decimal` parameter, a `Decimal` for an `Integer`, and
  a `Datetime` for a `Date` now report sfapex's "Method does not exist or
  incorrect signature" error. Numeric widening, `String`/`Id` interchange,
  `Date` to `Datetime`, an `Object` parameter, and null still bind.
- **Enum constants the enum does not declare are rejected at compile time.**
  Constants were validated for builtin enums only; a reference to an undeclared
  constant of an enum declared in source or shipped in a package passed the type
  check and failed at runtime. It now reports "Variable does not exist", the
  message builtin enums already produce.
- **A custom object component file whose path cannot name its object is
  rejected.** A field file placed directly under `objects/`, for instance,
  resolved to a meaningless object name and was silently dropped from a
  reference deployment. The required layout and the intended file are now
  reported.
- **Faster runs with heavy `@TestSetup` data and query volume.** Batch
  `UserRecordAccess` rebuilds share their batch-invariant state instead of
  re-querying per (record, user) pair.
- **A source path that expands to itself no longer re-runs the load.**
  Expanding a source path to its surrounding workspace resolves a package
  directory such as `force-app` back to that same directory, but expansion still
  reported that it had broadened the load set, so a type-check failure triggered
  a retry that re-parsed, re-analyzed, and re-checked the same classes to
  produce the same errors.

## v1.2.27 — 2026-07-29

- **Visible custom tabs are queryable through `TabDefinition`.** Permission set
  `tabSettings` and profile `tabVisibilities` were parsed and discarded, so a
  custom-object tab never appeared in a `TabDefinition` query even when the
  running user was assigned a permission set granting it. Tab visibility now
  becomes `PermissionSetTabSetting` rows — `Visible` maps to `DefaultOn`,
  `Available` to `DefaultOff`, and `None`/`Hidden` grant nothing — and
  `TabDefinition` reports a tab when any of the running user's permission sets
  (profile-owned, directly assigned, or group-expanded) exposes it, so both
  `--assign-perms` and `runAs`-based assignment work. Custom-object tabs report
  the object API name as `SobjectName` and the object's plural label as `Label`.
  Tabs shipped in a managed package are namespaced like object and field
  permissions, and a `.tab` file passed to `aer test` as an explicit argument is
  now parsed the same way the directory walker parses it, so its
  `CustomObject` relationship survives.
- **Inactive picklist values are rejected on restricted picklists.** aer
  accepted any entry in the value set regardless of its `isActive` flag;
  assigning an inactive value to a restricted picklist now fails with
  `INVALID_OR_NULL_FOR_RESTRICTED_PICKLIST`, matching Salesforce. Inactive
  values still drive save-time casing normalization, so a case variant of an
  inactive value is rewritten to the value set's casing first — the rejection
  message reports the normalized casing, and on an unrestricted picklist the
  value is stored with the inactive entry's casing. Describe results continue to
  exclude inactive entries.
- **`Database.insertImmediate` publishes standard platform events.**
  `Database.insertImmediate(new OrderSummaryCreatedEvent())` failed with
  "Argument must be of virtual sObject type." even though Salesforce accepts it.
  Standard platform events (those without an `__e` suffix) are now accepted by
  both insert variants: `insertImmediate` publishes them through the event bus,
  while `insertAsync` compiles and then throws `System.TypeException`
  ("Asynchronous DML not allowed on <EventName>") at runtime. Custom `__e`
  events are still rejected at compile time with the virtual-type error, and the
  update and delete variants keep the strict virtual-type requirement.
- **`ProductAttributeSet` metadata is queryable.** `productAttributeSets` files
  in a source tree were walked but never converted to records, so a SOQL query
  against `ProductAttributeSet` returned zero rows. The files are now loaded as
  `ProductAttributeSet` records, and their presence enables the B2B Commerce
  schema on its own — previously the object's table existed only when some class
  in the run referenced the type.

## v1.2.26 — 2026-07-29

- **Insert audit fields are assigned after validation rules evaluate.**
  `CreatedDate`, `CreatedById`, `LastModifiedDate`, `LastModifiedById`, and
  `SystemModstamp` were stamped before custom validation rules ran, so a rule
  such as `NOT(ISBLANK(CreatedDate))` blocked every create. Before-insert
  triggers and insert-time validation rules now see blank audit fields, while
  after-insert triggers, queries, and an update's validation rules see the
  assigned values. Values derived before the save — field defaults,
  auto-populated lookups, the default record type, and the `*State`/`*Country`
  labels derived from `*StateCode`/`*CountryCode` — are still populated before
  the rules evaluate, so a rule checking `ISBLANK(BillingState)` sees the label
  derived from `BillingStateCode`.
- **An Id assigned by a before-insert trigger is discarded instead of failing
  the insert.** An insert failed with "cannot specify Id in an insert call" when
  a before-insert trigger assigned a temporary Id to a `Trigger.new` record.
  sfapex runs the caller-Id check before triggers fire and ignores any Id a
  before-insert trigger sets, generating the persisted Id itself, so the insert
  now succeeds. An Id supplied by the caller is still rejected.
- **Picklist casing is applied when the record is saved, not when the field is
  assigned.** An in-memory SObject assigned `oPeN` now keeps `oPeN`, while the
  record handed to triggers and to storage reads `Open`, matching sfapex.
  Normalization also covers standard and unrestricted picklists. Multi-select
  picklists are stored in picklist-entry order rather than the order the
  selections were given, with selections outside the value set appended in the
  order supplied — inserting `'Phone;Email'` into a field whose entries are
  Email then Phone stores `'Email;Phone'`.
- **Assertions compare with `equals()` semantics rather than Apex `==`.**
  Strings differing only in case are no longer equal, including Strings nested
  in Lists, Sets, and Maps, and `System.assertEquals` no longer splits
  multi-select picklist field arguments on `;` to compare them as sets. Every
  assertion failure message now uses the sfapex wording, starting with
  "Assertion Failed", and `System.assertNotEquals` honors its message argument,
  which it had ignored. Two divergences the case-insensitive comparison had been
  hiding are also fixed: `Schema.FieldSet.getName()` returned a lowercased name,
  and builtin enums whose constants are normalized to upper case reported those
  constants from `name()`. `Metadata.DeployStatus` gets display names and
  sfapex's `values()` order.
- **Typed JSON deserialization of temporal values throws sfapex's
  messages.** `JSON.deserialize` into `Date`, `Datetime`, and `Time` targets
  throws `JSONException` with the Joda-style parse errors sfapex produces,
  including the document location of the failing member, and SObject fields
  throw the Jackson-style "Cannot deserialize instance of ..." errors. A JSON
  number in a `Date` position reads as a year, truncated ISO values default
  their missing components, and out-of-range SObject `Datetime` components roll
  over instead of throwing.
- **Querying `UserPermissionAccess` returns the running user's effective
  permissions.** One row describing the running user is rebuilt before each
  query against the object, so `WHERE` clauses, `COUNT()`, and field selection
  work normally. Every boolean `Permissions*` field aggregates across the user's
  permission sets and Profile.
- **Every `PermissionSetGroup` gets its companion `PermissionSet`.** sfapex
  exposes a generated permission set for each group, carrying the group's
  developer name and label, `Type = 'Group'`, and `PermissionSetGroupId`
  pointing back at the group. It is created with the group, deployed from
  source or inserted from Apex, and deleted with it, so code that queries
  `PermissionSet` by a group's developer name now finds a row.
  `PermissionSet.Type`, which aer left null on every permission set, is now
  derived: `Group` for a group's companion, `Profile` for a profile's permission
  set, `Session` for one that requires session activation, and `Regular`
  otherwise.
- **Permission set groups load with their components and their source status.**
  A group's `<permissionSets>` members are parsed at load and
  `PermissionSetGroupComponent` rows synthesized, so assigning a group grants
  the object and field permissions its component permission sets contribute and
  `Test.calculatePermissionSetGroup` has something to aggregate. Members that
  reference package permission sets by their `Namespace__Name` form now resolve
  instead of failing startup with "references unknown permission set". A group's
  `Status` comes from its source `<status>` element, defaulting to `Outdated`
  when none is declared; a group with no components loads `Updated`, since there
  is nothing to aggregate.
- **Overload resolution uses the declared type of a null argument.** A null read
  through a field or property carried no type, so every overload tied and the
  first declared constructor won. `new Target(wrapper.recordId)` with a null
  `Id` property now picks the `(Id)` overload rather than an `(Account)` one,
  matching how Salesforce selects from the argument's static type.
- **A list custom setting's `getInstance(String)` returns null for a null or
  blank name.** It previously selected an arbitrary record, so code looking up
  a setting by a nullable name silently received an unrelated configuration
  record, or a non-null empty instance when the table held no records.
- **Delivered platform events carry `CreatedById` and `CreatedDate`.** A
  subscriber trigger reading either field on an event published with
  `EventBus.publish` saw null, losing the identity of the publishing user. Both
  are now stamped on the delivered event and agree with the stored row:
  `CreatedById` is the running user and `CreatedDate` is the publish time at
  second precision. As with `ReplayId`, the publisher's own reference keeps them
  unset.
- **A SOSL `RETURNING` clause naming the same type twice produces one result
  group.** `RETURNING Account(Id), Account(Id)` returned one group per clause
  (`[[], []]`); Salesforce returns one group per distinct SObject type, for both
  `Search.query()` and inline `[FIND ...]` literals.
- **`BatchApexWorker` jobs report `LastProcessedOffset`.** A completed worker
  `AsyncApexJob` now reports the number of records its scope processed. The
  record is created with `-1`, which stays queryable when `execute()` throws,
  and is updated to the processed-record count when the scope completes. Worker
  records exist only for `QueryLocator`-based batches; an iterable-based batch
  runs `execute()` with no worker record and a null
  `BatchableContext.getChildJobId()`. The parent `BatchApex` job keeps a null
  `LastProcessedOffset`.
- **A legacy named credential's `<endpoint>` is loaded from its metadata.** A
  credential declaring `<endpoint>` directly, rather than a `Url` entry in
  `namedCredentialParameters`, was stored with an empty endpoint and queried
  back as null, and its principal type, protocol, and callout status were
  dropped. A `SecuredEndpoint` credential keeps its URL in a `Url` parameter and
  Salesforce leaves `Endpoint` null for that shape, so a `Url` parameter no
  longer populates `Endpoint`; `callout:` references still resolve. External
  credential metadata that a deploy would refuse is now rejected at load: a
  `Basic`, `Custom`, or `Jwt` credential may not carry `AuthParameter` entries.
- **`SObjectType.newSObject(recordTypeId, true)` populates the right fields for
  `User`.** The blanket heuristic defaulted every checkbox to false and applied
  all configured picklist and locale defaults, which is correct for objects like
  `Contact`, `Account`, and `Task` but wrong for `User`: sfapex populates
  only a small subset, with `IsActive` defaulting to true, and applies
  none of the configured picklist or locale defaults. In a multi-currency org,
  the currency fields are populated with the org default currency.
- **`DescribeSObjectResult.getLabelPlural()` is correct for a standard object
  extended from source.** An SFDX source directory that adds only a custom field
  to a standard object carries no `object-meta.xml`, and the schema merge
  overwrote the builtin object's label and plural label with the empty imported
  defaults, so the plural label came back as the singular. The merge now keeps
  the existing values when the imported ones are empty or merely the object
  name. Every builtin and optional-feature object declares its plural label on
  its own definition.
- **`ProductRelatedComponent` inserts are validated and `Name` is
  auto-numbered.** The parent `Product2` must be a bundle, set, or base product;
  a parent with no `Type` is rejected with `INVALID_INPUT` before any
  required-field check. When the parent is a static bundle, `IsDefaultComponent`
  must be true or the insert fails with `INVALID_INPUT` on that field. `Name` is
  a `PRC-000000000` auto-number so inserts succeed without one, and
  `ParentProductRole` and `ChildProductRole` are non-creatable and
  non-updateable, so assigning them is a compile-time "Field is not writeable"
  error.
- **`CartExtension.CartTestUtil.createCart()` exposes the `CartItem` it
  inserts.** `cart.getCartItems()` came back empty, so calculators that iterate
  cart items saw nothing; the returned cart now carries a `CartItem` (name 'My
  Cart Item', type Product) for the row it inserts. Two related defects in the
  parent/child model are fixed: child items were eagerly flattened into the
  cart's top-level item list so `getCartItems()` counted them, and
  `setParentCartItem` appended the child to the parent's `childCartItems` so
  `getCartItems(true)` included a child that was only parented. Child membership
  now comes from `getChildCartItems().add(...)`, and both `getCartItems()`
  overloads return top-level items only for an item that is parented but not
  added to any collection.
- **License management moved to `portal.aertest.com`.** The license management
  URL in the CLI, the license action hints, and the extension's
  subscription-management prompts changed from `kingdom.octoberswimmer.com`.

## v1.2.25 — 2026-07-24

- **Records on objects with row-level sharing get their implicit owner share
  row.** Inserting a record on an object whose org-wide default calculates
  sharing (Private or Public Read Only) now creates the share row sfapex
  maintains for the record owner: `RowCause` `Owner` with `UserOrGroupId` set to
  the record's `OwnerId`. The row moves to the new owner when ownership is
  transferred and is removed with the record. Objects whose default is Public
  Read/Write or Controlled by Parent do not get an owner row. Merging two
  records that each carry an owner share no longer fails with a share
  unique-constraint violation; the surviving record keeps its own owner row.
- **Changing `OwnerId` to a user without read access fails with
  `TRANSFER_REQUIRES_READ`.** An update that transfers a record to a user who
  cannot read the object now fails with that status code and the message "The new
  owner must have read permission". A nonexistent new owner still surfaces
  `INVALID_CROSS_REFERENCE_KEY`.
- **`System.hashCode(x)` returns the same value as `x.hashCode()`.**
  `System.hashCode` used a generic algorithm that diverged from the per-type
  `hashCode()` methods: `System.hashCode('aaaaaaa')` returned `88957452289`
  instead of `-1236860927` because the accumulator was not wrapped to 32 bits,
  and `Long`, `Double`, `Date`, and `Datetime` used entirely different
  algorithms than their own `hashCode()` builtins. Both paths now share one
  implementation per type, so `System.hashCode(x) == x.hashCode()` holds and the
  result is an `Integer`. `Integer`, `Decimal`, `Boolean`, and `Id` already
  agreed and are unchanged.
- **`System.Url` accessors match `java.net.URL`/`URI` for opaque URIs and absent
  components.** `getQuery()` and `getRef()` now return null, rather than an empty
  string, when the spec has no query or fragment. Opaque URIs whose
  scheme-specific part does not begin with a slash (`mailto:`, `tel:`, `data:,x`,
  `urn:`, `file:relative`) report a null path and return the whole
  scheme-specific part from `getFile()`, with empty host and authority;
  hierarchical URLs keep the escaped path and append the query to the file.
  A bare `scheme:` with no scheme-specific part is rejected with
  `System.StringException` "Expected scheme-specific part at index N: <spec>".
  `getProtocol()` and `toExternalForm()`/`toString()` preserve the authored case
  and percent-encoding instead of lowercasing the scheme and normalizing the
  spec.
- **Omni-Channel standard objects are modeled.** A relationship query whose
  lookup targeted `ServiceChannel` failed at runtime with "no such table:
  ServiceChannel". `ServiceChannel`, `QueueRoutingConfig`, `AgentWork`,
  `PendingServiceRouting`, `ServiceChannelStatus`, and `ServicePresenceStatus`
  are now available, enabled by `OmniChannelSettings.enableOmniChannel` and
  auto-enabled when Apex references one of them, the same way Person Accounts is
  detected. `AgentWork` and `PendingServiceRouting` carry `CurrencyIsoCode` in
  multi-currency orgs.
- **A custom lookup pointing at a missing object is reported at load time.** A
  custom field whose `referenceTo` names an object that is not part of the
  schema is now reported when metadata loads in `test`, `exec`, and `server`,
  instead of surfacing later as a SQL error during a relationship query.
- **Dynamic SOQL reports an unknown relationship with the sfapex message.**
  Querying a relationship that does not exist now fails with "Didn't understand
  relationship 'X' in field path...", and the relationship name keeps its
  original casing in the message.

## v1.2.24 — 2026-07-24

- **Content SObject describe metadata matches sfapex.** The
  `ContentDocumentLink`, `ContentDocument`, and `ContentVersion` describe
  surfaces were corrected. `ContentDocumentLink.ContentDocumentId` and
  `LinkedEntityId` are now non-nillable, non-updateable, and required,
  `LinkedEntityId.getRelationshipOrder()` returns null (it is a polymorphic
  lookup, not master-detail), `ShareType` and `Visibility` are restricted
  picklists with their full labels, and the `CreatedById`, `CreatedDate`,
  `LastModifiedById`, and `LastModifiedDate` fields the object does not have
  are dropped. Related Content objects get the matching restricted-picklist,
  updateable, and required corrections.
- **`ContentVersion.Title` is derived from `PathOnClient` when omitted on
  insert.** Inserting a `ContentVersion` with only `PathOnClient` and
  `VersionData` failed with `REQUIRED_FIELD_MISSING: [Title]`. sfapex
  populates `Title` from the file name, so the title is now derived from the
  last `PathOnClient` segment with its extension removed (`reports/MyReport.csv`
  yields `MyReport`); an explicit `Title` is left untouched.
- **`DescribeFieldResult.getRelationshipOrder()` reports order for standard
  master-detail fields.** It returned null for every standard master-detail
  field; it now returns `0` for a primary master-detail field and `1` for the
  secondary field on a junction object, and null for plain and polymorphic
  lookups (`OwnerId`, `CreatedById`, `LinkedEntityId`).
- **`OrderSummary.OrderedDate` and other feature-gated standard fields describe
  with their real type.** A retrieved standard-field stub with no `<type>` on a
  feature-gated object such as `OrderSummary` was left unknown-typed, so
  `OrderSummary.OrderedDate` described as `STRING` and `DAY_ONLY(OrderedDate)`
  was rejected with "only datetime fields support: DAY_ONLY". Such stubs are now
  replaced with the builtin field definition when the feature schema merges.
- **Flow formula `DATETIMEVALUE` and merge-field string conversion are fixed.**
  A record-triggered flow failed with "Illegal assignment from String to
  Datetime" because `DATETIMEVALUE` of a text expression converted to a bare
  `String`; it now yields a `Datetime`. Embedded `{!merge fields}` in
  record-create and record-update string assignments were emitted verbatim and
  never evaluated; they are now expanded and resolved. `LPAD` and `RPAD` are now
  supported in flow formulas.
- **Fields declared inline in a source-format `object-meta.xml` are imported.**
  Only decomposed `fields/*.field-meta.xml` files were read, so a `<fields>`
  element authored inline in an `object-meta.xml` was dropped, which in turn
  left the object without `CurrencyIsoCode` under MultiCurrency and failed
  dynamic SOQL with "No such column 'CurrencyIsoCode'". Inline fields are now
  imported, with decomposed entries taking precedence.
- **Custom labels with undeclared HTML entities load.** A `.labels` file
  containing an entity such as `&nbsp;` was rejected in full by the strict XML
  decoder, dropping every label in it and surfacing far away as unresolved
  `System.Label` references. Labels now parse through a tolerant parser that
  keeps bare ampersands and undeclared entities as literal text, and a labels
  parse failure is now a hard error instead of a silently swallowed warning.
- **An instance call resolves to an inherited instance method over a static
  method of the same signature.** A child class declaring a static method with
  the same signature as an inherited instance method dispatched to the static
  method, producing infinite recursion for a static convenience wrapper that
  called the inherited method on a fresh instance. A static method never shadows
  an inherited instance method, so the instance method is now invoked.
- **Feature auto-detection only fires on real SObject fields.** The
  MultiCurrency, PersonAccounts, and StateAndCountryPicklist auto-detectors
  matched their standard field names (`CurrencyIsoCode`, `IsPersonAccount`, the
  `*StateCode`/`*CountryCode` pattern) against any member with that name, so an
  ordinary Apex class with a `String CurrencyIsoCode` member wrongly enabled
  MultiCurrency and made `UserInfo.isMultiCurrencyOrganization()` return true.
  Detection now counts a match only when the target resolves to an SObject.
- **ConnectApi output representations inherit `success` and `errors`.** Order
  Management output representations such as `OrderSummaryOutputRepresentation`
  inherit `success` and `errors` from the abstract
  `ConnectApi.BaseOutputRepresentation`, but the flat stubs omitted them, so
  field access, value equality, and `createOrderSummary` did not expose them.
  Inherited fields now resolve at runtime and participate in `equals`,
  `hashCode`, and `toString`. `ConnectApi.UserProfiles.getPhoto` now returns the
  running user's photo for an existing user, throwing `NotFoundException` only
  when no matching `User` exists.
- **Auto-generated `User.CommunityNickname` is truncated to the field length.**
  Inserting a `User` without a `CommunityNickname` could auto-generate a
  nickname longer than 40 characters and fail with `STRING_TOO_LONG`. The
  auto-generated value is now truncated to the field length; an explicitly
  supplied value still fails validation when too long.
- **Two-argument `Type.forName` accepts an already-qualified class name.** In a
  namespaced org, `Class.class.getName()` returns `ns.Outer`, and Salesforce
  resolves `Type.forName(ns, 'ns.Outer')`; aer prepended the namespace
  unconditionally, producing an unresolvable `ns.ns.Outer` lookup and returning
  null. The class name is now used as-is when it already begins with the given
  namespace.
- **`Id.valueOf(String, Boolean)` restore-casing overload is supported.** The
  two-argument overload restores the canonical casing of an 18-character id
  from its checksum suffix when `restoreCasing` is true (a 15-character id
  throws `StringException`), and both the one-argument and
  `restoreCasing=false` forms now validate that an 18-character id carries the
  correct checksum suffix, throwing `StringException "Invalid id: ..."` on
  a mismatch.
- **`WebCart` and `CartItem` summary totals are maintained as platform
  roll-ups.** A freshly inserted `WebCart` returned null for `GrandTotalAmount`,
  `TotalAmount`, `TotalProductAmount`, and the other summary fields, so code like
  `Math.roundToLong(cart.GrandTotalAmount * 100)` threw a null pointer
  exception. A fresh cart now reads 0 for every total and count, and the
  summaries roll up from `CartItem`, `CartTax`, `CartItemPriceAdjustment`,
  `WebCartAdjustmentGroup`, and `CartValidationOutput` children sfapex does;
  none of the fields is writeable.
- **A class that extends its own inner class no longer hangs on method
  resolution.** `public class Outer extends Outer.AbstractType` sent method
  resolution for any method not declared in user code (such as the implicit
  `clone()`) into an infinite loop until the recursion guard tripped. Resolution
  now falls through to the default handlers instead.
- **`CartExtension.CartTestUtil.createCart` persists its records.** `createCart`
  built only an in-memory cart with a synthetic id, so it was invisible to SOQL
  and unusable as a lookup target; a `ProcessException` insert referencing
  `cart.getId()` failed the foreign-key check. It now inserts the record set
  sfapex creates (Account, Contact, WebStore, WebCart, CartDeliveryGroup, and
  CartItem), so the records are queryable and referenceable.
  `ProcessException.AttachedToId` now accepts every custom object (including
  custom settings) as a reference target.
- **Generic and array class-literal type names match sfapex.**
  `Map<Id, Id>.class` stringified as `"Map<Id, Id>"` from `toString()`/`getName()`
  while `Type.forName` and `StubProvider` paramTypes produced `"Map<Id,Id>"`, so
  the two spellings were unequal within a run and broke name-based stub matching
  in mocking frameworks. Class-literal names are now canonicalized everywhere:
  `Account[].class` yields `"List<Account>"` and generic parameter lists drop the
  space after commas. Chaining a plain `.` off a generic or array class literal
  (`Map<Id,Id>.class.toString()`) is now rejected as the compile error it is on
  sfapex.
- **`String.replace` with an empty target and `Map.keySet()` mutation match
  sfapex.** `String.replace('', repl)` now inserts the replacement between
  every character and at both ends instead of returning the string unchanged. A
  `Set` returned by `Map.keySet()` is now a read-only view: `add()` and a
  non-empty `addAll()` throw `System.FinalException: Collection is read-only`,
  an empty `addAll()` is a no-op, and removals still mutate the backing map.

## v1.2.23 — 2026-07-22

- **Record-triggered flow Get Records lookups bulkify for non-Id keys, custom
  lookups, and formulas.** A Get Records filtered by a record-dependent value
  ran one SOQL query per record unless the filter bound a field that looked
  like an Id: custom lookup (`__c`) fields, text and other non-Id fields, and
  flow-formula binds each disqualified the query, so a bulk insert of 200
  records threw `LimitException` ("Too many SOQL queries: 101"). All of these
  shapes now run one query per batch. A Get Records filtered by a per-record
  formula also no longer fails at runtime with "undefined bind variable", and
  fields referenced only from assignments, record creates, formula expressions,
  or text templates are now included in the generated query instead of throwing
  `SObjectException` when read.
- **Flow email alerts send through the alert's actual template.** An email
  alert action resolved its template at conversion time, guessing a template
  name when the lookup missed, and sent anyway when resolution failed, a
  bodyless send that failed the triggering DML with `REQUIRED_FIELD_MISSING`.
  The alert and its template now resolve at send time, a missing alert or
  template throws an `EmailException` naming it, and alert sends no longer
  count against `Limits.getEmailInvocations()`.
- **Flow deletes of null collections fault the way Salesforce does.** A Delete
  Records element whose input resolved to a null collection (an empty Get
  Records result) failed with the internal error "addAll() argument must be a
  List or Set, got null" in bulk runs. The save now fails with a
  `CANNOT_EXECUTE_FLOW_TRIGGER` `DmlException` that names the flow label and
  carries the fault detail "No records in Salesforce match your delete
  criteria.".
- **`SUBSTITUTE`, `DATEVALUE`, and old-variable `PRIORVALUE` convert in flow
  formulas.** A flow formula calling `SUBSTITUTE` or `DATEVALUE` failed to
  convert, so the raw formula text leaked into the generated Apex, assigning
  the literal formula string instead of the evaluated value. A Process Builder
  rule formula written against the old-record variable, such as
  `PRIORVALUE({!myVariable_old.Field__c})`, failed to convert and the process
  was silently skipped. All three now convert, and `PRIORVALUE` returns the
  current value while a record is being created.
- **Process Builder old-record references no longer fail on insert.** A
  process referencing its old-record variable threw a `NullPointerException`
  on every insert, even when the decision guarded with an `IsNull` check.
  The old-record variable now binds to null on insert, so the guard
  short-circuits and inserts proceed (reference-verified).
- **Process Builder flow invocations convert.** An action call with
  `actionType="flow"`, how Process Builder invokes a flow, aborted the whole
  flow's conversion with a "requires manual implementation" warning. It now
  converts to a subflow launch through `Flow.Interview`, passing the action's
  input parameters and assigning its output parameters.
- **Datetime values assigned to Date targets in flows convert implicitly.**
  Assigning a datetime value such as `$Flow.CurrentDateTime` to a Date
  variable or Date field aborted the run with "Illegal assignment from
  Datetime to Date". The Flow runtime's implicit conversion is now applied,
  so the assignment truncates to the date.
- **Tag objects are supported.** `TagDefinition`, the standard tag objects
  (`AccountTag`, `ContactTag`, …), and a `__Tag` companion for each custom
  object are registered when `SearchSettings` enables personal or public
  tagging, or automatically when Apex references a tag entity. Each taggable
  object gains a `Tags` child relationship.
- **Duplicate-matching engine queries no longer consume the SOQL query
  limit.** `Datacloud.FindDuplicates` runs the duplicate-matching engine, not
  SOQL, so it does not count toward `Limits.getQueries()` in Salesforce. aer
  counted each internal candidate scan, so per-record duplicate checks in a
  loop threw "Too many SOQL queries: 101" on bulk operations that succeed in a
  real org. The engine's internal queries are now excluded.
- **Namespace qualifiers written with trailing double underscores are
  accepted.** sfapex treats `pkg__.PkgApi` and `pkg__.Log.Logger` as
  equivalent to `pkg.PkgApi` and `pkg.Log.Logger`; aer rejected the `ns__.`
  spelling. It is now accepted in every type and expression position, and in
  single-argument `Type.forName`. The two-argument form `Type.forName('ns__',
  name)` returns null, matching sfapex.
- **Field describe metadata matches sfapex for length, nillability, and
  reference targets.** `DescribeFieldResult` reported a 255-character length
  for every field type; now only text-like fields report a length, id and
  reference fields report 18, and boolean, numeric, date, datetime, time, and
  base64 fields report zero. `isNillable()` applies the same rule as
  `FieldDefinition` (boolean and standard system fields are never nillable),
  `getDefaultValue()` returns false for booleans without an explicit default,
  base64 fields report the `BASE64BINARY` SOAP type and are no longer
  groupable or sortable, and integer fields report zero precision.
  `getReferenceTo()` and `isNamePointing()` now report lookup targets aer does
  not store against (for example `Attachment.OwnerId` reports `Calendar` and
  `User`). Attachment field metadata is updated to the org-verified shape.
- **Compound `Name` fields describe as not calculated and not nillable.**
  aer derives `Contact.Name`, `Lead.Name`, and `User.Name` internally, and
  that leaked into describe results: `isCalculated()` returned true,
  `getCalculatedFormula()` returned an internal formula, and `isNillable()`
  returned true. All three now match sfapex.
- **Share objects appear in their parent's child relationships.** The
  describe of a sharing-enabled custom object omitted the `Shares` child
  relationship, so `getChildRelationships()` had no entry for `Foo__Share`
  even though the share object existed. The relationship is now registered
  wherever a share object is synthesized.
- **`Case.CaseMilestones` subqueries resolve.** A child-relationship subquery
  like `(SELECT Id FROM CaseMilestones)` failed type-check with
  "'CaseMilestones' is not a valid child relationship of 'Case'" even though
  direct `CaseMilestone` queries worked. The standard relationship is now
  registered.
- **`ProductCategoryProduct` inserts derive `ProductToCategory` and
  `CatalogId`.** Inserting with only `ProductCategoryId` and `ProductId`
  failed with `REQUIRED_FIELD_MISSING` for `ProductToCategory`, but the
  platform derives that field. Both platform-managed fields are now populated
  on insert and rejected as caller-supplied values.
- **Subscriber parent relationships resolve in child subqueries from package
  code.** Package code running a subquery that traverses a subscriber-authored
  parent relationship, such as `(SELECT Entry__r.Note__c FROM Receipts__r)`,
  failed with "field not found" because the executing namespace was applied
  unconditionally to the derived foreign-key field. Subscriber relationships
  reached from package code now resolve.
- **Activity formulas referencing sibling-only fields evaluate as blank.**
  Activity custom formula fields exist on both Task and Event, but may
  reference standard fields present on only one, such as Task's
  `CallDurationInSeconds`. Salesforce evaluates such a reference as blank on
  the other object; aer failed with "field 'CallDurationInSeconds' not found"
  when computing the formula on Event. The reference now evaluates as the
  typed blank.

## v1.2.22 — 2026-07-21

- **`Datacloud.FindDuplicates` finds matching records through matching rules.**
  `findDuplicates()` and `findDuplicatesByIds()` returned the per-rule result
  skeleton but never ran any matching, so every `MatchResult` reported zero
  match records. Each duplicate rule's matching rules are now evaluated against
  the stored records: custom matching rules load from `MatchingRule` metadata
  alongside the existing `DuplicateRule` loading, and the standard Account,
  Contact, and Lead matching rules are built in. Values compare normalized,
  blank values match only when the rule item sets `MatchBlanks`, item
  results combine through the rule's boolean filter, and single-hop parent
  paths such as the Contact rule's `Account.Name` resolve. Fuzzy matching is
  not yet implemented.
- **`Account.OrderSummaries` and `ConnectApi.OrderSummaryCreation` are
  supported.** Querying the `Account.OrderSummaries` child relationship failed
  with "Didn't understand relationship 'OrderSummaries'"; it now resolves in
  inline subqueries, dynamic SOQL, and `getChildRelationships()`.
  `ConnectApi.OrderSummaryCreation.createOrderSummary` threw "method not
  found"; it now validates the order Id and creates the `OrderSummary` from the
  original Order, and throws the standard `UnsupportedOperationException` in
  data-siloed tests, matching sfapex. `OrderSummary.OriginalOrderId` is
  non-writeable, so assigning it is rejected at compile time.
- **`Database.convertLead` returns error results when `allOrNone` is false.**
  The `allOrNone` argument was ignored, so every row-level validation failure
  (invalid `convertedStatus`, missing `leadId`, missing `convertedStatus`,
  `contactId` without `accountId`) threw a `DmlException`. With `allOrNone`
  false, a failing row now yields a `Database.LeadConvertResult` with
  `isSuccess()` false and a populated `Database.Error`, and the remaining rows
  still convert. With `allOrNone` true, the exception message now reports the
  failing row's actual index instead of row 0.
- **Person Account inserts fall back to the org default record type.**
  Inserting an Account with person fields set and no explicit `RecordTypeId`
  failed with `REQUIRED_FIELD_MISSING` on `Name` unless the running user's
  profile named a default person account record type. The org default person
  account record type is now used as a fallback, matching sfapex's
  auto-assignment, so the insert succeeds and `Name` derives from
  `FirstName`/`LastName`. Non-string person fields (for example
  `PersonBirthdate` alone) now mark an insert as a person account, and
  `IsPersonAccount` reads correctly on a queried Account whose query selected
  neither `PersonContactId` nor `RecordTypeId`.
- **Flow conversions handle null values the way the Flow runtime does.**
  Decision conditions using Starts With, Ends With, or Contains evaluated
  against a null left operand threw a `NullPointerException`; they now
  evaluate to false. An Assignment element that sets a field on a
  still-null record or Apex-type variable implicitly creates the value
  instead of null-pointering. Flows whose only Update Records element targets
  a variable or filter criteria no longer fail on the handler-class path.
- **`$UserRole` and `$Label` flow references are supported.** `$UserRole.Id`
  converts to `UserInfo.getUserRoleId()`, and other `$UserRole` fields resolve
  through the running user's role, yielding null for users without a role.
  `$Label.<name>` references convert to `System.Label.<name>`; previously
  generated flow classes failed type checking with "Variable does not exist:
  $Label".
- **Constant-filter flow lookups are bulkified.** A Get Records element with
  only constant filters ran its query once per record in a trigger batch,
  exceeding the query limit on large batches. Such lookups now run once per
  batch.
- **`Flow.Interview` version selection follows the Manage Flow permission.** A
  user with the permission, including test-context administrators, runs the
  latest version of a flow started from Apex even when it is Draft or Obsolete.
  A user without it gets a `System.FlowException` whose message is the
  lowercased flow name. Record-triggered automation still runs only Active
  versions.
- **Dangling lookup values raise the sfapex error.** Setting a lookup to a
  nonexistent Id surfaced an internal foreign-key constraint error listing the
  statement's id values. `OwnerId` now raises `INVALID_CROSS_REFERENCE_KEY`
  with "invalid cross reference id", and any other lookup raises
  `INSUFFICIENT_ACCESS_ON_CROSS_REFERENCE_ENTITY` with the 15-character form of
  the dangling id, matching sfapex.
- **Deleted rows are excluded from custom setting `getValues`.** A List custom
  setting row deleted in the same transaction was still returned by
  `getValues(name)` with its stale field values. It now returns null once the
  row is deleted, matching sfapex, for both hierarchy and list settings.
- **`Crypto.sign` accepts RSA keys smaller than 1024 bits.** Signing with a
  512-bit key threw "crypto/rsa: 512-bit keys are insecure". sfapex imposes
  no minimum RSA key size, so the same key now signs successfully.
- **`ProcessException` inserts no longer require `StatusCategory`.** The field
  was marked required even though the platform derives it and rejects writes,
  so inserts failed with `REQUIRED_FIELD_MISSING`. It is now read-only and
  derived from `Status` on insert:  New and Triaged map to ACTIVE, Paused and
  Ignored to INACTIVE, Resolved and Voided to RESOLVED, and `Status` defaults
  to New. Describe results for `Status`, `Category`, `Priority`, `Severity`,
  and `OwnerId` now report defaulted-on-create the way sfapex does, and
  the `Category` picklist gains the missing Commerce Delivery Service value.
- **External Apps Experience Cloud licenses are portal-capable.** Inserting a
  User linked to a Contact under a profile with the "External Apps" or
  "External Apps Login" license failed with `FIELD_INTEGRITY_EXCEPTION` "only
  portal users can be associated to a contact". These licenses now resolve to
  the `CspLitePortal` user type and receive the implicit standard-object read
  access community users get. The full set of standard `UserLicense` records is
  now seeded, so a profile referencing any standard license resolves its
  `UserLicenseId`.
- **Namespaced stdlib method results match constructor overloads.** Passing a
  namespaced stdlib method's result directly to a constructor, such as
  `cart.getCartItems().iterator()` into a constructor taking
  `Iterator<CartExtension.CartItem>`, threw `IllegalArgumentException: no
  matching constructor found`, because stdlib return types referenced sibling
  types without their namespace. Those return types are now fully qualified, so
  the overload resolves.
- **Storage builds succeed for polymorphic lookups with very many targets.** A
  polymorphic lookup referencing every object in a large org — for example
  `ContentDocumentLink.LinkedEntityId` with over a thousand targets — failed
  storage creation with "Expression tree is too large". The validation now
  nests within the expression-depth limit regardless of target count.
- **Blank fields in bootstrap databases load as null.** A Salesforce export
  writes blank fields as empty strings. An empty string in a numeric column
  aborted storage-template creation outright, and an empty string in a lookup
  column became a dangling reference that failed foreign-key validation. Both
  now store null, so exported data with blank numeric and lookup fields loads.

## v1.2.21 — 2026-07-20

- **Parenthesized `when` labels in `switch` statements now match.** A `switch`
  `when` clause whose literal was wrapped in parentheses — `when ('Alpha')` —
  never equaled the switch argument, so control always fell through to
  `when else`. Parenthesized labels are now unwrapped and interpreted the same
  as bareword ones, including enum-constant handling, so they match as written.
- **`Database.convertLead` applies picklist defaults before validation rules.**
  The Account, Contact, Opportunity, and OpportunityContactRole records created
  during lead conversion did not have field defaults materialized until the
  storage write, after validation rules had already run. A validation rule
  guarding a field with a picklist default therefore failed on the blank value
  even though the default would have satisfied it. Defaults (and formula fields)
  are now applied before before-save flows, triggers, and validation rules,
  matching the ordering of a plain `insert`.

## v1.2.20 — 2026-07-20

- **`System.debug` and collection rendering honor a `toString` override.**
  `System.debug` formatted object arguments with the default `Type:[fields]`
  representation, ignoring a user-defined `toString()`. It now renders through
  the override, matching `String.valueOf`, and objects nested in collections,
  assertion messages, and string concatenation do too. Map rendering is shared
  across `String.valueOf`, `System.debug`, and `toString`:  numeric key
  ordering, `toString` dispatch on keys and values, and truncation at ten
  entries like `List` and `Set`. A `toString` override that returns null renders
  as an empty string from `String.valueOf`, an empty line from `System.debug`,
  and `null` when embedded in a collection or concatenation.
- **A static property getter sees writes made by a setter it triggers.** When a
  static property getter called a method that assigned the property, the getter
  returned a stale snapshot and, on exit, wrote that stale value back over the
  setter's write, so a lazy-init getter that delegated population to a helper
  returned and persisted the pre-helper value. The getter now observes the
  setter's write and returns the populated value.
- **Bare child relationships resolve in package dynamic queries.** Package code
  running a subscriber-supplied dynamic query failed to resolve an unnamespaced
  child relationship in a subquery, throwing `child relationship 'Bars__r' not
  found on object 'Foo__c'`. Bare child relationship names now resolve the same
  way as bare object and field references for both direct and nested subqueries.
- **`Lead.LastTransferDate` is supported.** The standard `Lead.LastTransferDate`
  field is now available. Salesforce describes it as a `DATE` while storing the
  full transfer datetime; it is stamped on insert and restamped when `OwnerId`
  changes.
- **Enums print as their constant name.** `System.debug` and other formatted
  output now print an enum value as its declared constant name in its declared
  casing, instead of an internal representation.
- **`isNillable()` reports the correct value for standard name fields.** When an
  org's source included a partial standard-field stub (for example an
  `Account/fields/Name.field-meta.xml` carrying only `trackFeedHistory`), the
  importer forced `Nillable` to true for every name field, so
  `Account.Name.getDescribe().isNillable()` returned true without Person
  Accounts. The builtin field's explicit value is now preserved, in both SFDX and
  MDAPI object formats.
- **Watch-mode errors explain inotify watch exhaustion.** When the per-user
  inotify watch budget (`fs.inotify.max_user_watches`) is exhausted, watch setup
  in test and server watch modes failed with a misleading "no space left on
  device". The error now names the limit, and a setup that ends with zero watches
  reports the underlying cause instead of a bare "no watchable paths found".

## v1.2.19 — 2026-07-18

- **`TEXT()` of a null value is an empty string in formulas.** A formula
  comparison like `TEXT(Picklist__c) <> 'literal'` no longer evaluates to null
  when the picklist is blank, so an `IF` guarded on it takes the expected branch.
  `TEXT(NULL)` is treated as an empty string, matching Salesforce.
- **Method calls work on the enum result of a binary expression.** Calling a
  method on the result of the null-coalescing operator with enum operands — for
  example `(q ?? System.Quiddity.UNDEFINED).name()` — now works, for both system
  and user-defined enums.
- **`MFLOOR` and `MCEILING` formula functions are supported.** Importing a
  package whose formula field uses `MFLOOR` or `MCEILING` no longer fails. `FLOOR`
  and `CEILING` are also corrected for negative inputs — `FLOOR` rounds toward
  zero and `CEILING` away from zero, while `MFLOOR` and `MCEILING` round toward
  negative and positive infinity — and `CEILING(3.0)` no longer returns 4.
- **Inserting a share for an inactive user is rejected.** A manual share whose
  `UserOrGroupId` is a deactivated user now fails with `INACTIVE_OWNER_OR_USER`
  and creates no row, matching sfapex. Group grantees are unaffected.
- **Picklist defaults are scoped to the record's record type.** A picklist
  default configured on one record type no longer leaks onto records inserted
  with a different record type; each record now receives only the default defined
  for its own record type.
- **`Database.convertLead` copies the standard field mappings.** Lead conversion
  now carries `LeadSource`, `Salutation`, `MobilePhone`, `Fax`, `DoNotCall`, and
  `HasOptedOutOfEmail` onto the new Contact, `Fax` and `Rating` onto the new
  Account, and `LeadSource` onto the new Opportunity, instead of copying only a
  subset.
- **`Contract.EndDate` is auto-calculated from `StartDate` and `ContractTerm`.**
  `EndDate` is now computed on insert and update as the day before `StartDate`
  advanced by `ContractTerm` months, capping at the end of shorter months, across
  all DML paths and before triggers. It is read-only, so describe results and
  dynamic `put()` reject writes, and a zero or negative `ContractTerm` is rejected
  with `FIELD_INTEGRITY_EXCEPTION`.
- **`Datacloud.FindDuplicates` returns a result per active rule.**
  `findDuplicates()` and `findDuplicatesByIds()` returned an empty
  `getDuplicateResults()` list, so the standard `[0]` access threw. They now emit
  one `DuplicateResult` per active duplicate rule that applies to the object, even
  when nothing matches, using `DuplicateRule` metadata loaded from source.
- **Inactive record types appear in `getRecordTypeInfos` describe results.** All
  four `getRecordTypeInfos*()` collections now include inactive record types, with
  `RecordTypeInfo.isActive()` reflecting the real state and `isAvailable()`
  returning false for them, instead of filtering them out and always reporting
  active. An object whose only custom record type is inactive still reports Master
  as the default.
- **`SObject.put` throws on a type-mismatched value into a text field.**
  `put(field, value)` silently stringified a numeric, Boolean, Date/Datetime/Time,
  or Blob value put into a text or picklist field. It now raises
  `System.SObjectException` ("Illegal assignment …") at assignment time, matching
  Salesforce; String, Id, and null remain accepted.

## v1.2.18 — 2026-07-17

- **Bare custom names resolve only through the executing namespace.** A bare
  custom object or field reference resolves the way the sfapex compiler does:
  first through the executing code's namespace, then by its exact authored name.
  Subscriber code never binds a bare name to a namespaced object or field, a
  managed object's namespace never substitutes for the caller's, and a namespaced
  reference is never matched by stripping its prefix. A subscriber field added to
  a managed object now stays distinct from the package's same-base-name field
  instead of colliding unpredictably. A `<packageVersions>` dependency pin in a
  class or trigger meta.xml no longer places the trigger in the pinned namespace,
  and namespace qualifiers written in a different case than the declared namespace
  still resolve.
- **Packaged objects, workflow rules, and flows resolve correctly from a
  `.pkg`.** Custom SObject types, workflow rules, and flows shipped in a package
  are now namespaced at load like the rest of the package. Previously a
  subscriber call passing a packaged list type could pick the wrong method
  overload, packaged workflow rules silently never fired, and flows were not
  saved in packages at all. Overload resolution, workflow rules, and
  record-triggered flows from packages now all behave as they do from source.
- **Packaged flows run under `aer server` and `Flow.Interview`.** Flows shipped
  in a `.pkg` previously ran under `aer test` and `aer exec` but not under `aer
  server`, and packaged flows could not be started with `Flow.Interview` from
  `aer exec` or `aer server`. Packaged record-triggered and autolaunched flows
  now run in all three commands.
- **Record-triggered flows run under `aer exec`.** `aer exec` did not collect
  flow files from `--path` directories, so DML in anonymous Apex skipped active
  record-triggered flows that `aer test` and `aer server` apply. Exec now runs
  them, and `aer exec --debug` loads the same triggers, flows, and packages as
  `aer exec`.
- **Deleting a share record now removes it.** Deleting a share previously left
  a soft-deleted row behind, so it still showed up in `ALL ROWS` queries and
  blocked re-sharing the same record with the same user. Shares are now
  hard-deleted, matching sfapex, so re-sharing works.
- **Protected custom metadata records are visible to the right namespaces.** A
  protected record authored in the subscriber org (not shipped in a package) is
  now visible to code in every namespace, and package code can read
  subscriber-authored protected records of its own type. A caching issue that
  could restore custom metadata records with protection stripped is also fixed.
- **Uppercase escape letters are accepted in string literals.** Salesforce
  accepts escape letters case-insensitively — `\F` is a form feed like `\f`, and
  `\U0041` decodes like `A`. aer previously rejected the uppercase forms with a
  token error; both cases now work.

## v1.2.17 — 2026-07-16

- **Method calls select the most specific matching overload.** Overload
  resolution bound a call to the first candidate whose parameters were merely
  compatible with the arguments, so a generic `List<SObject>` overload declared
  before a concrete-list one won purely by declaration order, and a generic
  overload that re-dispatched a concrete list value could select itself and
  recurse forever. Every matching candidate is now collected and the most
  specific applicable one is chosen.
- **Overriding a non-virtual, non-abstract method is rejected.** sfapex
  reports "Non-virtual, non-abstract methods cannot be overridden"; the type
  checker now does too. A method declared with `override` alone is non-virtual,
  so a further subclass cannot override it; only a `virtual override` method
  stays overridable.
- **Cross-namespace calls work through builtin global interfaces.** A class
  implementing a builtin system interface such as `System.Callable` or
  `System.Comparable` with a `public` (not `global`) method was rejected when
  code in another namespace invoked the method through the interface
  reference. Dispatch through an interface the caller can see now forwards to
  the concrete implementation regardless of the implementation method's own
  access modifier, matching the existing behavior for user-defined interfaces.
- **Protected custom metadata records are visible only to their owning
  namespace.** The `<protected>` flag on a custom metadata record was dropped
  at load time, so a managed package's protected records leaked to subscriber
  code and other namespaces. The flag is now carried through loading, package
  creation, and distribution, and SOQL (including subqueries), `getAll()`, and
  `getInstance()` all hide protected records from code executing outside the
  namespace that owns them. Unprotected records, and protected records owned
  by the executing namespace, remain visible.
- **Custom metadata records are no longer returned twice.** Because packaged
  custom metadata records carry no Id, reopening a persistent
  `--db` database inserted every record again, and merging the same package or
  source more than once accumulated duplicates, including when `aer exec`
  loads a package alongside a source directory that contributes schema
  metadata. Records are now upserted on their natural key and deduplicated
  during package schema merge. `aer package list` also gains a Custom Metadata
  Records section grouped by type, so the records a package carries are
  visible.
- **A custom metadata record's `NamespacePrefix` reflects the package that
  owns the record, not the type.** A record of `pkg__Config__mdt` authored in
  unpackaged subscriber code keeps a blank `NamespacePrefix` and an unprefixed
  `QualifiedApiName`; only records shipped inside a package take that
  package's namespace. Source directories loaded with `@ns[:pkg]` also behave
  like packages for custom metadata: a package directory shipping only
  metadata and no Apex forms a synthetic package so its types are visible to
  other namespaces, sibling directories sharing a namespace merge their
  schemas, and a subscriber loaded alongside a namespaced package stays in the
  default namespace.
- **Packaged permission records take their package's namespace, and same-name
  rows stay distinct.** A `CustomPermission`, `PermissionSet`, or
  `PermissionSetGroup` shipped in a package received a blank
  `NamespacePrefix`, so consumers filtering by namespace could not find it,
  and a subscriber and a package shipping a permission of the same base name
  collapsed into a single row. Both now match sfapex: packaged records are
  namespaced, same-name rows remain distinct, and reopening a persistent
  database does not duplicate them.
- **Lookup filter field references resolve to namespaced columns.** A lookup
  filter referencing a target field by its unprefixed API name generated SQL
  against a column that does not exist when the installed column carries a
  namespace prefix. Filter field references are now resolved to the
  canonical schema name at schema finalization, walking relationship segments
  to the object the final field lives on.
- **Person Accounts auto-enable when lookup filters reference person contact
  fields.** A lookup filter referencing a person contact field such as
  `Status__pc` only exists in orgs with Person Accounts enabled, but schema
  preparation ran before Person Accounts auto-detection and failed with an
  unresolvable lookup filter reference. A `__pc` or `__pr` segment in any
  lookup filter now enables Person Accounts before filter references are
  resolved, including the runtime behavior such as deriving `Account.Name` on
  person account inserts.
- **`Account.Name` describes as nillable when Person Accounts is enabled.**
  With Person Accounts, an Account's Name is derived from the person's
  FirstName and LastName, so its describe reports `isNillable` as true.
  Business account inserts still require Name.
- **Tracked field history values are stored as text.** Field history wrote a
  tracked field's raw stored value into the History object's `OldValue` and
  `NewValue` columns, so a tracked checkbox read back as "0"/"1" instead of
  "false"/"true" and a strict backend rejected the value outright. Tracked
  values are now rendered to their Apex string form, matching sfapex; null
  stays null.
- **Stale `FieldPermissions` rows no longer abort permission loading.** A
  persistent database can hold `FieldPermissions` rows for objects or fields
  no longer in the schema, such as a history object after field history
  tracking is turned off. Loading the current user's permissions failed
  outright on such a row; the stale rows are now skipped and valid permissions
  continue to load.
- **Locals stay in scope across multi-line `if` conditions ending in `break`
  or `continue`.** A loop-local referenced on the continuation line of a
  multi-line parenthesized condition (spanning `||` or `&&`) whose body was a
  `break` or `continue` was falsely reported as "Variable does not exist" when
  the `if` was the last statement in its block.
- **Debugger step-over no longer stops inside implicitly invoked method
  bodies.** A user-defined `hashCode` or `equals` invoked by a Map or Set
  operation, a static property accessor, and a class's static initializer all
  ran at the caller's call-stack depth, so step-over stopped inside them
  instead of stepping past the triggering statement. Each now runs one frame
  deeper, and exceptions thrown from these bodies carry the caller frame in
  their stack trace.

## v1.2.16 — 2026-07-15

- **String operands compare case-insensitively in relational operators.** `<`,
  `>`, `<=`, and `>=` now compare Strings with sfapex's case-insensitive
  ICU (English) collation, the same ordering applied to SOQL string columns,
  instead of raw code-point order, so `'a' < 'A'` is false rather than true and
  non-letter characters order by collation weight.
- **`String.escapeSingleQuotes` escapes backslashes.** The method added a
  backslash only before single quotes, leaving existing backslashes untouched;
  it now adds a backslash before each backslash as well.
- **`String.replaceAll` and `String.replaceFirst` support the full Java regex
  syntax.**
- **Custom-object share record DML matches sfapex.** Because all custom-object
  shares carry the `02c` key prefix, their DML forms behave specially.
  `Database.delete` on a share SObject accepts any share type as long as its id
  is filled in and deletes the row that id identifies, instead of failing with
  `INVALID_CROSS_REFERENCE_KEY`. Updating a share through a different share
  type that carries the row's id is rejected as `INVALID_ID_FIELD`, "invalid
  record id". Undelete fails on the entity type before any recycle-bin lookup:
  the keyword and `Database.undelete` SObject forms throw
  `CANNOT_INSERT_UPDATE_ACTIVATE_ENTITY`, "Entity type is not undeletable", and
  `Database.undelete` on a raw id throws the same "Invalid id" `TypeException`
  as `Database.delete`.
- **Deleting a parent purges its soft-deleted restrict-lookup children.** A
  delete blocked by `deleteConstraint=Restrict` children succeeds once the only
  remaining children are soft-deleted, and sfapex permanently removes those
  children from the recycle bin so they can no longer be undeleted, even after
  the parent is restored. aer left them restorable. The purge recurses into the
  children's own restrict children and applies to optional as well as required
  lookups, and to standard as well as custom parents.
- **`System.FlexQueue` move methods return a Boolean.**
  `moveJobToFront`, `moveJobToEnd`, `moveBeforeJob`, and `moveAfterJob`
  reordered the queue but evaluated to null, so `Assert.isTrue` failed after a
  successful move and `Assert.isFalse` passed by unboxing null. Each now
  returns whether the queue order actually changed.
- **A variable shadows a same-named type in a field-access chain.** When a
  variable's name case-insensitively matched a class name, the type checker
  resolved the chain as a type reference: a parameter `b` of type `A` shadowing
  a class `B` made `b.category` resolve to the nested type `B.Category` rather
  than the instance field, reporting "Comparison arguments must be compatible
  types". A local variable or parameter in scope now wins.
- **Multi-variable declarations work in property accessor bodies.** A
  declaration such as `Integer a = 0, b = 0;` in a getter or setter was scoped
  only to its own line, so a later reference in the same accessor failed with
  "Variable does not exist".
- **System-qualified builtin static methods resolve their return type.**
  `System.Date.today()` produced an empty return type where the unqualified
  `Date.today()` resolved to `Date`, so `Date d = System.Date.today() - 1` was
  rejected with "Illegal assignment from Integer to Date".

## v1.2.15 — 2026-07-14

- **Java regex property classes, whitespace, and class-set constructs.** Apex
  regular expressions use the `java.util.regex.Pattern` engine, and its
  `\p{...}`/`\P{...}` property names are now translated before compilation:
  POSIX names, `java.lang.Character` classes, Unicode general categories,
  scripts, blocks, and binary properties all resolve with Java's meaning
  instead of being rejected as "Illegal character range" or matched with the
  wrong definition. A bare script name such as `\p{Latin}` reproduces
  sfapex's `PatternSyntaxException`. The translation also covers `\h`,
  `\H`, `\v`, `\V`, and `\R`; a dot that excludes every Java line terminator
  and honors `(?s)`; Java-style octal escapes; character-class union,
  intersection, and subtraction (`[a-z&&[^aeiou]]`); the `(?U)` flag; and a
  default `\s` that includes the vertical tab. Two defects surfaced by the
  work are fixed: non-ASCII literals were corrupted by possessive-quantifier
  conversion (`Pattern.matches('café', 'café')` returned false), and full
  matches compared rune length against byte length, so some patterns failed to
  fully match multi-byte input.
- **Method definitions in anonymous Apex.** Anonymous blocks can define
  methods alongside classes, interfaces, and enums. Methods are static whether
  or not the keyword is written and are callable before their textual
  definition; top-level variable declarations hoist into shared static fields
  while declarations in nested blocks stay local; and stack frames format as
  `AnonymousBlock: line N, column 1`. `this` in a static context is rejected
  at compile time, and initializer blocks, field initializers, and property
  accessor bodies now receive the full statement checks.
- **`EntityDefinition` rows for associated entities.** Base objects have
  their boolean flags, sharing models, `PublisherId`, `DeploymentStatus`,
  URLs, `DefaultCompactLayoutId`, and durable id updated to match sfapex, and
  each emits rows for the associated History, Share, ChangeEvent, and Feed
  entities sfapex supports. Share records now use their real key prefix, and
  share tables without a dedicated prefix generate ids beginning with `02c`;
  because that prefix is shared, deleting a share by raw id throws
  a `TypeException`, matching sfapex, while typed DML and `getSObjectType()`
  still resolve. `SObjectType` gains `EnableSearch` and `DeploymentStatus`.
- **Flex queue rules for `Database.executeBatch`, with queryable
  `FlexQueueItem` rows.** Batch jobs report `Queued` until five occupy the
  batch queue and `Holding` after that, the rule now shared by
  `Database.executeBatch`, `System.scheduleBatch`, and
  `Test.enqueueBatchJobs`. `FlexQueueItem` rows are materialized for held jobs
  and reconciled as jobs are held, reordered, aborted, or start running, so
  SOQL observes the flex queue and `JobPosition` matches
  `Test.getFlexQueueOrder()`. Jobs held inside `startTest()`/`stopTest()`
  execute at `stopTest`.
- **Restrict delete constraints on lookup relationships.** Deleting a record
  still referenced through a lookup with `deleteConstraint=Restrict` fails
  with `DELETE_FAILED`, matching sfapex's message. The check runs in all
  delete paths after before-delete triggers and validates rows in
  statement order. `ChildRelationship.isRestrictedDelete()` now reports true
  for Restrict lookups loaded from explicit source paths, self-lookups with
  Restrict or Cascade are rejected at import, and field and object parse
  errors fail the load instead of silently dropping metadata.
- **`String.join` accepts any `Iterable`.** Joining a user-defined `Iterable`
  previously failed with "String.join() called with non-list argument".
- **License key and website links in CLI help and errors.** `aer --help`
  points to the main website, and the `license show` error and
  `license register --help` explain where to retrieve a key or start a trial.
- **Exceptions escaping a trigger carry the cause's full stack.** A
  `DmlException` raised by a trigger now embeds every frame from the throw
  site down to the trigger frame in `getMessage()`, instead of only the
  originating frame, and no longer leaks internal "Stack trace:" blocks or
  `file:line` prefixes. Trigger frames render as `Trigger.<ns>.<name>`,
  namespace-qualified in a namespaced org, and a trace captured while a
  trigger is on the stack stops at the trigger frame rather than descending
  into the code that performed the DML. Caller frames report the line of the
  call statement even inside a `try` block.
- **`Trigger.new` and `Trigger.old` expose no relationship data.** A parent
  relationship reads as null even when the DML'd record was queried with it,
  the null parent is traversable without throwing, and child relationships
  resolve as empty lists; the caller's own record keeps its data. Formula
  fields are still computed on `Trigger.new`, recalculated from pending values
  in before update. Workflow field-update and criteria formulas resolve
  spanning references such as `Parent__r.Field__c` against these records by
  hydrating the parent from the lookup id, evaluating to blank when the lookup
  is null instead of aborting the DML with a `DmlException`.
- **`Trigger.newMap` is null in before insert and `Trigger.oldMap` is null in
  after undelete**, in both Apex and flow trigger dispatch; aer exposed empty
  maps.
- **SObject collection conversion rules match sfapex.** Maps convert
  implicitly only from specific to generic, sets are invariant in both
  directions, and list elements keep the Id/String interchange. Casting a
  generic SObject map to a specific map, or a list to a list of a different
  SObject type, throws a `TypeException`; storing a mismatched SObject into a
  specific-element list or map throws a collection store exception. The `List`
  copy constructor requires an exactly matching element type, and
  `getSObjectType()` compiles only on SObject lists and SObject-valued maps.
- **`Database.QueryLocator` iterator dispatch.** Assigning a `QueryLocator` to
  `Iterable<SObject>` compiles, but dispatching `iterator()` through the
  `Iterable` interface, an enhanced for loop over a `QueryLocator`, or an
  `iterator()` call on an `Iterable<SObject>`-typed receiver, now raises the
  uncatchable "Internal Salesforce Error" sfapex raises, where the for
  loop previously iterated zero rows silently. Only the concrete
  `Database.QueryLocator.iterator()` binding works.
- **`Map<Object,V>` switches to identity lookup after enumeration.** Calling
  `keySet()`, `values()`, or `toString()` permanently switches an
  object-keyed map from `hashCode`/`equals` lookup to reference identity, so
  equivalent keys no longer match `get`, `containsKey`, or `remove` and an
  equivalent key creates a second entry. `size()` and `isEmpty()` do not
  trigger the switch, and maps with a concrete key type are unaffected.
- **List membership is type-strict for numeric values.** `List.contains` and
  `List.indexOf` compared elements with the lenient numeric equality used for
  Set and Map keys, so `new List<Long>{1}.contains(1)` wrongly returned true.
  Set and Map-key membership keep the lenient behavior. Map values are now
  coerced to the map's declared value type on literal construction and `put`,
  so `values()` membership checks compare the right runtime type.
- **Numeric `List` arguments bind to user-defined methods.** Passing
  `List<Integer>` or `List<Double>` to a `List<Decimal>` parameter was wrongly
  reported as "Method does not exist or incorrect signature"; sfapex
  accepts these. Set invariance and map value covariance are unchanged.
- **Built-in describe properties return what their getters return.** Property
  access on `Schema.ChildRelationship` returned null for
  `deprecatedAndHidden`, `junctionIdListNames`, and `junctionReferenceTo`, and
  `Schema.RecordTypeInfo` returned null for `active`, `available`, and
  `master`, because property access fell through to the raw runtime fields map.
  All fifteen built-in Schema describe types now route property access through
  the registered getter.
- **`AggregateResult.getPopulatedFieldsAsMap` includes null fields.** The map
  now holds a key for every selected expression, in alphabetical order,
  including group-by fields that are null for the row. `AggregateResult.get`
  of an expression that was not selected throws `System.SObjectException`
  instead of returning null.
- **A custom object's text Name field defaults to the 15-character Id.** aer
  stored the 18-character Id when the field was unset or null on insert, and
  persisted null when an update explicitly set it to null.
- **Invalid assignments and bare namespaced builtin references are rejected**,
  matching sfapex compile errors: assignments between incompatible final
  primitive types, assigning the result of a void builtin ("Illegal assignment
  from void"), a bare type or namespace name used as an assignment value
  ("Variable does not exist"), and bare stdlib namespace names used as
  variable types ("Invalid type").
- **The simulated CPU time limit applies only when governor limits are
  enforced.** The wall-clock deadline that simulates the CPU governor was
  applied to every execution, so execution overhead, including tracing,
  could abort a legitimately long-running script with "Apex CPU time limit
  exceeded" without `--enforce-governor-limits`. An explicit `--timeout` and
  the longer test-execution timeout are still honored, and per-query timeouts
  are unchanged.
- **Builtin trace spans exclude argument evaluation.** `l.add(slow())` showed
  `List.add` running for as long as `slow()`, because builtins evaluate their
  arguments lazily inside the traced function. Arguments are now evaluated
  before the span opens, exactly once, left to right.

## v1.2.14 — 2026-07-12

- **`Date.format` and `Date.parse` follow the running user's locale for every
  supported `LocaleSidKey`.**
- **`Datetime.valueOfGmt` and `valueOf` parse leniently, and `formatLong`
  renders CLDR zone names.** Out-of-range components roll over
  (`'2018-13-45 25:99:99'` becomes 2019-02-15 02:40:39 GMT) and structurally
  invalid strings throw `TypeException`. `formatLong` converts to the user's
  time zone and renders the short date with seconds and the CLDR short zone
  name: United States metazones keep abbreviations like PST while other zones
  render offsets like GMT+9.
- **Clones keep their query-row provenance.** `SObject.clone()` and
  `List.deepClone()` previously dropped queried-field tracking, so clones of
  SOQL rows stopped throwing `System.SObjectException` for unqueried field
  access, and clones of stub query rows lost their read-only enforcement.
- **The cache sweep can no longer delete the license key.** On Windows the
  default cache root shared its directory with `license.key`, so the
  opportunistic sweep deleted the license once it aged past
  `AER_CACHE_MAX_AGE`, and `aer cache clean` removed it outright. The cache
  root moves to a dedicated `aer/cache` subdirectory, the sweep never removes
  files directly at the cache root (protecting shared `AER_CACHE_DIR`
  directories too), and entries at the legacy root are migrated.

## v1.2.13 — 2026-07-11

- **Keyword DML in system-mode contexts skips user-mode permission checks.**
  Platform event triggers run as the Automated Process user in system mode,
  but bare `insert`/`update` statements in an API 67+ class threw
  `SecurityException` "Access to entity denied" where sfapex permits the
  operation. The `Database.*` builtins, SOQL `USER_MODE`, and
  `EventBus.publish` already skipped checks in system mode; keyword DML now
  does the same.
- **`EntityParticle` and `FieldDefinition` carry the values sfapex returns.**
  `Name` holds the `QualifiedApiName`, `ValueTypeId`/`ServiceDataTypeId` map
  field types to SOAP primitive names, `ByteLength` is computed per type,
  `ExtraTypeInfo` reports plain/rich text areas, `RelationshipName` is
  populated for lookups, and the boolean describe flags are computed from the
  schema with a verified table for standard system-field quirks. Custom
  entities get deterministic durable ids, `FieldDefinition.Id` equals the
  `DurableId`, `MetadataRelationship` joins switch to `QualifiedApiName`, an
  object without record types emits no `RecordTypeId` particle, and
  `USER_MODE` `EntityParticle` queries filter to readable entities.
- **Platform events describe like sfapex.** A platform event now carries
  only `ReplayId`, `EventUuid`, `CreatedDate`, and `CreatedById` plus its
  custom fields instead of the record system-field baseline, with no
  synthetic Master record type. `isCustom()` reports true for all
  custom-suffixed entities, publish-only flags (`isUpdateable()`,
  `isDeletable()`, `isUndeletable()`, `isQueryable()`) report false, and
  describe property access routes through the same logic as the getter
  methods so the two surfaces agree as permissions change.
  `DescribeSObjectResult.getFields()` is a compile error, matching sfapex:
  only the `fields` property exists.
- **Wrong-type lookup ids throw `FIELD_INTEGRITY_EXCEPTION`.** Assigning, for
  example, a User Id to `Contact.AccountId` surfaced as an
  `UNKNOWN_EXCEPTION` with a bare storage message; it now raises the
  sfapex-shaped error ("Account ID: id value of incorrect type: 005…:
  [AccountId]") on insert and update, for keyword DML and `Database` methods
  alike.
- **`Database.emptyRecycleBin`, savepoints, and `convertLead` match sfapex
  errors.** `emptyRecycleBin` supports only the real overloads, rejects a
  `List<String>` and an Id-less SObject with
  `InvalidParameterValueException`, fails per record with `INVALID_ID_FIELD`
  for ids not in the recycle bin, and throws the uncatchable
  `MISSING_ARGUMENT` exception for an empty batch. `Database.rollback(null)`
  throws an NPE, and rolling back to or releasing a released savepoint throws
  `TypeException` "Savepoint does not exist in this context". `convertLead`
  without a leadId or convertedStatus throws the specific `DmlException`s.
  `addError` maps a component field to its compound parent only for the
  `SObjectField` and field-name forms; DML results report compound-mapped
  errors with an empty field list while `getErrors()` keeps the compound
  name.
- **`TIMEVALUE` parses text arguments.** Text is parsed as
  `<hour>:<minute>:<second>.<millisecond>` integer components the way sfapex
  does (the fraction is a millisecond count, so `.1` is 1ms), empty or null
  returns null, invalid text throws `HandledException` "Invalid time format",
  and a formula field whose expression fails at runtime stores null instead
  of failing the operation.
- **JSON deserialization tightens to sfapex behavior.** Lenient
  `JSON.deserialize` rejects a float literal for an `Integer` target with a
  catchable `JSONException` (no lenient coercion exists),
  `@JsonAccess(…='samePackage')` denies access from code deployed as source,
  serializing a `JSONParser`/`JSONGenerator` reports the `common.apex.json.*`
  internal type names, and parent-relationship alias keys built from schema
  tokens carry the org namespace so qualified `__r` keys deserialize in lenient
  and strict modes.
- **`String.format` number patterns throw, and empty padding pads with
  spaces.** Arguments are stringified before formatting, so a `{N,number…}`
  element always throws `StringException` "Cannot format given Object as a
  Number" regardless of argument type, and `leftPad`/`rightPad`/`center` with
  an empty padding string pad with spaces instead of returning the input
  unchanged.
- **Expired monthly CI licenses renew automatically.** License validation
  renews an expired organization key against the renewal endpoint before
  failing, persisting the refreshed key when it is stored in the license file
  (an `AER_LICENSE_KEY` value is renewed for the current run only), and
  proactive background renewal now covers CI licenses as well as developer
  licenses.

## v1.2.12 — 2026-07-10

- **Updates and upserts write back all populated fields of a queried record.**
  sfapex persists every populated field of a record loaded from a query —
  fields queried with a non-null value or assigned in code (even to null) — not
  only the fields assigned since the query, so a queried non-null field
  overwrites a concurrent change made after the query. aer wrote only assigned
  fields. A field queried as null, or never queried nor assigned, is still not
  written, so a concurrent change to it survives.
- **Queried datetimes are whole seconds in subquery and parent paths.** Storage
  keeps sub-second precision for deterministic ordering, and the top-level
  query path truncated to seconds, but subquery children and parent
  relationship traversals returned sub-second values. The populated-field
  write-back then made `Trigger.new` differ from `Trigger.old` on datetime
  fields the code never assigned, creating phantom changes.
- **`SObjectType.newSObject(recordTypeId, true)` loads default values.** The
  `loadDefaults` flag was ignored. It now applies literal and formula
  defaults, sets non-calculated checkboxes to false, applies record-type-aware
  picklist defaults, and assigns the running user as `OwnerId`.
- **`DescribeFieldResult.isDefaultedOnCreate()` matches Apex describes.**
- **`DescribeFieldResult` property access covers every getter.** Properties
  like `unique`, `byteLength`, `calculatedFormula`, `localName`, `autoNumber`,
  and `externalId` returned null because the property-to-getter mapping listed
  only a subset of getters; every registered getter is now mapped.
- **`Messaging.renderEmailTemplate` resolves merge fields.** Body strings were
  returned verbatim; `{!Account.Name}`-style expressions now render against
  `whatId`, and `{!Contact.X}`/`{!recipient.X}` against `whoId`, using the same
  resolver as stored templates.
- **Unique and external ID fields match by case sensitivity and type.**
  Case-sensitive unique text fields were enforced case-insensitively on
  insert; their unique indexes now compare binary.
- **Timezone handling is consistent end to end.** The default timezone maps
  `UTC` to `GMT` so users insert in UTC environments now that `TimeZoneSidKey`
  is restricted, `exec --timezone` is applied before the default user is
  created and reflected by `UserInfo.getTimeZone()`, and a `TimeZone` prints as
  its zone ID instead of a Go handle.
- **After undelete triggers see `IsDeleted` = false.** Trigger records were
  read from the recycle bin with `IsDeleted` still true.
- **`addError()` state does not affect record equality.** The synthetic error
  markers leaked into `==`, `equals()`, `System.assertEquals`, and Set/Map
  hashing, so a record carrying an error compared unequal to an
  otherwise-identical record.
- **A custom object's Text Name field describes as nillable and defaulted on
  create.**
- **`Test.createStubQueryRow` supports child relationships and enforces
  read-only rows.** A key naming a child relationship stores the supplied List
  as a subquery result (readable via `getSObjects()`, static access, and JSON,
  with explicit empty lists preserved), and any non-List value throws
  `TypeException`. Stub rows are read-only — assignment and `put()` throw
  `SObjectException` "Record is read-only" — and `addError` outside a test
  throws the uncatchable `FinalException`, matching sfapex.
- **`Test.enqueueBatchJobs` inserts placeholder jobs.** It was a no-op, so
  `Test.getFlexQueueOrder()` stayed empty and fill-the-queue helpers looped
  until the CPU limit. The first five placeholders fill the batch queue and
  the rest hold in the flex queue; `System.scheduleBatch` also moves over-limit
  jobs to the flex queue instead of ignoring a capacity-check error.
- **Typed `JSON.deserialize` rejects `Object` targets.** A bare `Object` field,
  list element, or top-level target now throws
  `JSONException("Apex Type unsupported in JSON: Object")` for any non-null
  value instead of silently producing an untyped Map; explicit null is
  accepted.
- **A script-thrown `DmlException` reports the throw line.** The
  enclosing-try-line adjustment that applies to DML-raised exceptions was also
  applied to explicit `throw` statements, so `getStackTraceString()`
  disagreed with `getLineNumber()`.
- **Over-limit `String.repeat` throws the uncatchable `LimitException`.**
  sfapex rejects a repeat whose result would exceed the 6,000,000-character
  maximum up front with `LimitException("String is too long.")`, which a catch
  block cannot intercept; incremental concatenation keeps throwing the
  catchable `StringException`.
- **`Database.undelete` errors bind as `List<Database.Error>`.** The `errors`
  property on a failed `UndeleteResult` (and merge failures) held a raw value
  that could not be passed where a `List<Database.Error>` was expected.

## v1.2.11 — 2026-07-09

- **Date and DateTime roll-up summaries read back as the right type.** A
  min/max roll-up summary is assigned a placeholder type before its summarized
  field is known, so a roll-up over a `Date` or `DateTime` field was not
  materialized as an Apex `Date`/`Datetime` when read from storage. Field
  conversion now follows a roll-up summary to its summarized field and wraps the
  stored value with the effective type, so a max-over-`Date` roll-up is
  recognized as a `Date`.
- **Report, dashboard, document, and email-template folders are queryable.**
  Email-template folders shipped as metadata keep the `accessType` declared in
  their `.emailFolder-meta.xml` instead of defaulting to Hidden, so a query for
  a public folder returns it, and a `Folder.Type LIKE 'Email template'` filter
  matches the stored `Email`/`EmailTemplate` types. Report, dashboard, and
  document folder metadata is reconciled into queryable `Folder` rows carrying
  the folder Type.
- **`TimeZoneSidKey` is a restricted picklist.** sfapex rejects timezone
  keys outside its catalog with `INVALID_OR_NULL_FOR_RESTRICTED_PICKLIST`, but
  aer silently accepted invalid values like `"UTC"`. `User.TimeZoneSidKey` is
  now restricted, and the `Etc/GMT` offsets sfapex accepts but omits from
  describe (`Etc/GMT+1`..`+12`, `Etc/GMT-1`..`-14`, `Etc/UTC`) are honored.
- **`Flow.Interview.createInterview` for a missing flow reports `Invalid
  type`.** A flow that does not exist now throws a `TypeException` with the
  message `Invalid type: <namespace>.<name>` (dot-separated, and just `<name>`
  when no namespace is supplied), matching sfapex instead of the former
  "the flow is inactive or does not exist" wording.
- **Text-formula concatenation of lookup fields concatenates their Ids.** A
  reference operand was not treated as string-like, so `lookup + lookup`
  compiled to numeric addition and produced a silent zero. Reference operands
  in the `+` and `&` operators now concatenate as their 15-character record Id
  (`CASESAFEID` results stay 18), blank operands are treated as empty text.
- **Possessive quantifiers compile and match.** aer
  rejected `X*+`, `X++`, `X?+`, and `X{n,m}+` at `Pattern.compile`.
- **Builtin calls with the wrong argument count are rejected at compile time,
  and null arguments throw catchable exceptions.** The type checker previously
  validated only that a builtin method name existed, so calls like
  `Datetime.newInstanceGmt(2010, 1, 1, 14)` or `String.lastIndexOf()` failed
  only at runtime; declared overload arities are now enforced with the
  sfapex error format. A null `Integer` reaching a builtin argument
  conversion now throws a catchable `NullPointerException("Argument cannot be
  null.")` across the `Datetime`/`Date`/`Time` constructors, `add*` methods,
  and many other call sites instead of an uncatchable internal error, an NPE
  raised while evaluating a method argument is no longer swallowed except for
  the null-SObject-field-access quirk, and null list-index semantics
  (`get(null)`, null subscript writes/reads, `add(null, e)`) match sfapex.
- **Flow-generated queueables are transparent to the Apex async model.**
- **`Metadata.Operations.retrieve` supports `CustomMetadata`.** Requests for
  `Metadata.MetadataType.CustomMetadata` fell through to a default branch and
  returned an empty list, so code that indexed `[0]` into the result threw a
  List index exception..
- **`Schema.ChildRelationship.field` returns an `SObjectField`.** The lookup
  field was kept as a raw string, so direct property access
  (`relationship.field`) returned the string and `relationship.field.get`
  `Describe()` failed with "getDescribe() called on non-SObjectField instance".
- **Namespace-qualified builtin receivers keep their disambiguation.** The type
  checker resolved a method-call receiver twice, and the second pass discarded
  the `Schema.` qualifier — so inside a class declaring a nested type also named
  `AggregateResult`, a `Schema.AggregateResult` receiver collapsed onto the
  nested type and builtin methods like `getPopulatedFieldsAsMap` were reported
  nonexistent.

## v1.2.10 — 2026-07-07

- **`Matcher.find()` advances past zero-length matches.** A find loop over a
  boundary pattern like `\b` or an empty-capable pattern like `x?` returned the
  same empty match forever instead of yielding successive positions.
  A zero-length match now advances the continuation point by one, and
  continuation matches run against the full region with a start offset instead
  of a byte slice, so boundary and lookbehind constructs see the real left
  context. Group extraction now distinguishes an empty capture (empty string)
  from a group that did not participate in the match (null with -1 indices),
  matching sfapex.
- **Standard-field stubs in retrieved source resolve against builtin
  definitions.** Source retrievals include standard-field stubs (a bare
  `<fullName>` or a detail-less `<type>`), and the importer fabricated a text
  type for any stub it could not resolve. Stubs now fill from the builtin
  template.
- **Lowercase `label.X` custom-label references survive an inner class named
  `Label`.** An inner class named `Label` anywhere in the org hijacked the
  fallback resolution of lowercase `label` identifiers, failing unrelated
  classes with 'identifier "label" not found in binding'. An unbound `label`
  identifier now resolves deterministically to the custom-labels global; a
  `Label` type lexically in scope still shadows it, matching sfapex. This
  also fixes wrong-case label names returning a placeholder object instead of
  the label value.
- **`SingleEmailMessage.isUserMail()` derives from the target object type.** It
  was backed by a field nothing populated and always returned false, so code
  branching on it to read `getToAddresses()` hit a NullPointerException. It now
  returns true when `targetObjectId` is a User; Contact and Lead targets and
  messages addressed only via `setToAddresses` remain external.
- **Identity casts and `instanceof` work for builtin runtime values.** Casting
  a `TimeZone` held in an `Object` back to `TimeZone` threw "Invalid conversion
  from runtime type TimeZone to TimeZone"; the same affected `Url`, `Pattern`,
  `Matcher`, `UUID`, `Version`, `Location`, and `Type`. A cast is now accepted
  when the value's canonical runtime type equals the target type, `instanceof`
  matches builtin values (including `Schema.SObjectType` references) against
  their canonical type name, `Assert.isInstanceOfType` works for these types,
  and assert failure messages report the Apex type name instead of an internal
  Go type name.
- **Indexing a method-call result no longer leaks SOQL count semantics into the
  target.** `evalIndexExpression` set its evaluation context around the whole
  target, so a `COUNT()` query compared inside a method used as an index target
  (e.g. `pick_by_count()[0]`) failed with "invalid operands for >". The context
  now applies only when the target is a directly-nested SOQL literal; the
  direct `[SOQL][index]` behaviors are unchanged.
- **Custom metadata and custom setting accessor arguments bind to the right
  overload.** The resolver could not type the platform accessors
  (`getInstance`, `getOrgDefaults`, `getValues`, `getAll`) on custom metadata
  and custom setting types, so a call like
  `myMethod(Setting__mdt.getInstance(name))` treated the argument as a wildcard
  and bound to the first candidate — such as `myMethod(String)`, making the
  method call itself until the resolution limit reported infinite recursion.
  The accessor return types are now recorded and SObject-typed arguments are
  matched by canonical object name during overload selection.
- **Queueable chaining limits are enforced in tests.** `System.enqueueJob` from
  inside an executing queueable now throws `System.AsyncException` ("Maximum
  stack depth has been reached.") during a test unless the chain was started
  with an explicit `AsyncOptions.MaximumQueueableStackDepth`, and configuring a
  depth while a queueable is executing throws "Cannot reset Maximum Queueable
  Stack Depth in a queueable" — both reference-verified. The extra chained
  executions previously inflated `AsyncApexJob` counts and ran chain logic
  sfapex never executes in tests.
- **Non-formula picklist defaults are treated as literal values.** A package
  picklist defaulting to `"0.00"` parsed as a formula number, so records stored
  the numeric zero instead of the API value `0.00`. A picklist or multipicklist
  default whose formula evaluation yields a non-string is now taken as the
  literal API value.
- **Metadata-loaded permission set groups default to Status `Updated`.**
  Authored metadata normally omits `<status>`, so the record's Status was null
  and inserting a `PermissionSetAssignment` against the group failed with
  INVALID_CROSS_REFERENCE_KEY. Both importers now default Status to `Updated`
  (an explicit `<status>` still wins), matching a reference org at rest.
- **Custom sharing reasons are sorted in Share object `RowCause` picklists.**
  Map-iteration order made the schema differ from run to run, so the schema
  fingerprint keying the on-disk database template cache silently missed on
  warm runs and paid full template repopulation.
- **`ApexPages.Severity` constants declare in platform order.** aer used the
  alphabetized stub order, but the platform declares severity-descending:
  FATAL, ERROR, WARNING, INFO, CONFIRM (reference-verified). `values()`
  indexing and `ordinal()` threshold checks now return the right constants.
- **RunAs self-grants are suppressed at the top level of test methods.** A
  `PermissionSetAssignment` granting the running user a permission set inside a
  `System.runAs` block (commonly in `@TestSetup`) is honored on sfapex only
  inside a `System.runAs` block. aer honored it everywhere, so tests passing
  under aer failed on a real org. Such grants are now excluded at the
  synchronous top level of a test while remaining honored inside `runAs`
  blocks, in async job execution, and after `Test.startTest()`.
- **DataWeave collections written into Map-typed Apex fields are rejected.** An
  `application/apex` script writing an array into a `Map<String, Object>` field
  stored the list unvalidated, surfacing later as "method not found: List.put".
  It now throws `System.DataWeaveScriptException` ("cannot assign a collection
  to a field of type Map<String,ANY>").
- **`DescribeFieldResult.getReferenceTo()` reports targets for built-in
  polymorphic fields.** 31 reference fields, including `Attachment.ParentId`,
  `Note.ParentId`, `ContentDocumentLink.LinkedEntityId`, and
  `TopicAssignment.EntityId`, described an empty target list, so selectors that
  resolve relationship segments with `getReferenceTo()[0]` failed with
  a list-index error sfapex never produces. Polymorphic lists are now derived
  from the schema's child relationships and fixed-target lookups declare their
  verified targets.
- **MiddleName and Suffix auto-enable from Apex references.** References like
  `Lead.MiddleName` or `new Contact(Suffix = ...)` failed type checking with
  "Variable does not exist" unless NameSettings metadata was loaded. A
  compilation-unit scan now enables the matching setting when the field is
  referenced in an object context; an explicit `false` in a loaded NameSettings
  file is never overridden.
- **Short Text Area fields can be sorted in SOQL.** A custom `TextArea` field
  with no explicit length was rejected from ORDER BY (and `isSortable()`
  returned false) even though sfapex only forbids sorting Long and Rich
  Text Areas. Short text areas now sort; long text areas remain rejected.

## v1.2.9 — 2026-07-05

- **Record-triggered flows survive multi-row lookups again.** Several fixes to
  the flow converter and trigger optimizer that together broke record-triggered
  flows once multi-row lookup results were declared:
  - A Get Records element in collection mode that matches no records now sets its
    output to `null` (matching Salesforce, where an `IsNull` decision on the
    result takes the true branch) rather than an empty list. An Assignment that
    adds a `null` collection value appends a single `null` element, so a
    subsequent count yields 1, and Loop collections wrapped so a `null`
    collection iterates zero times.
  - The SOQL-in-loop bulkification kept only the first record per key even for
    collection-typed lookups, so a parent with several children silently lost
    every child after the first. Collection lookups keyed by a non-Id field now
    build a grouped `Map<Id, List<T>>` so every row is retained; Id-keyed parent
    traversals keep the single-record map.
  - A query filtering on a loop-variable equality plus another loop-variable
    comparison (e.g. `WHERE AccountId = :record.AccountId AND Id != :record.Id`)
    was hoisted before the loop with the second bind left referencing the vanished
    loop variable, failing with a bind-resolution error. Such queries now stay
    per-record inside the loop.
- **Flow interviews pause at Wait elements.** An autolaunched flow reaching a
  Wait element previously failed with "Unknown flow element". Elements before the
  Wait are now committed, control returns to the caller without an exception, and
  the wait-event connectors do not run during a test.
- **Rolled-back async jobs are skipped and aborted `AsyncApexJob` rows are
  kept.** A queueable enqueued after `Database.setSavepoint()` and then rolled
  back is now discarded entirely instead of failing the `Test.stopTest` drain
  with "AsyncApexJob not found". `System.abortJob` marks the `AsyncApexJob` row
  `Status=Aborted` and skips execution instead of deleting the row, and
  re-aborting an already-aborted queueable is a no-op.
- **A profile with no default record type falls back to the Master record
  type.** Inserting a record without a RecordTypeId no longer errors when the
  object has multiple active record types and the running user's profile has no
  registered default; the insert succeeds with a null RecordTypeId, matching
  sfapex.
- **Feature parameter names resolve against the executing namespace.**
  `FeatureManagement.checkPackage*Value` matched parameter names by exact string,
  so a namespaced package run threw `NoDataFoundException` for a
  namespace-qualified name seeded under its bare authored name. Names now match
  case-insensitively, a bare name resolves against the calling package's
  namespace, and a name qualified with a foreign namespace still throws.
- **`SetupEntityAccess.SetupEntityId` is polymorphic.** `SELECT
  SetupEntity.Name FROM SetupEntityAccess` failed with "No such column 'Name' on
  entity 'CustomPermission'". The field now declares its nine referents
  (ApexClass, ApexPage, ConnectedApplication, CustomPermission,
  ExternalClientApplication, ExternalDataSource, MessagingChannel,
  NamedCredential, OrgWideEmailAddress), so `SetupEntity.Name` resolves through
  the targets that have a Name field and returns null for CustomPermission rows.
- **User-mode DML field-access failures throw a catchable `DmlException`.**
  User-mode DML (`Database.insert`/`update` with `AccessLevel.USER_MODE`) that
  writes a field the running user lacks FLS on previously returned an
  uncatchable raw error and failed the whole call. It now throws a `DmlException`
  with status code `CANNOT_INSERT_UPDATE_ACTIVATE_ENTITY` listing all
  inaccessible fields in set order; with `allOrNone=false` the failing row
  becomes a failed `SaveResult` while the rest save, and object-level denial
  still throws `SecurityException`.

## v1.2.8 — 2026-07-04

- **A method call resolves to a method, not a constructor sharing the class
  name.** Apex lets a method share its class's name as long as it declares
  a return type. The typechecker matched method-call candidates by lowercase
  name alone, so such a call resolved to the constructor, whose empty return
  type was reported as `void` — producing errors like "Illegal assignment from
  void to Map<String, Object>" when the method's result was assigned.
  Constructors are now excluded from method-call lookups, and a same-named
  method is no longer treated as a constructor overload for `new` expressions.
- **Builtin maps and sets preserve insertion order so every entry is
  visible.** Several builtins wrote keyed entries without recording their
  order, so `size()` counted them but `values()`, `keySet()`, for-each, and
  `toString` never visited them. Fixed across six sites:
  `OrgLimits.getMap()`/`getAll()`, the `Auth.AuthToken`, `Auth.JWT`, and
  `Auth.SessionManagement` maps, `Approval.isLocked(List<Id>)`,
  `Slack.ViewReference.setParameter`, and `QueryException` inaccessible-field
  sets.
- **`WITH SECURITY_ENFORCED` failures report an empty inaccessible-field
  map.** A `QueryException` from `WITH SECURITY_ENFORCED` now has a non-null
  but empty `getInaccessibleFields()`; only `USER_MODE` failures carry the
  populated object-to-fields map.
- **`Auth.SessionManagement.getQrCode()` returns documented keys.** The
  result now uses the documented `qrCodeUrl` and `secret` keys.

## v1.2.7 — 2026-07-03

- **Static member access through a class name shadowed by a local variable is
  rejected.** Apex identifiers are case-insensitive, so a local variable or
  parameter named like a class (e.g. a `List<Account> statuses` parameter next
  to a `Statuses` class) shadows the class, and `Statuses.SOME_CONSTANT`
  resolves against the variable's type. aer previously fell back to the
  class's static member and compiled code that sfapex rejects; it now reports
  the same "Variable does not exist: <member>" error (reference-verified).
  Members that do exist on the variable's type, including inherited ones,
  still resolve through the variable.
- **Unknown methods on `List`, `Set`, and `Map` are compile errors.** A method
  call on a collection type that isn't in the collection's built-in method set
  previously passed type checking and failed at runtime with "method not
  found". It is now rejected.  This also covers static method calls through
  a shadowed class name when the method is missing from the shadowing
  variable's type.

## v1.2.6 — 2026-07-03

- **`Schema.SObjectType.<Object>` describe properties are typed as their real
  types.** Property access on the describe result (`name`, `label`,
  `keyPrefix`, `custom`, `childRelationships`, ...) previously left the whole
  chain typed as `Schema.DescribeSObjectResult`, so comparing
  `Schema.SObjectType.Contact.name` to a `String`, concatenating it, or calling
  String methods on it raised spurious "Comparison arguments must be compatible
  types" / "Method does not exist" errors, and
  `List<Schema.ChildRelationship> rels = Schema.SObjectType.My_Object__c.childRelationships;`
  failed with an illegal-assignment error. Properties now carry the
  String/Boolean/collection types from the describe result.
- **Describe record-type-info properties return real data.** The
  `recordTypeInfos`, `recordTypeInfosById`, `recordTypeInfosByName`, and
  `recordTypeInfosByDeveloperName` property forms on a describe result returned
  an empty placeholder while the equivalent `getRecordTypeInfos*()` methods
  returned the real data; both forms now share the same builders. The synthetic
  Master record type is also reported consistently by all four accessors —
  alongside custom record types, as sfapex does, where previously only
  `getRecordTypeInfosByDeveloperName()` included it.
- **Namespaced custom-SObject type arguments match in generic assignability.**
  In a managed-package context, assigning a class implementing
  `Database.Batchable<Custom_Object__c>` to a variable of that same type was
  rejected ("Illegal assignment from ns.MyBatch to
  Database.Batchable<ns__Custom_Object__c>") because the declared type argument
  was namespace-normalized while the implements-clause argument stayed bare.
  Two generic specs now unify when their SObject type arguments resolve to the
  same canonical name; `instanceof` "always true" detection is namespace-aware
  the same way.
- **Public override of a global abstract method is accepted.** A top-level
  global class overriding a global abstract method with a `public` method is
  valid (reference-org verified); aer wrongly rejected it with "Cannot reduce
  the visibility of method". `protected`/`private` overrides are still
  rejected.
- **Null-argument constructor overloads resolve by most-specific non-null
  position.** A constructor call with a null argument, such as
  `new Duration(123.1, null)`, was flagged "Ambiguous method signature"
  whenever two or more constructors were applicable. Matching sfapex, a
  null literal matches any reference type equally and cannot discriminate
  between overloads; the call is ambiguous only when no single constructor is
  strictly more specific than every other applicable one.
- **Array-literal elements are not treated as constructor arguments.** An
  array-literal initializer with a null element, such as
  `new Token[] { null }`, was resolved as the constructor call
  `new Token(null)` and reported "Constructor not defined" for a valid
  literal.
- **NUL bytes lex as whitespace.** Spurious NUL bytes (for example
  editor-introduced trailing padding) previously raised "token recognition
  error". sfapex treats NUL as a token separator, so the lexer now accepts
  it as whitespace.

## v1.2.5 — 2026-07-01

- **DataWeave `reduce` with a default accumulator.** DataWeave scripts using
  `reduce` with a default value on the accumulator parameter — e.g. `reduce
  ((item, acc = {}) -> ...)` — now parse and evaluate instead of failing with
  "no viable alternative at input 'reduce'". Lambda parameters may carry default
  values, and `reduce` seeds from the accumulator default, folding over every
  element and returning the default for an empty array.

## v1.2.4 — 2026-07-01

- **`Map.put` returns the previous value.** `Map.put(key, value)` now returns
  the value previously associated with the key (or `null` when the key was
  absent), matching sfapex. Assigning its result previously reported "Illegal
  assignment from void to ..." at compile time and produced no value at runtime.
- **`Trigger.new` is typed so it coalesces as a list.** The typechecker now
  gives the implicit `Trigger` context variable static types (`new`/`old` as
  `List<SObject>`, `newMap`/`oldMap` as `Map<Id, SObject>`, the context flags as
  `Boolean`, `size` as `Integer`, `operationType` as `System.TriggerOperation`).
  Without a type for `Trigger.new`, the null-coalescing operator collapsed
  `list ?? Trigger.new` to a singular `SObject` and raised a spurious "Illegal
  assignment from SObject to List<SObject>" error.
- **Custom setting fields are exempt from FLS in `USER_MODE` queries.** Once the
  running user can read a custom setting object, all of its own fields are
  readable under `USER_MODE` regardless of field-level security. A `WITH
  USER_MODE` / `Database.query(..., USER_MODE)` read of a custom-setting field
  previously failed with "No such column '<field>'". The object-level read
  check still gates access, and regular objects, relationship traversals, and
  the describe path are unchanged.
- **Trigger exception locations survive DML failures.** When an exception thrown
  inside a trigger surfaces from DML, the failure's `getMessage()` now carries
  the "caused by" chain — the System-qualified cause type and message followed
  by the originating `Trigger.<name>: line N` frame — so callers can tell which
  query failed. This is applied across the full DML matrix
  (insert/update/delete/undelete, DML statements and `Database` methods,
  `allOrNone` true/false); `Database.delete`/`undelete` and
  `Database.update(allOrNone=false)` are also corrected to wrap or report the
  error consistently instead of leaking the raw exception.
- **Inline SOQL in flow-generated triggers is namespaced.** Immediate
  record-triggered flows emit their Get Records / inline SOQL directly into the
  trigger body, so a bare custom object reference (e.g. `Queue__c`) previously
  survived un-namespaced and leaked into the object name reported by a
  `QueryException`. The default namespace is now applied to a trigger body's
  inline SOQL and local classes, mirroring the class path.
- **`aer exec` retains non-namespaced interfaces and enums.** When any source
  file carried a package namespace (a single packaged trigger is enough), exec's
  namespace-grouping dropped every unrelated non-namespaced interface and enum,
  so a class implementing such an interface failed with "unresolved interface".
  Non-namespaced interfaces and enums are now retained alongside classes and
  triggers.

## v1.2.3 — 2026-06-30

- **Flows publish platform events they create.** A Flow "Create Records" element
  targeting a platform event now publishes the event through `EventBus.publish`
  instead of compiling to a keyword `insert event;`, which threw at runtime
  ("Argument must be of internal sObject type") and silently dropped the event.
- **Flow stack traces point at the offending statement.** Flow-generated AST
  statements are stamped with their output line number while each class's source
  is generated, so an exception inside a generated Queueable now reports the real
  line instead of the default "line 1, column 1".
- **Time-based flow scheduled paths skip test execution.** Only time-based
  scheduled paths, which Salesforce defers to a future time and never runs during
  a test, are now skipped in the `@TestSetup` async flush; async-after-commit
  paths still run like ordinary async jobs, matching Salesforce.
- **VS Code extension adds downloaded aer to the terminal PATH.** When the
  extension downloads aer, its directory is appended to the integrated terminal's
  PATH (so a user-installed aer still wins) at both process creation and via shell
  integration, so it survives rc files that rebuild PATH.

## v1.2.2 — 2026-06-29

- **RecordTypes on objects with long qualified API names.**
  `RecordType.SobjectType` no longer rejects RecordTypes belonging to
  managed-package objects whose qualified API name exceeds 40 characters.
  sfapex reports the picklist length as 40 but accepts longer qualified
  names, so the field now keeps a describe-reported length of 40 while
  bypassing strict insert-time length validation, fixing a `STRING_TOO_LONG`
  failure during VM pool template creation.
- **`String.valueOf` truncates large collections.** `String.valueOf` and
  `toString` of a collection now render at most the first ten elements, followed
  by a `, ...` marker when the collection holds more than ten, matching
  sfapex. Object rendering folds the synthetic property backing field into
  the property's declared name so each property appears once.
- **"Regex too complicated" limit enforced with per-operation thresholds.**
  Regex-backed `String`/`Pattern` methods now throw the uncatchable
  `System.LimitException: Regex too complicated` past a character-based
  input-size limits: 500,000 chars for `split` / `split(rx, limit)`
  / `replaceFirst`, 1,000,000 chars for `replaceAll` / `Pattern.matches`
  / `Matcher` methods. `replaceAll` / `replaceFirst` backtracking is also
  bounded by a match timeout so a catastrophic pattern aborts instead of
  hanging.
- **CPU charged for string concatenation by result length.** The `+` operator is
  now billed proportional to the result length. A large string build that
  previously stayed cheap under the CPU limit in aer while exceeding it on
  sfapex now fails in aer.
- **License file written atomically.** Registering a license key now writes to a
  temporary file and renames it into place, so a concurrent aer process can no
  longer read the file in a truncated, zero-length state and report "license key
  not provided".

## v1.2.1 — 2026-06-25

- **Custom report types with joined child relationships.** `.reportType`
  metadata is parsed into report type definitions, and a report that references
  a custom report type by developer name resolves its base SObject. Joined child
  columns are selected as relationship subqueries and flattened into one detail
  row per child record, with an outer join keeping base records that have no
  children. Report permissions are enforced in user mode (`Run Reports`, and
  read access to filter columns).
- **Text block delimiters require a line terminator.** The opening `'''` of a
  triple-quoted text block must be immediately followed by a line feed, matching
  sfapex, which rejects the single-line `'''hello'''` form.
- **Deterministic package-merged child relationships.** Reverse child
  relationships are now built in canonical sorted order, so a name collision
  resolves the same way every run and the schema fingerprint (and the on-disk
  storage-template cache) stays stable instead of rebuilding each run. Report
  type definitions are carried through the schema merge and re-registered when a
  schema loads from the cache, fixing custom report types on a cache hit.
- **Opportunity probability derived from stage.** `Opportunity.Probability` is
  populated from the matched `OpportunityStage.DefaultProbability` on insert, and
  on update only when `StageName` is explicitly changed, leaving a
  user-customized probability untouched otherwise. Stage Won, Probability, and
  ForecastCategory attributes now flow through the standardValueSet import
  pipeline so custom org stages work end to end.
- **Visualforce rendering in `PageReference.getContent()`.**
  `getContent()` / `getContentAsPDF()` now render a page that has a custom
  controller, instantiating the controller and evaluating `{!...}` expressions
  with strict, save-time semantics so aer catches expression errors Salesforce
  rejects at save.
- **Faster server cold start.** The server's named in-memory database is now
  hydrated from the storage-template cache instead of running a full schema
  migration on every startup, substantially reducing cold start time.
- **Bare ampersands in picklist metadata.** Metadata containing bare `&`
  characters in picklist values or labels (for example `R&D`) now deploys
  instead of being silently dropped.

## v1.2.0

Changes in v1.2.0 that are not part of the v1.0.x patch line.

### New SObject and feature support

- **Salesforce Knowledge.** References to `__kav` article-version and `__ka`
  article objects now resolve, and the object family (article, version, share,
  data-category selection, view/vote stats, history, CaseArticle) is synthesized
  on demand. `KbManagement.PublishingService` implements the full lifecycle
  (publish, edit, archive, restore versions, schedule, and the translation
  methods), draft-version DML auto-creates the parent article, UrlName format and
  uniqueness are enforced, and queries hide superseded versions. `WITH DATA
  CATEGORY` SOQL, multi-language articles, the abstract `KnowledgeArticle` /
  `KnowledgeArticleVersion` entities, and gating on the "Manage Salesforce
  Knowledge" permission are all supported. `ConnectApi.Knowledge`
  trending-article methods are implemented.
- **Team Selling and Opportunity Splits.** `AccountTeamMember` (gated on Account
  Teams), `OpportunityTeamMember` (gated on Opportunity Teams), `OpportunitySplit`,
  and `OpportunitySplitType` can be described, queried, and manipulated. With
  Opportunity Splits enabled, the opportunity owner is auto-enrolled as an
  `OpportunityTeamMember` on insert, and the gated `Opportunity.IsSplit` field is
  exposed.
- **Enterprise Territory Management.** The Territory2 model, type, territory,
  assignment, and assignment-rule objects are registered when
  `Territory2Settings.enableTerritoryManagement2` is set, along with
  `Opportunity.Territory2Id` and `IsExcludedFromTerritory2Filter`. DML that
  Salesforce rejects at compile time is rejected, and territory metadata is loaded
  as queryable records.
- **Collaborative Forecasts.** `ForecastingType` and `ForecastingGroup` are
  registered when `ForecastingSettings.enableForecasts` is set, modeled as
  read-only setup objects, with the standard `ForecastingType` records seeded.
- **Case Team objects.** `CaseTeamMember`, `CaseTeamRole`, `CaseTeamTemplate`, and
  `CaseTeamTemplateMember`, with the DML restrictions Salesforce applies.
- **Education Cloud** (`--feature EducationCloud`). 88 Education Cloud objects plus
  fields added to standard objects, auto-enabled when an Education Cloud object or
  field is referenced.
- **High Velocity Sales** (`--feature HighVelocitySales`). The Sales Engagement
  cadence objects (`ActionCadence`, `ActionCadenceStep`, trackers, …),
  auto-enabled when referenced.
- **VideoCall** standard object (Einstein Conversation Insights), with its full
  field set, child relationships, and inserts.
- **Custom Address fields.** Address-type custom fields now generate their `__s`
  component fields (street, city, state/country codes, postal code, lat/long,
  geocode accuracy), so SOQL and Apex referencing the components compile and the
  compound reconstructs on query, including under a namespace.
- **Field history tracking.** Updating a field configured with `trackHistory` on a
  history-enabled object writes an `<Object>History` row when the change commits
  (integration tests, anonymous exec, and the server commit their data; standard
  `@IsTest` tests roll back and so create no history).
- **Seeded standard records.** `PermissionSetLicense`, `MilestoneType`,
  `ForecastingType`, and `SlaProcess` (Entitlement Process) now ship with the
  standard records a real org provides, so queries against them return rows
  instead of empty results.
- **Queue sharing with bosses.** A queue with `DoesIncludeBosses = true` now shares
  queue-owned records with users above queue members in the role hierarchy.

### New capabilities

- **`aer test` workspace auto-detection.** When type checking a file fails to
  resolve references and the file lives inside an `sfdx-project.json` or
  `package.xml` workspace, aer loads the surrounding workspace and retries once,
  while keeping test selection scoped to the files you named. The retry preserves
  explicitly-passed source paths that fall outside the workspace's package
  directories, honors a user-supplied `--filter-path` instead of widening it, and
  works under `--watch` (a lone file argument is watched through its parent
  directory, and the auto-loaded dependency sources are watched too).
- **`--runs N`.** Repeat each selected test method N times to surface flaky tests;
  each PASS/FAIL line is tagged with its run number.
- **`--integration-tests`.** Discover and run `@IntegrationTest` classes and
  methods, which commit their data and run `@TearDown` cleanup between methods.
  Compatible with `--debug` and the VS Code test panel, which now separates Unit
  and Integration tests.
- **Memory-constrained execution.** Large string concatenations become shared
  rope strings, and per-test in-memory SQLite databases automatically spill to
  temporary on-disk files under memory pressure
  (`--spill-to-disk=auto|never|always`). Pooled test VMs now share one immutable
  copy of the schema and copy-on-write only the objects a test mutates, cutting
  resident memory substantially on large suites. The wasm build keeps its SQLite
  template on disk to fit the 4 GiB linear-memory cap.
- **`System.EventBus.publishWithAccessLevel`.** Platform events can be published
  under an explicit `AccessLevel` (single event or list, with an optional publish
  callback). `USER_MODE` enforces only object-level Create access, not
  field-level security, matching sfapex, and a null `AccessLevel` raises a
  `NullPointerException`. `EventBus.publish` now runs in `USER_MODE` for
  callers on API version 67.0 or later and `SYSTEM_MODE` for earlier versions,
  matching the Summer '26 change.
- **Schema build caching.** Assembled schemas are fingerprinted from the metadata
  source directories and cached between runs. An age-based garbage collector
  reaps stale cache entries, configurable through `AER_CACHE_MAX_AGE`,
  `AER_CACHE_MAX_BYTES`, and `AER_CACHE_GC_INTERVAL`, with a new `aer cache`
  command (`info`, `prune`, `clean`).

### Correctness fixes

- **SObject field writeability** fixes false "Field
  is not writeable" errors on system fields that are writeable in memory. A share
  object's manual-sharing fields are writeable unless the shared object's sharing
  model is Controlled by Parent.
- **Platform-defaulted required fields.** Inserts that omit non-nillable fields
  the platform auto-populates now succeed instead of failing
  `REQUIRED_FIELD_MISSING` — `FeedItem` flags and counts, `EmailMessage` flags and
  `Status`, `AiJobRun.Status` (defaults to `New`), Territory2 access levels, and
  similar fields.
- **FeedItem** inserts now populate the audit timestamps and derive `Type`
  (`ContentPost` / `LinkPost` / `TextPost`), fixing `LAST_N_DAYS` filtering.
- **Change Data Capture `*ChangeEvent` objects.** In-memory field assignment is
  decoupled from describe flags (every field is settable when building a test
  event except auto-number, calculated, formula, and rollup fields), the field
  set matches Salesforce, and `EventBus.publish()` of a constructed change event
  raises the uncatchable "External Object Error" only when Change Data Capture is
  disabled. Once a test calls `Test.enableChangeDataCapture()`, publishing a
  constructed change event succeeds (assigning an `EventUuid` and in-memory
  `ReplayId`) without counting against the DML limits, and
  `Test.getEventBus().deliver()` / `Test.stopTest()` fire the change event's
  after-insert trigger.
- **SOQL string literals** of 4000 or more characters now raise a `QueryException`,
  matching Salesforce.
- **Unified schema loading.** The `test`, `exec`, and `server` commands all load
  schema through one walker that reads both source (SFDX) and metadata (MDAPI)
  format, so the server now resolves bare `.md` custom metadata records and
  embedded list views the same way the test command does.
- **Namespaced governor-limit reporting.** Limit exceptions raised in
  managed-package code are prefixed with the controlling namespace (for example
  `myns:Too many SOQL queries: 201`), and governor limits are always enforced
  during test execution on the server.

## v1.0.15 — 2026-06-20

- `DescribeSObjectResult.getKeyPrefix()` returns null for virtual or derived
  objects that are never addressable by a real Id prefix.

## v1.0.14 — 2026-06-19

- `Security.stripInaccessible(AccessType.READABLE, …)` now recurses into related
  records instead of stripping only top-level fields. Fields reached through a
  parent lookup (`record.Rel__r.Secret__c`) or a child subquery were previously
  copied verbatim, exposing data the running user cannot read; they are now
  removed at every depth, reported under their own object type in
  `getRemovedFields()`, and the parent's dotted queried paths are pruned so
  navigating to the relationship later does not re-mark a stripped field as
  queried.
- A `__r` token that names both a parent relationship and a child relationship
  now resolves the way sfapex does. The type checker and runtime
  resolve the child relationship first for such collisions, so an invalid deep
  parent traversal is rejected, while a scalar field read through a parent that
  was populated explicitly (via `putSObject` or a parent-traversal query)
  returns the parent record instead of throwing "List has no rows for
  assignment to SObject".
- Querying a fully-readable parent relationship and stripping the result no
  longer reports a false read-access violation. The internal `__rowid` join
  column SOQL attaches to nested relationship records is treated as a system
  field during the strip, so it is never surfaced in `getRemovedFields()` as a
  removed field with an empty name (previously printed as `{Object__c={}}`).

## v1.0.13 — 2026-06-18

- Subscriber (unmanaged) custom fields authored on a managed-package object keep
  their bare, unprefixed name instead of getting a synthesized namespaced alias.
  This removes phantom duplicate columns and fixes several downstream failures: a
  bare subscriber lookup relationship on a namespaced object resolves instead of
  failing "relationship not found"; a bare subscriber formula field used in a SOQL
  `WHERE`, `IN`, `ORDER BY`, or aggregate clause no longer fails with
  "AER_DEC128_UNPACK: expected BLOB, got float64"; and selecting an unmanaged field
  through a managed relationship resolves against the related object's namespace
  rather than the base object's.
- Field references inside a child subquery (for example
  `SELECT (SELECT … FROM Journal_Entries__r)`) are canonicalized against the
  subquery's related object, so a managed formula field referenced bare in a
  subquery no longer reaches SQL generation un-namespaced and fails with a doubled
  `AER_DEC128_UNPACK`.
- A cross-object Number/Currency formula that reads an integer-typed field on a
  related object (for example `Contract.ContractTerm` through a lookup) now reads
  the column directly instead of failing with "AER_DEC128_UNPACK: expected BLOB,
  got int64".
- A custom permission granted to the running user at runtime is honored inside
  `System.runAs(self)` when its `SetupEntityAccess` link was created in the current
  transaction, matching how Salesforce re-derives a user's effective custom
  permissions at the `runAs` boundary for `FeatureManagement.checkPermission` and
  `$Permission`; a link that pre-existed the transaction (loaded from metadata or
  created in `@TestSetup`) is not.
- `Security.stripInaccessible` field security now matches a real org in both
  directions: a top-level `stripInaccessible(READABLE)` no longer keeps fields based
  on a permission set the running user self-granted at runtime (those permissions
  are frozen at the top level and only re-derived inside `System.runAs`), and
  `UPDATABLE` no longer strips a user's own editable personal fields such as
  `Title` and `CompanyName`.
- SOQL `ORDER BY` over a custom metadata type breaks ties by `DeveloperName`,
  matching the canonical order of `<Type>.getAll().values()`, instead of returning
  tied rows in load order.
- Deleting the `CampaignMemberStatus` whose `IsDefault` is true is rejected (a
  failed `DeleteResult` under `allOrNone = false`, a thrown exception otherwise),
  matching Salesforce; deleting the parent Campaign still cascades through its
  statuses.
- A modified clone compares equal to an otherwise-identical record under `==`,
  `assertEquals`, and in `Set`/`Map`/`List` operations: the synthetic clone-tracking
  markers behind `isClone()`/`getCloneSourceId()` are excluded from SObject equality
  and hashing.

## v1.0.12 — 2026-06-16

- Strict SObject field-name resolution: an unresolved member no longer gets a
  fabricated canonical name, fixing a formula that traverses a namespaced
  managed-package relationship to an unmanaged field failing with "No such
  column". Aliasing a bare (non-aggregate) field in SOQL is now rejected ("only
  aggregate expressions use field aliasing"); a SOQL alias that is not a schema
  field is preserved in query results, JSON, and `SObject.get`.
- `User.UserType` is seeded (`Standard`), so querying it directly or through a
  relationship or formula returns a value, and `UserInfo.getUserId()` no longer
  crashes in anonymous exec or cache-hydrated runs.
- `ConnectApi` output objects compare equal by structure under `==`, in Sets, and
  as Map keys instead of by reference; `toString()` is namespace-qualified.
- `Entitlement.Status` is derived from `StartDate`/`EndDate` at insert
  (Active/Expired/Inactive) and is read-only — assigning it is a compile error.
- `SObject.getSObjects()` throws `System.SObjectException` for an invalid
  child-relationship name instead of returning null.
- `Label.get` and `Label.translationExists` accept a null namespace, the
  idiomatic way to reference an unmanaged custom label.
- A platform-event subscriber trigger that throws no longer fails the publishing
  transaction: the subscriber's DML rolls back and the publisher continues.
- `Lead.Status` defaults are applied at insert, not instantiation, so
  `new Lead()` and clones taken before insert have null `Status`.
- A typed-null array argument such as `(SObject[]) null` reports the declared
  `List<SObject>` parameter type to a stub, and array-form arguments select the
  matching `List<T>` overload.
- `Decimal` division keeps 33 significant digits (HALF_UP), fixing exact-equality
  comparisons over divided Decimals such as `1/31`.
- `new Map<Object, SObject>(list)` keys each record by its Id (previously produced
  an empty map).
- A field literally named `Type` (`Case.Type`, `Account.Type`, `Opportunity.Type`)
  resolves to its `SObjectField` token rather than `System.Type`.
- `Flow.Interview.<FlowName>.class` literals resolve to `System.Type`.
- Classes that use triple-quoted text blocks are stored in the parse cache instead
  of being re-parsed on every run.

## v1.0.11 — 2026-06-10

- Seeded the Customer Community Plus Login user license and profile, so projects
  that query or `runAs` this profile resolve it without explicit metadata.

## v1.0.10 — 2026-06-08

- Custom metadata records from `getAll()` / `getInstance()` resolve their field
  tokens correctly and are read-only; writing through `put()` or assignment throws
  an uncatchable `System.FinalException`. Constructed, cloned, and SOQL-queried
  records remain writable in memory.
- A permission-set assignment that grants a permission set to the running user is
  honored in a test only when inserted inside a `System.runAs` block (or once a
  `Test.startTest()` boundary is crossed), matching Salesforce, so
  custom-permission gates are no longer wrongly bypassed.
- Formula fields using `SUBSTITUTE` and `FIND` now translate to SQL instead of
  failing with "failed to convert formula to SQL" when referenced in a WHERE
  clause or SELECT list.
- Nested formula dependencies across relationships keep the full relationship
  path, fixing "relationship X not found" on deep formula chains.
- `UserRecordAccess` returns a no-access row (`HasReadAccess = false`,
  `HasEditAccess = false`) for record Ids that were never inserted, matching the
  virtual object, instead of returning zero rows.
- `Date.valueOf(String)` tolerates trailing characters after the date (for example
  the CPQ-serialized `2020-01-16 00:00:00:00`), applying Salesforce's lenient and
  strict parsing rules.
- DML `insert` of a concrete platform event throws and directs you to
  `EventBus.publish`; a generic `List<SObject>` of events publishes them through
  the event bus, and `Database.insert` publishes and returns successful results.
- Bare `BusinessHours`, `Site`, `Network`, and `Domain` resolve to the SObject
  ahead of the same-named System class (only `System.`-qualified names get the
  class), fixing `SObject.get`/`put` with field tokens on `new BusinessHours()`.
- `new BusinessHours(field = value)` creates the SObject and keeps its field
  assignments.
- Standard locale, state/country, timezone, language, and email-encoding picklists
  now use the full catalogues, with country-aware state resolution and the
  state-controlled-by-country dependency.

## v1.0.9 — 2026-06-05

- Custom metadata record filenames may include the `__mdt` suffix in the type
  segment (for example `Binding_Config__mdt.Account.md-meta.xml`) without
  producing a doubled `__mdt__mdt` type.

## v1.0.8 — 2026-06-04

- Added the standard `Order.OpportunityId` lookup, `Lead.DoNotCall`, and the Case
  Entitlement-management fields (`EntitlementId`, the SLA date fields, `IsStopped`,
  `MilestoneStatus`), with their reciprocal child relationships.

## v1.0.7 — 2026-06-02

- `Messaging.sendEmail` validates recipient addresses (raising
  `INVALID_EMAIL_ADDRESS` for malformed entries) and accepts `List<Id>` recipients
  such as User Ids.
- Flow `emailSimple` recipients are split on commas and semicolons before sending,
  fixing record-triggered flows that pass a joined recipient string.
- `FinalizerContext.getException()` reports the exception the way Salesforce does:
  user exceptions are detached (line -1, "External entry point"), a terminating
  `LimitException` becomes a `System.AsyncException`, and built-in exceptions keep
  their real source line.
- The email-template Visualforce controller is reset around each render so a stale
  controller cannot leak across emails or across tests sharing a pooled VM.

## v1.0.6 — 2026-05-31

- Email folder metadata in MDAPI (metadata) format is now loaded, so email
  templates resolve their `FolderName`.
- CPU governor timing was recalibrated to within roughly 1.1–1.7x of Salesforce
  (per-method and constructor cost, first-reference class-loading charge, and
  O(1) `Set` add/contains).

## v1.0.5 — 2026-05-30

- Constructor names are validated at compile time: a mismatched or `static`
  "constructor" is rejected instead of being silently coerced into a real
  constructor.
- `SetupOwnerId` is available on List custom settings, defaults to the
  Organization on insert, and rejects a Profile or User owner with
  `FIELD_INTEGRITY_EXCEPTION`.
- A failed all-or-none `upsert` that rolls back no longer nulls the in-memory Id of
  the existing records it updated.
- An async job whose error text exceeds the `ExtendedStatus` limit now records the
  failure (the job moves to Failed) instead of staying stuck in Processing.
- Before-save record-triggered flow fixes: spanning relationship references are
  traversed through each hop, flow-assigned fields are persisted on update, and
  before-save flows now run on bulk insert.

## v1.0.4 — 2026-05-28

- Queued platform events are delivered in a server-managed transaction after
  execute-anonymous and async-job commits, so event-trigger DML no longer leaves
  storage in an active transaction and blocks async workers.

## v1.0.3 — 2026-05-28

- The result type of a null-coalescing (`??`) expression is inferred, fixing
  false "Illegal assignment" errors such as `String s = (someString ?? '') + 5;`.

## v1.0.2 — 2026-05-28

- A relationship populated by an external Id that matches no parent now fails DML
  with `INVALID_FIELD` instead of silently inserting an unresolved foreign key,
  and `Database.upsert` defaults to all-or-none like `insert` and `update`.
- A scheduled job run during `Test.stopTest()` matches its cron expression in the
  user's timezone, fixing day-boundary skips.
- The type checker rejects unknown custom SObject and platform-event types,
  non-`static` `@IsTest`/`testMethod` methods, and class names containing two
  consecutive underscores.
- Namespaced field references that exceed a field's nominal length no longer fail
  `STRING_TOO_LONG`.

## v1.0.1 — 2026-05-27

- Hierarchy fields such as `Account.ParentId` are treated as self-referencing
  lookups, fixing "Illegal assignment from Account to String" on external-Id
  relationship assignment.
- Geolocation component fields (`__Latitude__s` / `__Longitude__s`) inherit their
  field-level security from the compound field, fixing USER_MODE queries that
  failed with "No such column".

## v1.0.0 — 2026-05-26

- Initial 1.0 release.
