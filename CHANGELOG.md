# Changelog

## [2.0.0](https://github.com/kastelo/ezapt/compare/v1.1.0...v2.0.0) (2025-08-20)


### ⚠ BREAKING CHANGES

* use config file, pool directory

### Features

* use config file, pool directory ([17605ab](https://github.com/kastelo/ezapt/commit/17605ab5dc7890d97bb3b58ff289c2379ee59485))


### Bug Fixes

* add ending newlines to clear/detached signatures ([2cd886a](https://github.com/kastelo/ezapt/commit/2cd886a75c9508d3486ece5a999fca91f4671f35))
* keep codename debian to avoid warning ([112b03f](https://github.com/kastelo/ezapt/commit/112b03f18c653863d0e1d73541895979bff9a651))

## [1.1.0](https://github.com/kastelo/ezapt/compare/v1.0.2...v1.1.0) (2024-12-16)


### Features

* separate sign command, keyring from env ([20ac295](https://github.com/kastelo/ezapt/commit/20ac295cb39b8cde1517e70b530e1915e67f1b1c))
* sign --detach --ascii ([6cbd3cc](https://github.com/kastelo/ezapt/commit/6cbd3cc219008a2aa2500cf8deadd159c2b8e888))

## [1.0.2](https://github.com/kastelo/ezapt/compare/v1.0.1...v1.0.2) (2024-11-26)


### Bug Fixes

* use updated OpenPGP implementation ([16ef709](https://github.com/kastelo/ezapt/commit/16ef70953bb3e4c9992ca17e46fdf26f7190ec25))

## [1.0.1](https://github.com/kastelo/ezapt/compare/v1.0.0...v1.0.1) (2024-11-24)


### Bug Fixes

* properly exit with code 1 on error ([68f9013](https://github.com/kastelo/ezapt/commit/68f90135c4d0bbf43e3dad0b8521768427839d25))
* use base image that allows running as root ([6594991](https://github.com/kastelo/ezapt/commit/6594991de665280947979ff6c51d277a0b07b5c2))
* use rename instead of link when adding ([2f8fb8c](https://github.com/kastelo/ezapt/commit/2f8fb8c2b88821a3a0c98d00de4c31daf829d23c))

## 1.0.0 (2024-11-24)


### Features

* command to add packages ([1c9a163](https://github.com/kastelo/ezapt/commit/1c9a1630fbc42824145971a3f7c1732eb7749d86))
* sign using multiple keys ([eafb30c](https://github.com/kastelo/ezapt/commit/eafb30c14be8693770adfac57ebe59acd6f9278f))
* use internal OpenPGP implementation ([014738c](https://github.com/kastelo/ezapt/commit/014738c1053704783377d0fc52448322f279d70f))
