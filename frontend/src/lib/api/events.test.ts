import { describe, it, expect, vi } from "vitest"

import { apiRequest, apiRequestWithMeta } from "./client"
import { listEvents, submitEvent, addDeadlines, cancelDeadline, supersedeDeadline } from "./events"
import type {
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

vi.mock("./client", () => ({
  apiRequest: vi.fn(),
  apiRequestWithMeta: vi.fn(),
}))

describe("listEvents", () => {
  it("builds a query string from the given params and returns data + meta", async () => {
    const events: EventListItem[] = []
    vi.mocked(apiRequestWithMeta).mockResolvedValue({ data: events, meta: { page: 1, total: 0 } })

    const params: ListEventsParams = { year: 2026, domain: "computer_science", status: "pending" }
    const result = await listEvents(params)

    expect(apiRequestWithMeta).toHaveBeenCalledWith(
      "/api/v1/events?year=2026&domain=computer_science&status=pending",
    )
    expect(result).toEqual({ data: events, meta: { page: 1, total: 0 } })
  })

  it("omits the query string entirely when called with no params", async () => {
    vi.mocked(apiRequestWithMeta).mockResolvedValue({ data: [], meta: { page: 1, total: 0 } })

    await listEvents({})

    expect(apiRequestWithMeta).toHaveBeenCalledWith("/api/v1/events")
  })
})

describe("submitEvent", () => {
  it("posts the submission to /api/v1/events/submit via apiRequest", async () => {
    const input: SubmitEventInput = {
      name: "Conf",
      slug: "conf",
      country: "Brazil",
      city: "Rio",
      latitude: -22.9,
      longitude: -43.2,
      start_date: "2026-06-01",
      end_date: "2026-06-03",
      website_url: "https://example.com",
      domain: "computer_science",
      submitter: { name: "Ana", email: "ana@example.com" },
    }
    const result: SubmitEventResult = {
      id: 1,
      name: "Conf",
      slug: "conf",
      country: "Brazil",
      city: "Rio",
      latitude: -22.9,
      longitude: -43.2,
      start_date: "2026-06-01",
      end_date: "2026-06-03",
      website_url: "https://example.com",
      domain: "computer_science",
      tier: "unranked",
      status: "pending",
      created_by: { id: 1, name: "Ana", email: "ana@example.com" },
      deadlines: [],
      created_at: "2026-01-01T00:00:00Z",
    }
    vi.mocked(apiRequest).mockResolvedValue(result)

    const data = await submitEvent(input)

    expect(apiRequest).toHaveBeenCalledWith("/api/v1/events/submit", {
      method: "POST",
      body: JSON.stringify(input),
    })
    expect(data).toEqual(result)
  })
})

describe("addDeadlines", () => {
  it("posts deadlines to /api/v1/events/{id}/deadlines via apiRequest", async () => {
    const input: AddDeadlinesInput = {
      submitter: { name: "Ana", email: "ana@example.com" },
      deadlines: [{ type: "abstract", description: "Abstract", date: "2026-03-01" }],
    }
    const result = {} as AddDeadlinesResult
    vi.mocked(apiRequest).mockResolvedValue(result)

    const data = await addDeadlines(1, input)

    expect(apiRequest).toHaveBeenCalledWith("/api/v1/events/1/deadlines", {
      method: "POST",
      body: JSON.stringify(input),
    })
    expect(data).toEqual(result)
  })
})

describe("cancelDeadline", () => {
  it("patches /api/v1/events/{eventId}/deadlines/{deadlineId}/cancel via apiRequest", async () => {
    const input: CancelDeadlineInput = {
      submitter: { name: "Ana", email: "ana@example.com" },
    }
    const result = {} as CancelDeadlineResult
    vi.mocked(apiRequest).mockResolvedValue(result)

    const data = await cancelDeadline(1, 2, input)

    expect(apiRequest).toHaveBeenCalledWith("/api/v1/events/1/deadlines/2/cancel", {
      method: "PATCH",
      body: JSON.stringify(input),
    })
    expect(data).toEqual(result)
  })
})

describe("supersedeDeadline", () => {
  it("posts to /api/v1/events/{eventId}/deadlines/{deadlineId}/supersede via apiRequest", async () => {
    const input: SupersedeDeadlineInput = {
      submitter: { name: "Ana", email: "ana@example.com" },
      date: "2026-04-01",
    }
    const result = {} as SupersedeDeadlineResult
    vi.mocked(apiRequest).mockResolvedValue(result)

    const data = await supersedeDeadline(1, 2, input)

    expect(apiRequest).toHaveBeenCalledWith("/api/v1/events/1/deadlines/2/supersede", {
      method: "POST",
      body: JSON.stringify(input),
    })
    expect(data).toEqual(result)
  })
})
