// Feedback management
let currentRating = 0;

async function submitFeedback() {
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
        const res = await fetch("/api/feedback", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ category, rating: currentRating, message })
        });
        if (!res.ok) throw new Error("Server error");
        showToast('Thank you for your feedback!');
        currentRating = 0;
        document.getElementById('feedbackCategory').value = '';
        document.getElementById('feedbackMessage').value = '';
        updateStarDisplay();
        showView('home');
    } catch (err) {
        console.error('Feedback submission error:', err);
        showToast('Failed to submit feedback', true);
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

    try {
        const response = await sb('feedback', 'GET', null, null, '*&order=created_at.desc');
        const feedback = Array.isArray(response) ? response : response.data || [];

        if (feedback.length === 0) {
            contentDiv.innerHTML = '<div style="padding: 20px; text-align: center; color: var(--muted);">No feedback yet</div>';
            return;
        }

        let html = `
            <div style="overflow-x: auto;">
                <table class="admin-table">
                    <thead>
                        <tr>
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
            const statusClass = item.status === 'new' ? 'tag-blue' : 'tag-gray';
            const categoryDisplay = item.category ? esc(item.category.charAt(0).toUpperCase() + item.category.slice(1)) : 'N/A';

            html += `
                <tr>
                    <td>${date}</td>
                    <td><span class="tag tag-gray">${categoryDisplay}</span></td>
                    <td style="white-space: nowrap; min-width: 140px;"><span style="font-size: 1.2rem;" title="${ratingText}">${stars}</span><span style="font-size: 0.9rem; color: var(--text); font-weight: 600; margin-left: 8px;">${ratingText}</span></td>
                    <td style="max-width: 300px; word-wrap: break-word;">${message}</td>
                    <td><span class="tag ${statusClass}">${item.status || 'new'}</span></td>
                    <td>
                        <button class="action-btn" onclick="toggleFeedbackStatus(${item.id}, '${item.status}')" title="Toggle status" style="font-size: 0.9rem;">${item.status === 'new' ? '✓ Mark read' : '↩ Mark new'}</button>
                        <button class="action-btn delete-btn" onclick="confirmAction('Delete this feedback?', () => deleteFeedback(${item.id}))" title="Delete">🗑</button>
                    </td>
                </tr>
            `;
        });

        html += `
                    </tbody>
                </table>
            </div>
        `;

        contentDiv.innerHTML = html;
    } catch (err) {
        console.error('Feedback render error:', err);
        contentDiv.innerHTML = '<div style="color: red; padding: 20px;">Error loading feedback</div>';
    }
}

async function toggleFeedbackStatus(id, currentStatus) {
    try {
        const newStatus = currentStatus === 'new' ? 'read' : 'new';
        await sb(`feedback?id=eq.${id}`, 'PATCH', { status: newStatus });
        renderAdminFeedback();
        loadReportsBadges();
        showToast(`Marked as ${newStatus}`);
    } catch (err) {
        console.error('Error updating feedback:', err);
        showToast('Failed to update feedback', true);
    }
}

async function deleteFeedback(id) {
    try {
        await sb(`feedback?id=eq.${id}`, 'DELETE');
        renderAdminFeedback();
        loadReportsBadges();
        showToast('Feedback deleted');
    } catch (err) {
        console.error('Error deleting feedback:', err);
        showToast('Failed to delete feedback', true);
    }
}
window.submitFeedback = submitFeedback; window.setRating = setRating; window.handleStarHover = handleStarHover; window.clearStarHover = clearStarHover;
