# Editorconfig guesser

This attempts to produce a valid Editorconfig file for a project it puts some work into trying to guess what the current
project configurations are. However, the generated config should be checked.

The algorithm used might change at a PRs notice. The one in place is good enough for now happy to take submissions.

The program is aware of:
* `.` hidden unix files (it avoids them)
* `.gitignore` files

Currently, all the supported file formats only support the most generic `editorconfig` arguments; as per https://editorconfig.org/. 
Happy to accept PRs that expand the scope of particular formats to; 

# Usage:

`ecguess` `[-save]` `[-verbose]` `[directories]`

By Pipe if you want to see the output without having to open the file individually

```bash
$ ecguess . | tee .editorconfig
```

To save it without viewing it (or using shell piping / output redirection) use:
```bash
$ ecguess . --save
```

# Support file formats

Currently:
* `*.ts;*.js`  - [Generic](fileformats/generic)
* `*.cpp;*.h;*.c`  - [Generic](fileformats/generic)
* `*.py`  - [Generic](fileformats/generic)
* `*.go;go.mod;go.sum` - [Custom](fileformats/go)
* `Makefile;*.mak` - [Custom](fileformats/gnumake)

Happy to accept PRs for more.

