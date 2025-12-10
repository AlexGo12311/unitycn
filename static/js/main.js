const API_URL = '/api';

async function loadPosts() {
    const response = await fetch(`${API_URL}/posts`);
    const posts = await response.json();
    
    const container = document.getElementById('posts');
    container.innerHTML = posts.map(post => `
        <div class="post" data-id="${post.id}">
            <div class="post-header">
                <strong>${post.user?.username || 'Товарищ'}</strong>
                <span class="slogan">${post.slogan}</span>
            </div>
            <p>${post.content}</p>
            <div class="post-footer">
                <button onclick="likePost(${post.id})">
                    👍 ${post.likes}
                </button>
                <span>${new Date(post.created_at).toLocaleString()}</span>
            </div>
        </div>
    `).join('');
}

async function createPost() {
    const content = document.getElementById('postContent').value;
    const token = localStorage.getItem('token');
    
    if (!token) {
        alert('Товарищ, авторизуйтесь!');
        return;
    }
    
    const response = await fetch(`${API_URL}/posts`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ content })
    });
    
    if (response.ok) {
        document.getElementById('postContent').value = '';
        loadPosts();
    }
}

// Инициализация
document.addEventListener('DOMContentLoaded', loadPosts);