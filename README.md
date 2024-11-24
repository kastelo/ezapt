# ezapt

A simple utility to generate a valid APT distribution.

`ezapt` combines the functionality of `dpkg-scanpackages`, `apt-ftparchive`,
`gpg`, and the required glue and scripting to tie them together, with
additional functionality to add new packages and remove outdated ones
automatically.

## Usage

```
Usage: ezapt --dists=STRING --keyring=STRING [flags]

Flags:
  -h, --help               Show context-sensitive help.
      --dists=STRING       Path to dists directory ($EZAPT_DISTS)
      --keep-versions=2    Number of versions to keep ($EZAPT_KEEP_VERSIONS)
      --keyring=STRING     Path to GPG keyring ($EZAPT_KEYRING)
      --add=STRING         Path to packages to add ($EZAPT_ADD)
```

## License

MIT