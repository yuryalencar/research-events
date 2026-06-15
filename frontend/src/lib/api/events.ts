import { apiRequest, apiRequestWithMeta } from "./client"
import type {
  ApiMeta,
  ListEventsParams,
  EventListItem,
  SubmitEventInput,
  SubmitEventResult,
  AddDeadlinesInput,
  AddDeadlinesResult,
  CancelDeadlineInput,
  CancelDeadlineResult,
  SupersedeDeadlineInput,
  SupersedeDeadlineResult,
} from "@/types/api"

// --- Public API ---

// listEvents fetches events for the globe/list views, filterable by year,
// domain, country, status, tier, first_deadline_month, bbox, and pagination.
// Public request — returns both the page of events and pagination meta.
async function listEvents(params: ListEventsParams): Promise<{ data: EventListItem[]; meta: ApiMeta }> {
  const query = buildQueryString(params)
  return apiRequestWithMeta<EventListItem[]>(`/api/v1/events${query}`)
}

// submitEvent creates a new event with status=pending, looking up or
// creating the contributor by email.
async function submitEvent(input: SubmitEventInput): Promise<SubmitEventResult> {
  return apiRequest<SubmitEventResult>("/api/v1/events/submit", {
    method: "POST",
    body: JSON.stringify(input),
  })
}

// addDeadlines adds one or more deadlines to an existing event.
async function addDeadlines(eventId: number, input: AddDeadlinesInput): Promise<AddDeadlinesResult> {
  return apiRequest<AddDeadlinesResult>(`/api/v1/events/${eventId}/deadlines`, {
    method: "POST",
    body: JSON.stringify(input),
  })
}

// cancelDeadline marks a deadline as is_active=false with no replacement.
async function cancelDeadline(
  eventId: number,
  deadlineId: number,
  input: CancelDeadlineInput,
): Promise<CancelDeadlineResult> {
  return apiRequest<CancelDeadlineResult>(`/api/v1/events/${eventId}/deadlines/${deadlineId}/cancel`, {
    method: "PATCH",
    body: JSON.stringify(input),
  })
}

// supersedeDeadline marks a deadline as is_active=false and creates a new
// deadline of the same type, linked via superseded_by_id.
async function supersedeDeadline(
  eventId: number,
  deadlineId: number,
  input: SupersedeDeadlineInput,
): Promise<SupersedeDeadlineResult> {
  return apiRequest<SupersedeDeadlineResult>(`/api/v1/events/${eventId}/deadlines/${deadlineId}/supersede`, {
    method: "POST",
    body: JSON.stringify(input),
  })
}

// --- Internals ---

// buildQueryString turns a ListEventsParams object into a "?key=value&..."
// string, omitting any params that are undefined. Returns "" (no leading
// "?") when every param is undefined.
function buildQueryString(params: ListEventsParams): string {
  const entries = Object.entries(params).filter(([, value]) => value !== undefined)
  if (entries.length === 0) {
    return ""
  }

  const search = new URLSearchParams(entries.map(([key, value]) => [key, String(value)]))
  return `?${search.toString()}`
}

// --- Export ---

export { listEvents, submitEvent, addDeadlines, cancelDeadline, supersedeDeadline }
