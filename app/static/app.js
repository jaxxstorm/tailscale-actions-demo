// Fetch and display user information
async function fetchUserInfo() {
    try {
        const response = await fetch('/api/user');
        const data = await response.json();
        
        const userInfoDiv = document.getElementById('user-info');
        
        if (data.connected) {
            userInfoDiv.innerHTML = `
                <div class="user-profile">
                    <div class="user-avatar">${data.first_initial || '?'}</div>
                    <div class="user-details">
                        <h3>${data.display_name || data.login_name || 'Unknown User'}</h3>
                        <p>${data.login_name || 'No login information'}</p>
                        <span class="badge badge-connected">✓ Connected via Tailscale</span>
                    </div>
                </div>
            `;
        } else {
            userInfoDiv.innerHTML = `
                <div class="user-profile">
                    <div class="user-avatar">?</div>
                    <div class="user-details">
                        <h3>Not Connected</h3>
                        <p>Connect via Tailscale to see your information</p>
                        <span class="badge badge-disconnected">✗ Not Connected</span>
                        ${data.error ? `<p class="error-message" style="margin-top: 10px;">${data.error}</p>` : ''}
                    </div>
                </div>
            `;
        }
        
        userInfoDiv.classList.remove('loading');
    } catch (error) {
        console.error('Error fetching user info:', error);
        document.getElementById('user-info').innerHTML = `
            <div class="error-message">
                <strong>Error:</strong> Failed to load user information. ${error.message}
            </div>
        `;
        document.getElementById('user-info').classList.remove('loading');
    }
}

// Fetch and display products
async function fetchProducts() {
    try {
        const response = await fetch('/api/products');
        const data = await response.json();
        
        const productsDiv = document.getElementById('products-info');
        
        if (data && data.length > 0) {
            const productsHTML = data.map(product => {
                // Handle created_at field if it exists
                let formattedDate = '';
                if (product.created_at) {
                    const date = new Date(product.created_at);
                    formattedDate = date.toLocaleDateString('en-US', { 
                        year: 'numeric', 
                        month: 'short', 
                        day: 'numeric' 
                    });
                }
                
                // Build stock badge if stock_quantity exists
                const stockBadge = product.stock_quantity !== undefined && product.stock_quantity !== null
                    ? `<span class="stock-badge ${product.stock_quantity > 50 ? 'stock-high' : product.stock_quantity > 0 ? 'stock-low' : 'stock-out'}">
                        ${product.stock_quantity > 0 ? `${product.stock_quantity} in stock` : 'Out of stock'}
                       </span>`
                    : '';
                
                // Build category badge if category exists
                const categoryBadge = product.category 
                    ? `<span class="category-badge">${product.category}</span>`
                    : '';
                
                // Parse price as a number (it comes as string from database)
                const price = parseFloat(product.price) || 0;
                
                // Build the product card
                return `
                    <div class="product-item">
                        <div class="product-header">
                            <h3>${product.name || 'Unnamed Product'}</h3>
                            ${categoryBadge}
                        </div>
                        <p>${product.description || 'No description available'}</p>
                        <div class="product-footer">
                            <div class="product-price">$${price.toFixed(2)}</div>
                            ${stockBadge}
                        </div>
                        ${formattedDate ? `<div class="product-date">Added: ${formattedDate}</div>` : ''}
                    </div>
                `;
            }).join('');
            
            productsDiv.innerHTML = `
                <div class="products-grid">
                    ${productsHTML}
                </div>
            `;
        } else {
            productsDiv.innerHTML = `
                <div class="no-data">
                    <p>No products found in the database.</p>
                    <p style="margin-top: 10px; font-size: 0.9rem;">Add products to see them here!</p>
                </div>
            `;
        }
        
        productsDiv.classList.remove('loading');
    } catch (error) {
        console.error('Error fetching products:', error);
        document.getElementById('products-info').innerHTML = `
            <div class="error-message">
                <strong>Error:</strong> Failed to load products. ${error.message}
            </div>
        `;
        document.getElementById('products-info').classList.remove('loading');
    }
}

// Fetch and display health status
async function fetchHealth() {
    try {
        const response = await fetch('/health');
        const data = await response.json();
        
        const healthDiv = document.getElementById('health-info');
        
        const getStatusClass = (status) => {
            if (status === 'ok' || status === 'connected' || status === 'Running') return 'status-ok';
            if (status === 'disconnected') return 'status-error';
            return 'status-warning';
        };
        
        healthDiv.innerHTML = `
            <div class="health-status">
                <div class="health-item">
                    <h3>Overall Status</h3>
                    <div class="health-value ${getStatusClass(data.status)}">
                        ${data.status.toUpperCase()}
                    </div>
                </div>
                <div class="health-item">
                    <h3>Database</h3>
                    <div class="health-value ${getStatusClass(data.database)}">
                        ${data.database === 'connected' ? '✓ Connected' : '✗ Disconnected'}
                    </div>
                </div>
                <div class="health-item">
                    <h3>Tailscale</h3>
                    <div class="health-value ${getStatusClass(data.tailscale)}">
                        ${data.tailscale === 'connected' ? '✓ Connected' : data.tailscale}
                    </div>
                </div>
            </div>
        `;
        
        healthDiv.classList.remove('loading');
    } catch (error) {
        console.error('Error fetching health:', error);
        document.getElementById('health-info').innerHTML = `
            <div class="error-message">
                <strong>Error:</strong> Failed to load health status. ${error.message}
            </div>
        `;
        document.getElementById('health-info').classList.remove('loading');
    }
}

// Initialize the app
document.addEventListener('DOMContentLoaded', () => {
    fetchUserInfo();
    fetchProducts();
    fetchHealth();
    
    // Refresh data every 30 seconds
    setInterval(() => {
        fetchUserInfo();
        fetchProducts();
        fetchHealth();
    }, 30000);
});
