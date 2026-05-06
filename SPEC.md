# Dancegobot Business Logic

Last reviewed: 2026-05-06.

This document describes product and business rules for Dancegobot. It is meant
to be used as a shared reference when changing code, reviewing behavior, or
discussing edge cases. It intentionally focuses on user-visible behavior, not
technical implementation details.

## Product Scope

Dancegobot is a Telegram bot for dance event registration.

The product supports two main personas:

- Organizer: creates an event post, configures event registration rules, and
  monitors registrations.
- Dancer: registers for an event as part of a couple or as a single looking for
  a partner.

The current product is optimized for events with two dance roles:

- Leader.
- Follower.

Other role models, same-role couples, switch roles, and role-less events are
not supported at this stage.

## Core Concepts

### Event

An event is the registration entity created from an inline query. It contains:

- Announcement text.
- Owner profile.
- Event settings.
- Registered couples.
- Single dancers looking for a partner.
- Original Telegram post information, if the event was published.

### Event Owner

The event owner is the Telegram user who created the event through the inline
query. Only the owner can manage event settings.

Current owner actions:

- Change event auto-pairing mode.
- Change couple limit.
- Close or reopen registration.

Organizer-side moderation of participants, such as removing another dancer,
editing participant names, or blocking dancers, is not currently planned.

### Dancer

A dancer is a person registered in an event.

A dancer may have:

- A Telegram profile, when the bot knows the user.
- Only a manually entered display name, when another dancer typed the partner
  name.

When a dancer has a Telegram profile, identity is based on Telegram user ID.
When a dancer has no profile, the bot can only infer identity from a Telegram
username if the manually entered name contains a valid `@username`.

Manually entered `@username` entries should be discoverable by the real Telegram
user when the username matches their current Telegram username. Telegram
username matching is case-insensitive. Plain manually entered names are not
claimable by the real Telegram user.

### Couple

A couple contains exactly two dancers:

- One leader.
- One follower.

A couple can be created by one dancer without explicit confirmation from the
partner. This is an intentional product rule.

The couple creation time defines its priority for event limits and waitlist
position.

### Single

A single is a dancer who is registered as looking for a partner.

Singles are not counted against the couple limit. Only couples are counted.

### Waitlist

The term "waitlist" refers to couples that are registered after the active
couple limit is full.

Singles looking for a partner are a separate concept and should not be confused
with the couple waitlist.

## Event Creation

### Inline Query

An organizer creates an event by typing an inline query:

```text
@dancegobot Event announcement text
```

An empty inline query does not create an event. It only returns a hint.

A non-empty inline query creates an event draft and returns one inline article
result that can be published into a Telegram chat or channel.

### Event Drafts

An event starts as a draft. It becomes a published event once Telegram provides
an inline message ID through either:

- Chosen inline result.
- Signup button callback from the event post.

Draft events without a published post, couples, or singles may be cleaned up
after the configured draft retention period.

### Announcement Text

The inline query text becomes the event announcement text, except for supported
inline settings shortcuts.

Announcement text containing only whitespace is treated as empty input.

The event announcement text is reused in:

- The original Telegram event post.
- Private `/my` event view.
- Notification messages.

### Couple Limit Shortcut

The inline query can contain a limit shortcut as a separate token:

```text
Event announcement /5
```

This sets the couple limit to 5 and removes the shortcut from the event
announcement.

Shortcut rules:

- Supported shortcut values are `/1` through `/99`.
- Values outside this range must not be accepted as a limit shortcut.
- Unsupported values such as `/0` and `/100` are treated as ordinary
  announcement text.
- The slash must be preceded by whitespace, so normal text such as dates
  `3/4/2026` is not interpreted as a limit shortcut.
- The shortcut should be clearly separated from announcement text and should not
  accidentally consume ordinary event text.

Current intended limit entry points:

- Organizer settings UI supports 0..20 couples.
- Inline shortcut supports 1..99 couples.
- Limit 0 means no couple limit.

## Event Settings

### Default Organizer Settings

Each organizer can configure default settings for newly created events:

- Auto-pairing on or off.
- Default couple limit.

