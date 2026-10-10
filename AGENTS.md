# AGENTS.md - three finger drag

## Project

This program provides a three finger drag utility for Linux systems. The utility works on Arch Linux and CachyOS. It replicates the exact drag behavior found in macOS. Users can move windows, select text, and drag files using three fingers without pressing down on the trackpad.

## Graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

Rules:
- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).

## Writing

- Load the `ste-writing` skill before your first reply.
- Follow its chat mode for every reply.
- Follow its applicable prose mode when writing or editing documentation.

## Commands

Use standard Go commands for formatting, testing and benchmarking. In addition use the Makefile for extra specific commands.

+ Run: go run
+ Unit tests: go test ./...
+ Integration tests: go test -tags=integration ./...
+ Lint/format: go fmt ./...

## Style and convensions

+ Go version: 1.27.1+
+ Code language: English
+ Code comments: English
+ Internal errors: English

## Rules

+ Read docs/constitution.md and the active spec before touching any code.
+ Update whenever a file or directory is added the tree in the [README.md](./README.md#project-structure) (e.g. `tree -I 'graphify-out'`).
+ When implementing code or documentation on either plan or build mode feel free to ask any clarifying questions only if they are needed to successfully implement.
+ After writing go code or protos run `make fmt` to avoid any code formatting or lint issues.
<!--+ <Limits: what not to touch, what not to add before asking>-->

## When completing any task

+ If go code was modified, run tests
