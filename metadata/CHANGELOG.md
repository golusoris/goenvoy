# Changelog

## [1.3.0](https://github.com/golusoris/goenvoy/compare/metadata/v1.2.0...metadata/v1.3.0) (2026-04-15)


### Features

* add 8 new service modules ([1dcd7bc](https://github.com/golusoris/goenvoy/commit/1dcd7bce5cc9cb04f5a14a7add647d1f0603f4e0))
* add Letterboxd API client (metadata/movie/letterboxd) ([773f4f2](https://github.com/golusoris/goenvoy/commit/773f4f2865d1da7820e0253c2f9584c6ff2452c2))
* bootstrap monorepo with category base modules ([ffd24fc](https://github.com/golusoris/goenvoy/commit/ffd24fc22b716b41b4a8d27a2ab80950f2c825b7))
* **metadata/anidb:** add AniDB HTTP XML API client ([072b76f](https://github.com/golusoris/goenvoy/commit/072b76f8da406c55458d23711762f6c1762625f4))
* **metadata/anilist:** add AniList GraphQL API client ([70bbb90](https://github.com/golusoris/goenvoy/commit/70bbb90267931b0ee5cb2d23c195924d3548bf52))
* **metadata/fanart:** add Fanart.tv API v3 client ([f9ee9e0](https://github.com/golusoris/goenvoy/commit/f9ee9e07ad2c51819feab8338adae47217540073))
* **metadata/kitsu:** add Kitsu JSON:API client ([5686576](https://github.com/golusoris/goenvoy/commit/56865760fb482d42e814f11adf5691333704bd9f))
* **metadata/mal:** add MyAnimeList API v2 client ([a6f331d](https://github.com/golusoris/goenvoy/commit/a6f331d17e0a8dc3e8ee68e160169160fb43b393))
* **metadata/omdb:** add OMDb API client ([1c83650](https://github.com/golusoris/goenvoy/commit/1c8365051a3ba52a5d5c013fc0e66e67d2eb4c2c))
* **metadata/tmdb:** add TMDb v3 API client ([3904e95](https://github.com/golusoris/goenvoy/commit/3904e9513d2ac0fe9f86ad73da76b2b08a352bd5))
* **metadata/tvdb:** add TheTVDB API v4 client ([93ce501](https://github.com/golusoris/goenvoy/commit/93ce501bd98c5671a77a035e7e496e285996bd4f))
* **metadata/tvmaze:** add TVmaze API client ([32b038c](https://github.com/golusoris/goenvoy/commit/32b038c26a957e5b7583926c3ce2dfe44e8bb18e))
* **metadata:** add shared BaseClient, migrate all 27 providers, restructure movie→video ([0026baa](https://github.com/golusoris/goenvoy/commit/0026baa54634fa25c00f67d9387e8013e5d70a6e))


### Bug Fixes

* BaseClient.Delete body support, URL injection, default timeouts ([59203c0](https://github.com/golusoris/goenvoy/commit/59203c0eda4c747e89c6567b91811e0cdee46e5b))
* name return values to satisfy golangci-lint unnamedResult ([9e9a97e](https://github.com/golusoris/goenvoy/commit/9e9a97e407f0b534ac82966d6d8da9cfc6e7949b))
* pass large structs by pointer to satisfy gocritic hugeParam ([b141cca](https://github.com/golusoris/goenvoy/commit/b141cca7f636ad055441e7b2b7b23b800999605a))
* resolve golangci-lint CI errors across 6 packages ([1ef79c1](https://github.com/golusoris/goenvoy/commit/1ef79c17f6c98a701f5001405a7686fb35ddc3b9))
* TMDb query/path injection, Kavita token race condition ([8be265a](https://github.com/golusoris/goenvoy/commit/8be265a4753653a2960c4e1bb4534efa44cd82c4))


### Code Refactoring

* restructure metadata into categories + fix error handling ([902861c](https://github.com/golusoris/goenvoy/commit/902861c594e4aad2ae8e249064082fa496be3f5a))
