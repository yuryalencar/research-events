# Graph Report - .  (2026-08-07)

## Corpus Check
- 325 files · ~351,070 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 2190 nodes · 5683 edges · 148 communities (88 shown, 60 thin omitted)
- Extraction: 79% EXTRACTED · 21% INFERRED · 0% AMBIGUOUS · INFERRED: 1174 edges (avg confidence: 0.8)
- Token cost: 0 input · 884,580 output

## Community Hubs (Navigation)
- Deadline Repository & Audit Tests
- Admin Event Review Handler Tests
- Admin User Handler Tests
- GORM Domain Models
- Auth Handler & Cookies
- Event Deadlines Service & Handler
- Event & Deadlines Input Mapping
- Admin User Service & Password Audit
- Auth Middleware Tests
- Admin Users Frontend Hook & API
- Admin Event Review Service
- Manage Login & Events Hook
- Event Submission Service
- Auth & Admin API Spec (JTI Design)
- Manage Review & Dashboard Components
- Deadline Management Frontend
- Review Wizard Steps
- Event Detail & Table Components
- Admin User List Query Parsing
- Mock User Repository
- Register User Form Components
- Admin/Moderator Page Routes
- Submission Wizard Deadlines Step
- Admin User Management Feature
- TypeScript Config
- Globe Homepage View Components
- Drawer/Sheet UI Primitives
- Server Route Registration Tests
- Health DB Checker Tests
- Frontend UI Dependencies
- Modal & Dialog Components
- Deadline Management API Client
- API Client & Auth Error Handling
- Submission Form & Filter Panel
- Health Handler
- Health Checker Registry Mock Tests
- Globe/Manage Filters Feature
- shadcn/ui Component Config
- Locale Layout & Language Selector
- Deadline Add/Cancel/Supersede Sessions
- Project Instructions & FP Principles
- Server Bootstrap & Auth Feature
- Deadline Management Frontend Planning
- OpenTelemetry Tracer Setup
- Backend Config Loading
- CORS Middleware
- Globe Homepage Feature Concepts
- Event Submission & Auth Sessions
- Frontend Dev Dependencies
- Info Button & Location Picker
- Frontend Feature Sessions
- Manage Review Frontend Session
- Update Password Backend Feature
- DB Tracing Integration Tests
- Globe Clustering Hook
- Backend Observability Libraries
- Manage Portal Frontend Session
- Globe Clustering & Loading Session
- Globe View Rendering
- Database Migrations (Core Tables)
- Deep-link & Info Modal Session
- Manage Dashboard Session
- Health Check Types
- npm Scripts
- OpenTelemetry Tracing Session
- Globe/Table View Toggle
- Session Guard Bug Fixes
- Admin User Frontend Components
- Frontend Build Tooling Deps
- Release Notes & README
- Year Filter Semantics Change
- Update Password Portal Session
- package.json Metadata
- Events List Feature Session
- Event Detail/Table Modifications
- Language Selector Feature
- Server Route Handler
- Globe Clusters Test
- Globe Homepage i18n Test
- ESLint Config
- Next.js Config
- Manage Layout
- Sitemap Generation
- Seed Events Script
- Go Testing Libraries
- eslint-config-next Dependency
- Leaflet Dependency
- Next.js Dependency
- React Dependency
- jsdom Dependency
- PostCSS Dependency
- Tailwind CSS Dependency
- Tailwind PostCSS Plugin
- Testing Library DOM
- Node Types Dependency
- React DOM Types Dependency
- Supercluster Types Dependency
- Vite Dependency
- Vitest Dependency
- PostCSS Config
- Errors i18n Test
- Tailwind Config
- Go Mod Tidy Concept
- Admin Events Review curl Script
- Admin Users Change Role curl Script
- Admin Users List curl Script
- Admin Users Register curl Script
- Admin Users Reset Password curl Script
- Admin Users Unlock curl Script
- Auth Login curl Script
- Auth Logout curl Script
- Auth Refresh Token curl Script
- Deadlines Time/Timezone curl Script
- Events Deadlines Add curl Script
- Events Deadlines Cancel curl Script
- Events Deadlines Supersede curl Script
- Events List curl Script
- Events Submit curl Script
- Server Bootstrap curl Script
- Users Me curl Script
- Users Update Password curl Script
- Admin API Client Helper
- godotenv Library
- Goose Migration Tool
- golang-jwt Library
- Go Channels Concept
- Go-to-FP Concepts Bridge
- Go Goroutines Concept
- Go Interfaces Concept
- Go Multiple Return Values Concept
- Go Pointer Receivers Concept
- Go Stack vs Heap Concept
- Globe Logo Image
- Open Source Logo Image
- Open Source Filled Logo Image
- App Icon Image
- Go Module Declaration

## God Nodes (most connected - your core abstractions)
1. `responseCode()` - 116 edges
2. `beginTx()` - 109 edges
3. `NewMockEventRepository()` - 86 edges
4. `NewMockUserRepository()` - 81 edges
5. `NewEventHandler()` - 73 edges
6. `NewEventRepository()` - 72 edges
7. `EventListItem` - 53 edges
8. `NewAdminUserHandler()` - 51 edges
9. `NewMockAuditRepository()` - 48 edges
10. `Event` - 44 edges

