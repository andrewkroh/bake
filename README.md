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

This determines the Go version that a project uses by reading the Go
version defined in the projects `.travis.yml` file.

`bake info --go-version`
