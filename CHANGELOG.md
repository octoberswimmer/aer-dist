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
