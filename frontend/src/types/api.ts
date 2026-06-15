/**
 * Hand-written request/response types for the backend API, transcribed from
 * specs/backend/*.yaml. This is a deliberate, temporary exception to the
 * "never hand-write API types" rule (see CLAUDE.md) — there is no OpenAPI
 * spec to run `make generate-types` against yet. Replace this file once that
 * tooling exists.
 */

// --- Shared ---

interface ApiMeta {
  page: number
  total: number
}

type EventStatus = "pending" | "approved" | "rejected"

type EventTier = "A*" | "A" | "B" | "C" | "unranked"

type DeadlineType = "abstract" | "paper" | "notification" | "camera_ready" | "other"

interface UserSummary {
  id: number
  name: string
  email: string
}

interface DeadlineResponse {
  id: number
  type: DeadlineType
  description: string
  date: string
  time: string | null
  timezone: string | null
  is_optional: boolean
  is_active: boolean
  superseded_by_id: number | null
}

// --- Events: GET /api/v1/events ---

interface ListEventsParams {
  year?: number
  domain?: string
  country?: string
  status?: EventStatus
  tier?: EventTier
  first_deadline_month?: number
  bbox?: string
  page?: number
  page_size?: number
  pagination?: "on" | "off"
}

interface EventListItem {
  id: number
  name: string
  slug: string
  country: string
  city: string
  latitude: number
  longitude: number
  start_date: string
  end_date: string
  website_url: string
  domain: string
  status: EventStatus
  tier: EventTier
  year: number
  created_by: UserSummary
  last_updated_by: UserSummary
  deadlines: DeadlineResponse[]
  created_at: string
  updated_at: string
}

// --- Events: POST /api/v1/events/submit ---

interface SubmitterInput {
  name: string
  email: string
}

interface DeadlineInput {
  type: DeadlineType
  description: string
  date: string
  is_optional?: boolean
  time?: string
  timezone?: string
}

interface SubmitEventInput {
  name: string
  slug: string
  country: string
  city: string
  latitude: number
  longitude: number
  start_date: string
  end_date: string
  website_url: string
  domain: string
  tier?: EventTier
  submitter: SubmitterInput
  deadlines?: DeadlineInput[]
}

interface SubmitEventResult {
  id: number
  name: string
  slug: string
  country: string
  city: string
  latitude: number
  longitude: number
  start_date: string
  end_date: string
  website_url: string
  domain: string
  tier: EventTier
  status: "pending"
  created_by: UserSummary
  deadlines: DeadlineResponse[]
  created_at: string
}

// --- Events: POST /api/v1/events/{id}/deadlines ---

interface AddDeadlinesInput {
  submitter: SubmitterInput
  deadlines: DeadlineInput[]
}

type AddDeadlinesResult = EventListItem

// --- Events: PATCH /api/v1/events/{eventId}/deadlines/{deadlineId}/cancel ---

interface CancelDeadlineInput {
  submitter: SubmitterInput
}

type CancelDeadlineResult = EventListItem

// --- Events: POST /api/v1/events/{eventId}/deadlines/{deadlineId}/supersede ---

interface SupersedeDeadlineInput {
  submitter: SubmitterInput
  date: string
  time?: string
  timezone?: string
}

type SupersedeDeadlineResult = EventListItem

// --- Auth: POST /api/v1/auth/login ---

interface LoginInput {
  email: string
  password: string
}

interface AuthResult {
  token: string
  role: string
  user: UserSummary
}

type LoginResult = AuthResult

// --- Auth: POST /api/v1/auth/refresh-token ---

type RefreshTokenResult = AuthResult

// --- Admin: PATCH /api/v1/admin/events/{id}/review ---

interface ReviewEventEditInput {
  name?: string
  slug?: string
  country?: string
  city?: string
  latitude?: number
  longitude?: number
  start_date?: string
  end_date?: string
  website_url?: string
  domain?: string
  tier?: EventTier
}

interface ReviewEventInput {
  action: "approve" | "reject"
  reason?: string
  event?: ReviewEventEditInput
}

type ReviewEventResult = EventListItem

// --- Admin: PATCH /api/v1/admin/users/{id}/unlock ---

interface UnlockUserResult {
  user: UserSummary & { role: string }
}

// --- Health: GET /health ---

interface HealthCheckResult {
  status: "healthy" | "unhealthy"
  latency_ms?: number
  error?: string
}

interface HealthResult {
  status: "healthy" | "unhealthy"
  version: string
  timestamp: string
  uptime: string
  checks: Record<string, HealthCheckResult>
}

// --- Export ---

export type {
  ApiMeta,
  EventStatus,
  EventTier,
  DeadlineType,
  UserSummary,
  DeadlineResponse,
  ListEventsParams,
  EventListItem,
  SubmitterInput,
  DeadlineInput,
  SubmitEventInput,
  SubmitEventResult,
  AddDeadlinesInput,
  AddDeadlinesResult,
  CancelDeadlineInput,
  CancelDeadlineResult,
  SupersedeDeadlineInput,
  SupersedeDeadlineResult,
  LoginInput,
  AuthResult,
  LoginResult,
  RefreshTokenResult,
  ReviewEventEditInput,
  ReviewEventInput,
  ReviewEventResult,
  UnlockUserResult,
  HealthCheckResult,
  HealthResult,
}