## Surprising Connections (you probably didn't know these)
- `Proposed Ownership and Update Rules` --semantically_similar_to--> `BuildRegisterAuditLog`  [INFERRED] [semantically similar]
  docs/roadmap/se-deadlines-importer.md → ai-sessions/2026-06-26-admin-user-management.md
- `Session: Server Bootstrap + Health Check` --references--> `Pair Programming Workflow (Phase 0–6)`  [INFERRED]
  ai-sessions/2026-06-08-server-bootstrap-health-check.md → CLAUDE.md
- `AGENTS.md — Codex Instructions` --references--> `CLAUDE.md — ReSEARCH Events Project Instructions`  [EXTRACTED]
  AGENTS.md → CLAUDE.md
- `Session: API Client + Error Handling (Frontend)` --references--> `CLAUDE.md — ReSEARCH Events Project Instructions`  [EXTRACTED]
  ai-sessions/2026-06-15-frontend-api-client-error-handling-feature.md → CLAUDE.md
- `CLAUDE.md — ReSEARCH Events Project Instructions` --references--> `docker-compose.yml — Local Postgres services`  [EXTRACTED]
  CLAUDE.md → docker-compose.yml

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **FP core principles enforced in internal/service/ (pure function, immutability, no side effects)** — claude_fp_pure_function, claude_fp_immutability, claude_fp_no_side_effects [EXTRACTED 1.00]
- **Deadline lifecycle management pattern (time/timezone, add, cancel, supersede)** — ai_sessions_2026_06_14_deadlines_time_timezone_feature_feature, ai_sessions_2026_06_14_event_deadlines_add_feature_feature, ai_sessions_2026_06_14_event_deadlines_cancel_feature_feature, ai_sessions_2026_06_14_event_deadlines_supersede_feature_feature [INFERRED 0.85]
- **Globe homepage frontend surface (pins, deep-link, filters, info modal)** — ai_sessions_2026_06_15_globe_homepage_feature_feature, ai_sessions_2026_06_16_globe_event_deeplink_feature_feature, ai_sessions_2026_06_16_globe_filters_feature_feature, ai_sessions_2026_06_16_info_modal_feature_feature [INFERRED 0.80]
- **Deadline Management Feature Flow (page, hook, cards, entry point)** — ai_sessions_2026_06_19_deadline_management_frontend_implementation_usedeadlinemanage, ai_sessions_2026_06_19_deadline_management_frontend_implementation_deadlinemanagepage, ai_sessions_2026_06_19_deadline_management_frontend_implementation_deadlinecard, ai_sessions_2026_06_19_deadline_management_frontend_implementation_adddeadlinecard, ai_sessions_2026_06_19_deadline_management_frontend_implementation_deadlinemanagesuccess, ai_sessions_2026_06_19_deadline_management_frontend_implementation_eventdetailview [EXTRACTED 1.00]
- **Manage Review Wizard Flow (3-step review: details, decision, deadlines)** — ai_sessions_2026_06_19_manage_review_frontend_usereviewwizard, ai_sessions_2026_06_19_manage_review_frontend_reviewwizard, ai_sessions_2026_06_19_manage_review_frontend_reviewstep1details, ai_sessions_2026_06_19_manage_review_frontend_reviewstep2decision, ai_sessions_2026_06_19_manage_review_frontend_reviewstep3deadlines, ai_sessions_2026_06_19_manage_review_frontend_approvemodal, ai_sessions_2026_06_19_manage_review_frontend_rejectmodal, ai_sessions_2026_06_19_manage_review_frontend_reviewsuccess [EXTRACTED 1.00]
- **Eager Session Validation & Auth Error Flow** — ai_sessions_2026_06_21_session_guard_eager_token_validation_usesessionguard, ai_sessions_2026_06_21_session_guard_eager_token_validation_validatesession, ai_sessions_2026_06_21_session_guard_eager_token_validation_get_users_me_endpoint, ai_sessions_2026_06_21_session_guard_eager_token_validation_userhandler_me, ai_sessions_2026_06_21_session_guard_eager_token_validation_updatepasswordcard [EXTRACTED 1.00]
- **Auth/Admin endpoints reading and writing the users table schema** — specs_backend_database_users_migration, specs_backend_auth_login_endpoint, specs_backend_auth_logout_endpoint, specs_backend_auth_middleware_requireauth, specs_backend_admin_users_unlock_endpoint, specs_backend_admin_users_change_role_endpoint, specs_backend_admin_users_register_endpoint, specs_backend_admin_users_list_endpoint [INFERRED 0.85]
- **Endpoints chained behind RequireAuth + RequireRole** — specs_backend_auth_middleware_requireauth, specs_backend_auth_middleware_requirerole, specs_backend_admin_events_review_endpoint, specs_backend_admin_users_change_role_endpoint, specs_backend_admin_users_list_endpoint, specs_backend_admin_users_register_endpoint, specs_backend_admin_users_unlock_endpoint [EXTRACTED 0.90]
- **Endpoints sharing preloadEventAssociations reload/response shape for deadlines** — specs_backend_events_deadlines_supersede_endpoint, specs_backend_events_list_endpoint, specs_backend_events_deadlines_add_endpoint, specs_backend_events_deadlines_cancel_endpoint [INFERRED 0.85]
- **Multi-Step Wizard Pattern (Submission, Review, Deadline Management)** — specs_frontend_event_submission_wizard, specs_frontend_manage_review, specs_frontend_deadline_management [INFERRED 0.85]
- **Globe Homepage View Ecosystem (rendering, filters, clustering, deep-link, loading, table toggle)** — specs_frontend_globe_homepage, specs_frontend_globe_filters, specs_frontend_globe_event_clustering, specs_frontend_globe_event_deeplink, specs_frontend_globe_loading_progression, specs_frontend_globe_table_toggle [INFERRED 0.85]
- **Management Portal Authenticated Session Flow (login, dashboard, review, password)** — specs_frontend_manage_portal, specs_frontend_manage_dashboard, specs_frontend_manage_review, specs_frontend_manage_update_password [INFERRED 0.85]
- **Admin User Management API Surface** — ai_sessions_2026_06_26_admin_user_management_register_endpoint, ai_sessions_2026_06_26_admin_user_management_changerole_endpoint, ai_sessions_2026_06_26_admin_users_list_endpoint_list_endpoint, ai_sessions_2026_06_26_admin_users_reset_password_resetpassword_endpoint, ai_sessions_2026_06_26_admin_user_management_frontend_implementation_feature [EXTRACTED 0.90]
- **Globe Homepage Progressive UX (clustering + loading)** — ai_sessions_2026_06_24_globe_event_clustering_useglobeclusters, ai_sessions_2026_06_24_globe_event_clustering_globeview, ai_sessions_2026_06_24_globe_loading_progression_useloadingmessage [INFERRED 0.75]
- **OpenTelemetry-to-Sentry Tracing Pipeline** — docs_backend_libraries_otel, docs_backend_libraries_sdktrace, docs_backend_libraries_sentryotel, docs_backend_libraries_otelhttp, docs_backend_libraries_gorm_otel_plugin [EXTRACTED 0.90]

