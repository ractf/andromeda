# Andromeda

![Build Release Badge](https://github.com/ractf/andromeda/workflows/Build%20Release/badge.svg) ![Build Container Badge](https://github.com/ractf/andromeda/workflows/Build%20Container/badge.svg)

Andromeda is a simple Docker container manager for CTFs.

## Usage

Build all the challenge containers:

```bash
andromeda build -f [challenge dir]
```

Run the server:

```bash
andromeda start -f [challenge dir] -b [server ip] -a [api ip:port]
```

## Config

Each challenge should be stored in its own subdirectory and have a challenge.json that looks like this:

```json
{
  "name": "challenge",
  "port": 3000,
  "mem_limit": 512,
  "user_limit": 100
}
```

This config file should be stored in a directory with the challenges docker file and other relevant files.
.
