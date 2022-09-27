# artifact-action

This is a GitHub Action and CLI tool which provides tooling for working with
[Ace Archive](https://acearchive.lgbt). This tool has three functions:

- Querying the archive to retrieve artifact metadata, including metadata for
  previous versions of artifacts.
- Optionally validating the syntax of artifact files.
- Optionally re-hosting the content in the archive on the IPFS network.

For background on how artifacts in the archive are stored and what an artifact
file is, you may want to check out
[acearchive/artifacts](https://github.com/acearchive/artifacts).

This action is used by
[acearchive/artifacts](https://github.com/acearchive/artifacts) to upload all
contributed content to [Web3.Storage](https://web3.storage), but it can be used
by anyone to help host the content on Ace Archive on the IPFS network. This
action could be used with any repository, as long as the artifact files conform
to the same schema.

This tool produces JSON output containing artifact metadata, including the
[CID](https://docs.ipfs.io/concepts/content-addressing/) of each file
associated with the artifact, which you can use to retrieve the content over
either the IPFS or HTTP protocols.

This action supports uploading content to Web3.Storage or any pinning service
that supports the [IPFS pinning service
API](https://ipfs.github.io/pinning-services-api-spec/).

## Inputs

### `path`

The path of the directory in the repository containing the artifact files.

### `mode`

The mode to operate in, either `validate` or `history`.

- In `validate` mode, artifact files are pulled from the working tree and their
  syntax is validated. If any artifact file in the working tree has invalid
  syntax, the action fails.
- In `history` mode, the entire commit history is traversed to pull each
  version of each artifact file, and syntax errors are ignored silently.

`validate` mode is useful for performing status checks on pull requests to
ensure submitted artifact files are valid and for uploading new artifact files
when commits are pushed or pull requests are merged.

This mode also creates a new UnixFS directory containing links to all the files
in artifacts in the repository, which will be uploaded to Web3.Storage or an
IPFS pinning service if the necessary parameters are set (see below).

`history` mode is useful for re-hosting content from the archive in bulk.
Because IPFS uses content-based addressing, links to files don't always
necessarily point to the latest version of that file. To ensure that old links
never go dead, it's prudent to not just host the content *currently* in Ace
Archive, but all the content that's *ever* been in Ace Archive.

Keep in mind that, by default,
[actions/checkout](https://github.com/actions/checkout) only fetches one
commit, so for `history` mode, you'll want to set `fetch-depth: 0` in its input
parameters to fetch the entire commit history (see examples below).

`history` mode does not validate artifact files beyond ensuring that they are
valid YAML. If they are not valid YAML, they are skipped silently. This is for
two reasons:

1. An error in a past version of an artifact file that is fixed in a subsequent
   commit should not cause the action to fail.
2. We may not support validation for previous artifact schema versions (see
   [acearchive/artifacts](https://github.com/acearchive/artifacts) for more
   information about schema versions).

### `ipfs-api`

The multiaddr of the API endpoint of the running IPFS node. This is required to
upload artifacts to either Web3.Storage or an IPFS pinning service. This is
also required to build a UnixFS directory containing the current version of
each file.

### `w3s-token`

The secret API token for Web3.Storage. If this is provided, all artifacts in
the repository are uploaded to Web3.Storage.

### `pin-endpoint`

The URL of the IPFS pinning service API endpoint to use. If this is provided,
all artifacts in the repository are pinned using this pinning service.

### `pin-token`

The secret bearer token for the configured IPFS pinning service. This is
required to pin artifacts using an IPFS pinning service.

## Web3.Storage

To upload content to Web3.Storage, an IPFS node must be running and you must
pass in your Web3.Storage API token and the mutiaddr of the IPFS node's API
endpoint. There's an example below for running an IPFS node in a GitHub Actions
workflow. The action is smart enough to skip any files already uploaded to your
Web3.Storage account in a previous run.

## Pinning services

To pin content with an IPFS pinning service, you must specify the API endpoint
of the pinning service and your bearer token. Note that pinning services that
have their own API may use a separate endpoint for the standardized pinning
service API. For example, the endpoint for [Pinata](https://www.pinata.cloud/)
is `https://api.pinata.cloud/psa`. The action is smart enough to skip any files
already pinned to your account in a previous run.

## Output

This tool produces two JSON outputs:

- `artifacts` is JSON document describing all the artifacts in the repository.
- `cids` is a JSON array containing a deduplicated set of all the CIDs
  contained in artifacts in the repository.

### `cids`

The `cids` output is provided for convenience if you just want to retrieve all
the content in the archive and don't need artifact metadata. In this list, CIDs
are deduplicated by their multihash, so if the repository contains a v0 CID and
a v1 CID with the same multihash, only one will be returned.

The `cids` output will always return the CIDs for all artifacts in the
repository, even through schema version changes.

### `artifacts`

The `artifacts` output looks like the example below. It contains the following
fields:

- `rootCid`: The CID of the UnixFS directory containing the current version of
  each file in the repository. If we're not running in `validate` mode, or if
  the `ipfs-api` input was not provided, then this will be `null`.
- `artifacts`: An array of all the artifacts in the repository.
  - `path` is the relative path of the artifact file from the root of the
    repository.
  - `slug` is the URL slug of the artifact, which is the file name of the
    artifact file without the file extension.
  - `commit` is the commit the artifact file was pulled from. In `validate`
    mode, this field is always `null`.
    - `commit.rev` is the commit hash.
    - `commit.date` is the author date in RFC 3339 format, normalized to UTC.
  - `entry` contains the actual contents of the artifact file, except as JSON
    instead of YAML. If a list value is omitted in the artifact file, it's
    serialized in the JSON output as `[]`. If a scalar value is omitted, it's
    serialized as `null`.

```json
{
  "rootCid": "bafybeiabf334qobjurmsp6kjytj2k4ociyor56oomdxhlwt4zvi64prmti",
  "artifacts": [
    {
      "path": "artifacts/orlando-the-asexual-manifesto.md",
      "slug": "orlando-the-asexual-manifesto",
      "commit": {
        "rev": "b9e7dc442ad8bb2ec30311825cb276179130bfde",
        "date": "2022-05-11T15:11:22Z"
      },
      "entry": {
        "version": 3,
        "title": "*The Asexual Manifesto*",
        "description": "A paper by the Asexual Caucus of the New York Radical Feminists\n",
        "longDescription": null,
        "files": [
          {
            "name": "Digital Scan",
            "mediaType": "application/pdf",
            "filename": "the-asexual-manifesto.pdf",
            "cid": "bafybeihsf4562gmmyoya7eh5buxv65lqcdoil3wsi5jf5fceskap7yzooi"
          },
          {
            "name": "Transcript",
            "mediaType": "text/html",
            "filename": "the-asexual-manifesto-transcript.html",
            "cid": "bafybeib2fu4qf44xiyduvhadog5raukc3ajdnd4qpsavyxaa2umzjeif5y"
          }
        ],
        "links": [
          {
            "name": "Internet Archive",
            "url": "https://archive.org/details/asexualmanifestolisaorlando"
          }
        ],
        "people": [
          "Lisa Orlando",
          "Barbara Getz"
        ],
        "identities": [
          "asexual"
        ],
        "fromYear": 1972,
        "toYear": null,
        "decades": [
          1970
        ],
        "aliases": []
      }
    }
  ]
}
```

## CLI

In addition to being available as a GitHub action, this tool provides a CLI. To
use the CLI, you must clone the Ace Archive repository yourself.

To use the CLI, you must first install [Go](https://go.dev/).

To run the CLI and see the help:

```
go run . --help
```

```
Host content from Ace Archive on the IPFS network.

To upload content to Web3.Storage, you must specify `--ipfs-api` and
`--w3s-token`.

To pin content with an IPFS pinning service, you must specify `--ipfs-api`,
`--pin-endpoint`, and `--pin-token`.

The multiaddr of your local IPFS node is most likely
`/dns/localhost/tcp/5001/http` by default.

Usage:
  artifact-action [flags]

Flags:
  -h, --help                 help for artifact-action
      --ipfs-api multiaddr   The multiaddr of your IPFS node
  -m, --mode string          The mode to operate in, either "validate" or "history" (default "validate")
  -o, --output string        The output to produce, either "artifacts", "cids", or "summary" (default "summary")
      --path path            The path of the artifact files in the repository (default "artifacts/")
      --pin-endpoint url     The url of the IPFS pinning service API endpoint to use
      --pin-token token      The secret bearer token for the configured IPFS pinning service
  -r, --repo path            The path of the git repo containing the artifact files (default ".")
      --w3s-token token      The secret API token for Web3.Storage
```

## Examples

Get the JSON output for the current version of each artifact.

```yaml
jobs:
  archive:
    name: "Get current artifacts"
    runs-on: ubuntu-latest
    steps:
      - name: "Checkout"
        uses: actions/checkout@v2
        with:
          repository: "acearchive/artifacts"
      - name: "Get artifacts"
        id: get_artifacts
        uses: acearchive/artifact-action@v0.3.0
      - name: "Do something with the artifacts"
        run: "echo ${{ steps.get_artifacts.outputs.artifacts }}"
```

Get the JSON output for all the artifacts in the history of the repo.

```yaml
jobs:
  archive:
    name: "Get all artifacts"
    runs-on: ubuntu-latest
    steps:
      - name: "Checkout"
        uses: actions/checkout@v2
        with:
          repository: "acearchive/artifacts"
          fetch-depth: 0
      - name: "Get artifacts"
        id: get_artifacts
        uses: acearchive/artifact-action@v0.3.0
        with:
          mode: history
      - name: "Do something with the artifacts"
        run: "echo ${{ steps.get_artifacts.outputs.artifacts }}"
```

Validate the artifact files in the working tree and upload the files to
Web3.Storage.

```yaml
jobs:
  archive:
    name: "Validate and upload curent artifacts"
    runs-on: ubuntu-latest
    services:
      ipfs-node:
        image: "ipfs/go-ipfs:latest"
        ports:
          - 4001:4001
          - 5001:5001
          - 8080:8080
    steps:
      - name: "Checkout"
        uses: actions/checkout@v2
        with:
          repository: "acearchive/artifacts"
      - name: "Upload artifacts"
        uses: acearchive/artifact-action@v0.3.0
        with:
          ipfs-api: "/dns/ipfs-node/tcp/5001/http"
          w3s-token: ${{ secrets.W3S_API_TOKEN }}
```

Pin all the files in the history of the repo with Pinata.

```yaml
jobs:
  archive:
    name: "Upload all artiacts"
    runs-on: ubuntu-latest
    services:
      ipfs-node:
        image: "ipfs/go-ipfs:latest"
        ports:
          - 4001:4001
          - 5001:5001
          - 8080:8080
    steps:
      - name: "Checkout"
        uses: actions/checkout@v2
        with:
          repository: "acearchive/artifacts"
          fetch-depth: 0
      - name: "Upload artifacts"
        uses: acearchive/artifact-action@v0.3.0
        with:
          mode: history
          ipfs-api: "/dns/ipfs-node/tcp/5001/http"
          pin-endpoint: "https://api.pinata.cloud/psa"
          pin-token: ${{ secrets.PINATA_API_TOKEN }}
```
