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

The mode to operate in, either `validate`, `history`, or `upload`.

- In `validate` mode, artifact files are pulled from the working tree and their
  syntax is validated. If any artifact file in the working tree has invalid
  syntax, the action fails.
- In `history` mode, the entire commit history is traversed to pull each
  version of each artifact file, and syntax errors are ignored silently.
- In `upload` mode, the entire commit history is traversed to pull each version
  of each artifact file, syntax errors are ignored silently, and files are
  uploaded to Web3.Storage or an IPFS pinning service.

`validate` mode is useful for performing status checks on pull requests to
ensure submitted artifact files are valid.

`history` mode is useful for querying artifact metadata, including previous
versions of artifacts.

`upload` mode is useful for re-hosting content from the archive in bulk. This
mode also creates a UnixFS directory containing links to the latest version of
each file in each artifact in the repository. This mode is smart enough to skip
any files already uploaded/pinned in a previous run.

Because IPFS uses content-based addressing, CIDs of files don't always
necessarily point to the latest version of that file. To ensure that old links
never go dead, it's prudent to not just host the content *currently* in Ace
Archive, but all the content that's *ever* been in Ace Archive.

`history` and `upload` mode do not validate artifact files beyond ensuring that
they are valid YAML. If they are not valid YAML, they are skipped silently.
This is for two reasons:

1. An error in a past version of an artifact file that is fixed in a subsequent
   commit should not cause the action to fail.
2. We may not support validation for previous artifact schema versions (see
   [acearchive/artifacts](https://github.com/acearchive/artifacts) for more
   information about schema versions).

Keep in mind that, by default,
[actions/checkout](https://github.com/actions/checkout) only fetches one
commit, so when using `history` or `upload` mode, you'll want to set
`fetch-depth: 0` in its input parameters to fetch the entire commit history
(see examples below).

### `ipfs-api`

The multiaddr of the API endpoint of the running IPFS node. This is required in
`upload` mode. The examples below show how to configure an IPFS node in a
GitHub Actions workflow.

The multiaddr of your local IPFS node is most likely
`/dns/localhost/tcp/5001/http` by default.

### `w3s-token`

The secret API token for Web3.Storage. In `upload` mode, this causes files to
be uploaded to Web3.Storage.

### `w3s-pin`

If this is `true`, then the Web3.Storage pinning service API will be used in
addition to their CAR upload endpoint for improved performance. Only enable
this if your Web3.Storage account has access to their pinning service API,
which is in closed beta at time of writing. If your account supports this
feature, then enabling it will likely improve performance.

This option is only valid in `upload` mode when `w3s-token` is set.

### `pin-endpoint`

The URL of the IPFS pinning service API endpoint to use. In `upload` mode, this
causes files to be pinned with the configured pinning service.

If you're using the Web3.Storage pinning service API, you should use
`--w3s-token` with `--w3s-pin` instead of `--pin-endpoint` and `--pin-token`.

### `pin-token`

The secret bearer token for the configured IPFS pinning service. This is
required to pin artifacts using an IPFS pinning service.

### `dry-run`

Prevents uploading files when used in `upload` mode. Legal in other modes, but
does nothing. Useful for testing.

## Output

This tool produces three outputs:

- `artifacts` is JSON document describing all the artifacts in the repository.
- `root-cid` is the CID of the UnixFS directory containing the current version
  of each file in the repository.
- `cids` is a JSON array containing a deduplicated set of all the CIDs
  contained in artifacts in the repository.

### `cids`

The `cids` output is provided for convenience if you just want to retrieve all
the content in the archive and don't need artifact metadata. In this list, CIDs
are deduplicated by their multihash, so if the repository contains a v0 CID and
a v1 CID with the same multihash, only one will be returned.

In `history` and `upload` mode, the `cids` output will always return the CIDs
for all artifacts in the history of the repository, even through schema version
changes.

### `root-cid`

The `root-cid` output is just the `rootCid` value from the `artifacts` output,
provided as a separate output to avoid the need to do JSON parsing when this is
the only value you need.

### `artifacts`

The `artifacts` output looks like the example below. It contains the following
fields:

- `rootCid`: The CID of the UnixFS directory containing the current version of
  each file in the repository. If we're not running in `upload` mode, then this
  will be `null`.
- `artifacts`: An array of all the artifacts in the repository.
  - `path` is the relative path of the artifact file from the root of the
    repository.
  - `slug` is the URL slug of the artifact, which is the file name of the
    artifact file without the file extension.
  - `commit` is the commit the artifact file was pulled from. In `validate`
    mode, this field is always `null`.
    - `commit.rev` is the commit hash.
    - `commit.date` is the committer date in RFC 3339 format, normalized to
      UTC.
  - `entry` contains the actual contents of the artifact file, except as JSON
    instead of YAML. If a list value is omitted in the artifact file, it's
    serialized in the JSON output as `[]`. If a scalar value is omitted, it's
    serialized as `null`.

```json
{
  "rootCid": "bafybeibvohqqj434rtvpfwutmnwtdes2qolqvpyiz7oqh7kitnsvf5ufyy",
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
            "filename": "the-asexual-manifesto-transcript",
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

```shell
go run . --help
```

```
Host content from Ace Archive on the IPFS network.

See the README for details.

Usage:
  artifact-action [flags]

Flags:
      --dry-run              Prevents uploading files when used in upload mode
  -h, --help                 help for artifact-action
      --ipfs-api multiaddr   The multiaddr of your IPFS node
  -m, --mode string          The mode to operate in, either "validate", "history", or "upload" (default "validate")
  -o, --output string        Print the given output to stdout instead of summary statistics
      --path path            The path of the artifact files in the repository (default "artifacts/")
      --pin-endpoint url     The url of the IPFS pinning service API endpoint to use
      --pin-token token      The secret bearer token for the configured IPFS pinning service
  -r, --repo path            The path of the git repo containing the artifact files (default ".")
      --w3s-pin              Use the pinning service provided by Web3.Storage
      --w3s-token token      The secret API token for Web3.Storage
```

## Examples

Validate the current version of each artifact and get the JSON output for them.

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
        uses: acearchive/artifact-action@main
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
        uses: acearchive/artifact-action@main
        with:
          mode: history
      - name: "Do something with the artifacts"
        run: "echo ${{ steps.get_artifacts.outputs.artifacts }}"
```

Upload all the files in the history of the repo to Web3.Storage.

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
          fetch-depth: 0
      - name: "Upload artifacts"
        uses: acearchive/artifact-action@main
        with:
          mode: upload
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
        uses: acearchive/artifact-action@main
        with:
          mode: upload
          ipfs-api: "/dns/ipfs-node/tcp/5001/http"
          pin-endpoint: "https://api.pinata.cloud/psa"
          pin-token: ${{ secrets.PINATA_API_TOKEN }}
```
