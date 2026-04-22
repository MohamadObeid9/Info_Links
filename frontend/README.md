# InfoLinks Frontend

This directory contains the user interface for InfoLinks, built with plain HTML, CSS, and Vanilla JavaScript.

## Structure
- `index.html`: The main entry point of the application.
- `main.js`: Frontend event wiring and app bootstrap.
- `js/`: Contains feature logic modules (`views.js`, `data.js`, etc.).
- `styles/`: Contains all CSS files for styling the application.
- `assets/`: Contains images, icons, and static assets.

## Deployment & Running
The frontend is designed to be served statically by the Go backend. All API requests are made using relative paths (`/api/...`). 

To run the frontend during development, start the Go backend from the repository root:
```bash
go run cmd/server/main.go
```
Then visit `http://localhost:8080/` in your browser.
