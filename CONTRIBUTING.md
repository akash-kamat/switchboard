# Contributing

Open an issue before a large behavior or configuration change. Keep pull requests
focused and preserve schema/API compatibility unless the change includes a
documented migration.

Required local checks:

```sh
go test ./...
go vet ./...
node --check internal/server/web/app.js
```

Use `gofmt` for Go files. Update user documentation and `CHANGELOG.md` with
user-visible changes. Do not add generated release artifacts to Git. Platform
changes should pass the manually triggered `Native platform integration`
workflow; packaging changes should pass `Linux package lifecycle`.

By contributing, you agree that your contribution is licensed under the MIT
license in this repository.