## Communities (148 total, 60 thin omitted)

### Community 0 - "Deadline Repository & Audit Tests"
Cohesion: 0.06
Nodes (130): addOneDeadline(), DB, T, Time, newApprovedEvent(), newSupersedingDeadline(), oneDeadline(), TestAuditLog_BatchDeadlinesAddedAction_IsAllowedByConstraint() (+122 more)

### Community 1 - "Admin Event Review Handler Tests"
Cohesion: 0.08
Nodes (119): Logger, NewAdminEventHandler(), existingEvent(), Handler, Request, T, rawReviewReq(), reviewMux() (+111 more)

### Community 2 - "Admin User Handler Tests"
Cohesion: 0.10
Nodes (88): NewAdminUserHandler(), changeRoleMux(), changeRoleReq(), Handler, Request, T, listUsersMux(), listUsersReq() (+80 more)

### Community 3 - "GORM Domain Models"
Cohesion: 0.05
Nodes (49): Model, Model, Time, Model, Time, Model, Time, Context (+41 more)

### Community 4 - "Auth Handler & Cookies"
Cohesion: 0.07
Nodes (67): clearAuthCookies(), generateRandomHex(), Context, Logger, Request, ResponseWriter, Time, NewAuthHandler() (+59 more)

### Community 5 - "Event Deadlines Service & Handler"
Cohesion: 0.08
Nodes (68): EventHandler, Request, ResponseWriter, BuildDeadlinesFromInput(), BuildSupersedingDeadline(), DetermineDeadlinesAuditAction(), DeadlineInput, SubmitterInput (+60 more)

### Community 6 - "Event & Deadlines Input Mapping"
Cohesion: 0.06
Nodes (65): AddDeadlinesInput, DeadlineInput, SupersedeDeadlineInput, toAddDeadlinesInput(), toDeadlineInputs(), toSupersedeDeadlineInput(), EventHandler, eventListItemResponse (+57 more)

### Community 7 - "Admin User Service & Password Audit"
Cohesion: 0.08
Nodes (59): Logger, Request, ResponseWriter, Time, toUserListItemResponse(), writeError(), Request, ResponseWriter (+51 more)

### Community 8 - "Auth Middleware Tests"
Cohesion: 0.07
Nodes (42): Context, Handler, ResponseWriter, NewAuthMiddleware(), RequireRole(), assertCode(), buildToken(), captureUser() (+34 more)

### Community 9 - "Admin Users Frontend Hook & API"
Cohesion: 0.09
Nodes (43): AdminUsersPhase, defaultFilters, meta, useAdminUsers(), UseAdminUsersReturn, changeUserRole(), listAdminUsers(), registerAdminUser() (+35 more)

### Community 10 - "Admin Event Review Service"
Cohesion: 0.12
Nodes (51): Request, ResponseWriter, toEventEditInput(), applyEventEdits(), ApplyReview(), BuildReviewAuditLog(), buildReviewDiff(), Time (+43 more)

