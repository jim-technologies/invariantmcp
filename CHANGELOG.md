# Changelog

## Unreleased

## v0.1.0 — 2026-10-11

First tagged release; `VERSION` records it and the README installs it.

- Upgrade `golang.org/x/net` to 0.60.0 to fix the reachable HTTP/2
  vulnerabilities found by the October 9 audit.
- Use the root Invariant Protocol v0.16.7 module and generated service
  registration. CLI, MCP and HTTP use the current context-aware projection
  lifecycle; MCP initialization reports the framework's standard identity.
- Pin the protobuf registrars and Go 1.27.2 toolchain and make `help` the
  default Make target.
- Generate bindings from the runtime descriptor without managed-option drift,
  check freshness in the ordinary gate, and add a secretless weekly audit.
- Flox now supplies Go 1.27.2 itself (`go_1_27`). Removed the
  `GOTOOLCHAIN=go1.27.2+auto` variable, which downloaded that toolchain at run
  time on top of a Flox Go 1.26.4 bootstrap.

The conversion request/response fields and configuration formats are unchanged.
Reverting the source and dependency commit restores the previous toolchain.
