# Invites

## Developer

Create with `invites.create` payload `{"roomId":"r1","targetPlayerId":"alice","ttlSeconds":60}`. A returned invite is project-scoped, expires, and may be accepted once by its target player. Accept or cancel only through the declared route with the appropriate identity. Treat invite codes as credentials.

## Administrator

Grant create, accept and cancel separately. The current operator routes require trusted host authorization of actual target/player membership before client exposure. The worker stores live invites in memory, bounds capacity and expires old entries; a worker restart is not durable invite recovery. Do not log invite codes.
