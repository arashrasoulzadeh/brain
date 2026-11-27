# Control-plane theme

The admin interface uses a neutral light surface, compact 224px sidebar, consistent eight-pixel surfaces, dense tables and 32px controls. Theme tokens are defined in `admin-nextjs/app/theme.css`; the specialized code editor retains its dark editing surface. Navigation is grouped into workspace, project tools, extensions and platform, filtered by the existing discovery rules. The menu search matches page names and plugin identifiers. Mobile uses an expandable navigation panel. Current-page links expose `aria-current`, and forms retain keyboard focus indicators.

Project selection stays in the existing Projects screen, and changes refresh the navigation without widening access. Plugin extensions retain their independently authorized pages. The persistent shell displays the active page's title, including dynamic model and plugin pages.