Changing default organizer settings affects only future events. It does not
change existing events.

### Event-Specific Settings

The owner can change settings for a specific event after it is created:

- Auto-pairing.
- Couple limit.
- Registration open or closed.

## Registration Entry Flow

Dancers start registration from the event post by choosing one of the role
buttons:

- Leader button.
- Follower button.

The button opens a private bot deep link. Registration continues in the private
chat with the bot.

If the event is removed or registration is closed, the bot shows a corresponding
message and does not show registration controls.

## Registering a Couple

A dancer can register a couple by:

- Typing the partner name.
- Sharing a Telegram contact/user, when supported by the Telegram client.
- Choosing an existing single dancer from the reply keyboard, when auto-pairing
  is disabled.
- Typing the visible single dancer number from the private reply keyboard, for
  example `1`, `1.`, or `1. Alice`.

The dancer's selected role is the dancer's own role. The partner is assigned the
opposite role.

Manual partner names containing only whitespace are rejected as empty input.
When a dancer types a single dancer number, the number must be greater than zero
and no greater than the number of visible single partner options.

Business rules:

- One dancer may register the whole couple without partner confirmation.
- A dancer who is already in a couple cannot register another couple without
  removing the current registration first.
- A partner who is already in another couple cannot be selected.
- A dancer cannot register with themselves.
- A dancer cannot register with a partner in the same role.
- If the dancer was single, they are removed from singles.
- If the selected partner was single, the partner is removed from singles and
  notified.
- If the selected partner was not single, no confirmation is required.
- If a couple is created after the active couple limit is full, it is still
  registered, but it is placed in the waitlist.

## Registering as Single

A dancer can register as single, meaning they are looking for a partner in the
opposite role.

Current behavior:

- If auto-pairing is enabled, the bot tries to immediately pair the dancer with
  the first available opposite-role single.
- If auto-pairing is disabled, dancers can manually choose from visible
  opposite-role singles.
- If no suitable partner is selected or found, the dancer remains in the singles
  list.

When auto-pairing is disabled and visible opposite-role singles are available,
the UI should not offer the "I am looking for a partner" action. In that case,
the dancer should choose an available single or enter a partner manually.

When auto-pairing is enabled, the "I am looking for a partner" action remains
available because it is the entry point for automatic matching.

## Auto-Pairing

Auto-pairing is an event setting.

When enabled:

- A new single dancer is paired with the first available opposite-role single.
- A restored single dancer may also be auto-paired after their previous couple
  is removed.
- The existing opposite-role single receives a notification that a partner was
  found.

When disabled:

- The bot does not create couples automatically.
- Dancers can choose available opposite-role singles manually.

Changing auto-pairing mode must not retroactively pair already waiting singles.
This is intentional: an organizer may toggle settings while preparing or
reviewing an event, and existing singles should not be changed by that alone.

## Removing Registration

A dancer can remove their own registration only while registration is open.

Closing registration intentionally blocks:

- New registrations.
- Changes to existing registrations.
- Registration removal.

If a single removes registration:

- The dancer is removed from the singles list.

If a dancer in a couple removes registration:

- The whole couple is removed.
- The removed dancer becomes not registered.
- The partner behavior depends on how the partner entered the couple.

Partner behavior after couple removal:

- If the partner originally registered as single, the partner is restored to the
  singles list or auto-paired with another suitable single.
- If the partner created the couple, the partner is notified that the other
  dancer cancelled.
- If the partner has no Telegram profile, the bot cannot notify them.

If the removed couple was inside the active limit and waitlisted couples exist,
the first waitlisted couple moves into the active list. Eligible dancers in that
couple are notified.

## Couple Limits and Waitlist

The couple limit is the maximum number of couples in the active participant
list.

Rules:

- Limit 0 means no limit.
- The limit applies to couples only.
- Singles do not count toward the limit.
- Couples are ordered by couple creation time.
- Couples with index lower than the limit are active participants.
- Couples after the limit are in the waitlist.
- Registration is not rejected when the limit is full.

When the owner changes the event limit:

