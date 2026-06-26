# Admin User Management

## Description

Admin-only UI for managing platform users. Two new routes under the existing management
area: a paginated, filterable user list with inline role change, password reset, and
unlock; and a registration form for new admin/moderator accounts.

Also adds a "Back to Globe" link to the existing events dashboard (`ManageDashboard`)
and a "Back to Events" link on the users list page to complete the navigation chain:
Globe ← Events Dashboard ← Users List ← Register.

## Backend prerequisites (all live)

| Endpoint | Purpose |
|---|---|
| `GET  /api/v1/admin/users` | Paginated list with filters |
| `POST /api/v1/admin/users` | Register new admin/moderator |
| `PATCH /api/v1/admin/users/{id}/role` | Change a user's role |
| `PATCH /api/v1/admin/users/{id}/password` | Reset a user's password |
| `PATCH /api/v1/admin/users/{id}/unlock` | Unlock a locked account |

## Routes

| Route | Page |
|---|---|
| `/[locale]/manage/admin/users` | User list |
| `/[locale]/manage/admin/users/register` | Register form |

Both are admin-only (`useSessionGuard("admin")`).

---

## Layout — Users List (`/manage/admin/users`)

```
┌──────────────────────────────────────────────────────┐
│ [Globe] Yury Lima                               [Y ▾]│  ← ManageHeader (unchanged)
├──────────────────────────────────────────────────────┤
│ ← Back to Events                  [Register User →]  │  ← breadcrumb row
├──────────────────────────────────────────────────────┤
│ [Search by name or email__________________________]  │
│ Roles: [admin] [moderator] [contributor]             │  ← multi-select chips
│ [Locked] [Deleted]                  [Reset] [Apply]  │  ← toggle badges + actions
├──────────────────────────────────────────────────────┤
│ ┌────────────────────────────────────────────────┐   │
│ │ Alice · alice@example.com                  [∨] │   │  ← collapsed
│ │ [admin]                                        │   │
│ └────────────────────────────────────────────────┘   │
│ ┌────────────────────────────────────────────────┐   │
│ │ Bob · bob@example.com          [moderator][∧]  │   │  ← expanded
│ ├────────────────────────────────────────────────┤   │
│ │ Change Role                                    │   │
│ │ [Contributor ▾]                  [Apply Role]  │   │
│ ├────────────────────────────────────────────────┤   │
│ │ Reset Password                                 │   │
│ │ New password  [________________] 👁            │   │
│ │ Confirm       [________________] 👁            │   │
│ │ ● 8+ chars  ● Uppercase  ● Lowercase ● Special │   │
│ │                             [Apply Password]   │   │
│ ├────────────────────────────────────────────────┤   │
│ │ Unlock Account  ← only when locked_at ≠ null  │   │
│ │ This account is currently locked.              │   │
│ │                               [Unlock User]    │   │
│ └────────────────────────────────────────────────┘   │
├──────────────────────────────────────────────────────┤
│ ← Prev      Page 1 of 3 (72 users)       Next →     │
└──────────────────────────────────────────────────────┘
```

---

## Filter area

Filters are **Apply-gated** — no fetch is triggered until the Apply button is clicked.
Apply resets to page 1. Reset clears all fields back to defaults (no fetch until Apply).

| Filter | Type | Default |
|---|---|---|
| Search | Text input (placeholder: "Search by name or email") | empty |
| Roles | Multi-select chips: admin / moderator / contributor | none selected (= all roles) |
| Locked | Toggle badge | off |
| Deleted | Toggle badge | off |

Role chips: each chip independently toggleable; multiple can be active simultaneously.
No chip active = no role filter (backend returns all roles).

Locked/Deleted badges: each independently toggleable.
- Locked ON → `?locked=true`
- Deleted ON → `?include_deleted=true`

---

## User card — collapsed state

One card per user. Shows:
- **Name** · **email** (truncated if long)
- **Role badge** (admin / moderator / contributor)
- **Locked badge** — shown only when `locked_at !== null`
- **Deleted badge** — shown only when `deleted_at !== null`
- **Chevron toggle** (expand / collapse)

