# SE Deadlines Importer Roadmap

## Status

Proposed feature roadmap. This document captures the current direction and the
decisions that must be approved before the formal backend feature spec and TDD
implementation plan are created.

## Goal

Build an idempotent Go importer for the conference dataset maintained at
<https://github.com/se-deadlines/se-deadlines.github.io>.

For every import run, the application will clone the upstream project into a
temporary directory, parse `_data/conferences.yml`, normalize and enrich its
records, and persist compatible events and deadlines in ReSEARCH Events. Later
runs must detect existing imported records and synchronize changes instead of
creating duplicates.

The importer must preserve the project's existing rules:

- Event and deadline changes are audited.
- Deadlines are immutable: a changed deadline creates a replacement and
  deactivates the old row.
- Imports are repeatable and safe to resume.
- Ambiguous source data is reported for human review rather than guessed.
- A dry run shows intended changes without modifying the database.

## Current Compatibility Assessment

The upstream dataset is partially compatible with the current database, but
records cannot be inserted directly.

Fields that map directly or with simple conversion:

| Upstream | ReSEARCH Events | Adaptation |
|---|---|---|
| `name` | `events.name` | Direct mapping |
| `year` | `events.year` | Validate against the parsed start date |
| `link` | `events.website_url` | Direct mapping |
| Deadline date/time | `deadlines.date`, `deadlines.time` | Split the timestamp |
| AoE deadline convention | `deadlines.timezone` | Store `AoE` |

Fields requiring normalization or enrichment:

| Upstream | ReSEARCH Events | Problem |
|---|---|---|
| `date` | `start_date`, `end_date` | Source is free-form text |
| `place` | `city`, `country` | Source is free-form and sometimes incomplete |
| No coordinates | `latitude`, `longitude` | Coordinates are mandatory locally |
| No slug | `slug` | Must be generated with collision handling |
| No domain | `domain` | Initially infer `computer_science` |
| No tier | `tier` | Default to `unranked` |
| No contributor | `created_by_id` | Use a dedicated importer system user |
| Deadline meaning | `deadlines.type` | Source timestamps usually have no explicit type |

Fields that cannot currently be preserved:

- Conference descriptions: the local event model has no description column.
- Notes: the local event model has no notes column.
- Venue and track tags: the local model has no normalized taxonomy.
- Import provenance: there is no source record identity or synchronization
  metadata.
- Journal and virtual entries: the current event schema requires a physical
  location and coordinates.

## Proposed Source Acquisition

Each run should:

1. Create a unique temporary directory using the operating system's temporary
   directory mechanism.
2. Shallow-clone the default branch of the upstream repository.
3. Record the resolved commit SHA.
4. Read `_data/conferences.yml` and `_data/types.yml`.
5. Perform parsing, validation, normalization, enrichment, and persistence.
6. Remove the temporary clone when the command finishes, including after an
   error when possible.

The source repository URL and branch should be configurable, which allows tests
to use a local fixture repository without network access.

If an upstream record disappears, the importer should not delete the local
event automatically. It should mark the source record as missing and include it
in the import report for administrator review.

## Proposed Source Identity and Provenance

The upstream dataset does not provide a stable record ID. A generated identity
is therefore required.

Initial source key proposal:

```text
sha256(normalized name + year + canonical URL)
```

Because all three values can change, matching should occur in stages:

1. Exact stored source key.
2. Exact canonical URL match within the same source.
3. Exact normalized name and year match.
4. Possible fuzzy match reported for manual review; never update automatically.

The importer should store enough provenance to synchronize safely:

- Source name and repository URL.
- Source key.
- Upstream commit SHA.
- Original source payload or its checksum.
- First-imported and last-synchronized timestamps.
- Last-seen upstream timestamp.
- Synchronization state such as active, missing, invalid, or needs review.
- Link to the local event when one has been created or matched.

Potential manually submitted duplicates must be reported. They must not be
linked to an upstream record or overwritten without administrator approval.

## Proposed Ownership and Update Rules

Imported fields and community-maintained fields need explicit ownership so a
sync does not erase reviewed local corrections.

Proposed rules:

- Record the upstream payload independently from the local event.
- On first import, populate all safe mapped fields from upstream.
- On later imports, compare the previous upstream snapshot with the new one.
- Automatically update a local field only if it has not been changed locally
  since the previous synchronization.
