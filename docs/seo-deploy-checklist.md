# SEO deploy checklist

After merging SEO changes to `main` and deploying production:

1. Set `SITE_BASE_URL` in production env to your live domain (no trailing slash), e.g. `https://infolinks.example.com`.
2. Redeploy the Go server and rebuilt frontend if applicable.
3. Verify pages:
   - `/robots.txt` references your sitemap URL
   - `/sitemap.xml` lists `/`, `/courses`, `/course/{code}`, `/program/{slug}`
   - `/course/{known-code}` returns HTML with links and FAQ
   - `/?highlight={code}` opens home with search and scroll
4. [Google Search Console](https://search.google.com/search-console): verify property, submit `sitemap.xml`.
5. Request indexing for your top ~20 course codes (`/course/nfa035`, etc.).
6. Share course URLs in [@Info_Links9](https://t.me/Info_Links9) and year-group chats (not only `/`).
7. Monthly: review Search Console queries; tune page titles for pages ranking positions 8–15.

While finishing `go-backend-migration`, merge `main` into that branch weekly to keep SEO routes in sync.
