# Admin page contract

Every authenticated admin page uses the single shared shell in `admin-nextjs/app/components/admin-chrome.tsx`:

- `AdminPageFrame` owns the layout.
- `AdminSidebar` is the only sidebar implementation.
- `AdminHeader` and `AdminFooter` are shared as well.
- Plugin-backed pages are visible and accessible only when their plugin is ready and enabled.
- The Plugins catalog and plugin detail pages remain accessible so an operator can install and enable a plugin.

The page notes below are the route-by-route contract used for manual verification.
