package web

// Obsidian Dark Theme CSS
const obsidianCSS = `
:root {
    --bg-primary: #0D0D0D;
    --bg-secondary: #0D0D0D;
    --bg-tertiary: #171717;
    --bg-card: #141414;
    --bg-card-hover: #1A1A1A;
    --bg-row: #1A1A1A;
    --bg-input: #1E1E1E;
    --border-color: #2A2A2A;
    --text-primary: #FFFFFF;
    --text-secondary: #888888;
    --text-muted: #505050;
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
    border-right: none;
    padding: 28px 0;
    position: fixed;
    height: 100vh;
    display: flex;
    flex-direction: column;
}

.sidebar-brand {
    padding: 0 24px 48px;
}

.sidebar-brand h1 {
    font-size: 16px;
    font-weight: 700;
    color: var(--text-primary);
    letter-spacing: 0;
}

.sidebar-brand .version {
    font-size: 11px;
    color: var(--text-muted);
    margin-top: 4px;
}

.sidebar-nav {
    flex: 1;
    padding: 0 12px;
}

.nav-item {
    display: flex;
    align-items: center;
    padding: 12px 16px;
    color: var(--text-muted);
    text-decoration: none;
    font-size: 12px;
    font-weight: 500;
    letter-spacing: 0.5px;
    transition: all 0.15s ease;
    border-radius: 10px;
    margin-bottom: 2px;
}

.nav-item:hover {
    color: var(--text-secondary);
}

.nav-item.active {
    background: var(--bg-tertiary);
    color: var(--text-primary);
}

.nav-item svg {
    width: 18px;
    height: 18px;
    margin-right: 12px;
    opacity: 0.5;
}

.nav-item:hover svg {
    opacity: 0.7;
}

.nav-item.active svg {
    opacity: 0.9;
}

.sidebar-footer {
    padding: 18px 20px;
    margin: 12px;
    background: var(--bg-tertiary);
    border-radius: 24px;
}

.node-status {
    font-size: 9px;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 1px;
    margin-bottom: 8px;
}

.node-name {
    display: flex;
    align-items: center;
    font-size: 12px;
    color: var(--text-secondary);
}

.node-name::before {
    content: '';
    width: 6px;
    height: 6px;
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
    padding: 16px 40px;
    border-bottom: none;
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
    font-size: 16px;
    font-weight: 500;
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
    font-weight: 400;
    transition: all 0.2s;
}

.header-tab:hover {
    color: var(--text-secondary);
}

.header-tab.active {
    color: var(--text-primary);
}

.search-box {
    display: flex;
    align-items: center;
    background: var(--bg-tertiary);
    border: none;
    border-radius: 25px;
    padding: 12px 20px;
    min-width: 220px;
}

.search-box svg {
    opacity: 0.4;
    margin-right: 10px;
    flex-shrink: 0;
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
    gap: 12px;
}

.btn {
    padding: 12px 24px;
    border-radius: 50px;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s ease;
    border: none;
    text-decoration: none;
}

.btn-primary {
    background: #F5F5F5;
    color: #0D0D0D;
}

.btn-primary:hover {
    background: #FFFFFF;
}

.btn-secondary {
    background: var(--bg-tertiary);
    color: var(--text-primary);
    border: 1px solid var(--border-color);
}

.btn-secondary:hover {
    background: var(--bg-card-hover);
}

.btn-danger {
    background: var(--accent-red);
    color: white;
}

.notification-btn {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background: transparent;
    border: 1px solid var(--border-color);
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    color: var(--text-muted);
}

.notification-btn:hover {
    border-color: var(--text-muted);
    color: var(--text-secondary);
}

/* Page Content */
.page-content {
    padding: 32px 40px;
}

.page-header {
    margin-bottom: 0;
}

.page-subtitle {
    font-size: 10px;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 3px;
    margin-bottom: 8px;
}

.page-title {
    font-size: 38px;
    font-weight: 500;
    font-style: italic;
    margin-bottom: 8px;
    letter-spacing: -1px;
}

.page-description {
    color: var(--text-secondary);
    font-size: 15px;
    max-width: 600px;
    line-height: 1.6;
}

/* Stats Cards */
.stats-grid {
    display: grid;
    grid-template-columns: repeat(5, 1fr);
    gap: 10px;
    margin: 24px 0;
}

.stat-card {
    background: #141414;
    border: 1px solid #1E1E1E;
    border-radius: 48px;
    padding: 24px 28px;
    min-height: 130px;
    display: flex;
    flex-direction: column;
}

.stat-card .icon {
    width: 22px;
    height: 22px;
    opacity: 0.45;
    flex-shrink: 0;
}

.stat-card .label {
    font-size: 9px;
    color: #606060;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-top: auto;
    margin-bottom: 8px;
}

.stat-card .value {
    font-size: 28px;
    font-weight: 500;
    color: var(--text-primary);
    letter-spacing: -0.5px;
    line-height: 1;
}

/* Charts Section */
.charts-section {
    display: grid;
    grid-template-columns: 2fr 1fr;
    gap: 16px;
    margin-bottom: 28px;
}

.chart-card {
    background: #111111;
    border: 1px solid #1A1A1A;
    border-radius: 48px;
    padding: 28px 32px;
}

.chart-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 20px;
}

.chart-title {
    font-size: 14px;
    font-weight: 500;
    margin-bottom: 4px;
}

.chart-subtitle {
    font-size: 11px;
    color: #505050;
}

.chart-controls {
    display: flex;
    gap: 6px;
}

.chart-btn {
    padding: 8px 16px;
    border-radius: 50px;
    font-size: 10px;
    font-weight: 500;
    background: transparent;
    border: 1px solid #2A2A2A;
    color: #606060;
    cursor: pointer;
    transition: all 0.15s ease;
}

.chart-btn:hover {
    border-color: #3A3A3A;
    color: #888888;
}

.chart-btn.active {
    background: #1A1A1A;
    border-color: #2A2A2A;
    color: var(--text-primary);
}

.chart-placeholder {
    height: 180px;
    display: flex;
    align-items: flex-end;
    padding: 16px 0;
}

.chart-line {
    width: 100%;
    height: 60%;
    background: linear-gradient(180deg, transparent 0%, rgba(255,255,255,0.02) 100%);
    border-top: 1px solid #333;
    position: relative;
}

/* Quick Task Panel */
.quick-task {
    background: #141414;
    border: 1px solid #1E1E1E;
    border-radius: 48px;
    padding: 28px 32px;
}

.quick-task-title {
    font-size: 15px;
    font-weight: 600;
    margin-bottom: 4px;
}

.quick-task-subtitle {
    font-size: 12px;
    color: #606060;
    margin-bottom: 24px;
}

.form-group {
    margin-bottom: 18px;
}

.form-label {
    font-size: 9px;
    color: #505050;
    text-transform: uppercase;
    letter-spacing: 1px;
    margin-bottom: 10px;
    display: block;
}

.form-select, .form-input {
    width: 100%;
    padding: 14px 20px;
    background: #1E1E1E;
    border: none;
    border-radius: 16px;
    color: var(--text-primary);
    font-size: 13px;
    outline: none;
}

.form-select {
    appearance: none;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%23666' stroke-width='2'%3E%3Cpath d='M6 9l6 6 6-6'/%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right 20px center;
    padding-right: 44px;
}

.form-select:focus, .form-input:focus {
    background: #222222;
}

.form-input::placeholder {
    color: #505050;
}

.btn-block {
    width: 100%;
    padding: 16px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 1.5px;
    border-radius: 50px;
    margin-top: 12px;
    background: #1A1A1A;
    border: 1px solid #2A2A2A;
    color: var(--text-primary);
}

.btn-block:hover {
    background: #222222;
    border-color: #333333;
}

/* Tables */
.table-section {
    background: #111111;
    border: 1px solid #1A1A1A;
    border-radius: 48px;
    overflow: hidden;
}

.table-section table {
    background: transparent;
}

.table-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 28px 36px 16px;
    border-bottom: none;
}

.table-title {
    font-size: 16px;
    font-weight: 600;
}

.table-action {
    font-size: 10px;
    color: var(--text-muted);
    text-decoration: none;
    text-transform: uppercase;
    letter-spacing: 1.5px;
    font-weight: 500;
}

.table-action:hover {
    color: var(--text-primary);
}

table {
    width: 100%;
    border-collapse: separate;
    border-spacing: 0 12px;
    padding: 0 28px 32px;
}

th {
    text-align: left;
    padding: 12px 24px;
    font-size: 10px;
    color: #6B6B6B;
    text-transform: uppercase;
    letter-spacing: 1.5px;
    font-weight: 500;
    border-bottom: none;
}

td {
    padding: 20px 24px;
    font-size: 14px;
    font-weight: 500;
    color: #ACABAA;
    border: none;
    background-color: #1F2020;
}

tbody tr td:first-child {
    border-top-left-radius: 9999px !important;
    border-bottom-left-radius: 9999px !important;
    padding-left: 28px;
}

tbody tr td:last-child {
    border-top-right-radius: 9999px !important;
    border-bottom-right-radius: 9999px !important;
    padding-right: 28px;
}

tbody tr td[colspan] {
    border-radius: 9999px !important;
    padding-left: 28px;
    padding-right: 28px;
}

tbody tr td:only-child {
    border-radius: 9999px !important;
    padding-left: 28px;
    padding-right: 28px;
}

tbody tr {
    transition: all 0.15s ease;
}

tbody tr:hover td {
    background-color: #252626;
}

/* Agent Row */
.agent-row {
    background: #1A1A1A;
    border-radius: 10px;
    margin: 8px 16px;
}

.agent-icon {
    width: 32px;
    height: 32px;
    background: #252626;
    border-radius: 9999px;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-right: 16px;
    flex-shrink: 0;
}

.agent-icon svg {
    opacity: 0.7;
    width: 14px;
    height: 14px;
    color: #E7E5E4;
}

.agent-info {
    display: flex;
    align-items: center;
}

.agent-name {
    font-weight: 700;
    font-size: 14px;
    color: #E7E5E4;
}

.agent-id {
    font-size: 11px;
    color: #505050;
    font-family: 'SF Mono', Monaco, 'Courier New', monospace;
    letter-spacing: 0.3px;
}

/* Status Badges */
.status {
    display: inline-flex;
    align-items: center;
    font-size: 12px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.6px;
    background: transparent;
    padding: 0;
}

.status::before {
    content: '';
    width: 8px;
    height: 8px;
    border-radius: 9999px;
    margin-right: 8px;
    flex-shrink: 0;
}

.status-active, .status-healthy, .status-completed {
    color: #E7E5E4;
}

.status-active::before, .status-healthy::before, .status-completed::before {
    background: #FCF9F8;
}

.status-polling, .status-running, .status-pending, .status-processing {
    color: #E7E5E4;
}

.status-polling::before, .status-running::before, .status-pending::before, .status-processing::before {
    background: #FCF9F8;
}

.status-idle, .status-inactive {
    color: #ACABAA;
}

.status-idle::before, .status-inactive::before {
    background: #606060;
}

.status-failed, .status-error {
    color: #F87171;
}

.status-failed::before, .status-error::before {
    background: #F87171;
}

/* Capability Tags */
.capability-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
}

.capability-tag {
    padding: 5px 10px;
    background: transparent;
    border: 1px solid #2A2A2A;
    border-radius: 4px;
    font-size: 9px;
    color: #888888;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    font-weight: 500;
}

/* Progress Bar */
.progress-bar {
    width: 96px;
    height: 6px;
    background: #252626;
    border-radius: 9999px;
    overflow: hidden;
}

.progress-fill {
    height: 100%;
    background: #C6C6C7;
    border-radius: 9999px;
}

/* Action Menu */
.action-btn {
    width: 32px;
    height: 32px;
    border-radius: 6px;
    background: transparent;
    border: none;
    color: #505050;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 16px;
    letter-spacing: 1px;
}

.action-btn:hover {
    color: var(--text-primary);
}

/* Detail Panel */
.detail-panel {
    position: fixed;
    right: 0;
    top: 0;
    width: 360px;
    height: 100vh;
    background: #0D0D0D;
    border-left: 1px solid #1A1A1A;
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
    font-size: 9px;
    color: #505050;
    text-transform: uppercase;
    letter-spacing: 1.5px;
}

.close-btn {
    width: 28px;
    height: 28px;
    border-radius: 6px;
    background: transparent;
    border: none;
    color: #505050;
    cursor: pointer;
    font-size: 16px;
    text-decoration: none;
    display: flex;
    align-items: center;
    justify-content: center;
}

.close-btn:hover {
    color: var(--text-primary);
}

.detail-visual {
    width: 100%;
    aspect-ratio: 1;
    background: #141414;
    border-radius: 12px;
    margin-bottom: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.visual-placeholder {
    width: 100px;
    height: 100px;
    border: 1px solid #2A2A2A;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #505050;
}

.detail-label {
    font-size: 9px;
    color: #505050;
    text-transform: uppercase;
    letter-spacing: 1px;
    margin-bottom: 6px;
}

.detail-value {
    font-size: 16px;
    font-weight: 500;
    margin-bottom: 20px;
}

.detail-section {
    margin-bottom: 24px;
}

.detail-section-title {
    font-size: 9px;
    color: #505050;
    text-transform: uppercase;
    letter-spacing: 1.5px;
    margin-bottom: 14px;
}

.health-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 14px 16px;
    background: #141414;
    border-radius: 10px;
    margin-bottom: 8px;
}

.health-label {
    font-size: 12px;
    color: #888888;
}

.health-value {
    font-size: 13px;
    font-weight: 600;
}

.property-item {
    display: flex;
    justify-content: space-between;
    padding: 12px 0;
    border-bottom: 1px solid #1A1A1A;
}

.property-item:last-child {
    border-bottom: none;
}

.property-label {
    font-size: 9px;
    color: #505050;
    text-transform: uppercase;
    letter-spacing: 1px;
}

.property-value {
    font-size: 12px;
    color: var(--text-primary);
}

/* Discovery Cards */
.discovery-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 16px;
    margin-top: 28px;
}

.discovery-card {
    background: #111111;
    border: 1px solid #1A1A1A;
    border-radius: 48px;
    padding: 28px 32px;
    transition: all 0.15s ease;
}

.discovery-card:hover {
    border-color: #2A2A2A;
}

.discovery-card-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 16px;
}

.discovery-icon {
    width: 44px;
    height: 44px;
    background: #1A1A1A;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.match-rate {
    text-align: right;
}

.match-label {
    font-size: 8px;
    color: #505050;
    text-transform: uppercase;
    letter-spacing: 1px;
}

.match-value {
    font-size: 20px;
    font-weight: 500;
    color: var(--text-primary);
}

.discovery-status {
    font-size: 8px;
    color: #4ADE80;
    text-transform: uppercase;
    letter-spacing: 1px;
    margin-bottom: 8px;
}

.discovery-status::before {
    content: '';
    display: inline-block;
    width: 5px;
    height: 5px;
    background: #4ADE80;
    border-radius: 50%;
    margin-right: 6px;
}

.discovery-name {
    font-size: 20px;
    font-weight: 500;
    margin-bottom: 10px;
}

.discovery-description {
    font-size: 12px;
    color: #888888;
    line-height: 1.6;
    margin-bottom: 20px;
}

.discovery-actions {
    display: flex;
    gap: 8px;
}

.discovery-actions .btn {
    flex: 1;
    border-radius: 50px;
    padding: 14px;
    font-size: 11px;
}

/* Filter Tabs */
.filter-tabs {
    display: flex;
    gap: 8px;
    margin-top: 20px;
}

.filter-tab {
    padding: 10px 20px;
    border-radius: 50px;
    font-size: 10px;
    font-weight: 500;
    background: transparent;
    border: 1px solid #2A2A2A;
    color: #888888;
    cursor: pointer;
    transition: all 0.15s ease;
    text-decoration: none;
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.filter-tab:hover {
    border-color: #3A3A3A;
    color: var(--text-primary);
}

.filter-tab.active {
    background: var(--text-primary);
    border-color: var(--text-primary);
    color: #0D0D0D;
}

/* Task Timeline */
.timeline {
    padding: 12px 0;
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
    background: #1A1A1A;
}

.timeline-item:last-child::before {
    display: none;
}

.timeline-dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: #505050;
    margin-right: 16px;
    flex-shrink: 0;
    margin-top: 4px;
}

.timeline-dot.success {
    background: #4ADE80;
}

.timeline-dot.error {
    background: #F87171;
}

.timeline-content {
    flex: 1;
}

.timeline-title {
    font-size: 12px;
    font-weight: 500;
    margin-bottom: 3px;
}

.timeline-subtitle {
    font-size: 10px;
    color: #505050;
}

/* JSON Display */
.json-display {
    background: #141414;
    border-radius: 10px;
    padding: 16px;
    font-family: 'SF Mono', 'Monaco', 'Menlo', monospace;
    font-size: 11px;
    overflow-x: auto;
    line-height: 1.6;
}

.json-key {
    color: #888888;
}

.json-string {
    color: #4ADE80;
}

.json-string.error {
    color: #F87171;
}

.json-number {
    color: #60A5FA;
}

/* Footer */
.footer {
    padding: 24px 40px;
    border-top: none;
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 10px;
    color: #505050;
    margin-top: 40px;
}

.footer-links {
    display: flex;
    gap: 32px;
}

.footer-link {
    color: #505050;
    text-decoration: none;
    text-transform: uppercase;
    letter-spacing: 1.5px;
    font-size: 9px;
}

.footer-link:hover {
    color: #888888;
}

/* Status Bar */
.status-bar {
    display: flex;
    align-items: center;
    gap: 24px;
    padding: 14px 40px;
    background: #0D0D0D;
    border-top: none;
    font-size: 10px;
}

.status-item {
    display: flex;
    align-items: center;
    gap: 8px;
    color: #505050;
    letter-spacing: 0.5px;
}

.status-item::before {
    content: '';
    width: 5px;
    height: 5px;
    background: #4ADE80;
    border-radius: 50%;
}

/* Empty State */
.empty-state {
    text-align: center;
    padding: 48px 20px;
    color: #505050;
}

.empty-state-icon {
    font-size: 40px;
    margin-bottom: 14px;
    opacity: 0.3;
}

.empty-state-title {
    font-size: 14px;
    font-weight: 500;
    margin-bottom: 6px;
    color: #888888;
}

.empty-state-description {
    font-size: 12px;
    max-width: 360px;
    margin: 0 auto;
    color: #505050;
}

/* Responsive */
@media (max-width: 1400px) {
    .stats-grid {
        grid-template-columns: repeat(3, 1fr);
    }
}

@media (max-width: 1200px) {
    .stats-grid {
        grid-template-columns: repeat(2, 1fr);
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

    .page-content {
        padding: 24px;
    }

    .header {
        padding: 16px 24px;
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
