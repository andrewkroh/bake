bake
====

An attempt at simplifying the building of Go projects across Unix and Windows
development environments.

gvm
---

Go version management. This is for bootstrapping Go once you have a
release of `bake` downloaded. This installs Go and sets GOROOT and PATH.

bash:

`eval "$(./bake gvm 1.7.4)"`

windows cmd.exe:

`FOR /f "tokens=*" %i IN ('"bake.exe" gvm 1.7.4') DO %i`

windows powershell.exe:

`bake gvm --powershell 1.7.4 | Invoke-Expression`

Or using the project's Go version. For example:

`eval "$(bake gvm --project-go)"`

info
----

Get project info.

`--project-go`

This determines the Go version that a project uses by reading the Go
version defined in the projects `.travis.yml` file.

`bake info --go-version`

`--go-files`

This provides a list of the project's Go files minus any vendored dependencies.
This is useful with commands like `goimports -w -l $(bake info --go-files)`.

`--go-packages`

This provides a list of the project's packages minus any vendored dependencies.
This is useful with commands like `go test $(bake info --go-packages)`.

check
-----

Run checks on the project.

`bake check fmt`

This checks that all `.go` files (except vendored dependencies) in the project
are formatted according to `gofmt -s`.
