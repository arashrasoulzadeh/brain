# Admin page contract

Every authenticated admin page uses the shared persistent shell in `admin-nextjs/app/components/admin-shell.tsx`:

- `AdminPageFrame` registers its heading and checks required-plugin access; the persistent shell owns the layout.
- The [shared theme](theme.md) defines compact navigation, tables, forms and focus states.
- `AdminSidebar` is the only sidebar implementation.
- `AdminHeader` and `AdminFooter` are shared as well.
- Plugin-backed pages are visible and accessible only when their plugin is ready and enabled.
- The Plugins catalog and plugin detail pages remain accessible so an operator can install and enable a plugin.

The page notes below are the route-by-route contract used for manual verification.
