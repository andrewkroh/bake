bake
====

gvm
---

Go version management. This installs and sets GOROOT and PATH.

bash:

`eval "$(./bake gvm 1.7.4)"`

windows cmd.exe:

`FOR /f "tokens=*" %i IN ('".\bake.exe" gvm 1.7.4') DO %i`

Or using the project's Go version as determined by `libbeat/docs/versions.asciidoc` or `.travis.yml`. For example:

`eval "$(./bake gvm --project-go)"`

info
----

Get project info.

`./bake info --go-version`
