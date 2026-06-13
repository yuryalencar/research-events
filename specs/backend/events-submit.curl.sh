#!/bin/bash
# Events: Submit
# Spec: specs/backend/events-submit.yaml

BASE_URL="${BASE_URL:-http://localhost:8080}"

# 201 — successful submission, no deadlines, no tier (defaults to "unranked")
echo "==> 201 submit without deadlines"
curl -s -X POST "$BASE_URL/api/v1/events/submit" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "International Conference on Model-Driven Engineering",
    "slug": "MODELS2026",
    "country": "Brazil",
    "city": "Recife",
    "latitude": -8.0476,
    "longitude": -34.8770,
    "start_date": "2026-09-21",
    "end_date": "2026-09-25",
    "website_url": "https://models2026.example.org",
    "domain": "computer_science",
    "submitter": {"name": "Ana Silva", "email": "ana@example.com"}
  }' | jq .

# 201 — successful submission with an explicit tier
echo "==> 201 submit with tier"
curl -s -X POST "$BASE_URL/api/v1/events/submit" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "International Conference on Software Engineering",
    "slug": "ICSE2027TIER",
    "country": "USA",
    "city": "Boston",
    "latitude": 42.3601,
    "longitude": -71.0589,
    "start_date": "2027-05-10",
    "end_date": "2027-05-18",
    "website_url": "https://icse2027.example.org",
    "domain": "computer_science",
    "tier": "A*",
    "submitter": {"name": "John Doe", "email": "john@example.com"}
  }' | jq .

# 201 — successful submission with deadlines
echo "==> 201 submit with deadlines"
curl -s -X POST "$BASE_URL/api/v1/events/submit" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "International Conference on Software Engineering",
    "slug": "ICSE2027",
    "country": "USA",
    "city": "Boston",
    "latitude": 42.3601,
    "longitude": -71.0589,
    "start_date": "2027-05-10",
    "end_date": "2027-05-18",
    "website_url": "https://icse2027.example.org",
    "domain": "computer_science",
    "submitter": {"name": "John Doe", "email": "john@example.com"},
    "deadlines": [
      {"type": "abstract", "description": "Research track abstract", "date": "2026-08-15", "is_optional": true},
      {"type": "paper", "description": "Research track full paper", "date": "2026-08-22", "is_optional": false}
    ]
  }' | jq .

# 201 — past event dates are accepted
echo "==> 201 submit with past dates"
curl -s -X POST "$BASE_URL/api/v1/events/submit" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Past Conference Example",
    "slug": "PASTCONF2020",
    "country": "Brazil",
    "city": "Sao Paulo",
    "latitude": -23.5505,
    "longitude": -46.6333,
    "start_date": "2020-03-01",
    "end_date": "2020-03-05",
    "website_url": "https://pastconf.example.org",
    "domain": "computer_science",
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"}
  }' | jq .

# 400 — missing required fields
echo "==> 400 missing required fields"
curl -s -X POST "$BASE_URL/api/v1/events/submit" \
  -H "Content-Type: application/json" \
  -d '{"name": "Incomplete Event"}' | jq .

# 400 — invalid submitter email
echo "==> 400 invalid email"
curl -s -X POST "$BASE_URL/api/v1/events/submit" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Invalid Email Event",
    "slug": "INVALIDEMAIL2026",
    "country": "Brazil",
    "city": "Recife",
    "latitude": -8.0476,
    "longitude": -34.8770,
    "start_date": "2026-09-21",
    "end_date": "2026-09-25",
    "website_url": "https://example.org",
    "domain": "computer_science",
    "submitter": {"name": "Ana Silva", "email": "not-an-email"}
  }' | jq .

# 400 — invalid website_url (no scheme)
echo "==> 400 invalid website_url"
curl -s -X POST "$BASE_URL/api/v1/events/submit" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Invalid URL Event",
    "slug": "INVALIDURL2026",
    "country": "Brazil",
    "city": "Recife",
    "latitude": -8.0476,
    "longitude": -34.8770,
    "start_date": "2026-09-21",
    "end_date": "2026-09-25",
    "website_url": "models2026.example.org",
    "domain": "computer_science",
    "submitter": {"name": "Ana Silva", "email": "ana@example.com"}
  }' | jq .

