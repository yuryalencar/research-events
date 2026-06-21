# Update Password — Management Portal

## Description
Authenticated admins and moderators can change their own password from within
the management portal. Accessed via the avatar dropdown menu in the sticky
header. The form validates all complexity rules in real time (checklist below
the new password field) before the API is called.

## Behaviour
- Clicking the avatar in ManageHeader opens the dropdown. A new "Update
  password" item appears above "Sign out".
- Clicking "Update password" navigates to `/manage/admin/password` or
  `/manage/moderator/password` (role-determined from localStorage).
- The page shows a centered card with three password fields:
    - Current password (show/hide eye toggle)
    - New password (show/hide eye toggle) + live complexity checklist below
    - Confirm new password (show/hide eye toggle)
- Complexity checklist updates on every keystroke (onChange):
    - ✓/✗  At least 8 characters
    - ✓/✗  At least one uppercase letter (A–Z)
    - ✓/✗  At least one lowercase letter (a–z)
    - ✓/✗  At least one special character
- The "Save" button is disabled until the checklist is fully satisfied AND
  confirm matches new password AND current password is non-empty.
- On submit, calls `PATCH /api/v1/users/me/password` via apiPrivateRequest.
- On success: replaces the form with a success state card containing a
  "Back to dashboard" button (→ `/manage/admin` or `/manage/moderator`).
- On API error: handleApiError → sonner toast; form stays open so the user
  can correct and resubmit.
- Page guards itself: if localStorage.manage_user is missing or invalid,
  redirects to /manage.

## Rules
- Complexity rules (validated client-side and enforced server-side):
    - Minimum 8 characters
    - At least one uppercase letter (A–Z)
    - At least one lowercase letter (a–z)
    - At least one special character (non-alphanumeric)
- new_password and new_password_confirmation must match before submit is enabled
- current_password must be non-empty before submit is enabled
- "Save" button disabled while any rule fails — no submit with invalid state
- Submission uses apiPrivateRequest (JWT cookie); on TOKEN_EXPIRED the client
  auto-refreshes and retries exactly once (existing client behaviour)
- No AuditLog or UI audit trail — password changes are not displayed in the portal

## Permissions
- admin and moderator roles only — same session guard as all other manage pages
- Contributors cannot access this page (they have no JWT)

## Error cases
| Scenario | Expected behaviour |
|---|---|
| Missing/empty localStorage session | Redirect to /manage |
| Wrong session role for this route | Redirect to correct role's dashboard |
| current_password is wrong (400 INVALID_CURRENT_PASSWORD) | sonner toast with error message; form stays open |
| Any field empty on submit (400 VALIDATION_ERROR) | Prevented by disabled Save button; should not reach the API |
| New password fails complexity (400 VALIDATION_ERROR) | Prevented by disabled Save button; should not reach the API |
| Passwords don't match (400 VALIDATION_ERROR) | Prevented by disabled Save button; should not reach the API |
| Rate limit exceeded (429) | sonner toast RATE_LIMIT_EXCEEDED |
| Expired JWT (401 TOKEN_EXPIRED) | client auto-refreshes; if still 401, sonner toast + redirect to /manage |
| Network error | sonner toast NETWORK_ERROR |

## Border / corner cases
- All three eye toggles are independent — toggling one does not affect the others
- Checklist items turn green (✓) as each rule is met; red (✗) while unmet
- Confirm field shows an inline mismatch hint below it (not part of the checklist)
- Navigating away mid-form loses unsaved changes — no confirmation dialog needed
- After success, pressing "Back to dashboard" does NOT re-submit anything

## Definition of done
- [ ] Avatar dropdown shows "Update password" above "Sign out" in ManageHeader
- [ ] Clicking "Update password" navigates to role-appropriate password page
- [ ] Password page redirects to /manage when session is missing
- [ ] Card shows three fields: current, new, confirm — each with show/hide toggle
- [ ] Complexity checklist renders below new password field and updates on every keystroke
- [ ] Each checklist item turns green when satisfied, red while unmet
- [ ] Confirm field shows mismatch hint when confirm ≠ new password
- [ ] Save button is disabled until all rules pass + confirm matches + current non-empty
- [ ] Successful submit shows success state with "Back to dashboard" button
- [ ] "Back to dashboard" routes to the correct role dashboard
- [ ] Wrong current password → sonner toast INVALID_CURRENT_PASSWORD
- [ ] Rate limit exceeded → sonner toast RATE_LIMIT_EXCEEDED
- [ ] All i18n keys present in en / pt / es / de
- [ ] useUpdatePassword hook tested (validation logic + API call)
- [ ] users.ts API function tested
