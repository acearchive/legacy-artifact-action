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

Out of the box, this action supports uploading content to Web3.Storage, but the
JSON output could be used to write CI tooling for uploading content to any IPFS
pinning service.

To upload content to Web3.Storage, an IPFS node must be running and you must
pass in your Web3.Storage API token and the mutiaddr of the IPFS node's API
endpoint. The action is smart enough to skip any files already uploaded to your
Web3.Storage account in a previous run.

## Output

The JSON output of this action looks like this. It mirrors the schema of
artifact files, with the addition of the `slug` field containing the URL slug
of the artifact. Fields which are optional in the artifact file are serialized
as `null` in the JSON output.

```json
{
  "artifacts": [
    {
      "slug": "orlando-the-asexual-manifesto",
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