### Community 11 - "Manage Login & Events Hook"
Cohesion: 0.07
Nodes (28): LoginPhase, baseFilters, CURRENT_YEAR, event, toListEventsParams(), useEvents(), UseEventsReturn, AUTH_ERROR_CODES (+20 more)

### Community 12 - "Event Submission Service"
Cohesion: 0.13
Nodes (47): BuildEventFromInput(), BuildSubmission(), BuildSubmitterFromInput(), DeadlineInput, SubmitterInput, Time, isValidWebsiteURL(), normalizeTier() (+39 more)

### Community 13 - "Auth & Admin API Spec (JTI Design)"
Cohesion: 0.06
Nodes (47): PATCH /api/v1/admin/events/{id}/review, PATCH /api/v1/admin/users/:id/role, GET /api/v1/admin/users, POST /api/v1/admin/users, PATCH /api/v1/admin/users/:id/unlock, JTI (JWT ID), Refresh Token (opaque, SHA-256 hashed), Access Token Validation Flow (+39 more)

### Community 14 - "Manage Review & Dashboard Components"
Cohesion: 0.09
Nodes (30): Step1Search(), Step1SearchProps, EventReviewCard(), EventReviewCardProps, ManageDashboard(), ManageDashboardProps, SessionUser, STATUS_OPTIONS (+22 more)

### Community 15 - "Deadline Management Frontend"
Cohesion: 0.10
Nodes (26): AddDeadlineCard(), AddDeadlineCardProps, DEADLINE_TYPES, DeadlineCard(), DeadlineCardProps, DeadlineManageContent(), DeadlineManagePage(), DeadlineManageSuccess() (+18 more)

### Community 16 - "Review Wizard Steps"
Cohesion: 0.10
Nodes (29): ApproveModal(), RejectModal(), FieldProps, LocationPicker, ReviewStep1Details(), ReviewStep1DetailsProps, STATUS_STYLES, StatusBadge() (+21 more)

### Community 17 - "Event Detail & Table Components"
Cohesion: 0.10
Nodes (25): DeadlineListProps, EventDetailContent(), EventDetailContentProps, EventDetailViewProps, EventTableCard(), EventTableCardProps, EventTableView(), EventTableViewProps (+17 more)

### Community 18 - "Admin User List Query Parsing"
Cohesion: 0.14
Nodes (34): Values, toRawListUsersQuery(), parseIncludeDeleted(), parseLockedFilter(), parseUserRoles(), T, TestValidateListUsersQuery_AllThreeRoles_ReturnsAllRoles(), TestValidateListUsersQuery_IncludeDeletedFalse_ReturnsFalse() (+26 more)

### Community 19 - "Mock User Repository"
Cohesion: 0.13
Nodes (6): Call, Context, Controller, Time, MockUserRepository, MockUserRepositoryMockRecorder

### Community 20 - "Register User Form Components"
Cohesion: 0.07
Nodes (21): ConfirmModalProps, FieldProps, RegisterUserFormProps, ROLE_OPTIONS, SuccessScreenProps, ChipProps, ROLE_OPTIONS, UserFilters() (+13 more)

### Community 21 - "Admin/Moderator Page Routes"
Cohesion: 0.11
Nodes (20): AdminReviewPage(), AdminDashboardPage(), AdminUpdatePasswordPage(), AdminUsersPage(), AdminRegisterUserPage(), ModeratorReviewPage(), ModeratorDashboardPage(), ModeratorUpdatePasswordPage() (+12 more)

### Community 22 - "Submission Wizard Deadlines Step"
Cohesion: 0.11
Nodes (20): DEADLINE_TYPES, DeadlineRowProps, Step3Deadlines(), Step3DeadlinesProps, StepIndicator(), StepIndicatorProps, SubmitSuccess(), SubmitSuccessProps (+12 more)

### Community 23 - "Admin User Management Feature"
Cohesion: 0.08
Nodes (30): BuildRegisterAuditLog, BuildRegisterUser, BuildRoleChangedAuditLog, PATCH /api/v1/admin/users/{id}/role (ChangeRole), ClearTokens session-invalidation mechanism, ExistsByEmail repository method, Admin User Management Frontend, useAdminUsers hook (+22 more)

### Community 24 - "TypeScript Config"
Cohesion: 0.07
Nodes (27): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+19 more)

### Community 25 - "Globe Homepage View Components"
Cohesion: 0.12
Nodes (17): GlobeView, Page(), EventDetailView(), AddEventButton(), ClusterEventDrawer(), ViewToggleButton(), ViewToggleButtonProps, EventFilters (+9 more)

### Community 26 - "Drawer/Sheet UI Primitives"
Cohesion: 0.16
Nodes (16): Drawer(), DrawerContent(), DrawerDescription(), DrawerFooter(), DrawerHeader(), DrawerOverlay(), DrawerTitle(), Sheet() (+8 more)

### Community 27 - "Server Route Registration Tests"
Cohesion: 0.33
Nodes (20): BuildHandler(), DB, Logger, TracerProvider, T, TestBuildHandler_AdminListUsersRouteRegistered(), TestBuildHandler_AdminResetPasswordRouteRegistered(), TestBuildHandler_AdminUnlockRouteRegistered() (+12 more)

