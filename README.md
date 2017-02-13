bake
====

gvm
---

Go version management. This installs and sets GOROOT and PATH.

bash:

`eval "$(.\bake gvm 1.7.4)"`

windows cmd.exe:

`FOR /f "tokens=*" %i IN ('".\bake.exe" gvm 1.7.4') DO set %i`