- Some couples may enter the waitlist.
- Some couples may leave the waitlist.
- The owner receives a private notice listing affected couples.
- The owner can choose whether to notify affected dancers.

## Closing and Reopening Registration

When registration is closed:

- The event post shows a closed marker.
- The event post no longer exposes registration role buttons.
- New registrations are blocked.
- Existing registration changes are blocked.
- Registration removal is blocked.

When registration is reopened:

- Registration role buttons are available again.
- Dancers can register, change by removing and registering again, or remove
  their own registration.

## Event Post Rendering

The event post contains:

- Announcement text.
- Couples list, when at least one couple exists.
- Waitlist section, when a couple limit exists and registered couples exceed
  the limit.
- Singles list, when at least one single exists.

Couples are shown as:

```text
Leader - Follower
```

Singles are grouped by role. The current rendering chooses which role heading
appears first based on which role has more singles.

Dancers with known Telegram profiles are rendered as profile links. Manually
entered dancers without profiles are rendered as plain text.

## `/my`

The `/my` command shows events related to the current Telegram user:

- Published events owned by the user.
- Published events where the user is a participant.

Draft events and removed events are not shown.

The private `/my` view can include:

- Event text.
- Current participant lists.
- Registration action buttons, when registration is open.
- Link to the original chat post, when the bot knows the original chat message.
- Event settings button for the owner.

Participant lookup is based on:

- Telegram user ID stored in structured participant profiles.
- Telegram username stored in structured participant profiles.
- Valid manually entered `@username` values that match the real user's current
  Telegram username.

Manual plain-name entries do not appear in the real user's `/my`.

## Notifications

The bot sends notifications only to participants who explicitly interacted with
the bot in the context of the event.

Eligible notification recipients:

- The dancer who created a couple through the bot.
- A dancer who registered as single through the bot.

A participant with a Telegram profile who was only added or shared by another
dancer is not notified solely because their profile is known.

Within those eligibility rules, the bot sends notifications in these business
cases:

- A dancer who was single is selected by someone as a partner.
- A dancer who was single is auto-paired.
- A dancer who was restored from a cancelled couple is auto-paired with a new
  partner.
- A dancer who was originally single is restored to singles after partner
  cancellation.
- The creator of a couple is notified when the other dancer cancels.
- A waitlisted couple leaves the waitlist after an active couple is removed.
- Affected dancers are notified after a limit change, if the owner chooses to
  send notifications.

Notification delivery requires the bot to be able to message the recipient.
Notification delivery also requires the recipient to be eligible under the
event-interaction rule above.

Notifications may include a link to the original event post when:

- The bot knows the original chat and message.
- Telegram can produce a usable link for that chat type.

## Removed Events

An event can be marked as removed when the bot repeatedly fails to render the
original inline event post and the failure indicates that the post probably no
longer exists.

Removed events:

- Are hidden from `/my`.
- Cannot accept new registrations.
- Cannot accept registration changes.
- Are not rendered again.

There is currently no owner-facing restore or republish workflow for removed
events.

## Accepted Platform and Product Constraints

The following constraints are accepted for the current product stage:

- One event has one original event post.
- Forwarded copies of the event post are not updated.
- A forwarded event post may show stale participant information.
- The bot is not guaranteed to observe all forwarded or copied Telegram posts.
  This is accepted as a Telegram platform limitation.
- The private `/my` view acts as the reliable current view outside the original
  post.
- Partner confirmation is not required.
- Organizer participant moderation is not planned.
- Roles are strictly leader and follower.
- Role changes are not a supported workflow. A dancer must remove registration
  and register again when registration is open.
- Organizer settings UI supports limits up to 20 couples.
- Inline limit shortcut supports limits up to 99 couples.

## Known Product Follow-Up Candidates

These are not all confirmed bugs. They are product/business topics that should
be reviewed before related changes.

- Avoid using "waitlist" wording for singles looking for a partner.
- Make unsupported role switching clearer in the private registration scene.
- Make limit-change notification prompts resilient to multiple consecutive
  limit edits.
- Provide an owner-facing explanation or recovery path when an event becomes
  removed because the original post cannot be rendered.
