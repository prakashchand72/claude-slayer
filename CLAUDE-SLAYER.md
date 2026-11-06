# Claude Slayer

Run **Claude Code** on a confidential security engagement without the target's
real identifiers ever reaching the API.

A local proxy sits between Claude Code and the API. On the way out it rewrites
every real identifier - domains, subdomains, IPs, org names, emails, file paths,
cloud buckets, AWS account IDs - to stable fiction, and on the way back it
restores them. You keep working with the real values on your machine; the
upstream only ever sees fiction.

Built on [tanuki](https://github.com/Pigyon/tanuki) (Apache-2.0), with a
cloak-closed leak policy so a hunt is **never hard-blocked** mid-session: a value
the auto-mapper missed is cloaked on the fly instead of refusing the request.

> It is an anonymization/confidentiality tool. It does **not** bypass the API's
> safety systems, and it does not change what Claude can or cannot do - it only
> changes what leaves your machine. Only use it on targets you are authorized to
> test.

## Requirements

- Docker (Docker Desktop, or Colima on macOS - Colima is auto-started)
- [Claude Code](https://claude.com/claude-code) on your `PATH` (optional - the
  proxy still runs without it)

Nothing else. The proxy image is built from source on first run, so no Go
toolchain, no `buildx`, and no prebuilt image are needed.

## Quick start

```sh
git clone https://github.com/prakashchand72/claude-slayer
cd claude-slayer

# start the proxy, register a target (and all its subdomains), launch Claude
./claude-slayer '*.example.com'
```

That one command:

1. starts the container engine (Colima) if it is not already up,
2. builds the proxy image from source the first time, then starts it on
   `127.0.0.1:18080`,
3. registers `example.com` and `*.example.com` as the engagement target,
4. launches Claude Code already wired to the proxy (via `.claude/settings.json`).

Then just use Claude Code as usual. Everything it sends is anonymized.

## Commands

| Command | What it does |
| --- | --- |
| `./claude-slayer example.com` | Start proxy, register the target, launch Claude. The `*.` on a wildcard is optional - `'*.example.com'` and `example.com` are the same. |
| `./claude-slayer stop` | Stop the proxy. Your mappings are kept, so restarting resumes the same fiction. |
| `./claude-slayer reset` | Wipe **all** data - every mapping, every target - for a clean slate. |
| `./claude-slayer reset example.com` | Wipe everything, then register a fresh target and launch. |
| `./claude-slayer build` | Rebuild the proxy image from source (after a `git pull`). |

A target argument accepts a bare domain, a wildcard, or a full URL - the `*.`,
the scheme and any path are stripped, so all of these register `example.com`:

```
example.com        *.example.com        https://example.com/login
```

## Checking it works

With the proxy running:

```sh
docker compose -f claude-slayer.yml exec tanuki /tanuki status   # engagement + mapping count
docker compose -f claude-slayer.yml exec tanuki /tanuki map      # the mapping table (real -> fiction)
docker compose -f claude-slayer.yml logs -f tanuki               # live request log
```

Each forwarded request logs `anonymized=true`. A request is never forwarded with
a real value in it: anything the rewriter misses is cloaked first.

## How targets are registered

- `./claude-slayer <domain>` maps the domain and scopes `*.<domain>`, so every
  subdomain Claude discovers is anonymized automatically.
- Anything else Claude turns up during the hunt - a new domain in tool output, an
  IP, an org name, an AWS account - is detected and mapped on the fly.
- To add more by hand while running:

  ```sh
  docker compose -f claude-slayer.yml exec tanuki /tanuki add other-target.com
  docker compose -f claude-slayer.yml exec tanuki /tanuki add-ip 203.0.113.10
  ```

## Notes

- The proxy binds to `127.0.0.1` only - it is not reachable off your machine.
- Mappings live in a Docker volume (`claude-slayer-data`) and persist across
  `stop`/start. `reset` removes that volume.
- `TANUKI_ON_LEAK=cloak` (the default here) means never hard-block. Set it to
  `refuse` in `claude-slayer.yml` for a strict pre-submission run that stops on
  any unmapped value, or `warn` to forward unchanged and only log.

## Credits & license

Claude Slayer is a fork of [tanuki](https://github.com/Pigyon/tanuki) by Pigyon,
licensed under Apache-2.0. See [`LICENSE`](LICENSE). Changes here: a
cloak-closed default leak policy with a guaranteed residual sweep (never blocks a
hunt), bound-ordering and own-fiction fixes, a self-contained source build, and
the `claude-slayer` one-command wrapper.
