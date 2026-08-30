import { AppState } from "./state.js";
import { sb, apiRequest, formatApiError, logApiError } from "./supabase.js";
import { showToast } from "./export.js";
import { esc, setBtnLoading, adminCell } from "./ui.js";
import { loadReportsBadges } from "./data.js";
import { getAdminTableSkeleton } from "./skeleton.js";
import { loadStudentDirectory, senderDetail } from "./students.js";

// Feedback management
const FEEDBACK_PAGE_SIZE = 10;
let currentRating = 0;
let feedbackPage = 0;
let feedbackHasNext = false;

function resetAdminFeedbackPage() {
    feedbackPage = 0;
    feedbackHasNext = false;
}

function _renderFeedbackPager() {
    const pageNum = feedbackPage + 1;
    const prev = feedbackPage > 0
        ? `<button type="button" class="action-btn pager-btn" onclick="setAdminFeedbackPage(-1)">← Prev</button>`
        : `<button type="button" class="action-btn pager-btn" disabled>← Prev</button>`;
    const next = feedbackHasNext
        ? `<button type="button" class="action-btn pager-btn" onclick="setAdminFeedbackPage(1)">Next →</button>`
        : `<button type="button" class="action-btn pager-btn" disabled>Next →</button>`;
    return `<div class="admin-pager">${prev}<span>Page ${pageNum}</span>${next}</div>`;
}

function setAdminFeedbackPage(delta) {
    if (delta < 0 && feedbackPage === 0) return;
    if (delta > 0 && !feedbackHasNext) return;
    feedbackPage = Math.max(0, feedbackPage + delta);
    renderAdminFeedback();
}

async function submitFeedback() {
    if (!window.requireStudent(submitFeedback)) return;
    const btn = document.querySelector("#view-feedback .btn-primary");
    const category = document.getElementById('feedbackCategory').value;
    if (!category) {
        showToast('Please select a category', true);
        return;
    }

    if (currentRating === 0) {
        showToast('Please select a rating', true);
        return;
    }

    const message = document.getElementById('feedbackMessage').value.trim();

    setBtnLoading(btn, true, "Submitting…");
    try {
        await apiRequest("/api/feedback", {
            method: "POST",
            body: { category, rating: currentRating, message }
        });
        showToast('Thank you for your feedback!');
        currentRating = 0;
        document.getElementById('feedbackCategory').value = '';
        document.getElementById('feedbackMessage').value = '';
        updateStarDisplay();
    } catch (err) {
        if (window.handleStudentAuthError?.(err, submitFeedback)) return;
        logApiError(err, 'submitFeedback');
        showToast(formatApiError(err, 'Failed to submit feedback'), true);
    } finally {
        setBtnLoading(btn, false);
    }
}

function setRating(rating) {
    currentRating = rating;
    updateStarDisplay();
}

function handleStarHover(rating) {
    const stars = document.querySelectorAll('#starRating .star');
    stars.forEach((star) => {
        const value = Number(star.dataset.rating || 0);
        star.classList.toggle('hovered', value <= rating);
    });
}

function clearStarHover() {
    document.querySelectorAll('#starRating .star').forEach((star) => star.classList.remove('hovered'));
}

function updateStarDisplay() {
    document.querySelectorAll('#starRating .star').forEach((star) => {
        const value = Number(star.dataset.rating || 0);
        star.classList.toggle('active', value <= currentRating);
    });

    const displayDiv = document.getElementById('ratingDisplay');
    if (currentRating > 0) {
        displayDiv.textContent = `${currentRating} out of 5 stars`;
        displayDiv.style.color = 'var(--accent)';
    } else {
        displayDiv.textContent = 'Select a rating';
        displayDiv.style.color = 'var(--muted)';
    }
}

// Note: star hover events are attached inline in body.html since it loads dynamically.

