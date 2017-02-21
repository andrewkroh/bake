bake
====

Usage
-----

```
usage: bake [<flags>] <command> [<args> ...]

Utility for working with Beats projects

Flags:
  -h, --help   Show context-sensitive help (also try --help-long and --help-man).
  -d, --debug  Enable debug logging

Commands:
  help [<command>...]
    Show help.


  test [<flags>] [<packages>]
    Run tests.

    --cover           Generate code coverage output and HTML report
    --race            Enable race detector while testing
    --junit           Generate JUnit XML report summarizing test results
    --tests=unit ...  Test types to execute. Options are unit (default), benchmark, integ, and system.

  crosscompile
    Cross-compile the beat without CGO


  docs
    Build the Elastic asciidoc book for the Beat


  ci
    Run all checks and tests.


  check [<checks>...]
    Run checks on the project. The options are fmt, vet, and notice. By default all checks are run.


  fmt
    Run gofmt -s on non-vendor Go files


  notice [<flags>] [<dirs>...]
    Create a NOTICE file containing the licenses of the project's vendored dependencies.

    -b, --beat="Elastic Beats"  Beat name
    -c, --copyright="Elasticsearch BV"  
                                Copyright owner
    -y, --year=2014             Copyright begin year
    -o, --output=NOTICE         Output file

  docker [<flags>] [<script>]
    Start test services powered by Docker and open a shell on the host where environment variables point to services.

    -p, --project=PROJECT  Specify an alternate project name (default: directory name)
    -f, --file=docker-compose.yml ...  
                           Specify an alternate compose file (default: docker-compose.yml)
    -o, --log=LOG          Specify log output file
```