### Community 28 - "Health DB Checker Tests"
Cohesion: 0.18
Nodes (14): NewDatabaseChecker(), T, TestDatabaseChecker_Check_ReturnsHealthyWhenPingSucceeds(), TestDatabaseChecker_Check_ReturnsUnhealthyWhenContextCancelled(), TestDatabaseChecker_Check_ReturnsUnhealthyWhenPingFails(), TestDatabaseChecker_Name_ReturnsDatabaseString(), Call, Context (+6 more)

### Community 29 - "Frontend UI Dependencies"
Cohesion: 0.10
Nodes (21): clsx, dependencies, clsx, globe.gl, lucide-react, next-intl, @radix-ui/react-dialog, react-dom (+13 more)

### Community 30 - "Modal & Dialog Components"
Cohesion: 0.19
Nodes (10): InfoModalProps, PinLegendRowProps, ApproveModalProps, RejectModalProps, Dialog(), DialogContent(), DialogDescription(), DialogHeader() (+2 more)

### Community 31 - "Deadline Management API Client"
Cohesion: 0.15
Nodes (18): addDeadlines(eventId, input) client function, cancelDeadline(eventId, deadlineId) client function, DeadlineManagePage component (main orchestrator), DeadlineManageSuccess component, EventDetailContent component, EventReviewCard component (normal + greyed-out own-event variants), Leaflet map location picker (LocationPicker), reviewEvent(id, input) — PATCH /api/v1/admin/events/{id}/review client (+10 more)

### Community 32 - "API Client & Auth Error Handling"
Cohesion: 0.15
Nodes (18): ApiError class, apiPrivateRequest (cookie-authenticated fetch client with refresh-retry), apiRequest (public fetch client), Bug fix: non-JSON response mismapped to NETWORK_ERROR, Bug fix: Chrome third-party cookie blocking caused TOKEN_MISSING, decodeJwt utility (client-side JWT payload decoder), errorMessageKey (error code to i18n key mapper), getHealth() — GET /health client (+10 more)

### Community 33 - "Submission Form & Filter Panel"
Cohesion: 0.17
Nodes (12): FieldProps, LocationPicker, Step2Details(), Step2DetailsProps, FilterPanel(), FilterPanelProps, TierChipProps, UseFiltersReturn (+4 more)

### Community 34 - "Health Handler"
Cohesion: 0.18
Nodes (11): Request, ResponseWriter, Time, NewHealthHandler(), T, TestHealthHandler_AllChecksPass_Returns200WithHealthyStatus(), TestHealthHandler_DatabaseUnhealthy_Returns503WithErrorPopulated(), TestHealthHandler_NewCheckerUnhealthy_PropagatesTopLevelStatus() (+3 more)

### Community 35 - "Health Checker Registry Mock Tests"
Cohesion: 0.19
Nodes (10): Call, Context, Controller, NewMockChecker(), T, TestRegistry_RunAll_AllCheckersPass_ReturnsHealthy(), TestRegistry_RunAll_NoCheckers_ReturnsHealthy(), TestRegistry_RunAll_OneCheckerFails_ReturnsUnhealthy() (+2 more)

### Community 36 - "Globe/Manage Filters Feature"
Cohesion: 0.17
Nodes (17): Floating '+' AddEventButton (submission entry point), COUNTRIES constant (lib/countries.ts), DOMAINS constant (hardcoded domain list), EventFilters type (applied filter state shape), EventListItem type (API response item shape), FilterPanel component (globe/table homepage floating filters), ManageDashboard component (filter row + list + pagination), ReviewFilters type (admin/moderator review queue filter state) (+9 more)

### Community 37 - "shadcn/ui Component Config"
Cohesion: 0.12
Nodes (16): aliases, components, hooks, lib, ui, utils, rsc, $schema (+8 more)

### Community 38 - "Locale Layout & Language Selector"
Cohesion: 0.15
Nodes (8): Props, LanguageSelector(), LocaleOption, LOCALES, Toaster(), { Link, useRouter, usePathname, redirect, permanentRedirect }, routing, config

### Community 39 - "Deadline Add/Cancel/Supersede Sessions"
Cohesion: 0.16
Nodes (16): Session: Deadlines — Add time + timezone Fields (cross-cutting), Deadlines: time + timezone fields (cross-cutting prerequisite for supersede), date stays separate; time/timezone are independent nullable additions, Session: Add Deadlines to an Approved Event (POST /api/v1/events/{id}/deadlines), Add Deadlines to an Approved Event feature (collaborative keep-deadlines-fresh), One audit row per logical change, not per row written, No automatic supersession (documented limitation), Session: Cancel a Deadline (PATCH .../deadlines/{deadlineId}/cancel) (+8 more)

### Community 40 - "Project Instructions & FP Principles"
Cohesion: 0.18
Nodes (14): AGENTS.md — Codex Instructions, FP-annotated service layer for event submission (ValidateSubmitEventInput, BuildEventFromInput, BuildDeadlinesFromInput, BuildSubmitterFromInput, BuildSubmission), Session: Admin/Moderator Event Review (PATCH /api/v1/admin/events/{id}/review), GORM Omit(clause.Associations) fix, Reusing ValidateEditedEvent for partial updates, Shared parseDateField helper, CLAUDE.md — ReSEARCH Events Project Instructions, FP: Immutability principle (+6 more)

