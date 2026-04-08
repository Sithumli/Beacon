package web

// Obsidian Dark Theme CSS
const obsidianCSS = `
:root {
    --bg-primary: #0D0D0D;
    --bg-secondary: #141414;
    --bg-tertiary: #1A1A1A;
    --bg-card: #1E1E1E;
    --bg-card-hover: #252525;
    --border-color: #2A2A2A;
    --text-primary: #FFFFFF;
    --text-secondary: #A0A0A0;
    --text-muted: #666666;
    --accent-green: #22C55E;
    --accent-red: #EF4444;
    --accent-blue: #3B82F6;
    --accent-yellow: #F59E0B;
    --accent-purple: #8B5CF6;
}

* {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
}

body {
    font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    background: var(--bg-primary);
    color: var(--text-primary);
    min-height: 100vh;
}

.app-container {
    display: flex;
    min-height: 100vh;
}

/* Sidebar */
.sidebar {
    width: 200px;
    background: var(--bg-primary);
    border-right: 1px solid var(--border-color);
    padding: 20px 0;
    position: fixed;
    height: 100vh;
    display: flex;
    flex-direction: column;
}

.sidebar-brand {
    padding: 0 20px 30px;
}

.sidebar-brand h1 {
    font-size: 16px;
    font-weight: 600;
    color: var(--text-primary);
    letter-spacing: -0.5px;
}

.sidebar-brand .version {
    font-size: 11px;
    color: var(--text-muted);
    margin-top: 2px;
}

.sidebar-nav {
    flex: 1;
}

.nav-item {
    display: flex;
    align-items: center;
    padding: 12px 20px;
    color: var(--text-secondary);
    text-decoration: none;
    font-size: 14px;
    transition: all 0.2s;
    border-left: 3px solid transparent;
}

.nav-item:hover {
    background: var(--bg-secondary);
    color: var(--text-primary);
}

.nav-item.active {
    background: var(--bg-tertiary);
    color: var(--text-primary);
    border-left-color: var(--text-primary);
}

.nav-item svg {
    width: 18px;
    height: 18px;
    margin-right: 12px;
    opacity: 0.7;
}

.nav-item.active svg {
    opacity: 1;
}

.sidebar-footer {
    padding: 20px;
    border-top: 1px solid var(--border-color);
}

.node-status {
    font-size: 11px;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 8px;
}

.node-name {
    display: flex;
    align-items: center;
    font-size: 13px;
    color: var(--text-secondary);
}

.node-name::before {
    content: '';
    width: 8px;
    height: 8px;
    background: var(--accent-green);
    border-radius: 50%;
    margin-right: 8px;
}

/* Main Content */
.main-content {
    flex: 1;
    margin-left: 200px;
    min-height: 100vh;
}

/* Header */
.header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 32px;
    border-bottom: 1px solid var(--border-color);
    background: var(--bg-primary);
    position: sticky;
    top: 0;
    z-index: 100;
}

.header-left {
    display: flex;
    align-items: center;
    gap: 24px;
}

.header-brand {
    font-size: 18px;
    font-weight: 600;
    color: var(--text-primary);
}

.header-tabs {
    display: flex;
    gap: 8px;
}

.header-tab {
    padding: 8px 16px;
    color: var(--text-muted);
    text-decoration: none;
    font-size: 14px;
    border-radius: 6px;
    transition: all 0.2s;
}

.header-tab:hover {
    color: var(--text-secondary);
}

.header-tab.active {
    color: var(--text-primary);
    background: var(--bg-tertiary);
}

.search-box {
    display: flex;
    align-items: center;
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    padding: 8px 16px;
    width: 280px;
}

.search-box input {
    background: transparent;
    border: none;
    color: var(--text-primary);
    font-size: 14px;
    width: 100%;
    outline: none;
}

.search-box input::placeholder {
    color: var(--text-muted);
}

.header-right {
    display: flex;
    align-items: center;
    gap: 16px;
}

.btn {
    padding: 10px 20px;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    border: none;
    text-decoration: none;
}

.btn-primary {
    background: var(--text-primary);
    color: var(--bg-primary);
}

.btn-primary:hover {
    background: #E0E0E0;
}

.btn-secondary {
    background: var(--bg-tertiary);
    color: var(--text-primary);
    border: 1px solid var(--border-color);
}

.btn-secondary:hover {
    background: var(--bg-card);
}

.btn-danger {
    background: var(--accent-red);
    color: white;
}

.notification-btn {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
}

/* Page Content */
.page-content {
    padding: 32px;
}

.page-header {
    margin-bottom: 8px;
}

.page-subtitle {
    font-size: 11px;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 1px;
    margin-bottom: 4px;
}

.page-title {
    font-size: 32px;
    font-weight: 600;
    font-style: italic;
    margin-bottom: 8px;
}

.page-description {
    color: var(--text-secondary);
    font-size: 14px;
    max-width: 600px;
}

/* Stats Cards */
.stats-grid {
    display: grid;
    grid-template-columns: repeat(5, 1fr);
    gap: 16px;
    margin: 32px 0;
}

.stat-card {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    padding: 20px;
}

.stat-card .icon {
    width: 24px;
    height: 24px;
    margin-bottom: 16px;
    opacity: 0.6;
}

.stat-card .label {
    font-size: 11px;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 8px;
}

.stat-card .value {
    font-size: 28px;
    font-weight: 600;
    color: var(--text-primary);
}

/* Charts Section */
.charts-section {
    display: grid;
    grid-template-columns: 2fr 1fr;
    gap: 24px;
    margin-bottom: 32px;
}

.chart-card {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    padding: 24px;
}

.chart-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 24px;
}

.chart-title {
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 4px;
}

.chart-subtitle {
    font-size: 13px;
    color: var(--text-muted);
}

.chart-controls {
    display: flex;
    gap: 8px;
}

.chart-btn {
    padding: 6px 12px;
    border-radius: 6px;
    font-size: 12px;
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    color: var(--text-secondary);
    cursor: pointer;
}

.chart-btn.active {
    background: var(--bg-primary);
    color: var(--text-primary);
}

.chart-placeholder {
    height: 200px;
    display: flex;
    align-items: flex-end;
    padding: 20px 0;
}

.chart-line {
    width: 100%;
    height: 60%;
    background: linear-gradient(180deg, transparent 0%, rgba(255,255,255,0.05) 100%);
    border-top: 2px solid var(--text-secondary);
    position: relative;
}

/* Quick Task Panel */
.quick-task {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    padding: 24px;
}

.quick-task-title {
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 4px;
}

.quick-task-subtitle {
    font-size: 13px;
    color: var(--text-muted);
    margin-bottom: 24px;
}

.form-group {
    margin-bottom: 20px;
}

.form-label {
    font-size: 11px;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 8px;
    display: block;
}

.form-select, .form-input {
    width: 100%;
    padding: 12px 16px;
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    color: var(--text-primary);
    font-size: 14px;
    outline: none;
}

.form-select:focus, .form-input:focus {
    border-color: var(--text-muted);
}

.btn-block {
    width: 100%;
    padding: 14px;
    font-size: 14px;
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

/* Tables */
.table-section {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    overflow: hidden;
}

.table-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 24px;
    border-bottom: 1px solid var(--border-color);
}

.table-title {
    font-size: 18px;
    font-weight: 600;
}

.table-action {
    font-size: 12px;
    color: var(--text-secondary);
    text-decoration: none;
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.table-action:hover {
    color: var(--text-primary);
}

table {
    width: 100%;
    border-collapse: collapse;
}

th {
    text-align: left;
    padding: 16px 24px;
    font-size: 11px;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    font-weight: 500;
    border-bottom: 1px solid var(--border-color);
}

td {
    padding: 16px 24px;
    font-size: 14px;
    border-bottom: 1px solid var(--border-color);
}

tr:last-child td {
    border-bottom: none;
}

tr:hover {
    background: var(--bg-card-hover);
}

/* Agent Row */
.agent-row {
    background: var(--bg-tertiary);
    border-radius: 8px;
    margin: 8px 16px;
}

.agent-icon {
    width: 36px;
    height: 36px;
    background: var(--bg-card);
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-right: 12px;
}

.agent-info {
    display: flex;
    align-items: center;
}

.agent-name {
    font-weight: 500;
}

.agent-id {
    font-size: 12px;
    color: var(--text-muted);
}

/* Status Badges */
.status {
    display: inline-flex;
    align-items: center;
    padding: 6px 12px;
    border-radius: 20px;
    font-size: 12px;
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.3px;
}

.status::before {
    content: '';
    width: 6px;
    height: 6px;
    border-radius: 50%;
    margin-right: 8px;
}

.status-active, .status-healthy, .status-completed {
    background: rgba(34, 197, 94, 0.1);
    color: var(--accent-green);
}

.status-active::before, .status-healthy::before, .status-completed::before {
    background: var(--accent-green);
}

.status-polling, .status-running, .status-pending {
    background: rgba(59, 130, 246, 0.1);
    color: var(--accent-blue);
}

.status-polling::before, .status-running::before, .status-pending::before {
    background: var(--accent-blue);
}

.status-idle, .status-inactive {
    background: rgba(160, 160, 160, 0.1);
    color: var(--text-muted);
}

.status-idle::before, .status-inactive::before {
    background: var(--text-muted);
}

.status-failed, .status-error {
    background: rgba(239, 68, 68, 0.1);
    color: var(--accent-red);
}

.status-failed::before, .status-error::before {
    background: var(--accent-red);
}

/* Capability Tags */
.capability-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
}

.capability-tag {
    padding: 4px 12px;
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    font-size: 11px;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.3px;
}

/* Progress Bar */
.progress-bar {
    width: 120px;
    height: 4px;
    background: var(--bg-tertiary);
    border-radius: 2px;
    overflow: hidden;
}

.progress-fill {
    height: 100%;
    background: var(--text-secondary);
    border-radius: 2px;
}

/* Action Menu */
.action-btn {
    width: 32px;
    height: 32px;
    border-radius: 6px;
    background: transparent;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
}

.action-btn:hover {
    background: var(--bg-card);
    color: var(--text-primary);
}

/* Detail Panel */
.detail-panel {
    position: fixed;
    right: 0;
    top: 0;
    width: 360px;
    height: 100vh;
    background: var(--bg-secondary);
    border-left: 1px solid var(--border-color);
    padding: 24px;
    overflow-y: auto;
}

.detail-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
}

.detail-title {
    font-size: 11px;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.close-btn {
    width: 32px;
    height: 32px;
    border-radius: 6px;
    background: transparent;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    font-size: 18px;
}

.detail-visual {
    width: 100%;
    aspect-ratio: 1;
    background: var(--bg-tertiary);
    border-radius: 12px;
    margin-bottom: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.visual-placeholder {
    width: 120px;
    height: 120px;
    border: 2px solid var(--border-color);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-muted);
}

.detail-label {
    font-size: 11px;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 4px;
}

.detail-value {
    font-size: 18px;
    font-weight: 600;
    margin-bottom: 24px;
}

.detail-section {
    margin-bottom: 24px;
}

.detail-section-title {
    font-size: 11px;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 16px;
}

.health-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px;
    background: var(--bg-tertiary);
    border-radius: 8px;
    margin-bottom: 8px;
}

.health-label {
    font-size: 14px;
    color: var(--text-secondary);
}

.health-value {
    font-size: 14px;
    font-weight: 600;
}

.property-item {
    display: flex;
    justify-content: space-between;
    padding: 12px 0;
    border-bottom: 1px solid var(--border-color);
}

.property-label {
    font-size: 11px;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.property-value {
    font-size: 13px;
    color: var(--text-primary);
}

/* Discovery Cards */
.discovery-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 24px;
    margin-top: 32px;
}

.discovery-card {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    padding: 24px;
    transition: all 0.2s;
}

.discovery-card:hover {
    border-color: var(--text-muted);
}

.discovery-card-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 16px;
}

.discovery-icon {
    width: 48px;
    height: 48px;
    background: var(--bg-tertiary);
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.match-rate {
    text-align: right;
}

.match-label {
    font-size: 10px;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.match-value {
    font-size: 20px;
    font-weight: 600;
    color: var(--accent-green);
}

.discovery-status {
    font-size: 10px;
    color: var(--accent-green);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 8px;
}

.discovery-status::before {
    content: '';
    display: inline-block;
    width: 6px;
    height: 6px;
    background: var(--accent-green);
    border-radius: 50%;
    margin-right: 6px;
}

.discovery-name {
    font-size: 20px;
    font-weight: 600;
    margin-bottom: 12px;
}

.discovery-description {
    font-size: 14px;
    color: var(--text-secondary);
    line-height: 1.5;
    margin-bottom: 24px;
}

.discovery-actions {
    display: flex;
    gap: 8px;
}

.discovery-actions .btn {
    flex: 1;
}

/* Filter Tabs */
.filter-tabs {
    display: flex;
    gap: 12px;
    margin-top: 24px;
}

.filter-tab {
    padding: 10px 20px;
    border-radius: 20px;
    font-size: 13px;
    background: transparent;
    border: 1px solid var(--border-color);
    color: var(--text-secondary);
    cursor: pointer;
    transition: all 0.2s;
    text-decoration: none;
}

.filter-tab:hover {
    border-color: var(--text-muted);
    color: var(--text-primary);
}

.filter-tab.active {
    background: var(--text-primary);
    border-color: var(--text-primary);
    color: var(--bg-primary);
}

/* Task Timeline */
.timeline {
    padding: 16px 0;
}

.timeline-item {
    display: flex;
    padding: 12px 0;
    position: relative;
}

.timeline-item::before {
    content: '';
    position: absolute;
    left: 5px;
    top: 28px;
    bottom: -12px;
    width: 1px;
    background: var(--border-color);
}

.timeline-item:last-child::before {
    display: none;
}

.timeline-dot {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    background: var(--text-muted);
    margin-right: 16px;
    flex-shrink: 0;
    margin-top: 4px;
}

.timeline-dot.success {
    background: var(--accent-green);
}

.timeline-dot.error {
    background: var(--accent-red);
}

.timeline-content {
    flex: 1;
}

.timeline-title {
    font-size: 14px;
    font-weight: 500;
    margin-bottom: 4px;
}

.timeline-subtitle {
    font-size: 12px;
    color: var(--text-muted);
}

/* JSON Display */
.json-display {
    background: var(--bg-tertiary);
    border-radius: 8px;
    padding: 16px;
    font-family: 'Monaco', 'Menlo', monospace;
    font-size: 12px;
    overflow-x: auto;
}

.json-key {
    color: var(--text-secondary);
}

.json-string {
    color: var(--accent-green);
}

.json-string.error {
    color: var(--accent-red);
}

.json-number {
    color: var(--accent-blue);
}

/* Footer */
.footer {
    padding: 24px 32px;
    border-top: 1px solid var(--border-color);
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 12px;
    color: var(--text-muted);
}

.footer-links {
    display: flex;
    gap: 32px;
}

.footer-link {
    color: var(--text-muted);
    text-decoration: none;
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.footer-link:hover {
    color: var(--text-secondary);
}

/* Status Bar */
.status-bar {
    display: flex;
    align-items: center;
    gap: 24px;
    padding: 12px 32px;
    background: var(--bg-secondary);
    border-top: 1px solid var(--border-color);
    font-size: 12px;
}

.status-item {
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--text-muted);
}

.status-item::before {
    content: '';
    width: 6px;
    height: 6px;
    background: var(--accent-green);
    border-radius: 50%;
}

/* Empty State */
.empty-state {
    text-align: center;
    padding: 60px 20px;
    color: var(--text-muted);
}

.empty-state-icon {
    font-size: 48px;
    margin-bottom: 16px;
    opacity: 0.3;
}

.empty-state-title {
    font-size: 18px;
    font-weight: 600;
    margin-bottom: 8px;
    color: var(--text-secondary);
}

.empty-state-description {
    font-size: 14px;
    max-width: 400px;
    margin: 0 auto;
}

/* Responsive */
@media (max-width: 1200px) {
    .stats-grid {
        grid-template-columns: repeat(3, 1fr);
    }

    .discovery-grid {
        grid-template-columns: repeat(2, 1fr);
    }
}

@media (max-width: 768px) {
    .sidebar {
        display: none;
    }

    .main-content {
        margin-left: 0;
    }

    .stats-grid {
        grid-template-columns: 1fr;
    }

    .charts-section {
        grid-template-columns: 1fr;
    }

    .discovery-grid {
        grid-template-columns: 1fr;
    }
}
`

const appJS = `
// Real-time updates via SSE
const eventSource = new EventSource('/events');

eventSource.onmessage = function(event) {
    const data = JSON.parse(event.data);
    // Update stats if elements exist
    const agentCount = document.getElementById('agent-count');
    if (agentCount) agentCount.textContent = data.agents;

    const taskCount = document.getElementById('task-count');
    if (taskCount) taskCount.textContent = data.tasks;
};

// Detail panel toggle
function showDetail(agentId) {
    const panel = document.getElementById('detail-panel');
    if (panel) panel.classList.add('active');
}

function hideDetail() {
    const panel = document.getElementById('detail-panel');
    if (panel) panel.classList.remove('active');
}

// Quick task form
document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('quick-task-form');
    if (form) {
        form.addEventListener('submit', async function(e) {
            e.preventDefault();
            const formData = new FormData(form);
            const data = {
                capability: formData.get('task_type'),
                payload: formData.get('parameters')
            };

            try {
                const response = await fetch('/api/v1/route', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
                if (response.ok) {
                    alert('Task initialized successfully!');
                    form.reset();
                }
            } catch (err) {
                console.error('Failed to send task:', err);
            }
        });
    }
});
`