---

## User card — expanded state

Three independent sections separated by horizontal dividers. Each section has its own
action button, confirmation modal, and success/error banner. Sections never share state.

### Section 1 — Change Role

- Dropdown pre-selected with current role (admin / moderator / contributor)
- **Apply Role** button — disabled when dropdown value equals the user's current role
- On click → confirmation modal: "Change role from {currentRole} to {newRole}?"
- On confirm → `PATCH /api/v1/admin/users/{id}/role`
- On success → green banner "Role updated to {newRole}" inside this section; badge in
  collapsed summary updates to the new role
- On error → red banner inside this section; dropdown resets to original role
- `ROLE_UNCHANGED` from backend → red banner (should not happen — Apply is disabled in this case, but handled defensively)

### Section 2 — Reset Password

- **New password** — text/password input with visibility toggle
- **Confirm password** — text/password input with visibility toggle
- **Complexity checklist** (live, inline below the new password field):
  - ● 8+ characters
  - ● Uppercase letter
  - ● Lowercase letter
  - ● Special character
- **Apply Password** button — disabled until:
  - New password is non-empty AND meets all complexity rules
  - Confirm matches new password
- If confirm is filled but doesn't match → inline error below confirm field; Apply disabled
- On click → confirmation modal: "Password will be updated for {name}."
- On confirm → `PATCH /api/v1/admin/users/{id}/password`
- On success → green banner "Password updated"; both fields cleared; complexity resets
- On error → red banner inside this section; fields remain editable

### Section 3 — Unlock Account

- **Only visible** when `locked_at !== null`
- Description text: "This account is currently locked."
- **Unlock User** button
- No confirmation modal — unlock is immediately called on button click
- On click → `PATCH /api/v1/admin/users/{id}/unlock`
- On success → green banner "Account unlocked"; Locked badge disappears from collapsed
  summary; this section disappears (account is no longer locked)
- On error → red banner inside this section; button re-enabled

### Deleted user behaviour

When `deleted_at !== null`, the card is still expandable but all three section controls
are **disabled** (dropdown greyed out, password inputs greyed out, no unlock section shown).
A banner at the top of the expanded area reads: "This user account has been deleted."

---

## Layout — Register Page (`/manage/admin/users/register`)

```
┌──────────────────────────────────────────────────────┐
│ [Globe] Yury Lima                               [Y ▾]│  ← ManageHeader (unchanged)
├──────────────────────────────────────────────────────┤
│ ← Back to Users                                      │
│                                                      │
│ Register User                                        │
│ ┌──────────────────────────────────────────────────┐ │
│ │ Name     [__________________________________]    │ │
│ │ Email    [__________________________________]    │ │
│ │ Role     [Moderator ▾]  (admin | moderator only)│ │
│ │ Password [__________________________________] 👁  │ │
│ │ ● 8+  ● Uppercase  ● Lowercase  ● Special       │ │
│ │ Confirm  [__________________________________] 👁  │ │
│ │                             [Register User]      │ │
│ └──────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────┘
```

### Form fields

| Field | Type | Validation |
|---|---|---|
| Name | Text input | Required, non-empty |
| Email | Email input | Required, non-empty |
| Role | Dropdown | Required; options: Admin / Moderator only (contributor excluded) |
| Password | Password input + visibility toggle | Required, must meet complexity |
| Confirm password | Password input + visibility toggle | Required, must match password |

Client-side validation runs continuously; **Register User** button is disabled until
all fields are valid and passwords match.

### Submit flow

1. User clicks **Register User** (only enabled when all client-side checks pass)
2. Confirmation modal: "Create {role} account for {email}?" — Cancel / Create
3. On Create → `POST /api/v1/admin/users`
4. On 201 → replace form with success screen (same route, no navigation)
5. On error → red banner above the form; fields stay editable; modal closes

### Success screen (replaces form, same URL)

```
      ✓  (green check icon)
      User registered
      {name} has been added as {role}.

   ┌─────────────────────────────┐
   │ Name:   Alice               │
   │ Email:  alice@example.com   │
   │ Role:   Admin               │
   └─────────────────────────────┘

   [Register another user]   [Back to users list]
```