### Community 41 - "Server Bootstrap & Auth Feature"
Cohesion: 0.14
Nodes (13): Session: Server Bootstrap + Health Check, Server Bootstrap + Health Check feature (GET /health, extensible checker pattern), CORS hand-written over rs/cors library, DBPinger interface over *sql.DB directly, Account lockout after 5 failed login attempts, Session: Auth Feature (Login, Refresh Token, Logout, Account Unlock), CORS default changed from * to http://localhost:3000, In-memory token bucket rate limiter (+5 more)

### Community 42 - "Deadline Management Frontend Planning"
Cohesion: 0.20
Nodes (14): Deadline Management Frontend Planning, Add / Supersede / Cancel Deadline Operations, onManageDeadlines Navigation Prop Pattern, sessionStorage State Passing (no GET /events/:id endpoint), AddDeadlineCard component, Deadline Management Frontend Implementation, DeadlineCard component, DeadlineManagePage component (+6 more)

### Community 43 - "OpenTelemetry Tracer Setup"
Cohesion: 0.26
Nodes (11): main(), TracerProvider, InitTracerProvider(), T, TestInitTracerProvider_EmptyDSN_ReturnsProviderWithoutCallingSentryInit(), TestInitTracerProvider_MalformedDSN_ReturnsError(), TestInitTracerProvider_ValidDSN_InitializesSentryAndRegistersPropagator(), TestTracesSampleRate_DevelopmentReturns1_0() (+3 more)

### Community 44 - "Backend Config Loading"
Cohesion: 0.32
Nodes (12): getEnv(), Load(), T, TestConfig_Load_OverridesDefaultsWithEnvVars(), TestConfig_Load_ReturnsDatabaseURLFromEnv(), TestConfig_Load_ReturnsErrorWhenDatabaseURLMissing(), TestConfig_Load_ReturnsErrorWhenJWTSecretMissing(), TestConfig_Load_ReturnsJWTSecretFromEnv() (+4 more)

### Community 45 - "CORS Middleware"
Cohesion: 0.35
Nodes (12): CORS(), Handler, Handler, T, nextHandler(), TestCORS_DoesNotSetAllowCredentialsForWildcard(), TestCORS_NonOptionsRequestCallsNextHandler(), TestCORS_PreflightOptionsReturns204AndDoesNotCallNext() (+4 more)

### Community 46 - "Globe Homepage Feature Concepts"
Cohesion: 0.21
Nodes (14): EventDetailView component (side panel / bottom sheet), Globe.gl (3D WebGL globe library), handleApiError (centralized error-to-toast handler), InfoModal component (floating info button + legend), lib/constants.ts — APP_VERSION, GITHUB_REPO_URL, GITHUB_PROFILE_URL, ADMIN_EMAIL, lib/events.ts — pin colors, PIN_COLOR_CLUSTER, listEvents(params) — GET /api/v1/events client, supercluster (geographic clustering library) (+6 more)

### Community 47 - "Event Submission & Auth Sessions"
Cohesion: 0.18
Nodes (13): Auth Feature (login, refresh-token, logout, unlock, JWT middleware, rate limiting), Session: Event Submission Feature (POST /api/v1/events/submit), Event Submission Feature (public, contributor lookup/creation, slug reuse), AuditLog construction stays in repository, not service, Contributor lookup/creation/rename on submission, Partial unique index for slug reuse, Events List Feature (filterable, paginated, bbox, tier, first_deadline_month), Admin/Moderator Event Review feature (approve/reject, partial edit, reviewer reason) (+5 more)

### Community 48 - "Frontend Dev Dependencies"
Cohesion: 0.15
Nodes (13): eslint, devDependencies, eslint, @testing-library/react, @types/leaflet, @types/react, typescript, @vitejs/plugin-react (+5 more)

### Community 49 - "Info Button & Location Picker"
Cohesion: 0.23
Nodes (7): InfoButton(), InfoButtonProps, InfoModal(), LocationPicker(), LocationPickerProps, useMediaQuery(), COUNTRY_COORDINATES

### Community 50 - "Frontend Feature Sessions"
Cohesion: 0.20
Nodes (12): Session: API Client + Error Handling (Frontend), API Client + Error Handling feature (typed fetch client, error-to-toast mapping), eslint downgraded from ^10 to ^9, Hand-written src/types/api.ts as temporary exception, Session: Globe Homepage (Frontend), Globe Homepage feature (3D pins, Sheet/Drawer detail view, loading/empty states), generateMetadata fix for Next.js 16 hydration mismatch, middleware.ts renamed to src/proxy.ts (+4 more)

### Community 51 - "Manage Review Frontend Session"
Cohesion: 0.17
Nodes (12): ApproveModal component, buildEventPatch (pure function, diffs only changed fields), CartoDB Voyager Tiles Decision (English map labels), LocationPicker Async Init Race Bug Fix (latLngRef), Management Event Review Frontend, RejectModal component, ReviewStep1Details component, ReviewStep2Decision component (+4 more)

