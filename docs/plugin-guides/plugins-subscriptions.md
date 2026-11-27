# Subscriptions

## Developer

Open `subscriptions.open` with `{"channel":"rooms.connected"}`. Use the returned ticket and cursor with `subscriptions.receive`; if the cursor is older than retained events, use `subscriptions.resync`. Close with the ticket and channel. The worker retains a bounded stream and validates tickets against the trusted project and channel.

## Administrator

The current polling adapter is operator-only and supports the room-connected channel, not general public WebSocket authorization. Grant the observation capability and all required operations. Monitor persisted observation delivery separately from the in-memory ticket and stream state. Reopen subscriptions after a worker restart.
