// Free Games Bot Documentation JavaScript

class BotDocumentation {
    constructor() {
        this.apiBaseUrl = '';
        this.init();
    }

    init() {
        this.loadBotStats();
        this.setupEventListeners();
        this.setupAnimations();
        this.setupNavigation();
    }

    // Load bot statistics from API
    async loadBotStats() {
        try {
            const response = await fetch('/api/status');
            const data = await response.json();
            
            this.updateStats(data);
            this.updateStatus(data.status);
        } catch (error) {
            console.error('Failed to load bot stats:', error);
            this.showError('Failed to load bot statistics');
        }
    }

    // Update statistics display
    updateStats(data) {
        const serverCountEl = document.getElementById('server-count');
        const gameCountEl = document.getElementById('game-count');
        
        if (serverCountEl) {
            this.animateNumber(serverCountEl, data.server_count || 0);
        }
        
        if (gameCountEl) {
            this.animateNumber(gameCountEl, data.game_count || 0);
        }

        // Update last updated time
        const lastUpdateEl = document.getElementById('last-update');
        if (lastUpdateEl && data.last_update) {
            const date = new Date(data.last_update);
            lastUpdateEl.textContent = date.toLocaleString();
        }
    }

    // Update bot status indicator
    updateStatus(status) {
        const statusElements = document.querySelectorAll('.bot-status');
        statusElements.forEach(el => {
            el.className = `status ${status === 'online' ? 'status-online' : 'status-offline'}`;
            el.textContent = status.charAt(0).toUpperCase() + status.slice(1);
        });
    }

    // Animate number counting
    animateNumber(element, targetNumber) {
        const startNumber = parseInt(element.textContent) || 0;
        const duration = 1000; // 1 second
        const startTime = performance.now();

        const updateNumber = (currentTime) => {
            const elapsed = currentTime - startTime;
            const progress = Math.min(elapsed / duration, 1);
            
            // Easing function for smooth animation
            const easeOutQuart = 1 - Math.pow(1 - progress, 4);
            const currentNumber = Math.round(startNumber + (targetNumber - startNumber) * easeOutQuart);
            
            element.textContent = currentNumber.toLocaleString();
            
            if (progress < 1) {
                requestAnimationFrame(updateNumber);
            }
        };

        requestAnimationFrame(updateNumber);
    }