### Community 52 - "Update Password Backend Feature"
Cohesion: 0.18
Nodes (12): GET /api/v1/users/me endpoint, UserHandler.Me handler method, validateSession() function, bcrypt Cost 12 (consistent with auth flow), HashPassword (FP: no side effects), PATCH /api/v1/users/me/password endpoint, Rate Limit Shared With Login (10/min/IP, wraps RequireAuth), Update User Password Backend Feature (+4 more)

### Community 53 - "DB Tracing Integration Tests"
Cohesion: 0.26
Nodes (10): T, TracerProvider, TestBuildHandler_DBQueryProducesChildSpanUnderHTTPSpan(), TestBuildHandler_DBSpanDoesNotLeakRawQueryParameters(), tracerForTest(), beginTx(), DB, M (+2 more)

### Community 54 - "Globe Clustering Hook"
Cohesion: 0.20
Nodes (11): supercluster, ClusterPoint, ClusterProps, PointProps, SuperclusterFeature, toGeoJSONFeatures(), toGlobePoints(), useGlobeClusters() (+3 more)

### Community 55 - "Backend Observability Libraries"
Cohesion: 0.18
Nodes (11): GORM (gorm.io/gorm), gorm.io/plugin/opentelemetry/tracing, OpenTelemetry (go.opentelemetry.io/otel), otelhttp (contrib/instrumentation/net/http), pgx (jackc/pgx/v5), sdktrace (otel/sdk/trace), sentry-go, sentryotel (sentry-go/otel) (+3 more)

### Community 56 - "Manage Portal Frontend Session"
Cohesion: 0.29
Nodes (10): Management Dashboard Frontend Implementation, Own-Event Restriction (moderator cannot review own submission), credentials: include Fix on login() (cross-origin cookie bug), decodeJwt utility, JWT Decoded Client-Side via atob() (no library), /manage/admin welcome dashboard (stub), /manage login page, /manage/moderator welcome dashboard (stub) (+2 more)

### Community 57 - "Globe Clustering & Loading Session"
Cohesion: 0.20
Nodes (10): altitudeToZoom function, ClusterEventDrawer component, EventDetailView component, Globe Event Clustering, GlobeView component, InfoModal (violet PinLegendRow), supercluster library, useGlobeClusters hook (+2 more)

### Community 58 - "Globe View Rendering"
Cohesion: 0.39
Nodes (6): GlobeView(), GlobeViewProps, GlobePoint, isCluster(), getClusterPinRadius(), altitudeToZoom()

### Community 59 - "Database Migrations (Core Tables)"
Cohesion: 0.32
Nodes (4): users, audit_logs, events, deadlines

### Community 60 - "Deep-link & Info Modal Session"
Cohesion: 0.29
Nodes (7): Session: Globe Event Deep-link via URL Slug (Frontend), Globe Event Deep-link feature (?event=SLUG, globe rotation, useSelectedEvent URL sync), Resolve slug from in-memory events list, no extra API call, router.replace instead of router.push for ?event= param, Session: Info Modal — Floating Info Button (Frontend), Info Modal feature (floating ⓘ button, version/legend/submission-flow/credits), New ui/dialog.tsx primitive added

### Community 61 - "Manage Dashboard Session"
Cohesion: 0.43
Nodes (7): EventReviewCard component, GET /api/v1/events endpoint, ManageDashboard component, ManageHeader component, useReviewEvents hook, Management Dashboard Frontend Planning, specs/frontend/manage-dashboard.md

### Community 62 - "Health Check Types"
Cohesion: 0.33
Nodes (4): Context, Context, CheckResult, HealthResponse

### Community 63 - "npm Scripts"
Cohesion: 0.29
Nodes (7): scripts, build, dev, lint, start, test, typecheck

### Community 64 - "OpenTelemetry Tracing Session"
Cohesion: 0.40
Nodes (6): Session: OpenTelemetry Tracing — Spec + Plan (partial), OpenTelemetry Tracing feature (HTTP + DB spans exported to Sentry via sentryotel), ParentBased(TraceIDRatioBased) sampling, rate 1.0 dev / 0.2 prod, sentryotel bridge chosen over manual OTLP exporter, Sentry double-sampling bug: unset TracesSampleRate silently drops all transactions, Session: OpenTelemetry Tracing — Sentry Double-Sampling Bug Fix

### Community 65 - "Globe/Table View Toggle"
Cohesion: 0.33
Nodes (6): useRef + Version Counter Pattern (sync error reads), Globe/Table Mutual Exclusion (never render simultaneously), Globe / Table View Toggle Frontend, onForcedGlobe Callback-via-Ref Pattern (avoids effect loops), useViewMode hook, ViewToggleButton component

### Community 66 - "Session Guard Bug Fixes"
Cohesion: 0.47
Nodes (6): No Redirect When Token/Refresh Token Both Invalid Bug, Session Guard + Eager Token Validation, errors.errors.TOKEN_MISSING Double-Namespace Bug, UpdatePasswordCard component (modified), useSessionGuard hook, useUpdatePassword hook (modified, onAuthError param)

### Community 67 - "Admin User Frontend Components"
Cohesion: 0.33
Nodes (6): checkPasswordComplexity function, PasswordField component, RegisterUserForm component, useRegisterUser hook, UserTableCard component, useUserCard hook

