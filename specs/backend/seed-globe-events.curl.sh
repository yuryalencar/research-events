#!/bin/bash
# Seed fake approved events around the world for testing the globe homepage.
# Not tied to a single spec — exercises auth-login, events-submit, and
# admin-events-review together.

BASE_URL="${BASE_URL:-http://localhost:8080}"
COOKIE_JAR="${COOKIE_JAR:-/tmp/cookies.txt}"

echo "==> login as admin"
curl -sc "$COOKIE_JAR" -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"your-email@domain.com","password":"your-admin-password"}' | jq .

submit() {
  curl -s -X POST "$BASE_URL/api/v1/events/submit" \
    -H "Content-Type: application/json" \
    -d "$1"
}

# --- Submit events (status=pending) ---

echo "==> submit ICSE2026 (Rio de Janeiro, Brazil)"
submit '{
  "name": "International Conference on Software Engineering 2026",
  "slug": "ICSE2026",
  "country": "Brazil",
  "city": "Rio de Janeiro",
  "latitude": -22.9068,
  "longitude": -43.1729,
  "start_date": "2026-04-13",
  "end_date": "2026-04-19",
  "website_url": "https://conf.researchr.org/home/icse-2026",
  "domain": "computer_science",
  "tier": "A*",
  "submitter": {"name": "Your Name", "email": "your-email@domain.com"}
}' | jq .

echo "==> submit NEURIPS2026 (San Diego, USA)"
submit '{
  "name": "Conference on Neural Information Processing Systems 2026",
  "slug": "NEURIPS2026",
  "country": "United States",
  "city": "San Diego",
  "latitude": 32.7157,
  "longitude": -117.1611,
  "start_date": "2026-12-08",
  "end_date": "2026-12-14",
  "website_url": "https://neurips.cc/",
  "domain": "computer_science",
  "tier": "A*",
  "submitter": {"name": "Your Name", "email": "your-email@domain.com"}
}' | jq .

echo "==> submit ECOOP2026 (Lisbon, Portugal)"
submit '{
  "name": "European Conference on Object-Oriented Programming 2026",
  "slug": "ECOOP2026",
  "country": "Portugal",
  "city": "Lisbon",
  "latitude": 38.7223,
  "longitude": -9.1393,
  "start_date": "2026-07-06",
  "end_date": "2026-07-10",
  "website_url": "https://2026.ecoop.org/",
  "domain": "computer_science",
  "tier": "A",
  "submitter": {"name": "Your Name", "email": "your-email@domain.com"}
}' | jq .

echo "==> submit PLDI2026 (Tokyo, Japan)"
submit '{
  "name": "Programming Language Design and Implementation 2026",
  "slug": "PLDI2026",
  "country": "Japan",
  "city": "Tokyo",
  "latitude": 35.6762,
  "longitude": 139.6503,
  "start_date": "2026-06-20",
  "end_date": "2026-06-25",
  "website_url": "https://pldi26.sigplan.org/",
  "domain": "computer_science",
  "tier": "A*",
  "submitter": {"name": "Your Name", "email": "your-email@domain.com"}
}' | jq .

echo "==> submit AFRISE2026 (Cape Town, South Africa)"
submit '{
  "name": "Africa Software Engineering Symposium 2026",
  "slug": "AFRISE2026",
  "country": "South Africa",
  "city": "Cape Town",
  "latitude": -33.9249,
  "longitude": 18.4241,
  "start_date": "2026-09-14",
  "end_date": "2026-09-17",
  "website_url": "https://afrise.example.org/2026",
  "domain": "computer_science",
  "tier": "B",
  "submitter": {"name": "Your Name", "email": "your-email@domain.com"}
}' | jq .

echo "==> submit OZCHI2026 (Sydney, Australia)"
submit '{
  "name": "Australian Conference on Human-Computer Interaction 2026",
  "slug": "OZCHI2026",
  "country": "Australia",
  "city": "Sydney",
  "latitude": -33.8688,
  "longitude": 151.2093,
  "start_date": "2026-11-30",
  "end_date": "2026-12-03",
  "website_url": "https://ozchi.org/2026",
  "domain": "computer_science",
  "tier": "B",
  "submitter": {"name": "Your Name", "email": "your-email@domain.com"}
}' | jq .

echo "==> submit ICALP2026 (Berlin, Germany)"
submit '{
  "name": "International Colloquium on Automata, Languages and Programming 2026",
  "slug": "ICALP2026",
  "country": "Germany",
  "city": "Berlin",
  "latitude": 52.5200,
  "longitude": 13.4050,
  "start_date": "2026-07-13",
  "end_date": "2026-07-17",
  "website_url": "https://icalp2026.eu/",
  "domain": "computer_science",
  "tier": "A",
  "submitter": {"name": "Your Name", "email": "your-email@domain.com"}
}' | jq .

echo "==> submit TGCSS2026 (Toronto, Canada)"
submit '{
  "name": "Toronto Graduate Computer Science Symposium 2026",
  "slug": "TGCSS2026",
  "country": "Canada",
  "city": "Toronto",
  "latitude": 43.6532,
  "longitude": -79.3832,
  "start_date": "2026-10-05",
  "end_date": "2026-10-06",
  "website_url": "https://tgcss.example.org/2026",
  "domain": "computer_science",
  "submitter": {"name": "Your Name", "email": "your-email@domain.com"}
}' | jq .

# --- Approve all of the above (admin/moderator only) ---
# There is no admin-only "list pending events" endpoint — the public
# GET /api/v1/events?status=pending (events-list.yaml) shows pending events
# to everyone regardless of auth ("filtering controls what's shown, not
# authorization"). We use it here to find the IDs of the slugs just submitted.

echo "==> find pending event IDs for the slugs just submitted"
SLUGS="ICSE2026 NEURIPS2026 ECOOP2026 PLDI2026 AFRISE2026 OZCHI2026 ICALP2026 TGCSS2026"
IDS=$(curl -s "$BASE_URL/api/v1/events?status=pending&pagination=off" \
  | jq -r --arg slugs "$SLUGS" '
      ($slugs | split(" ")) as $wanted
      | .data[] | select(.slug as $s | $wanted | index($s)) | .id
    ')

for id in $IDS; do
  echo "==> approving event $id"
  curl -s -b "$COOKIE_JAR" -X PATCH "$BASE_URL/api/v1/admin/events/$id/review" \
    -H "Content-Type: application/json" \
    -d '{"action":"approve","reason":"seed data for globe homepage testing"}' | jq '.data | {id, slug, status}'
done
