# Delivery contract (v2)

A producer writes a delivery package. Publisher validates it and reads its assets.
Tool and executable spec: [delivery.py](delivery.py). If this document and the
script disagree, fix both and bump `contract_version`.

## Package

```
deliveries/<episode_id>/
  manifest.json
  video/   <episode_id>_<variant>.mp4
  text/    <lang>.json, timings.json
```

Asset paths are **relative to the manifest's folder**. Package video and text assets inside that folder. Publisher supports a read-only compatibility exception for the reserved pilot ID `catsoup-poc-streetlight`.

## manifest.json

Use [manifest.template.json](manifest.template.json) to author a package. For a working example, run the [synthetic demo generator](../../apps/publisher/support/create_demo.py) after [setup](https://publisher.meruw.dev/guides/setup/). It creates `deliveries/catsoup-demo/manifest.json` with ES and EN video files and `publishable: false`.

| Field | Content |
|---|---|
| `contract`, `contract_version` | `"catsoup-delivery"`, `2` |
| `episode_id` | `catsoup-<series>-<nn>-<slug>`; equals the folder name |
| `version` | episode version (`v1`, `v2`…); goes up with every export that changes a video |
| `exported_on`, `languages` | ISO date; e.g. `["es", "en"]` |
| `status` | episode status (below) |
| `publishable`, `not_publishable_reason` | `false` requires a reason |
| `review` | `{by, date, notes, files}`: `files` maps every item id to the sha256 that was reviewed |
| `approval` | `{by, date, files}`: same, for the approval |
| `videos[]` | `id`, `lang`, `kind` (`final` / `clean`), `narration`, `burned_text`, `status`, `path`, `sha256`, `measured{duration_s, resolution, fps, audio}` |
| `texts[]` | `id`, `lang`, `path`, `sha256` |
| `source` | `project`, `episode`, `blender`, `command`, `provenance`, `notes` |
| `history[]` | events such as voided reviews or approvals (written by `complete`) |

`sha256` and `measured` are written by `delivery.py complete`; never by hand.
There must be at least one `final` video per language. `clean` versions
(no burned-in text) are delivered when useful for platform subtitles or reuse.

## Statuses

| Status | Meaning | Set by | Requires |
|---|---|---|---|
| `exported` | Files generated and technically checked | producer | `validate` passes |
| `reviewed` | Watched and listened to by a person, with notes | owner: `delivery.py review --by NAME` | `review.files` equal to the current sha256 of every item |
| `approved_for_publishing` | Authorised to publish | owner: `delivery.py approve --by NAME` | a matching review, `publishable: true`, `approval.files` equal to the current sha256 |

- Generating or regenerating a file **never** raises a status. When a file
  changes, `complete` voids the review and/or approval that covered the old
  version, records it in `history` and drops the status back.
- `validate` prints two verdicts:
  - **VALID**: structure, files, hashes and measurements are consistent.
    Publisher may ingest and preview a valid manifest in any status (useful
    for development with the PoC).
  - **PUBLISHABLE**: valid, `status: approved_for_publishing` and
    `publishable: true`. Only then may Publisher publish.
  `validate --require-publishable` exits 1 unless publishable.
