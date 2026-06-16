const APP_VERSION = "0.1.0"
const GITHUB_REPO_URL = "https://github.com/yuryalencar/research-events"
const GITHUB_PROFILE_URL = "https://github.com/yuryalencar"
const ADMIN_EMAIL = "yuryalencar19@gmail.com"

// --- Filter constants ---

const MIN_FILTER_YEAR = 2000

// DOMAINS must stay in sync with allowedDomains in backend/internal/service/event.go.
const DOMAINS: readonly string[] = ["computer_science"]

// TIERS mirrors the events_tier_check CHECK constraint in the backend.
const TIERS: readonly string[] = ["A*", "A", "B", "C", "unranked"]

export { APP_VERSION, GITHUB_REPO_URL, GITHUB_PROFILE_URL, ADMIN_EMAIL, MIN_FILTER_YEAR, DOMAINS, TIERS }
