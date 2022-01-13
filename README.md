# artifact-action

This repository is a GitHub Action which provides tooling for hosting content
contributed to [Ace Archive](https://acearchive.lgbt).

This action parses a repository for [artifact
files](https://acearchive.lgbt/docs/contributing/artifact-files/), validates
their syntax, and outputs a JSON document containing metadata about each
artifact, including the [IPFS
CID](https://docs.ipfs.io/concepts/content-addressing/) of each file.

This action is used by
[acearchive/acearchive.lgbt](https://github.com/acearchive/acearchive.lgbt) to
upload all contributed content to [Web3.Storage](https://web3.storage), but it
can be used by anyone to help host the content on Ace Archive on the IPFS
network. This action can also be used with forks of the repository as long as
the format of the artifact files is the same.

Out of the box, this action supports uploading content to Web3.Storage or any
pinning service that supports the [IPFS pinning service
API](https://ipfs.github.io/pinning-services-api-spec/), but the JSON output
could be used to write CI tooling for hosting the content anywhere. It's
possible to upload content to both Web3.Storage and a pinning service by
specifying all the necessary input parameters.

## Previous versions of artifacts

Sometimes, the contents of an artifact file change. For example, a
transcription might be replaced with a more accurate one. However, because IPFS
uses content-based addressing, links to files don't always necessarily point to
the latest version of that file. To ensure that these links never go dead, it's
prudent to not just host the content *currently* in Ace Archive, but all the
content that's *ever* been in Ace Archive. Because artifact files are version
controlled using git, we can do this fairly easily.

Out of the box, this action traverses the git commit history and parses every
version of every artifact file in the history of the repository. You can
control how much of the history you want to include by configuring the
`fetch-depth` input of the
[actions/checkout](https://github.com/actions/checkout) action. The default
behavior is to only fetch a single commit, but you can set `fetch-depth: 0` to
fetch the entire commit history.

Something important to note is that when this action validates the syntax of
artifact files, it only validates the syntax for the current (HEAD) versions.
If a previous version of an artifact file has invalid syntax, it is just
skipped silently.

## Web3.Storage

To upload content to Web3.Storage, an IPFS node must be running and you must
pass in your Web3.Storage API token and the mutiaddr of the IPFS node's API
endpoint. The action is smart enough to skip any files already uploaded to your
Web3.Storage account in a previous run.

## Pinning services

To pin content with an IPFS pinning service, you must specify the API endpoint
of the pinning service and your bearer token. Note that pinning services that
support the standardized API may use a separate endpoint for it. For example,
the endpoint for [Pinata](https://www.pinata.cloud/) is
`https://api.pinata.cloud/psa`. The action is smart enough to skip any files
already pinned to your account in a previous run.

## Output

The JSON output of this action looks like this. It mirrors the schema of
artifact files, with the addition of the following fields:

- `slug` contains the URL slug of the artifact
- `rev` contains the git commit hash of the commit the artifact file was pulled
  from

Fields which are optional in the artifact file are serialized as `null` in the
JSON output.

```json
{
  "artifacts": [
    {
      "slug": "orlando-the-asexual-manifesto",
      "rev": "43470f27477e20154be40a6cb3f8ee444ffc0467",
      "title": "<em>The Asexual Manifesto</em>",
      "description": "A paper by the Asexual Caucus of the New York Radical Feminists",
      "longDescription": null,
      "files": [
        {
          "name": "The Asexual Manifesto",
          "mediaType": "application/pdf",
          "filename": "the-asexual-manifesto.pdf",
          "cid": "bafybeihsf4562gmmyoya7eh5buxv65lqcdoil3wsi5jf5fceskap7yzooi"
        },
        {
          "name": "The Asexual Manifesto (Transcript)",
          "mediaType": "text/html",
          "filename": "the-asexual-manifesto-transcript.html",
          "cid": "bafkreie5hknsonewqxuyf6vzlauhn2qwm2og5yjcqrltv5yumyqdvdm4sm"
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
      ]
    }
  ]
}
```

## Examples

### Just get the JSON output

```yaml
jobs:
  archive:
    name: "Upload artifacts"
    runs-on: ubuntu-latest
    steps:
      - name: "Checkout"
        uses: actions/checkout@v2
        with:
          repository: "acearchive/acearchive.lgbt"
      - name: "Get artifacts"
        id: get_artifacts
        uses: acearchive/artifact-action@main
      - name: "Upload artifacts"
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
          repository: "acearchive/acearchive.lgbt"
      - name: "Upload artifacts"
        uses: acearchive/artifact-action@main
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
          repository: "acearchive/acearchive.lgbt"
      - name: "Upload artifacts"
        uses: acearchive/artifact-action@main
        with:
          pin-endpoint: "https://api.pinata.cloud/psa"
          pin-token: ${{ secrets.PINATA_API_TOKEN }}
```
