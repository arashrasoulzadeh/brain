# Matchmaking

## Developer

Queue a ticket with `matchmaking.queue` payload `{"id":"ticket1"}`. Use the same ID for status, cancellation or assignment. The ticket ID is the resource for grant checks; separate matchmaking settings from this identity. The Rooms integration is optional.

## Administrator

Grant queue/status/cancel/assign operations only as appropriate and verify the host storage adapter for durable tickets. Do not expose assignment as an unrestricted user operation. Reconcile assigned and cancelled tickets before changing plugin versions; test that project A cannot access project B's ticket.
