# Agent Tooling

Windows/WSL helper scripts for agent workflows in this repository.

## Purpose

These scripts are convenience wrappers for cases where the repo lives on a Windows-mounted path and the agent runs from WSL.

They are useful only when Windows interop is available from the current session.

## Scripts

- `tools/agent/doctor`
  - preflight for WSL + Windows interop + Docker Desktop + required `.env` keys
- `tools/agent/win`
  - generic wrapper to run a Windows command from WSL through PowerShell
- `tools/agent/dockw`
  - Docker Desktop wrapper
- `tools/agent/gitw`
  - Windows Git wrapper
- `tools/agent/nodew`
  - Windows Node wrapper
- `tools/agent/npmw`
  - Windows npm wrapper

## Usage Notes

- These scripts do not guarantee that the current Codex session can reach your Windows host.
- They work only when `powershell.exe` and the relevant Windows binaries are exposed in the active WSL shell.
- Run `bash tools/agent/doctor` first before assuming the wrappers are usable.

## Documentation Rule

When these scripts change:
- update their inline usage comments
- update this file if behavior or prerequisites change
- update `README.md` and relevant files under `docs/` if the workflow expectation changes
