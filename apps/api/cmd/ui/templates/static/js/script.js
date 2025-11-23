// CarFlow UI JavaScript

// Wait for document to be fully loaded
document.addEventListener('DOMContentLoaded', function() {
    
    // Auto-dismiss alerts after 5 seconds
    const alerts = document.querySelectorAll('.alert:not(.alert-permanent)');
    alerts.forEach(function(alert) {
        setTimeout(function() {
            const fadeEffect = setInterval(function() {
                if (!alert.style.opacity) {
                    alert.style.opacity = 1;
                }
                if (alert.style.opacity > 0) {
                    alert.style.opacity -= 0.1;
                } else {
                    clearInterval(fadeEffect);
                    alert.style.display = 'none';
                }
            }, 50);
        }, 5000);
    });
    
    // Enable tooltips everywhere
    const tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
    tooltipTriggerList.map(function(tooltipTriggerEl) {
        return new bootstrap.Tooltip(tooltipTriggerEl);
    });
    
    // Add year validation to relevant form fields
    const yearInputs = document.querySelectorAll('input[type="number"][name="year"]');
    yearInputs.forEach(function(input) {
        input.addEventListener('change', function() {
            const currentYear = new Date().getFullYear();
            const minYear = 1900;
            const value = parseInt(this.value);
            
            if (value < minYear) {
                this.value = minYear;
                showToast('Year adjusted to minimum allowed: ' + minYear);
            } else if (value > currentYear + 1) {
                this.value = currentYear + 1;
                showToast('Year adjusted to maximum allowed: ' + (currentYear + 1));
            }
        });
    });
    
    // Set active navigation based on current page
    const currentPath = window.location.pathname;
    const navLinks = document.querySelectorAll('.navbar-nav .nav-link');
    
    navLinks.forEach(function(link) {
        if (currentPath === link.getAttribute('href')) {
            link.classList.add('active');
        } else if (currentPath.includes('/cars/') && link.getAttribute('href') === '/cars') {
            link.classList.add('active');
        }
    });
});

// Helper function to show a toast notification
function showToast(message) {
    const toastContainer = document.getElementById('toast-container');
    
    if (!toastContainer) {
        const container = document.createElement('div');
        container.id = 'toast-container';
        container.className = 'position-fixed bottom-0 end-0 p-3';
        container.style.zIndex = '11';
        document.body.appendChild(container);
    }
    
    const id = 'toast-' + Date.now();
    const html = `
        <div id="${id}" class="toast" role="alert" aria-live="assertive" aria-atomic="true">
            <div class="toast-header">
                <strong class="me-auto">CarFlow</strong>
                <small>Just now</small>
                <button type="button" class="btn-close" data-bs-dismiss="toast" aria-label="Close"></button>
            </div>
            <div class="toast-body">
                ${message}
            </div>
        </div>
    `;
    
    document.getElementById('toast-container').innerHTML += html;
    const toastElement = document.getElementById(id);
    const toast = new bootstrap.Toast(toastElement);
    toast.show();
    
    setTimeout(function() {
        toastElement.remove();
    }, 5000);
} 