"Register another user" resets all form state (fields, errors, complexity).
"Back to users list" navigates to `/manage/admin/users`.

---

## ManageDashboard addition

Add a **"Back to Globe"** link in the breadcrumb row below ManageHeader, navigating
to `/{locale}`. Placed on the left, styled as a plain text link (same style as
"← Back to Events" on the users page).

---

## Permissions

- Both users pages: `useSessionGuard("admin")` — non-admins redirect to `/manage`
- Register User button navigates to `/manage/admin/users/register` (admin-only route)

---

## API — new functions in `lib/api/admin.ts`

```typescript
listAdminUsers(params: ListAdminUsersParams): Promise<{ data: AdminUserListItem[]; meta: ApiMeta }>
registerAdminUser(input: RegisterAdminUserInput): Promise<RegisterAdminUserResult>
changeUserRole(id: number, role: string): Promise<ChangeUserRoleResult>
resetUserPassword(id: number, input: ResetUserPasswordInput): Promise<ResetUserPasswordResult>
// unlockUser already exists
```

## Types — additions to `types/api.ts`

```typescript
interface AdminUserListItem {
  id: number
  name: string
  email: string
  role: string
  created_at: string
  locked_at: string | null
  deleted_at: string | null
}

interface ListAdminUsersParams {
  search?: string
  roles?: string        // comma-separated: "admin,moderator"
  locked?: boolean
  include_deleted?: boolean
  page?: number
  page_size?: number
  pagination?: "on" | "off"
}

interface RegisterAdminUserInput {
  name: string
  email: string
  password: string
  role: string
}

interface RegisterAdminUserResult {
  user: { id: number; name: string; email: string; role: string; created_at: string }
}

interface ChangeUserRoleResult {
  user: { id: number; name: string; email: string; role: string }
}

interface ResetUserPasswordInput {
  new_password: string
  new_password_confirmation: string
}

interface ResetUserPasswordResult {
  user: { id: number; name: string; email: string; role: string }
}
```

---

## Shared component extraction

`PasswordField` and `ComplexityItem` are currently private sub-components in
`UpdatePasswordCard.tsx`. Both the register form and the reset-password card section
need the same UI. Extract them to `components/ui/PasswordField.tsx` and update
`UpdatePasswordCard` to import from there.

