# Admin home

Route: `/`

The home page is the operator entry point. Use the shared admin shell to establish an operator session and navigate the authorized project and plugin pages. Project and installation information comes from the authenticated host API; unavailable or disabled plugins do not grant access through navigation.

Verify an unauthenticated visit, successful operator sign-in, session revocation and authenticated navigation. Do not treat visible links as an authorization control: every API action is independently checked by the host.
