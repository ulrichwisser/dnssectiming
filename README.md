[![GPLv3 License](https://img.shields.io/badge/License-GPL%20v3-yellow.svg)](https://www.gnu.org/licenses/gpl-3.0.en.html)

# dnssectiming

Find DNSSEC timing parameters for domains

# STATUS

This project is a work in progress!
Currently many parts are under construction.

Part 1 - Resolving domains and saving data to database: initial working version 
Part 2 - Extracting timing data: TBD

# Copyright

All code is licensed under [![GPLv3 License](https://img.shields.io/badge/License-GPL%20v3-yellow.svg)](https://www.gnu.org/licenses/gpl-3.0.en.html).
Artwork is licensed under [![Creative Commons BYNC-ND 4.0](https://i.creativecommons.org/l/by-nc-nd/4.0/80x15.png)](http://creativecommons.org/licenses/by-nc-nd/4.0/)


# Contributing

Contributions are always welcome!

Please note that all submission must be licensed under [![GPLv3 License](https://img.shields.io/badge/License-GPL%20v3-yellow.svg)](https://www.gnu.org/licenses/gpl-3.0.en.html).

Easiest way to contribute is via pull-request, open an issue or contact the author.

## Configuration

Configuration is done in a config file in YAML format.
Files `~/.dnssect` is loaded followed by `./.dnssect.`
Currently only the MySQL DSN can be given in the config file. (see example config file)

### Command Line Arguments

|            |    | Description |
|------------|----|----------------------------------------------------------------------------|
|--verbose   | -v | increase the level of verbosity (1=error,2=warnings,3=info,4=debug)
|--resolvers | -r | ip address of a resolver (can be given several times)
|--concurrent| -c | number of concurrent resolver threads

# Compiling for Synology NAS

```
sudo docker run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:alpine ./syno.sh
```