async function renderAdminFeedback() {
    const contentDiv = document.getElementById('adminContent');
    contentDiv.innerHTML = getAdminTableSkeleton();
    const q = (AppState.adminSearch || "").trim();
    const offset = feedbackPage * FEEDBACK_PAGE_SIZE;

    try {
        const [response] = await Promise.all([
            sb(`feedback?limit=${FEEDBACK_PAGE_SIZE + 1}&offset=${offset}&q=${encodeURIComponent(q)}`, 'GET'),
            loadStudentDirectory(),
        ]);
        const fetched = Array.isArray(response) ? response : (response && response.data) || [];
        feedbackHasNext = fetched.length > FEEDBACK_PAGE_SIZE;
        const feedback = fetched.slice(0, FEEDBACK_PAGE_SIZE);
        if (feedbackPage > 0 && feedback.length === 0) {
            feedbackPage -= 1;
            renderAdminFeedback();
            return;
        }

        let html = `<input class="admin-search" placeholder="🔍 Search feedback…" value="${esc(AppState.adminSearch)}" oninput="AppState.adminSearch=this.value;resetAdminFeedbackPage();renderAdminFeedback()"/>`;
        if (feedback.length === 0) {
            const emptyMsg = q ? `No feedback matching "${esc(q)}" found.` : "No feedback yet.";
            contentDiv.innerHTML = html + `<div style="padding: 20px; text-align: center; color: var(--muted);">${emptyMsg}</div>`;
            return;
        }
        html += `
                <table class="admin-table">
                    <thead>
                        <tr>
                            <th>Sender</th>
                            <th>Date</th>
                            <th>Category</th>
                            <th>Rating</th>
                            <th>Message</th>
                            <th>Status</th>
                            <th>Action</th>
                        </tr>
                    </thead>
                    <tbody>
        `;

        feedback.forEach(item => {
            const date = new Date(item.created_at).toLocaleDateString();
            const filledStars = '★'.repeat(item.rating);
            const emptyStars = '★'.repeat(5 - item.rating);
            const stars = `<span style="color: gold;">${filledStars}</span><span style="color: #999;">${emptyStars}</span>`;
            const ratingText = `${item.rating}/5`;
            const message = esc(item.message) || '(no message)';
            const statusClass = item.status === 'new'
                ? 'tag-blue'
                : item.status === 'rejected'
                    ? 'tag-rejected'
                    : 'tag-resolved';
            const categoryDisplay = item.category ? esc(item.category.charAt(0).toUpperCase() + item.category.slice(1)) : 'N/A';

            html += `
                <tr class="admin-row">
                    ${senderDetail(item.user_id)}
                    ${adminCell("admin-detail", "Date", date)}
                    ${adminCell("admin-detail", "Category", `<span class="tag tag-gray">${categoryDisplay}</span>`)}
                    ${adminCell("admin-pri", "Rating", `<span style="font-size: 1.1rem;" title="${ratingText}">${stars}</span><span style="font-size: 0.9rem; color: var(--text); font-weight: 600; margin-left: 8px;">${ratingText}</span>`)}
                    ${adminCell("admin-sec", "Message", message)}
                    ${adminCell("admin-meta", "Status", `<span class="tag ${statusClass}">${esc(item.status || 'new')}</span>`)}
                    ${adminCell("admin-actions action-btns", "Actions", _feedbackActions(item))}
                </tr>
            `;
        });

        html += `
                    </tbody>
                </table>
        `;
        html += _renderFeedbackPager();

        contentDiv.innerHTML = html;
    } catch (err) {
        console.error('Feedback render error:', err);
        contentDiv.innerHTML = '<div style="color: red; padding: 20px;">Error loading feedback</div>';
    }
}

function _feedbackActions(item) {
    if (item.status === 'rejected') {
        return `<button class="action-btn" onclick="setFeedbackStatus(${item.id},'new','Feedback reopened.')">↩ Reopen</button>
            <button class="action-btn delete-btn del" onclick="confirmAction('Delete this feedback permanently?', () => deleteFeedback(${item.id}))">🗑</button>`;
    }
    if (item.status === 'read') {
        return `<button class="action-btn" onclick="setFeedbackStatus(${item.id},'new','Marked as new.')">↩ Mark new</button>`;
    }
    return `<button class="action-btn" style="color:var(--success); border-color:var(--success)" onclick="setFeedbackStatus(${item.id},'read','Marked as read.')">✓ Mark read</button>
        <button class="action-btn del" onclick="confirmAction('Reject this feedback? It stays in the list as rejected.',()=>setFeedbackStatus(${item.id},'rejected','Feedback rejected.'),'Reject')">✕ Reject</button>`;
}

async function setFeedbackStatus(id, status, toast) {
    try {
        await sb(`feedback?id=eq.${id}`, 'PATCH', { status });
        renderAdminFeedback();
        loadReportsBadges();
        if (toast) showToast(toast);
    } catch (err) {
        logApiError(err, 'updateFeedback');
        showToast(formatApiError(err, 'Failed to update feedback'), true);
    }
}

async function deleteFeedback(id) {
    try {
        await sb(`feedback?id=eq.${id}`, 'DELETE');
        renderAdminFeedback();
        loadReportsBadges();
        showToast('Feedback deleted');
    } catch (err) {
        logApiError(err, 'deleteFeedback');
        showToast(formatApiError(err, 'Failed to delete feedback'), true);
    }
}
Object.assign(window, {
  submitFeedback,
  setRating,
  handleStarHover,
  clearStarHover,
  renderAdminFeedback,
  setAdminFeedbackPage,
  resetAdminFeedbackPage,
  toggleFeedbackStatus: setFeedbackStatus,
  setFeedbackStatus,
  deleteFeedback,
});

export { updateStarDisplay, renderAdminFeedback };