# 400 — tier not in allowed enum
echo "==> 400 invalid tier"
curl -s -X POST "$BASE_URL/api/v1/events/submit" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Invalid Tier Event",
    "slug": "INVALIDTIER2026",
    "country": "Brazil",
    "city": "Recife",
    "latitude": -8.0476,
    "longitude": -34.8770,
    "start_date": "2026-09-21",
    "end_date": "2026-09-25",
    "website_url": "https://example.org",
    "domain": "computer_science",
    "tier": "S",
    "submitter": {"name": "Ana Silva", "email": "ana@example.com"}
  }' | jq .

# 400 — domain not in allowed enum
echo "==> 400 invalid domain"
curl -s -X POST "$BASE_URL/api/v1/events/submit" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Invalid Domain Event",
    "slug": "INVALIDDOMAIN2026",
    "country": "Brazil",
    "city": "Recife",
    "latitude": -8.0476,
    "longitude": -34.8770,
    "start_date": "2026-09-21",
    "end_date": "2026-09-25",
    "website_url": "https://example.org",
    "domain": "medicine",
    "submitter": {"name": "Ana Silva", "email": "ana@example.com"}
  }' | jq .

# 400 — latitude out of range
echo "==> 400 latitude out of range"
curl -s -X POST "$BASE_URL/api/v1/events/submit" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Invalid Latitude Event",
    "slug": "INVALIDLAT2026",
    "country": "Brazil",
    "city": "Recife",
    "latitude": 200,
    "longitude": -34.8770,
    "start_date": "2026-09-21",
    "end_date": "2026-09-25",
    "website_url": "https://example.org",
    "domain": "computer_science",
    "submitter": {"name": "Ana Silva", "email": "ana@example.com"}
  }' | jq .

# 400 — end_date before start_date
echo "==> 400 end_date before start_date"
curl -s -X POST "$BASE_URL/api/v1/events/submit" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Invalid Date Range Event",
    "slug": "INVALIDDATES2026",
    "country": "Brazil",
    "city": "Recife",
    "latitude": -8.0476,
    "longitude": -34.8770,
    "start_date": "2026-09-25",
    "end_date": "2026-09-21",
    "website_url": "https://example.org",
    "domain": "computer_science",
    "submitter": {"name": "Ana Silva", "email": "ana@example.com"}
  }' | jq .

# 400 — deadline missing required field
echo "==> 400 deadline missing description"
curl -s -X POST "$BASE_URL/api/v1/events/submit" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Bad Deadline Event",
    "slug": "BADDEADLINE2026",
    "country": "Brazil",
    "city": "Recife",
    "latitude": -8.0476,
    "longitude": -34.8770,
    "start_date": "2026-09-21",
    "end_date": "2026-09-25",
    "website_url": "https://example.org",
    "domain": "computer_science",
    "submitter": {"name": "Ana Silva", "email": "ana@example.com"},
    "deadlines": [{"type": "paper", "date": "2026-08-22"}]
  }' | jq .

# 400 — deadline with invalid type
echo "==> 400 deadline invalid type"
curl -s -X POST "$BASE_URL/api/v1/events/submit" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Bad Deadline Type Event",
    "slug": "BADDEADLINETYPE2026",
    "country": "Brazil",
    "city": "Recife",
    "latitude": -8.0476,
    "longitude": -34.8770,
    "start_date": "2026-09-21",
    "end_date": "2026-09-25",
    "website_url": "https://example.org",
    "domain": "computer_science",
    "submitter": {"name": "Ana Silva", "email": "ana@example.com"},
    "deadlines": [{"type": "keynote", "description": "Keynote announcement", "date": "2026-08-22"}]
  }' | jq .

# 400 — malformed JSON
echo "==> 400 malformed JSON"
curl -s -X POST "$BASE_URL/api/v1/events/submit" \
  -H "Content-Type: application/json" \
  -d '{"name": "Broken JSON"' | jq .

# 409 — slug already used by a pending/approved event (run the first request twice)
echo "==> 409 duplicate slug"
curl -s -X POST "$BASE_URL/api/v1/events/submit" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Duplicate Slug Event",
    "slug": "MODELS2026",
    "country": "Brazil",
    "city": "Recife",
    "latitude": -8.0476,
    "longitude": -34.8770,
    "start_date": "2026-09-21",
    "end_date": "2026-09-25",
    "website_url": "https://models2026.example.org",
    "domain": "computer_science",
    "submitter": {"name": "Ana Silva", "email": "ana@example.com"}
  }' | jq .

# 429 — rate limit exceeded (run this script's first request 51+ times in under a minute)
echo "==> 429 rate limit exceeded (requires looping 51+ requests in <60s)"