- If both upstream and a local user changed the same field, record a conflict
  for administrator review.
- Never overwrite community-maintained coordinates with a lower-confidence
  geocoding result.
- Every applied change writes an AuditLog entry.
- Use a dedicated system user, such as
  `se-deadlines-importer@system.local`, as the audit actor.

Newly imported records should initially be created as `pending`. Automatic
approval of trusted future synchronizations may be added as an explicit,
configurable policy after the importer has been validated in production.

## Coordinate and Location Enrichment

### Normalization pipeline

1. Trim whitespace and normalize punctuation.
2. Split known place forms into city, optional region, and country.
3. Normalize country names to a consistent representation.
4. Check a local location cache using the normalized place.
5. Query a geocoding provider only on a cache miss.
6. Store the result, provider metadata, and confidence.
7. Route ambiguous or low-confidence results to manual review.

Examples:

- `Dublin, Ireland` becomes city `Dublin`, country `Ireland`.
- `Oakland, California, United States` becomes city `Oakland`, region
  `California`, country `United States`.
- `Rio de Janeiro` is ambiguous in the source representation. A provider may
  infer Brazil, but the importer must record that inference and its confidence.

### Geocoding abstraction

Use a provider-neutral interface so the importer is not coupled to one vendor.
Nominatim is the proposed first implementation.

The interface should support:

- A normalized query containing city, region, and country.
- Latitude and longitude.
- Provider name and provider result ID.
- Confidence or match quality.
- Provider-normalized city, region, and country.
- Raw response metadata needed for diagnostics, without sensitive data.

Nominatim-specific behavior must respect its rate limit and usage policy. Calls
should be serialized or throttled, identify the application with an appropriate
User-Agent, retry only transient errors with backoff, and rely on the local
cache to avoid repeated queries.

Provider failures must not abort the entire import. The record remains staged
with a failure reason and can be resumed later.

## Event Date Adaptation

The source stores event dates as free-form strings. The parser should recognize
formats incrementally, including:

- `October 5, 2026`
- `April 25 - May 1, 2027`
- `July 5 - 9, 2026`
- Values with suffixes such as `with ICSE 2026` or `co-located with FSE`

Rules:

- A single date maps to identical start and end dates.
- A range maps to separate start and end dates.
- If the first part omits a month, inherit the end date's month.
- If the first part omits a year, inherit the explicit year.
- Remove only recognized contextual suffixes; preserve the original text in
  provenance.
- Never fabricate a date for `TBD`, a missing date, or an unrecognized format.
- Compare the parsed start year with the upstream `year` and report conflicts.
- Prefer the parsed start year for the local `events.year` only after the
  conflict policy is approved.

Records without a usable event date remain staged and are not inserted into the
current `events` table.

## Journal and Locationless Records

The upstream dataset includes journal special issues that may have neither an
event date nor a physical location. Supporting them properly likely requires:

- An `event_kind` field, initially conference, workshop, or journal.
- Nullable event location for journal or virtual entries.
- A clear rule for whether start/end dates may be absent.

Recommended direction: expand the model so the platform can represent these
records accurately. Until that schema is approved, locationless or dateless
records should remain in staging rather than receive invented values.

## Descriptions, Notes, and Tags

Recommended schema capabilities:

- `events.description` for the upstream long name or description.
- `events.notes` for source notes.
- `events.event_kind` for conference, workshop, journal, and future kinds.
- Normalized `tags` and `event_tags` tables for venue and track taxonomy.

Unknown upstream tags should be retained and reported as warnings rather than
blocking the record. This prevents data loss when upstream adds a taxonomy value
before ReSEARCH Events knows how to display it.

## Deadline Adaptation and Synchronization

For each upstream deadline:

- Parse `YYYY-MM-DD HH:MM` into local date and time fields.
- Store timezone as `AoE` according to the upstream convention.
- Preserve source order in provenance.
- Use `type=other` unless the source provides enough unambiguous information to
  select abstract, paper, notification, or camera-ready.
- Use a meaningful fallback description such as `Submission deadline`.
- Treat `TBA` and rolling `%y`/`%Y` templates as staged values until explicit
  product behavior is approved.

On synchronization:

- An unchanged deadline remains unchanged.
- A changed deadline uses the existing supersede workflow: insert a new active
  row and deactivate the old row.
