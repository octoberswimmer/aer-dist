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
