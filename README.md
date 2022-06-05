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

## Modes

This action has two modes of operation, which you specify via the `mode` input
parameter. The default mode is `validate`.

### Validate mode

In `validate` mode, artifact files are pulled from the working tree of the
repository and their syntax is validated. If any artifact file in the working
tree has invalid syntax, the action fails.

This mode is useful or performing status checks on pull requests to ensure
submitted artifact files are valid and for uploading new artifact files when
commits are pushed or pull requests are merged.

### History mode

Sometimes, the contents of an artifact file changes. For example, a file
containing a transcript might be replaced with a more accurate one. However,
because IPFS uses content-based addressing, links to files don't always
necessarily point to the latest version of that file. To ensure that old links
never go dead, it's prudent to not just host the content *currently* in Ace
Archive, but all the content that's *ever* been in Ace Archive. Because
artifact files are version controlled using git, we can do this fairly easily.

In `history` mode, the commit history of the repository is traversed and each
version of each artifact file is pulled from the commit history. However, in
this mode, invalid artifact files are skipped silently. Otherwise, an invalid
artifact file that is committed to the repository and then fixed in a
subsequent commit would cause the action to fail, which we don't want.

This mode is useful for hosting artifacts from the archive in bulk, including
previous versions of artifact files that are no longer in the working tree.
Keep in mind that, by default,
[actions/checkout](https://github.com/actions/checkout) only fetches one
commit, so you'll want to set `fetch-depth: 0` in its input parameters to fetch
the entire commit history (see examples below).

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

To upload content to Web3.Storage, you must specify `--w3s-token` and
`--ipfs-api`.

To pin content with an IPFS pinning service, you must specify `--pin-endpoint`
and `--pin-token`.

Usage:
  artifact-action [flags]

Flags:
  -h, --help                    help for artifact-action
      --ipfs-api multiaddr      The multiaddr of your IPFS node (default "/dns/localhost/tcp/5001/http")
  -m, --mode string             The mode to operate in, either "validate" or "history" (default "validate")
  -o, --output string           The output to produce, either "artifacts", "cids", or "summary" (default "summary")
      --path string             The path of the artifact files in the repository (default "artifacts")
      --pin-endpoint endpoint   The URL of the IPFS pinning service API endpoint to use
      --pin-token token         The bearer token for the configured IPFS pinning service
  -r, --repo path               The path of the git repo containing the artifact files (default ".")
      --w3s-token token         The secret API token for Web3.Storage
```

## Output

This tool produces two JSON outputs:

- `artifacts` is JSON document describing all the artifacts in the repository.
- `cids` is a JSON array containing a deduplicated list of all the CIDs
  contained in artifacts in the repository.

The `cids` output is provided for convenience if you just want to retrieve all
the content in the archive and don't need artifact metadata. In this list, CIDs
are deduplicated by their multihash, so if the repository contains a v0 CID and
a v1 CID with the same multihash, only one will be returned.

The `artifacts` output looks like the example below. It contains an array of
objects with the following fields:

- `path` is the relative path of the artifact file from the root of the
  repository.
- `slug` is the URL slug of the artifact, which is the file name of the
  artifact file without the file extension.
- `commit` is the commit the artifact file was pulled from. In `validate` mode,
  this field is always `null`.
  - `commit.rev` is the commit hash.
  - `commit.date` is the author date in RFC 3339 format, normalized to UTC.
- `entry` contains the actual contents of the artifact file. It mirrors the
  [schema of artifact
  files](https://acearchive.lgbt/docs/contributing/artifact-files/), except as
  JSON instead of YAML. If a list value is omitted in the artifact file, it's
  serialized in the JSON output as `[]`. If a scalar value is omitted, it's
  serialized as `null`.

```json
{
  "artifacts": [
    {
      "path": "artifacts/orlando-the-asexual-manifesto.md",
      "slug": "orlando-the-asexual-manifesto",
      "commit": {
        "rev": "b9e7dc442ad8bb2ec30311825cb276179130bfde",
        "date": "2022-05-11T15:11:22Z"
      },
      "entry": {
        "version": 1,
        "title": "\u003cem\u003eThe Asexual Manifesto\u003c/em\u003e",
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
            "filename": null,
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

## Examples

### Just get the JSON output (validate mode)

```yaml
jobs:
  archive:
    name: "Upload artifacts"
    runs-on: ubuntu-latest
    steps:
      - name: "Checkout"
        uses: actions/checkout@v2
        with:
          repository: "acearchive/artifacts"
      - name: "Get artifacts"
        id: get_artifacts
        uses: acearchive/artifact-action@v0.1.0
      - name: "Do something with the artifacts"
        run: "echo ${{ steps.get_artifacts.outputs.artifacts }}"
```

### Just get the JSON output (history mode)

```yaml
jobs:
  archive:
    name: "Upload artifacts"
    runs-on: ubuntu-latest
    steps:
      - name: "Checkout"
        uses: actions/checkout@v2
        with:
          repository: "acearchive/artifacts"
          fetch-depth: 0
      - name: "Get artifacts"
        id: get_artifacts
        uses: acearchive/artifact-action@v0.1.0
        with:
          mode: history
      - name: "Do something with the artifacts"
        run: "echo ${{ steps.get_artifacts.outputs.artifacts }}"
```

### Upload to Web3.Storage

```yaml
jobs:
  archive:
    name: "Upload artiacts"
    runs-on: ubuntu-latest
    services:
      ipfs:
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
        uses: acearchive/artifact-action@v0.1.0
        with:
          w3s-token: ${{ secrets.W3S_API_TOKEN }}
          ipfs-api: "/dns/ipfs/tcp/5001/http"
```

### Pin with a pinning service

```yaml
jobs:
  archive:
    name: "Upload artiacts"
    runs-on: ubuntu-latest
    steps:
      - name: "Checkout"
        uses: actions/checkout@v2
        with:
          repository: "acearchive/artifacts"
      - name: "Upload artifacts"
        uses: acearchive/artifact-action@v0.1.0
        with:
          pin-endpoint: "https://api.pinata.cloud/psa"
          pin-token: ${{ secrets.PINATA_API_TOKEN }}
```
