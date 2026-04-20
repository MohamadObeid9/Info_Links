# 📚 Info Links

A modern, open-source platform designed to help students at Le CNAM Lebanon's Computer Science Department discover and organize course materials and learning resources in one centralized, user-friendly hub.

## 🌟 What is Info Links?

**Info Links** started as a simple idea to help students find course materials more easily. Today, it's evolved into a comprehensive platform covering **50+ courses** with hundreds of curated links — making it an essential resource for students pursuing their License and Master's degrees in Computer Science at Le CNAM Lebanon.

### The Growth Story
📈 **From Humble Beginnings to Impact**
- Started with just **4 courses** covering basic materials
- Grew to **50+ courses** with hundreds of curated resources
- Serving **300+ students** in under a year
- Recently launched **new website** to diversify and improve user experience
- **Open-sourcing the project** to empower the community and encourage contributions

### The Impact
- **50+ Courses Covered** — From foundational to advanced subjects
- **Multi-disciplinary Resources** — Supports both License and partial Master's program courses  
- **Growing Community** — Trusted by 300+ students across semesters
- **Regularly Updated** — New resources added consistently
- **Telegram Channel** — Parallel channel at [@Info_Links9](https://t.me/Info_Links9) for real-time updates

---

## ✨ Features

### For Students
- 🔍 **Smart Search** — Find courses by name or code instantly with keyboard shortcut (`/` or `Ctrl+K`)
- 📋 **Organized by Program** — Sorted by year, semester, and specialization  
- 🏷️ **Easy Navigation** — Filter courses and identify optional vs. mandatory classes
- 🔗 **Multiple Resource Types** — Google Drive, Google Classroom, Telegram, and more — each with a color-coded badge
- ⭐ **My Courses (Favorites)** — Star courses to save them locally in your browser for quick access
- 🏷️ **Content Type Labels** — See what each link contains at a glance: TD, Cours, Videos, Sessions, Exams
- 🔗 **Multiple Content Types** — Links can have multiple content categories (e.g., TD + Cours + Videos)
- 🌓 **Light/Dark Mode** — Comfortable viewing in any lighting, with automatic system detection and persistence
- 📱 **Fully Responsive** — Works seamlessly on desktop, tablet, and mobile
- 💬 **Report & Contribute** — Report broken links with course/link dropdowns, or contribute new resources with link type identification
- ⭐ **Feedback System** — Rate the platform (1-5 stars) by category and share suggestions
- 🔍 **Deep Linking** — Hash-based routing for direct view access (e.g., `#report-submit`)
- 🌐 **Multi-language Notes** — Important announcements available in English, French, and Arabic
- ⌨️ **Keyboard Shortcuts** — `/` or `Ctrl+K` to search, `Esc` to close modals
- 📱 **PWA Support** — Installable as a Progressive Web App with service worker

### For Admins
- ➕ **Full Course Management** — Add, edit, delete, and organize courses with program/year/semester placement
- 🏷️ **Advanced Labeling** — Mark courses as optional or mandatory
- 🔁 **Sibling Course Detection** — Courses shared across programs auto-sync names, codes, and links
- 🔗 **Multi-Content Link Management** — Assign multiple content types per link (TD, Cours, Videos, Sessions, Exams)
- 📊 **Analytics Dashboard** — View visitor statistics:
  - Daily visitor counts with interactive bar chart
  - 7-day, 30-day, 90-day range selection
  - Top clicked links tracking
  - Export data as JSON for further analysis
- ✅ **Smart Contribution Review** — Approve user-submitted links with:
  - Grouped course selector (by program)
  - Pre-filled link type from contributor
  - Automatic sibling course detection for multi-program courses
- 🚨 **Report Management** — Handle user reports and improve content quality
- 💬 **Feedback Management** — View and manage user feedback with star ratings by category
- 🔐 **Secure Admin Panel** — Password-protected access with Supabase authentication
- 📦 **Extra Resources** — Manage additional resource sections beyond regular courses

### Content Type Legend
Each link can display one or more content type badges:
| Badge | Meaning |
|-------|---------|
| ✏️ TD | Travaux Dirigés (exercises/tutorials) |
| 📄 Cours | Course materials/lectures |
| 🎬 Videos | Video recordings |
| 🎤 Sessions | Live session recordings |
| 📝 Exams | Exam papers and solutions |
| 📦 Other | Other types of content |

### Link Type Legend
| Badge | Meaning |
|-------|---------|
| **TG** | Telegram |
| **GD** | Google Drive |
| **GC** | Google Classroom |
| **OT** | Other / External |

---

## 🛠️ Tech Stack

- **Frontend:** HTML5, CSS3, JavaScript (Vanilla)
- **Backend:** Go (Golang) — A high-performance REST API and static file server
- **Styling:** Modern, custom CSS with theme support (light/dark) and responsive design
- **Backend Database:** Supabase (PostgreSQL) with Row Level Security
- **Authentication:** Supabase Auth (email/password)
- **Deployment:** Render (Go Web Service) + Supabase
- **PWA:** Service Worker for offline support
- **Version Control:** Git & GitHub

---

## 🚀 Getting Started

### Prerequisites
- A modern web browser
- Git (for cloning the repository)
- Go (Golang) installed on your system
- Supabase account (if self-hosting the database)

### Installation & Running Locally

1. **Clone the repository**
   ```bash
   git clone https://github.com/MohamadObeid9/Info_Links.git
   cd Info_Links
   ```

2. **Configure Environment**
   - Copy `backend/.env.example` to `backend/.env`
   - Set your `DATABASE_URL` to your Supabase Postgres connection string.

3. **Run the Full Stack**
   ```bash
   cd backend
   go run cmd/server/main.go
   ```
   - The Go server will start the API and serve the frontend files natively.
   - Open your browser to `http://localhost:8080`

### Deployment (Render)
InfoLinks is configured to be deployed as a single Web Service on Render. 
- **Build Command:** `cd backend && go build -o server cmd/server/main.go`
- **Start Command:** `cd backend && ./server`
- Don't forget to set your `DATABASE_URL` environment variable!

---

## 📖 How to Use

### For Students
1. **Browse Courses** — Start on the home page to see all available courses
2. **Search** — Use the search bar (or press `/`) to find specific courses
3. **Filter** — Use program tabs and year/semester filters to narrow down
4. **Access Resources** — Click on resource links (Google Drive, Telegram, Classroom, etc.)
5. **Save Favorites** — Click the ★ star on any course to save it to "My Courses"
6. **Report Issues** — Use the "Report" section to report broken links by selecting course and link from dropdowns
7. **Contribute** — Submit new links with URL, link type, and notes
8. **Toggle Theme** — Switch between light and dark mode using the theme button
9. **Provide Feedback** — Rate the platform by category and share suggestions

### For Admins
1. **Log In** — Access the admin panel via the "Admin" button
2. **Manage Courses** — Add, edit, or remove courses and resources
3. **Add Links** — Assign multiple content types to each link
4. **Sibling Sync** — When editing shared courses, choose to update all occurrences
5. **Review Contributions** — Approve user-submitted links with smart course matching
6. **View Analytics** — Track visitor data, top clicked links, and usage patterns
7. **Export Data** — Download analytics as JSON for reporting
8. **Manage Feedback** — Review and respond to user feedback and ratings

---

## 🎨 Design Highlights

- **Modern Interface** — Clean, sleek design with smooth animations
- **Accessibility** — Easy navigation and clear typography (Inter font family)
- **Performance** — Lightweight with client-side caching (1-hour TTL)
- **Responsive Design** — Adaptive layout for all screen sizes
- **Theme Support** — Automatic system detection with manual toggle and persistence
- **Interactive Elements** — Smooth hover effects, micro-animations, and loading states
- **Content Security Policy** — XSS protection via CSP headers

---

## 🤝 Contributing

We love contributions from the community! Whether it's bug fixes, new features, or additional resources:

1. **For Resources** — Use the "Report / Contribute" feature in the app
2. **For Code** — Fork the repository and submit a pull request:
   ```bash
   git checkout -b feature/your-feature-name
   git commit -m "Add your feature"
   git push origin feature/your-feature-name
   ```

---

## 📞 Connect With Us

- **GitHub** — [MohamadObeid9/Info_Links](https://github.com/MohamadObeid9/Info_Links)
- **Telegram Channel** — [@Info_Links9](https://t.me/Info_Links9) — Real-time updates and discussions
- **Live Site** — Check the CNAME for current domain

---

## 📜 License

This project is **open source** and available under the MIT License. See the [LICENSE](LICENSE) file for more details.

---

## 🙏 Acknowledgments

- **Claude AI** — Built and continuously improved using Claude
- **Supabase** — Reliable database infrastructure
- **Le CNAM Lebanon** — For the Computer Science program and student community
- **All Contributors** — Thank you for making this project a success!
- **The Student Community** — For the feedback, support, and belief in this project

---

## 📊 Project Milestones

| Phase | Achievement |
|-------|-------------|
| **Phase 1** | Started with 4 courses covering basics |
| **Phase 2** | Expanded to 25+ courses |
| **Phase 3** | Reached 50+ courses with multiple resources per course |
| **Phase 4** | Serving 300+ students in under a year |
| **Phase 5** | Launched new website for better UX |
| **Phase 6** | Open-sourced project for community contributions |
| **Phase 7** | Added favorites, content types, analytics, and PWA support |
| **Phase 8** | Fully migrated to a high-performance Go backend with native static serving |

---

## 💡 Future Roadmap

- [x] Advanced filtering and categorization
- [x] Personalized bookmarks (My Courses / Favorites)
- [x] Multi-language support (EN/FR/AR notes)
- [ ] Mobile app (iOS/Android)
- [ ] Integration with more platforms
- [ ] AI-powered course recommendations
- [x] Community rating system for resources (Feedback)
- [x] Offline mode support (PWA / Service Worker)
- [ ] Push notifications for new resources
- [ ] Course schedule integration

---

**Made with ❤️ for Le CNAM Lebanon Computer Science Students**

