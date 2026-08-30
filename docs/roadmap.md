# Roadmap

Project history and planned work. 

---

## Milestones

| Phase | Achievement |
|-------|-------------|
| **Phase 1** | Started with 4 courses covering basics |
| **Phase 2** | Expanded to 25+ courses |
| **Phase 3** | Reached 50+ courses with multiple resources per course |
| **Phase 4** | Serving 300+ students in under a year |
| **Phase 5** | Launched new website for better UX |
| **Phase 6** | Open-sourced project for community contributions |
| **Phase 7** | Favorites, content types, analytics, and PWA support |
| **Phase 8** | Go backend with layered architecture, observability, CI, and SEO |
| **Phase 9** | Student identity without passwords, synced favorites, and unique-user analytics |
| **Phase 10** | Community services, agent/API discovery, graceful shutdown, integration tests, Cloudflare cache |

---

## Future

- [x] Advanced filtering and categorization
- [x] Personalized bookmarks (My Courses / Favorites)
- [ ] Multi-language support
- [x] Community rating system for resources (Feedback)
- [x] Offline mode support (PWA / Service Worker)
- [x] Production Go backend with tests and observability
- [x] Student accounts without email or password (name + number 1-100)
- [x] Guest sessions claimed at signup so pre-signup activity is kept
- [x] Favorites synced to the account instead of one browser
- [x] Unique-user analytics aggregated in SQL (active students per range, top students)
- [x] Admin Students directory with per-student activity timeline
- [x] Cleanup job for stale unclaimed guest rows
- [x] Community services (student businesses / tutoring) with click tracking
- [x] HTTP server timeouts and graceful shutdown on SIGTERM
- [x] Postgres integration tests (repo + HTTP, CI-gated)
- [x] Cloudflare CDN cache for static assets and `GET /api/content` (kept warm with a 10-minute ping)
- [ ] Mobile app (iOS/Android)
- [ ] Push notifications for new resources
- [ ] Course schedule integration