- A newly added upstream deadline creates a new active deadline.
- A removed upstream deadline is marked inactive with an audit record; it is
  never physically deleted.
- Ambiguous matching among multiple deadlines must be reported rather than
  resolved by position alone.

## Slugs and Duplicate Handling

Generate a slug from the normalized conference name and event year. If that
slug conflicts, add a stable short suffix derived from the source key.

Duplicate detection should report:

- Exact source-key matches.
- Exact canonical-URL matches.
- Exact normalized name/year matches.
- Same name/year with different URLs.
- Likely matches against manually submitted events.

Automatic merging is allowed only for an existing record already linked to the
same upstream source identity. Every other possible duplicate requires review.

## Import Command

Proposed command:

```bash
cd backend
go run ./cmd/import-se-deadlines --dry-run
go run ./cmd/import-se-deadlines --apply
go run ./cmd/import-se-deadlines --apply --resume <run-id>
```

Initial modes:

- `--dry-run`: clone, parse, normalize, enrich from cache where possible, and
  report changes without database writes or paid/external geocoding unless
  explicitly enabled.
- `--apply`: persist staging results and apply safe creates/updates.
- `--resume`: continue an interrupted run without repeating completed records.

An HTTP endpoint is intentionally out of scope for the first version. A local
or scheduled CLI reduces authentication and operational risk.

## Transactions, Idempotency, and Recovery

- Create an import-run record before processing source records.
- Process each conference in its own database transaction so one invalid record
  does not roll back the entire dataset.
- Store per-record status and error details.
- Use source keys and checksums to make rerunning the same commit idempotent.
- A second run of the same unchanged commit must produce zero event/deadline
  mutations.
- Retry transient clone, geocoder, and database failures with bounded backoff.
- Do not retry deterministic YAML, validation, or parsing errors automatically.
- Support continuation from the last safely completed record.
- Preserve the original payload or checksum for diagnostics and conflict
  detection.

## Import Report

Every run should produce a human-readable and machine-readable summary:

- Upstream repository and commit SHA.
- Run ID, start/end time, and execution mode.
- Total records discovered.
- Created, updated, unchanged, skipped, missing, and failed counts.
- Records awaiting location, date, duplicate, or conflict review.
- Deadlines added, superseded, deactivated, and unchanged.
- Geocoder cache hits, provider calls, ambiguous results, and failures.
- Warnings for year/date mismatches and unknown tags.

The command should return a non-zero exit status for source-wide or
infrastructure failures. Individual invalid records should be reported but
should not necessarily fail the entire run.

## Security and Operational Requirements

- Never execute scripts or code from the cloned repository.
- Read only the expected data files.
- Clone into a unique, bounded temporary directory.
- Set clone and geocoding timeouts.
- Limit YAML input size and record counts to prevent accidental resource
  exhaustion.
- Validate URLs and all parsed values before persistence.
- Do not log database credentials or geocoder secrets.
- Follow the upstream repository's license and attribution requirements. The
  current repository does not visibly include a license file, so redistribution
  permission must be clarified before production import.

## Development Phases

Each development phase below should become one or more independently approved
Red-Green-Refactor cycles in the formal implementation plan. The scope is
intentionally small enough to complete incrementally over multiple days.

### Day 1 — Source acquisition

- Define the source-fetcher interface.
- Implement safe shallow clone into a temporary directory.
- Resolve and record the commit SHA.
- Guarantee cleanup.
- Add tests using a local Git fixture.

Deliverable: the command can acquire a source snapshot without parsing it.

### Day 2 — YAML parsing

- Define raw upstream conference and tag types.
- Parse `conferences.yml` and `types.yml`.
- Validate required source fields.
- Retain raw values for diagnostics.
- Report YAML and structural errors.

Deliverable: a typed, validated in-memory source dataset.

### Day 3 — Event-date parsing

- Implement pure parsers for single dates and date ranges.
- Handle recognized co-location suffixes.
- Detect `TBD`, missing, malformed, and year-conflicting values.
- Add table-driven tests from real upstream examples.

Deliverable: normalized start/end dates or an explicit review reason.

### Day 4 — Place normalization

- Parse common city/country and city/region/country forms.
- Normalize country aliases.
- Represent ambiguous and incomplete places explicitly.
- Add table-driven tests using real source values.

