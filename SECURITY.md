# Security Policy

## Supported versions

This project is pre-1.0; security fixes land on `main`.

## Reporting a vulnerability

Please **do not** open a public issue for security problems.

Report privately via [GitHub's security advisories][advisories] ("Report a
vulnerability") or by email to **theeazyfit@gmail.com**. Include a description,
reproduction steps, and the impact you foresee.

We aim to acknowledge reports within 72 hours and to agree a disclosure timeline
once the issue is confirmed.

## Scope notes

- Secrets (`.env`, API keys) are gitignored and must never be committed.
- Uploaded document bytes and credentials are never written to logs or returned
  in error messages; please flag any regression here as a security issue.

[advisories]: https://github.com/TheBraveByte/bloom-parser/security/advisories/new
