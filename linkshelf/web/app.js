// Client-side script that fetches links and renders them as <li> elements
// On load, fetches /api/links and renders each entry as an <li> showing the title (as an anchor to the URL), the raw URL, and a "Delete" button
// Submits the form via POST /api/links, clears inputs on success, and refreshes the list
// On clicking "Delete", sends DELETE /api/links/{id}, then refreshes the list

// Get the links container element
const linksContainer = document.getElementById('links');

// Fetch links on page load
window.onload = async () => {
  try {
    const response = await fetch('/api/links');
    const links = await response.json();
    renderLinks(links);
  } catch (error) {
    console.error('Error fetching links:', error);
  }
};

// Render links as <li> elements
function renderLinks(links) {
  linksContainer.innerHTML = '';
  links.forEach((link) => {
    const linkElement = document.createElement('li');
    const titleAnchor = document.createElement('a');
    titleAnchor.href = link.url;
    titleAnchor.textContent = link.title;
    const urlText = document.createTextNode(` (${link.url})`);
    const deleteButton = document.createElement('button');
    deleteButton.textContent = 'Delete';
    deleteButton.onclick = async () => {
      try {
        await fetch(`/api/links/${link.id}`, { method: 'DELETE' });
        window.location.reload();
      } catch (error) {
        console.error('Error deleting link:', error);
      }
    };
    linkElement.appendChild(titleAnchor);
    linkElement.appendChild(urlText);
    linkElement.appendChild(deleteButton);
    linksContainer.appendChild(linkElement);
  });
}

// Handle form submission
document.getElementById('link-form').addEventListener('submit', async (event) => {
  event.preventDefault();
  try {
    const title = document.getElementById('title').value;
    const url = document.getElementById('url').value;
    const response = await fetch('/api/links', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ title, url }),
    });
    if (response.ok) {
      document.getElementById('title').value = '';
      document.getElementById('url').value = '';
      window.location.reload();
    } else {
      console.error('Error creating link:', response.status);
    }
  } catch (error) {
    console.error('Error creating link:', error);
  }
});