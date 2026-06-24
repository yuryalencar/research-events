# Globe Loading Message Progression

## Description

When the globe homepage fetches events, the loading overlay shows a progression of messages
instead of a single static label. After a short delay the message shifts to communicate that
the server may be cold-starting (Render free tier spins down after 15 minutes of inactivity),
so users understand the wait and stay on the page.

## Scope

Globe view on the homepage only — the `isLoading` overlay in `app/[locale]/page.tsx`.
No other loading states are affected.

## Behaviour

- When `isLoading` becomes `true`, show message 1 immediately.
- After **2500 ms** of continuous loading, switch to message 2.
- Every **1000 ms** after that, advance to the next message.
- When the last message is reached, stay on it — no looping.
- When `isLoading` becomes `false`, all timers are cleared and the overlay disappears
  (identical to current behaviour — no success state added).
- If `isLoading` flips `true` again (new filter fetch), timers reset and progression
  restarts from message 1.

## Message sequence (en)

| # | Delay from start | Message |
|---|---|---|
| 1 | 0 ms | "Loading events…" |
| 2 | 2500 ms | "Waking up our server…" |
| 3 | 3500 ms | "First load takes a moment…" |
| 4 | 4500 ms | "Almost there, hang tight…" |

## i18n

All four locales must have every key — no missing keys allowed.

| Key | en | pt | es | de |
|---|---|---|---|---|
| `home.loadingStep1` | Loading events… | Carregando eventos… | Cargando eventos… | Ereignisse werden geladen… |
| `home.loadingStep2` | Waking up our server… | Acordando nosso servidor… | Despertando nuestro servidor… | Server wird gestartet… |
| `home.loadingStep3` | First load takes a moment… | O primeiro carregamento demora um pouco… | La primera carga tarda un momento… | Erster Ladevorgang dauert einen Moment… |
| `home.loadingStep4` | Almost there, hang tight… | Quase lá, aguarde… | Casi listo, un momento… | Fast fertig, kurz warten… |

The existing `home.loading` key is replaced by these four keys. No other translation
files outside of `messages/` are affected.

## Styling

Identical to the current loading overlay — no visual changes to the container,
spinner, or text style.

## Permissions

Public — no authentication required.

## Error cases

| Scenario | Expected behaviour |
|---|---|
| `isLoading` becomes `false` before 2500 ms | No message switch ever happens; overlay disappears normally |
| New filter fetch starts mid-progression | Timers reset; progression restarts from message 1 |
| Component unmounts while loading | All timers cleared immediately (no memory leak, no state update on unmounted component) |
| User stays on last message (4500 ms+) | Stays on message 4 — no looping back to message 1 |

## Implementation note

Logic lives in a new hook `hooks/useLoadingMessage.ts`. The hook receives
`isLoading: boolean` and returns `message: string`. The page component replaces
the static `t("loading")` call with `message` from the hook.

## Definition of done

- [ ] `useLoadingMessage` hook created and unit-tested
- [ ] Message starts at index 0 immediately when `isLoading = true`
- [ ] Switches to index 1 after exactly 2500 ms of continuous loading
- [ ] Advances +1 every 1000 ms until the last message
- [ ] Stays on last message — does not loop
- [ ] Timers cleared when `isLoading` becomes `false`
- [ ] Timers reset when `isLoading` flips `true` again (new fetch)
- [ ] `home.loading` key replaced by `home.loadingStep1`–`home.loadingStep4` in all 4 locales
- [ ] No missing translation keys in any locale file
- [ ] Styling of the loading overlay is unchanged
- [ ] All error cases in the table above have a corresponding test
