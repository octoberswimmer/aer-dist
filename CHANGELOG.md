# Changelog

## v1.4 (in development)

### New capabilities

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
