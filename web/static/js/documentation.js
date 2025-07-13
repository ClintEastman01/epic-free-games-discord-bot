// Documentation-specific JavaScript functionality

class Documentation {
    constructor() {
        this.currentSection = 'overview';
        this.init();
    }

    init() {
        this.setupNavigation();
        this.setupSmoothScrolling();
        this.loadBotStats();
        this.setupAutoRefresh();
        this.handleInitialHash();
    }

    setupNavigation() {
        const navLinks = document.querySelectorAll('.nav-link');
        const sections = document.querySelectorAll('.content-section');

        navLinks.forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const targetSection = link.getAttribute('data-section');
                this.showSection(targetSection);
                this.updateURL(targetSection);
            });
        });
    }

    showSection(sectionId) {
        // Hide all sections
        const sections = document.querySelectorAll('.content-section');
        sections.forEach(section => {
            section.classList.remove('active');
        });

        // Show target section
        const targetSection = document.getElementById(sectionId);
        if (targetSection) {
            targetSection.classList.add('active');
        }

        // Update navigation
        const navLinks = document.querySelectorAll('.nav-link');
        navLinks.forEach(link => {
            link.classList.remove('active');
            if (link.getAttribute('data-section') === sectionId) {
                link.classList.add('active');
            }
        });

        this.currentSection = sectionId;
        this.scrollToTop();
    }

    updateURL(sectionId) {
        history.pushState({ section: sectionId }, '', `#${sectionId}`);
    }

    handleInitialHash() {
        const hash = window.location.hash.substring(1);
        if (hash && document.getElementById(hash)) {
            this.showSection(hash);
        }
    }

    setupSmoothScrolling() {
        // Handle browser back/forward
        window.addEventListener('popstate', (e) => {
            if (e.state && e.state.section) {
                this.showSection(e.state.section);
            } else {
                const hash = window.location.hash.substring(1);
                if (hash) {
                    this.showSection(hash);
                } else {
                    this.showSection('overview');
                }
            }
        });

        // Handle anchor links within content
        document.addEventListener('click', (e) => {
            if (e.target.tagName === 'A' && e.target.getAttribute('href').startsWith('#')) {
                e.preventDefault();
                const targetId = e.target.getAttribute('href').substring(1);
                if (document.getElementById(targetId)) {
                    this.showSection(targetId);
                    this.updateURL(targetId);
                }
            }
        });
    }

    scrollToTop() {
        window.scrollTo({
            top: 0,
            behavior: 'smooth'
        });
    }

    async loadBotStats() {
        try {
            const response = await fetch('/api/status');
            if (response.ok) {
                const data = await response.json();
                this.updateStats(data);
            } else {
                console.warn('Failed to load bot stats');
                this.updateStats({
                    server_count: '-',
                    game_count: '-',
                    status: 'Unknown'
                });
            }
        } catch (error) {
            console.error('Error loading bot stats:', error);
            this.updateStats({
                server_count: '-',
                game_count: '-',
                status: 'Offline'
            });
        }
    }

    updateStats(data) {
        // Update header stats
        const serverCountEl = document.getElementById('server-count');
        const gameCountEl = document.getElementById('game-count');
        const statusEl = document.getElementById('status');

        if (serverCountEl) serverCountEl.textContent = data.server_count || '-';
        if (gameCountEl) gameCountEl.textContent = data.game_count || '-';
        if (statusEl) statusEl.textContent = data.status || 'Unknown';

        // Update footer stats
        const footerStatusEl = document.getElementById('footer-status');
        const lastUpdatedEl = document.getElementById('last-updated');

        if (footerStatusEl) {
            footerStatusEl.textContent = data.status || 'Unknown';
            footerStatusEl.className = data.status === 'online' ? 'status-online' : 'status-offline';
        }

        if (lastUpdatedEl && data.last_update) {
            const date = new Date(data.last_update);
            lastUpdatedEl.textContent = date.toLocaleString();
        }
    }

    setupAutoRefresh() {
        // Refresh stats every 5 minutes
        setInterval(() => {
            this.loadBotStats();
        }, 5 * 60 * 1000);
    }

    // Utility method to copy code to clipboard
    copyToClipboard(text) {
        if (navigator.clipboard) {
            navigator.clipboard.writeText(text).then(() => {
                this.showNotification('Copied to clipboard!');
            }).catch(err => {
                console.error('Failed to copy: ', err);
            });
        } else {
            // Fallback for older browsers
            const textArea = document.createElement('textarea');
            textArea.value = text;
            document.body.appendChild(textArea);
            textArea.select();
            try {
                document.execCommand('copy');
                this.showNotification('Copied to clipboard!');
            } catch (err) {
                console.error('Failed to copy: ', err);
            }
            document.body.removeChild(textArea);
        }
    }

    showNotification(message) {
        // Create notification element
        const notification = document.createElement('div');
        notification.className = 'notification';
        notification.textContent = message;
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: var(--success-color);
            color: white;
            padding: 15px 20px;
            border-radius: 8px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
            z-index: 1000;
            animation: slideIn 0.3s ease-out;
        `;

        document.body.appendChild(notification);

        // Remove after 3 seconds
        setTimeout(() => {
            notification.style.animation = 'slideOut 0.3s ease-in';
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.parentNode.removeChild(notification);
                }
            }, 300);
        }, 3000);
    }

    // Method to highlight search results (for future search functionality)
    highlightText(element, searchTerm) {
        if (!searchTerm) return;

        const walker = document.createTreeWalker(
            element,
            NodeFilter.SHOW_TEXT,
            null,
            false
        );

        const textNodes = [];
        let node;
        while (node = walker.nextNode()) {
            textNodes.push(node);
        }

        textNodes.forEach(textNode => {
            const parent = textNode.parentNode;
            if (parent.tagName === 'SCRIPT' || parent.tagName === 'STYLE') return;

            const text = textNode.textContent;
            const regex = new RegExp(`(${searchTerm})`, 'gi');
            
            if (regex.test(text)) {
                const highlightedHTML = text.replace(regex, '<mark>$1</mark>');
                const wrapper = document.createElement('span');
                wrapper.innerHTML = highlightedHTML;
                parent.replaceChild(wrapper, textNode);
            }
        });
    }

    // Method to clear highlights
    clearHighlights() {
        const highlights = document.querySelectorAll('mark');
        highlights.forEach(mark => {
            const parent = mark.parentNode;
            parent.replaceChild(document.createTextNode(mark.textContent), mark);
            parent.normalize();
        });
    }
}

// Enhanced code block functionality
class CodeBlockEnhancer {
    constructor() {
        this.init();
    }

    init() {
        this.addCopyButtons();
        this.addSyntaxHighlighting();
    }

    addCopyButtons() {
        const codeBlocks = document.querySelectorAll('pre code, .command-example code, .api-response pre');
        
        codeBlocks.forEach(block => {
            const container = block.closest('pre') || block.closest('.command-example') || block.closest('.api-response');
            if (!container) return;

            // Don't add button if it already exists
            if (container.querySelector('.copy-button')) return;

            const copyButton = document.createElement('button');
            copyButton.className = 'copy-button';
            copyButton.innerHTML = '📋 Copy';
            copyButton.style.cssText = `
                position: absolute;
                top: 10px;
                right: 10px;
                background: rgba(255, 255, 255, 0.1);
                border: 1px solid rgba(255, 255, 255, 0.2);
                color: white;
                padding: 5px 10px;
                border-radius: 4px;
                cursor: pointer;
                font-size: 0.8rem;
                transition: background 0.3s ease;
            `;

            copyButton.addEventListener('click', () => {
                const text = block.textContent;
                if (window.documentation) {
                    window.documentation.copyToClipboard(text);
                }
            });

            copyButton.addEventListener('mouseenter', () => {
                copyButton.style.background = 'rgba(255, 255, 255, 0.2)';
            });

            copyButton.addEventListener('mouseleave', () => {
                copyButton.style.background = 'rgba(255, 255, 255, 0.1)';
            });

            // Make container relative for absolute positioning
            container.style.position = 'relative';
            container.appendChild(copyButton);
        });
    }

    addSyntaxHighlighting() {
        // Basic syntax highlighting for JSON
        const jsonBlocks = document.querySelectorAll('pre code');
        
        jsonBlocks.forEach(block => {
            let content = block.innerHTML;
            
            // Highlight JSON syntax
            content = content.replace(/"([^"]+)":/g, '<span style="color: #e06c75;">"$1"</span>:');
            content = content.replace(/:\s*"([^"]+)"/g, ': <span style="color: #98c379;">"$1"</span>');
            content = content.replace(/:\s*(\d+)/g, ': <span style="color: #d19a66;">$1</span>');
            content = content.replace(/:\s*(true|false|null)/g, ': <span style="color: #56b6c2;">$1</span>');
            
            block.innerHTML = content;
        });
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.documentation = new Documentation();
    window.codeBlockEnhancer = new CodeBlockEnhancer();
});

// Add CSS for animations
const style = document.createElement('style');
style.textContent = `
    @keyframes slideIn {
        from {
            transform: translateX(100%);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }

    @keyframes slideOut {
        from {
            transform: translateX(0);
            opacity: 1;
        }
        to {
            transform: translateX(100%);
            opacity: 0;
        }
    }

    .copy-button:hover {
        background: rgba(255, 255, 255, 0.2) !important;
    }

    .status-offline {
        color: #f04747 !important;
    }
`;
document.head.appendChild(style);