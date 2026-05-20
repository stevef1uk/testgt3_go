document.addEventListener('DOMContentLoaded', () => {
    const bookmarkList = document.getElementById('bookmark-list');
    const addBookmarkForm = document.getElementById('add-bookmark-form');

    // Load bookmarks on page load
    loadBookmarks();

    // Handle form submission
    addBookmarkForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const title = document.getElementById('title').value.trim();
        const url = document.getElementById('url').value.trim();

        if (!title || !url) {
            alert('Please fill in both title and URL');
            return;
        }

        try {
            const response = await fetch('/api/bookmarks', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ title, url })
            });

            if (response.ok) {
                // Clear form
                addBookmarkForm.reset();
                // Reload bookmarks
                loadBookmarks();
            } else {
                const error = await response.json();
                alert('Failed to add bookmark: ' + (error.error || response.statusText));
            }
        } catch (err) {
            alert('Network error: ' + err.message);
        }
    });

    // Load bookmarks from API
    async function loadBookmarks() {
        try {
            const response = await fetch('/api/bookmarks');
            if (!response.ok) {
                throw new Error('Failed to fetch bookmarks');
            }
            const bookmarks = await response.json();
            renderBookmarks(bookmarks);
        } catch (err) {
            console.error(err);
            bookmarkList.innerHTML = '<li>Error loading bookmarks</li>';
        }
    }

    // Render bookmarks list
    function renderBookmarks(bookmarks) {
        if (bookmarks.length === 0) {
            bookmarkList.innerHTML = '<li>No bookmarks yet. Add one above!</li>';
            return;
        }

        bookmarkList.innerHTML = bookmarks.map(b => `
            <li>
                <div class="bookmark">
                    <h3><a href="${b.url}" target="_blank" rel="noopener">${b.title}</a></h3>
                    <p>${b.url}</p>
                </div>
            </li>
        `).join('');
    }
});