    // Setup event listeners
    setupEventListeners() {
        // Refresh stats button
        const refreshBtn = document.getElementById('refresh-stats');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => {
                this.loadBotStats();
                this.showSuccess('Statistics refreshed!');
            });
        }

        // Copy code blocks
        this.setupCodeCopyButtons();

        // Setup command examples
        this.setupCommandExamples();

        // Setup FAQ toggles
        this.setupFAQToggles();
    }

    // Setup code copy functionality
    setupCodeCopyButtons() {
        const codeBlocks = document.querySelectorAll('.code-block');
        codeBlocks.forEach(block => {
            const copyBtn = document.createElement('button');
            copyBtn.className = 'copy-btn';
            copyBtn.innerHTML = '📋 Copy';
            copyBtn.style.cssText = `
                position: absolute;
                top: 10px;
                right: 10px;
                background: var(--primary-color);
                color: white;
                border: none;
                padding: 5px 10px;
                border-radius: 4px;
                cursor: pointer;
                font-size: 0.8rem;
            `;
            
            block.style.position = 'relative';
            block.appendChild(copyBtn);
            
            copyBtn.addEventListener('click', () => {
                const code = block.textContent.replace('📋 Copy', '').trim();
                navigator.clipboard.writeText(code).then(() => {
                    copyBtn.innerHTML = '✅ Copied!';
                    setTimeout(() => {
                        copyBtn.innerHTML = '📋 Copy';
                    }, 2000);
                });
            });
        });
    }

    // Setup command examples with interactive features
    setupCommandExamples() {
        const commandCards = document.querySelectorAll('.command-card');
        commandCards.forEach(card => {
            card.addEventListener('click', () => {
                card.classList.toggle('expanded');
                const details = card.querySelector('.command-details');
                if (details) {
                    details.style.display = details.style.display === 'none' ? 'block' : 'none';
                }
            });
        });
    }

    // Setup FAQ toggles
    setupFAQToggles() {
        const faqItems = document.querySelectorAll('.faq-item');
        faqItems.forEach(item => {
            const question = item.querySelector('.faq-question');
            const answer = item.querySelector('.faq-answer');
            
            if (question && answer) {
                question.addEventListener('click', () => {
                    const isOpen = answer.style.display === 'block';
                    answer.style.display = isOpen ? 'none' : 'block';
                    question.classList.toggle('active');
                });
            }
        });
    }

    // Setup animations
    setupAnimations() {
        // Intersection Observer for fade-in animations
        const observerOptions = {
            threshold: 0.1,
            rootMargin: '0px 0px -50px 0px'
        };

        const observer = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    entry.target.classList.add('fade-in');
                }
            });
        }, observerOptions);

        // Observe all cards and sections
        document.querySelectorAll('.card, .section, .setup-step').forEach(el => {
            observer.observe(el);
        });
    }

    // Setup navigation
    setupNavigation() {
        const navLinks = document.querySelectorAll('.nav-link');
        const sections = document.querySelectorAll('.content-section');

        navLinks.forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const targetId = link.getAttribute('href').substring(1);
                
                // Update active nav
                navLinks.forEach(l => l.classList.remove('active'));
                link.classList.add('active');
                
                // Show target section
                sections.forEach(section => {
                    section.style.display = section.id === targetId ? 'block' : 'none';
                });
                
                // Smooth scroll to top
                window.scrollTo({ top: 0, behavior: 'smooth' });
            });
        });
    }

    // Utility functions
    showSuccess(message) {
        this.showNotification(message, 'success');
    }

    showError(message) {
        this.showNotification(message, 'error');
    }

    showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.textContent = message;
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 15px 20px;
            border-radius: 8px;
            color: white;
            font-weight: 500;
            z-index: 1000;
            animation: slideInRight 0.3s ease;
            background: ${type === 'success' ? 'var(--success-color)' : 
                       type === 'error' ? 'var(--danger-color)' : 'var(--primary-color)'};
        `;
        
        document.body.appendChild(notification);
        
        setTimeout(() => {
            notification.style.animation = 'slideOutRight 0.3s ease';
            setTimeout(() => {
                document.body.removeChild(notification);
            }, 300);
        }, 3000);
    }

    // Load games data
    async loadGamesData() {
        try {
            const response = await fetch('/api/games');
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to load games data:', error);
            return null;
        }
    }

    // Format date for display
    formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'long',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    }

    // Check if bot is online
    async checkBotStatus() {
        try {
            const response = await fetch('/api/status');
            return response.ok;
        } catch (error) {
            return false;
        }
    }
}

// Additional CSS animations
const additionalStyles = `
    @keyframes slideInRight {
        from { transform: translateX(100%); opacity: 0; }
        to { transform: translateX(0); opacity: 1; }
    }
    
    @keyframes slideOutRight {
        from { transform: translateX(0); opacity: 1; }
        to { transform: translateX(100%); opacity: 0; }
    }
    
    .faq-question {
        cursor: pointer;
        padding: 15px;
        background: var(--light-bg);
        border: 1px solid var(--border-color);
        border-radius: var(--border-radius);
        margin-bottom: 5px;
        transition: all 0.3s ease;
    }
    
    .faq-question:hover {
        background: var(--primary-color);
        color: white;
    }
    
    .faq-question.active {
        background: var(--primary-color);
        color: white;
    }
    
    .faq-answer {
        display: none;
        padding: 15px;
        background: #f8f9fa;
        border-radius: var(--border-radius);
        margin-bottom: 15px;
    }
    
    .command-details {
        display: none;
        margin-top: 15px;
        padding-top: 15px;
        border-top: 1px solid var(--border-color);
    }
    
    .notification {
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    }
`;

// Inject additional styles
const styleSheet = document.createElement('style');
styleSheet.textContent = additionalStyles;
document.head.appendChild(styleSheet);

// Initialize the documentation app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.botDocs = new BotDocumentation();
});

// Export for use in other scripts
window.BotDocumentation = BotDocumentation;