Deliverable: normalized location queries without external network calls.

### Day 5 — Coordinate provider and cache

- Define the provider-neutral geocoder interface.
- Add the location-cache model and repository.
- Implement Nominatim with throttling, identification, timeouts, and bounded
  retries.
- Define confidence thresholds and manual-review results.

Deliverable: repeatable coordinates with provider metadata and caching.

### Day 6 — Import staging and provenance schema

- Add import runs and per-source-record tables.
- Store source identity, checksum, commit, status, and local event link.
- Add conflict/review reason representation.
- Test migration constraints and repositories.

Deliverable: source records can be staged without creating public events.

### Day 7 — Event schema extensions

- Add approved description, notes, event-kind, taxonomy, and nullable-location
  changes.
- Update models, validation, API responses, and frontend types as required.
- Preserve behavior for existing events.

Deliverable: the local model can represent upstream data without avoidable loss.

This day depends on product approval of the schema decisions.

### Day 8 — New-event import

- Build normalized event values using pure service functions.
- Generate collision-safe slugs.
- Create or resolve the importer system user.
- Insert pending events transactionally.
- Write creation audit records.

Deliverable: valid staged records can create pending local events.

### Day 9 — Initial deadline import

- Convert timestamps to date, time, and AoE timezone.
- Apply safe deadline-type and description defaults.
- Insert deadlines and audit entries.
- Handle multiple and unsupported deadline values.

Deliverable: newly imported events include their compatible deadlines.

### Day 10 — Existing-record synchronization

- Match exact source identities.
- Compare previous source snapshots to new snapshots.
- Detect local-versus-upstream field conflicts.
- Apply only safe importer-owned updates.
- Write field-level audit diffs.

Deliverable: changed upstream events update safely and unchanged runs are
idempotent.

### Day 11 — Immutable deadline synchronization

- Match existing imported deadlines.
- Add new deadlines.
- Supersede changed deadlines.
- Deactivate removed deadlines without deletion.
- Report ambiguous deadline matches.

Deliverable: upstream deadline changes preserve complete local history.

### Day 12 — Duplicate and manual-review workflow

- Detect likely matches with manually submitted events.
- Record date, location, tag, and ownership conflicts.
- Provide a review report or administrative integration contract.
- Prevent automatic merges outside exact linked source identities.

Deliverable: ambiguous records are actionable and never silently merged.

### Day 13 — CLI orchestration and reporting

- Connect acquisition, parsing, staging, enrichment, and persistence.
- Implement dry-run, apply, and resume modes.
- Add text and JSON reports.
- Define meaningful process exit codes.

Deliverable: an operator can run and inspect a complete import safely.

### Day 14 — End-to-end verification and operations

- Run integration tests against a test PostgreSQL database and fixture Git
  repository.
- Verify rerun idempotency.
- Verify recovery from partial failure.
- Test representative malformed and ambiguous records.
- Document local execution, scheduled execution, geocoder policy, monitoring,
  and rollback procedures.

Deliverable: the importer is ready for a controlled first production dry run.

## Decisions Required Before the Formal Spec

1. Should newly imported records remain pending, or may trusted records be
   automatically approved?
2. When an upstream URL changes, may normalized name/year matching reconnect
   the record automatically, or must an administrator approve it?
3. May upstream synchronization overwrite fields previously edited by a local
   user?
4. Is Nominatim approved as the first coordinate provider?
5. What geocoding confidence threshold requires manual review?
6. Should journal and virtual entries be supported through nullable locations,
   or skipped in the first version?
7. Are description, notes, event kind, tags, and provenance schema additions
   approved?
8. Should unknown tags be stored with a warning or block the record?
9. Should a deadline with no unambiguous semantic type default to `other`?
10. Should an upstream-removed event remain visible, become pending, or only be
    flagged for review?
11. Who owns and reviews the dedicated importer system user?
12. Do we have permission from the upstream maintainers to import and
    redistribute their dataset?
13. Are there any edge cases not covered by this roadmap?

## Next Step

Review and answer the decisions above. Once the rules are confirmed, convert
this roadmap into the formal backend spec under `specs/backend/`, run the
project's five-gate spec approval checklist, and obtain explicit `spec approved`
confirmation before implementation planning or production code begins.
