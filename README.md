# ezapt

Simple utility to generate a valid APT distribution.

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