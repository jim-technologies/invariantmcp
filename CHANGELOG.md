# Changelog

## Unreleased

- Upgrade `golang.org/x/net` to 0.60.0 to fix the reachable HTTP/2
  vulnerabilities found by the October 9 audit.
- Use the root Invariant Protocol v0.16.7 module and generated service
  registration. CLI, MCP and HTTP use the current context-aware projection
  lifecycle; MCP initialization reports the framework's standard identity.
- Pin the protobuf registrars and Go 1.27.2 toolchain and make `help` the
  default Make target.
- Generate bindings from the runtime descriptor without managed-option drift,
  check freshness in the ordinary gate, and add a secretless weekly audit.

The conversion request/response fields and configuration formats are unchanged.
Reverting the source and dependency commit restores the previous toolchain.