A pure utility function `checkPasswordComplexity(password: string): PasswordComplexity`
will live in `lib/utils.ts` (extracted from `useUpdatePassword`'s useMemo logic).

---

## i18n keys (`manage.users.*` namespace, all 4 locales)

```
manage.users.backToEvents              — "← Back to Events"
manage.users.registerButton            — "Register User"
manage.users.backToGlobe              — "← Back to Globe"  (used in ManageDashboard)
manage.users.filters.searchPlaceholder — "Search by name or email"
manage.users.filters.rolesLabel        — "Roles"
manage.users.filters.lockedBadge       — "Locked"
manage.users.filters.deletedBadge      — "Deleted"
manage.users.filters.apply             — "Apply"
manage.users.filters.reset             — "Reset"
manage.users.list.loading              — "Loading users…"
manage.users.list.error                — "Failed to load users. Try again."
manage.users.list.empty                — "No users match the current filters."
manage.users.list.pageLabel            — "Page {page} of {totalPages} ({total} users)"
manage.users.list.prevButton           — "← Prev"
manage.users.list.nextButton           — "Next →"
manage.users.card.lockedBadge          — "Locked"
manage.users.card.deletedBadge         — "Deleted"
manage.users.card.expand               — "Expand user"
manage.users.card.collapse             — "Collapse user"
manage.users.card.deletedBanner        — "This user account has been deleted."
manage.users.card.roleSection          — "Change Role"
manage.users.card.roleLabel            — "Role"
manage.users.card.applyRole            — "Apply Role"
manage.users.card.roleSuccess          — "Role updated to {role}"
manage.users.card.roleConfirmTitle     — "Change role"
manage.users.card.roleConfirmBody      — "Change role from {from} to {to}?"
manage.users.card.passwordSection      — "Reset Password"
manage.users.card.newPasswordLabel     — "New password"
manage.users.card.confirmPasswordLabel — "Confirm password"
manage.users.card.passwordMismatch     — "Passwords do not match"
manage.users.card.applyPassword        — "Apply Password"
manage.users.card.passwordSuccess      — "Password updated"
manage.users.card.passwordConfirmTitle — "Reset password"
manage.users.card.passwordConfirmBody  — "Password will be updated for {name}."
manage.users.card.unlockSection        — "Unlock Account"
manage.users.card.unlockDescription    — "This account is currently locked."
manage.users.card.unlockButton         — "Unlock User"
manage.users.card.unlockSuccess        — "Account unlocked"
manage.users.card.confirmCancel        — "Cancel"
manage.users.card.confirmConfirm       — "Confirm"
manage.users.register.pageTitle        — "Register User"
manage.users.register.backToUsers      — "← Back to Users"
manage.users.register.nameLabel        — "Name"
manage.users.register.emailLabel       — "Email"
manage.users.register.roleLabel        — "Role"
manage.users.register.rolePlaceholder  — "Select role"
manage.users.register.passwordLabel    — "Password"
manage.users.register.confirmLabel     — "Confirm password"
manage.users.register.passwordMismatch — "Passwords do not match"
manage.users.register.submitButton     — "Register User"
manage.users.register.confirmTitle     — "Confirm registration"
manage.users.register.confirmBody      — "Create {role} account for {email}?"
manage.users.register.confirmCancel    — "Cancel"
manage.users.register.confirmCreate    — "Create"
manage.users.register.errorBanner      — "Registration failed. Please try again."
manage.users.register.success.title    — "User registered"
manage.users.register.success.subtitle — "{name} has been added as {role}."
manage.users.register.success.nameLabel — "Name"
manage.users.register.success.emailLabel — "Email"
manage.users.register.success.roleLabel  — "Role"
manage.users.register.success.registerAnother — "Register another user"
manage.users.register.success.backToUsers     — "Back to users list"
```

New `errors.*` keys (add to existing errors namespace):
```
errors.EMAIL_ALREADY_EXISTS      — "A user with this email already exists."
errors.CANNOT_CHANGE_OWN_PASSWORD — "Use your own profile settings to change your password."
errors.PASSWORD_TOO_WEAK         — "Password must be at least 8 characters with uppercase, lowercase, and a special character."
errors.ROLE_UNCHANGED            — "User already has this role."
errors.CANNOT_CHANGE_OWN_ROLE    — "You can't change your own role."
```

---

## Files to create / modify

| File | Action |
|---|---|
| `src/components/ui/PasswordField.tsx` | **Create** — extract shared PasswordField + ComplexityItem |
| `src/lib/utils.ts` | **Modify** — add `checkPasswordComplexity` pure function |
| `src/components/manage/UpdatePasswordCard.tsx` | **Modify** — import PasswordField from ui/ |
| `src/types/api.ts` | **Modify** — add 6 new types |
| `src/lib/api/admin.ts` | **Modify** — add 4 new API functions |
| `src/lib/api/admin.test.ts` | **Modify** — add tests for new functions |
| `src/hooks/useAdminUsers.ts` | **Create** — filter + fetch + pagination (same pattern as useReviewEvents) |
| `src/hooks/useAdminUsers.test.ts` | **Create** — unit tests |
| `src/hooks/useUserCard.ts` | **Create** — per-card state (role change, password reset, unlock) |
| `src/hooks/useUserCard.test.ts` | **Create** — unit tests |
| `src/hooks/useRegisterUser.ts` | **Create** — form state + confirmation + submit |
| `src/hooks/useRegisterUser.test.ts` | **Create** — unit tests |
| `src/components/manage/users/UserTableCard.tsx` | **Create** — expandable card with 3 sections |
| `src/components/manage/users/UserFilters.tsx` | **Create** — search + role chips + badges |
| `src/components/manage/users/UserListPage.tsx` | **Create** — list + filters + pagination |
| `src/components/manage/users/RegisterUserForm.tsx` | **Create** — form page (form + success screen) |
| `src/app/[locale]/manage/admin/users/page.tsx` | **Create** — users list route |
| `src/app/[locale]/manage/admin/users/register/page.tsx` | **Create** — register route |
| `src/components/manage/ManageDashboard.tsx` | **Modify** — add "← Back to Globe" link |
| `src/messages/en.json` | **Modify** — add manage.users.* + errors.* keys |
| `src/messages/pt.json` | **Modify** — translated keys |
| `src/messages/es.json` | **Modify** — translated keys |
| `src/messages/de.json` | **Modify** — translated keys |

---

## Error cases

| Scenario | Behaviour |
|---|---|
| List fetch error | Red message in place of card list |
| Role change error | Red banner inside Change Role section; dropdown resets |
| Password reset error | Red banner inside Reset Password section |
| Unlock error | Red banner inside Unlock Account section; button re-enabled |
| Register 409 EMAIL_ALREADY_EXISTS | Red banner above form: "A user with this email already exists." |
| Register 400 VALIDATION_ERROR | Red banner above form with backend message |
| Register generic error | Red banner above form |
| Password mismatch (client) | Inline error below confirm; Apply Password / Register disabled |
| Complexity not met (client) | Checklist shows unmet items; Apply Password / Register disabled |
| Session expired mid-action | Redirect to `/manage` |
| Admin changes own role via UI | 422 from backend; red banner "You can't change your own role." |
| Admin resets own password via UI | 422 from backend; red banner in Reset Password section |

---

## Border / corner cases

- Role unchanged → Apply Role button disabled (client-side guard; 409 handled defensively)
- Password empty + confirm empty → Apply Password button disabled (no API call)
- Unlock section disappears after successful unlock; Locked badge removed from summary
- Deleted user card: all controls disabled; deleted banner at top of expanded section
- No role chips selected → no `roles` param sent → all roles returned by backend
- Apply while filters unchanged → refetch page 1 (same behaviour as ManageDashboard)
- "Register another user" resets all fields and complexity state back to blank
- Long names/emails in card collapsed row truncate with ellipsis

---

## Definition of done

- [ ] `/manage/admin/users` renders user list; filters are Apply-gated; Apply resets to page 1
- [ ] Search by name/email works
- [ ] Role chips allow multi-select; none selected = all roles shown
- [ ] Locked / Deleted badges filter independently
- [ ] Collapsed card: name, email, role badge, Locked badge (if locked), Deleted badge (if deleted)
- [ ] Expanded card: three sections visible (unlock only when locked)
- [ ] Change Role: dropdown pre-selected; Apply disabled when role unchanged; confirmation modal; success/error banners; summary badge updates
- [ ] Reset Password: Apply disabled until complexity met + match; confirmation modal; success/error banners; fields cleared on success
- [ ] Unlock: immediate call on button click; success removes section + badge; error re-enables button
- [ ] Deleted user: all controls disabled; deleted banner shown
- [ ] Pagination: Prev/Next + page label; same pattern as ManageDashboard
- [ ] "← Back to Events" link on users list page
- [ ] "Register User →" button navigates to register page
- [ ] `/manage/admin/users/register` form validates all fields client-side
- [ ] Complexity checklist live on password field
- [ ] Confirmation modal before POST
- [ ] 201 → success screen with name/email/role summary
- [ ] "Register another user" resets form; "Back to users list" navigates
- [ ] 409 EMAIL_ALREADY_EXISTS → error banner with correct message
- [ ] ManageDashboard has "← Back to Globe" link
- [ ] Both pages protected by `useSessionGuard("admin")`
- [ ] `PasswordField` extracted to `components/ui/` and `UpdatePasswordCard` updated
- [ ] `checkPasswordComplexity` in `lib/utils.ts` with unit tests
- [ ] All new API functions have unit tests
- [ ] `useAdminUsers`, `useUserCard`, `useRegisterUser` all have unit tests
- [ ] All `manage.users.*` and new `errors.*` keys in all 4 locales (en, pt, es, de)
- [ ] `pnpm typecheck` passes; `pnpm test` passes