### Community 68 - "Frontend Build Tooling Deps"
Cohesion: 0.33
Nodes (6): pnpm, onlyBuiltDependencies, @parcel/watcher, @sentry/cli, sharp, @swc/core

### Community 69 - "Release Notes & README"
Cohesion: 0.80
Nodes (6): README.md — Project Index (setup, specs, sessions, releases), RELEASES.md — Release Notes, v0.1.0 — Initial Public Release (June 23, 2026), v0.1.1 — Globe Event Clustering (June 24, 2026), v0.1.2 — Globe Loading Message Progression (June 24, 2026), v0.1.3 — Admin User Management (June 26, 2026)

### Community 70 - "Year Filter Semantics Change"
Cohesion: 0.60
Nodes (5): FilterPanel year stepper (globe, no clear button, min = currentYear-2), ListEventsFilter.Year (int -> *int), Step1Search year input (submission wizard, min = currentYear), Year Filter "From Year" Semantics Change, Year Filter Changed From Exact-Match to "year >= value" Semantics

### Community 71 - "Update Password Portal Session"
Cohesion: 0.40
Nodes (5): ManageHeader component (sticky header + avatar menu), sonner toast library / shadcn Toaster, PATCH /api/v1/users/me/password endpoint, useUpdatePassword hook (validation + API call), Update Password — Management Portal (spec)

### Community 72 - "package.json Metadata"
Cohesion: 0.40
Nodes (4): name, packageManager, private, version

### Community 73 - "Events List Feature Session"
Cohesion: 0.50
Nodes (4): Session: Events List Feature (GET /api/v1/events), first_deadline_month via correlated EXISTS+MIN subquery, GORM zero-value bool + gorm:default tag gotcha, Pagination defaults must always be present

### Community 74 - "Event Detail/Table Modifications"
Cohesion: 0.50
Nodes (4): EventDetailContent component (modified), EventDetailView component (modified), EventTableCard component, EventTableView component

### Community 75 - "Language Selector Feature"
Cohesion: 0.50
Nodes (4): src/i18n/navigation.ts (createNavigation), Language Selector Feature, LanguageSelector component, Locale Path Preservation via router.replace(pathname, {locale})

### Community 76 - "Server Route Handler"
Cohesion: 0.67
Nodes (3): Handler, routeFromPattern(), traced()

## Knowledge Gaps
- **297 isolated node(s):** `github.com/yuryalencar/research-events`, `listMeta`, `eventListItemResponse`, `authContextKey`, `eventRepository` (+292 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **60 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `BuildHandler()` connect `Server Route Registration Tests` to `Deadline Repository & Audit Tests`, `Admin Event Review Handler Tests`, `Admin User Handler Tests`, `Health Handler`, `Auth Handler & Cookies`, `GORM Domain Models`, `Auth Middleware Tests`, `OpenTelemetry Tracer Setup`, `Server Route Handler`, `CORS Middleware`, `Backend Config Loading`, `DB Tracing Integration Tests`?**
  _High betweenness centrality (0.086) - this node is a cross-community bridge._
- **Why does `NewEventHandler()` connect `Admin Event Review Handler Tests` to `GORM Domain Models`, `Server Route Registration Tests`, `Event & Deadlines Input Mapping`?**
  _High betweenness centrality (0.044) - this node is a cross-community bridge._
- **Why does `User` connect `GORM Domain Models` to `Deadline Repository & Audit Tests`, `Auth Handler & Cookies`, `Admin User Service & Password Audit`, `Auth Middleware Tests`, `Admin Event Review Service`, `Event Submission Service`, `Mock User Repository`?**
  _High betweenness centrality (0.042) - this node is a cross-community bridge._
- **Are the 96 inferred relationships involving `responseCode()` (e.g. with `TestAdminEventHandler_Review_ApproveNoEdits_Returns200()` and `TestAdminEventHandler_Review_ApproveWithEdits_RecomputesYear()`) actually correct?**
  _`responseCode()` has 96 INFERRED edges - model-reasoned connections that need verification._
- **Are the 106 inferred relationships involving `beginTx()` (e.g. with `TestAuditRepository_Create_PersistsAuditLog()` and `TestAuditRepository_Create_PreservesJSONBDiff()`) actually correct?**
  _`beginTx()` has 106 INFERRED edges - model-reasoned connections that need verification._
- **Are the 83 inferred relationships involving `NewMockEventRepository()` (e.g. with `TestAdminEventHandler_Review_AdminReviewingOwnEvent_Returns200()` and `TestAdminEventHandler_Review_ApproveNoEdits_Returns200()`) actually correct?**
  _`NewMockEventRepository()` has 83 INFERRED edges - model-reasoned connections that need verification._
- **Are the 78 inferred relationships involving `NewMockUserRepository()` (e.g. with `TestAdminUserHandler_ChangeRole_Returns200AndClearsTokensOnSuccess()` and `TestAdminUserHandler_ChangeRole_Returns400ForInvalidRole()`) actually correct?**
  _`NewMockUserRepository()` has 78 INFERRED edges - model-reasoned connections that need verification._