# Auth Design: Why JTI Instead of Storing the Full Access Token

## What is JTI?

JTI stands for **JWT ID** — a standard claim (`jti`) defined in [RFC 7519](https://www.rfc-editor.org/rfc/rfc7519#section-4.1.7). It is a unique identifier (UUID v4) embedded inside the JWT payload. It identifies *this specific token issuance*, not the user.

Example access token payload for this project:

```json
{
  "jti": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "sub": "42",
  "role": "admin",
  "name": "Alice",
  "email": "alice@example.com",
  "exp": 1748000000,
  "iat": 1747998200
}
```

`jti` is what we store in `users.access_token_jti`. Not the full token.

---

## Why not store the full token?

Three reasons.

### 1. The token is already self-validating

The middleware already has the full token from the HTTP-only cookie. HMAC-SHA256 signature
verification proves the token was issued by us and was not tampered with — no database
needed for that step. The database lookup only needs to answer one question:
**"has this token been revoked?"**

A 36-character UUID is sufficient to answer that. Storing the full 300–500 character
token in the database adds nothing — the middleware already has it.

### 2. Index performance

We hit `users.access_token_jti` on **every authenticated request** to check revocation.
A UUID index (36 chars) is substantially smaller and faster than a full JWT index
(300–500 chars). At scale this matters on a hot index.

### 3. Keeping secret material out of the database

If the database is compromised:

| What's in DB | What the attacker gets |
|---|---|
| Full JWT token | An immediately usable credential — valid for up to 30 minutes |
| JTI (UUID only) | A UUID — useless without `JWT_SECRET` to forge a signature |

The JWT secret never touches the database. The token itself never touches the database.
Only its identifier does.

---

## The validation flow

```
Request arrives with access_token cookie
  │
  ├─ 1. Decode JWT locally (no DB) — verify HMAC-SHA256 signature with JWT_SECRET
  │       Invalid signature → 401 TOKEN_INVALID (fast, no DB hit)
  │
  ├─ 2. Check exp claim — is the token expired?
  │       Expired → 401 TOKEN_EXPIRED (no DB hit — saves a query on the common case)
  │
  ├─ 3. SELECT access_token_jti, locked_at FROM users WHERE id = sub
  │
  ├─ 4. Compare DB jti to JWT jti claim
  │       Mismatch → 401 TOKEN_INVALID (token was revoked/rotated)
  │
  └─ 5. Check locked_at IS NULL
          Locked → 423 ACCOUNT_LOCKED
          OK → attach AuthUser to context, proceed
```

Expiry is checked before the DB lookup intentionally — most rejections will be
expired tokens on the hot path, and this avoids a database query for them.

---

## Refresh token uses a different approach

The **refresh token** is an opaque random 32-byte hex string — not a JWT. Because it
carries no claims and is never decoded, there is no `jti` to extract. Instead:

- `SHA-256(token)` is stored in `users.refresh_token_hash`
- On validation: compute `SHA-256(cookie value)` and compare to the stored hash

**Why SHA-256 and not bcrypt?**

bcrypt is designed for low-entropy secrets like passwords — it is deliberately slow to
resist brute-force attacks. A refresh token is already a 32-byte cryptographically
random value (256 bits of entropy). Brute-force is computationally impossible regardless
of hash speed. Using bcrypt here would add ~100ms of CPU cost to every refresh request
for no security benefit. SHA-256 is the right tool.

---

## Why not Redis / a token blacklist?

A blacklist stores revoked JTIs in Redis until their expiry, supporting multiple
concurrent sessions per user. This is a common production pattern but adds operational
complexity — another dependency, cache invalidation logic, TTL management, and cache
miss fallback.

For this project, a single `access_token_jti` column in `users` is the right trade-off:
simpler to operate, one valid session per user at a time, revoked immediately on logout
or rotation. If multi-session support becomes a requirement, the column approach can be
migrated to a sessions table without changing the middleware contract.

---

## Summary

| Token | What's stored in DB | Type | Why |
|---|---|---|---|
| Access token | `access_token_jti` (UUID, 36 chars) | Column in users | Self-validating JWT; only revocation status needed |
| Refresh token | `refresh_token_hash` (SHA-256 hex, 64 chars) | Column in users | Opaque token; hash prevents usable value exposure |
