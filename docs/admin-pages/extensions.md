# Manifest-driven admin pages

A plugin can contribute operator-only action pages with `admin.version: 1` and `admin.pages` in plugin.yaml. Each page has an ID, title and actions referencing declared routes. No React import or host product code is needed.

The shared sidebar discovers pages for the selected project, exact enabled/ready installation and granted route capabilities. The generic page accepts JSON input, requires explicit confirmation, and renders JSON or a table of returned rows. The new `plugin.admin.invoke` operation revalidates the project, installation and manifest declaration before calling the existing registry broker. The operator session and audit boundary are unchanged. Metadata never grants access.

Existing specialized pages remain available. This is the first migration layer, not a replacement for existing resource browsers, function editors, or lifecycle controls. Automatic data sources, schema-driven forms, event-driven refresh, cross-plugin slots and custom visual components are follow-up work. Never use a mutating route as an automatic refresh source.

Only plugins with implemented runtime handlers contribute action pages. Arena, invites and subscriptions require runtime route implementations before their admin pages can be enabled.
