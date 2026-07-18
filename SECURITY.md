# Security policy

Before 1.0, security fixes target the latest released minor version. This policy will name supported major versions after the first stable release.

## Report a vulnerability

Use GitHub's private [security advisory form](https://github.com/pawnkit/pawnkit-core/security/advisories/new). If it is unavailable, contact a maintainer listed in `CODEOWNERS` or the PawnKit organization profile.

Include the affected version or commit, likely impact, and a small reproduction when possible. Do not open a public issue before a fix is available.

## Scope

Security-sensitive input includes:

- arbitrary file content passed to line and position conversion;
- untrusted edit sets passed to `textedit`;
- diagnostic JSON decoded by `protocol`.

A panic or excessive resource use in a long-running editor or CI process is in scope even when Go's memory safety prevents memory corruption.

Core does not execute Pawn code, load AMX bytecode, or spawn subprocesses. Runtime and native plugin issues belong to `goamx` or `pawn-plugin-host`.
