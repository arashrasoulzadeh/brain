# Analytics

## Developer

Capture an event with `analytics.capture` and payload `{"name":"room.opened","fields":{"members":2}}`. Names and field names are bounded; values must be finite numbers, booleans or bounded strings. The worker supplies the trusted project. The room-connected observer receives `roomId` and `reconnected` from the host, not a forged client event. Avoid credentials and personal information in event fields.

## Administrator

Grant the capture capability and, if observing Rooms, the declared observation capability. Install the publisher in the same project. Verify observation audit delivery and use an explicit durable sink if durable analytics retention is required; the default TypeScript event sink is in memory. Do not treat successful capture as proof of external analytics delivery.
