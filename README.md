# w3s-upload

This repository is a GitHub Action used by
[acearchive/acearchive.lgbt](https://github.com/acearchive/acearchive.lgbt) to
upload contributed content to [Web3.Storage](https://web3.storage).

This action parses the repository for [artifact
files](https://acearchive.lgbt/docs/contributing/artifact-files/) and validates
their syntax. Optionally, it can also extract the CIDs from those artifact
files and upload the content to Web3.Storage. In this latter case, an IPFS node
must be running and you must pass in the multiaddr of the node's API endpoint.

See
[here](https://github.com/acearchive/acearchive.lgbt/blob/main/.github/workflows/validate-artifacts.yml)
for an example of how to use this action to just validate artifact files and
[here](https://github.com/acearchive/acearchive.lgbt/blob/main/.github/workflows/upload-artifacts.yml)
for an example of how to use it to upload the content to Web3.Storage.
