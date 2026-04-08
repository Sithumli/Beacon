package web

const sidebarHTML = `
<aside class="sidebar">
    <div class="sidebar-brand">
        <h1>Beacon</h1>
        <div class="version">v1.0</div>
    </div>
    <nav class="sidebar-nav">
        <a href="/" class="nav-item {{if eq .Page "dashboard"}}active{{end}}">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="3" width="7" height="7"></rect>
                <rect x="14" y="3" width="7" height="7"></rect>
                <rect x="14" y="14" width="7" height="7"></rect>
                <rect x="3" y="14" width="7" height="7"></rect>
            </svg>
            DASHBOARD
        </a>
        <a href="/agents" class="nav-item {{if eq .Page "agents"}}active{{end}}">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
                <circle cx="9" cy="7" r="4"></circle>
                <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
                <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
            </svg>
            AGENTS
        </a>
        <a href="/tasks" class="nav-item {{if eq .Page "tasks"}}active{{end}}">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                <polyline points="14 2 14 8 20 8"></polyline>
                <line x1="16" y1="13" x2="8" y2="13"></line>
                <line x1="16" y1="17" x2="8" y2="17"></line>
            </svg>
            TASKS
        </a>
        <a href="/discovery" class="nav-item {{if eq .Page "discovery"}}active{{end}}">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="11" cy="11" r="8"></circle>
                <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
            </svg>
            DISCOVERY
        </a>
        <a href="/health" class="nav-item {{if eq .Page "health"}}active{{end}}">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M22 12h-4l-3 9L9 3l-3 9H2"></path>
            </svg>
            HEALTH
        </a>
    </nav>
    <div class="sidebar-footer">
        <div class="node-status">SERVER STATUS</div>
        <div class="node-name">Running</div>
    </div>
</aside>
`

const dashboardTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Beacon - Dashboard</title>
    <link rel="stylesheet" href="/static/style.css">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
</head>
<body>
    <div class="app-container">
        ` + sidebarHTML + `

        <main class="main-content">
            <header class="header">
                <div class="header-left">
                    <span class="header-brand">Beacon</span>
                    <div class="search-box">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <circle cx="11" cy="11" r="8"></circle>
                            <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
                        </svg>
                        <input type="text" placeholder="Search...">
                    </div>
                </div>
                <div class="header-right">
                    <a href="/tasks" class="btn btn-primary">Send Task</a>
                </div>
            </header>

            <div class="page-content">
                <div class="page-header">
                    <div class="page-subtitle">OVERVIEW</div>
                    <h1 class="page-title">Dashboard</h1>
                </div>

                <div class="stats-grid">
                    <div class="stat-card">
                        <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
                            <circle cx="9" cy="7" r="4"></circle>
                            <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
                            <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
                        </svg>
                        <div class="label">TOTAL AGENTS</div>
                        <div class="value" id="agent-count">{{.TotalAgents}}</div>
                    </div>
                    <div class="stat-card">
                        <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M22 12h-4l-3 9L9 3l-3 9H2"></path>
                        </svg>
                        <div class="label">HEALTHY</div>
                        <div class="value">{{.ActiveAgents}}</div>
                    </div>
                    <div class="stat-card">
                        <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <polyline points="9 11 12 14 22 4"></polyline>
                            <path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"></path>
                        </svg>
                        <div class="label">TASKS COMPLETED</div>
                        <div class="value" id="task-count">{{.TasksDone}}</div>
                    </div>
                    <div class="stat-card">
                        <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <circle cx="12" cy="12" r="10"></circle>
                            <polyline points="12 6 12 12 16 14"></polyline>
                        </svg>
                        <div class="label">RUNNING</div>
                        <div class="value">{{.RunningTasks}}</div>
                    </div>
                    <div class="stat-card">
                        <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <circle cx="12" cy="12" r="10"></circle>
                            <line x1="15" y1="9" x2="9" y2="15"></line>
                            <line x1="9" y1="9" x2="15" y2="15"></line>
                        </svg>
                        <div class="label">FAILED</div>
                        <div class="value">{{.FailedTasks}}</div>
                    </div>
                </div>

                <div class="charts-section">
                    <div class="quick-task" style="max-width: 400px;">
                        <div class="quick-task-title">Quick Task</div>
                        <div class="quick-task-subtitle">Route a task to capable agents</div>

                        <form id="quick-task-form" action="/api/v1/route" method="POST">
                            <div class="form-group">
                                <label class="form-label">CAPABILITY</label>
                                <input type="text" class="form-input" name="capability" placeholder="e.g. echo, code-generation">
                            </div>
                            <div class="form-group">
                                <label class="form-label">PAYLOAD (JSON)</label>
                                <input type="text" class="form-input" name="payload" placeholder='{"message": "hello"}'>
                            </div>
                            <button type="submit" class="btn btn-secondary btn-block">SEND TASK</button>
                        </form>
                    </div>
                </div>

                <div class="table-section">
                    <div class="table-header">
                        <h2 class="table-title">Registered Agents</h2>
                        <a href="/agents" class="table-action">VIEW ALL</a>
                    </div>
                    <table>
                        <thead>
                            <tr>
                                <th>NAME</th>
                                <th>CAPABILITIES</th>
                                <th>ENDPOINT</th>
                                <th>STATUS</th>
                            </tr>
                        </thead>
                        <tbody>
                            {{range .Agents}}
                            <tr>
                                <td>
                                    <div class="agent-info">
                                        <div class="agent-icon">
                                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                                <circle cx="12" cy="12" r="10"></circle>
                                            </svg>
                                        </div>
                                        <div>
                                            <div class="agent-name">{{.Name}}</div>
                                            <div class="agent-id">{{.Version}}</div>
                                        </div>
                                    </div>
                                </td>
                                <td>
                                    <div class="capability-tags">
                                        {{range .Capabilities}}<span class="capability-tag">{{.Name}}</span>{{end}}
                                    </div>
                                </td>
                                <td>{{.Endpoint.Host}}:{{.Endpoint.Port}}</td>
                                <td><span class="status status-{{statusColor .Status}}">{{.Status}}</span></td>
                            </tr>
                            {{else}}
                            <tr>
                                <td colspan="4">
                                    <div class="empty-state">
                                        <div class="empty-state-title">No agents registered</div>
                                        <div class="empty-state-description">Use the SDK or API to register an agent</div>
                                    </div>
                                </td>
                            </tr>
                            {{end}}
                        </tbody>
                    </table>
                </div>
            </div>

            <footer class="footer">
                <div>Beacon</div>
                <div class="footer-links">
                    <a href="/api/v1/agents" class="footer-link">API: /api/v1/agents</a>
                    <a href="/api/v1/tasks" class="footer-link">API: /api/v1/tasks</a>
                </div>
            </footer>
        </main>
    </div>
    <script src="/static/app.js"></script>
</body>
</html>
`

const agentsTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Beacon - Agents</title>
    <link rel="stylesheet" href="/static/style.css">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
</head>
<body>
    <div class="app-container">
        ` + sidebarHTML + `

        <main class="main-content">
            <header class="header">
                <div class="header-left">
                    <span class="header-brand">Beacon</span>
                </div>
                <div class="header-right">
                    <div class="search-box">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <circle cx="11" cy="11" r="8"></circle>
                            <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
                        </svg>
                        <input type="text" placeholder="Search agents...">
                    </div>
                    <a href="/tasks" class="btn btn-primary">Send Task</a>
                </div>
            </header>

            <div class="page-content">
                <div class="page-header">
                    <h1 class="page-title">Agent Registry</h1>
                    <p class="page-description">View and manage all registered agents in the platform.</p>
                </div>

                <div class="table-section" style="margin-top: 32px;">
                    <table>
                        <thead>
                            <tr>
                                <th>AGENT</th>
                                <th>STATUS</th>
                                <th>CAPABILITIES</th>
                                <th>ENDPOINT</th>
                                <th>LAST HEARTBEAT</th>
                            </tr>
                        </thead>
                        <tbody>
                            {{range .Agents}}
                            <tr onclick="window.location='/agents/{{.ID}}'">
                                <td>
                                    <div class="agent-info">
                                        <div class="agent-icon">
                                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                                <circle cx="12" cy="12" r="10"></circle>
                                            </svg>
                                        </div>
                                        <div>
                                            <div class="agent-name">{{.Name}}</div>
                                            <div class="agent-id">{{truncateID .ID}}</div>
                                        </div>
                                    </div>
                                </td>
                                <td><span class="status status-{{statusColor .Status}}">{{.Status}}</span></td>
                                <td>
                                    <div class="capability-tags">
                                        {{range .Capabilities}}
                                        <span class="capability-tag">{{.Name}}</span>
                                        {{end}}
                                    </div>
                                </td>
                                <td>{{.Endpoint.Protocol}}://{{.Endpoint.Host}}:{{.Endpoint.Port}}</td>
                                <td>{{formatTime .LastHeartbeat}}</td>
                            </tr>
                            {{else}}
                            <tr>
                                <td colspan="5">
                                    <div class="empty-state">
                                        <div class="empty-state-title">No agents registered</div>
                                        <div class="empty-state-description">Register an agent using the SDK or POST to /api/v1/agents</div>
                                    </div>
                                </td>
                            </tr>
                            {{end}}
                        </tbody>
                    </table>
                </div>
            </div>
        </main>

        {{if .Agent}}
        <aside class="detail-panel" id="detail-panel">
            <div class="detail-header">
                <span class="detail-title">AGENT DETAILS</span>
                <a href="/agents" class="close-btn">x</a>
            </div>

            <div class="detail-section">
                <div class="detail-label">NAME</div>
                <div class="detail-value">{{.Agent.Name}}</div>
            </div>

            <div class="detail-section">
                <div class="detail-label">ID</div>
                <div class="property-value" style="font-family: monospace; font-size: 12px;">{{.Agent.ID}}</div>
            </div>

            <div class="detail-section">
                <div class="detail-label">DESCRIPTION</div>
                <div class="property-value">{{.Agent.Description}}</div>
            </div>

            <div class="detail-section">
                <div class="detail-section-title">ENDPOINT</div>
                <div class="property-item">
                    <span class="property-label">HOST</span>
                    <span class="property-value">{{.Agent.Endpoint.Host}}</span>
                </div>
                <div class="property-item">
                    <span class="property-label">PORT</span>
                    <span class="property-value">{{.Agent.Endpoint.Port}}</span>
                </div>
                <div class="property-item">
                    <span class="property-label">PROTOCOL</span>
                    <span class="property-value">{{.Agent.Endpoint.Protocol}}</span>
                </div>
            </div>

            <div class="detail-section">
                <div class="detail-section-title">METADATA</div>
                <div class="property-item">
                    <span class="property-label">VERSION</span>
                    <span class="property-value">{{.Agent.Version}}</span>
                </div>
                <div class="property-item">
                    <span class="property-label">REGISTERED</span>
                    <span class="property-value">{{formatTime .Agent.RegisteredAt}}</span>
                </div>
                <div class="property-item">
                    <span class="property-label">LAST HEARTBEAT</span>
                    <span class="property-value">{{formatTime .Agent.LastHeartbeat}}</span>
                </div>
            </div>

            <div class="detail-section">
                <div class="detail-section-title">CAPABILITIES</div>
                <div class="capability-tags" style="margin-top: 8px;">
                    {{range .Agent.Capabilities}}
                    <span class="capability-tag">{{.Name}}</span>
                    {{end}}
                </div>
            </div>
        </aside>
        {{end}}
    </div>
    <script src="/static/app.js"></script>
</body>
</html>
`

const tasksTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Beacon - Tasks</title>
    <link rel="stylesheet" href="/static/style.css">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
</head>
<body>
    <div class="app-container">
        ` + sidebarHTML + `

        <main class="main-content">
            <header class="header">
                <div class="header-left">
                    <span class="header-brand">Beacon</span>
                    <div class="search-box">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <circle cx="11" cy="11" r="8"></circle>
                            <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
                        </svg>
                        <input type="text" placeholder="Search tasks...">
                    </div>
                </div>
                <div class="header-right">
                    <a href="#" class="btn btn-primary" onclick="document.getElementById('task-modal').style.display='block'">New Task</a>
                </div>
            </header>

            <div class="page-content">
                <div class="page-header">
                    <h1 class="page-title">Tasks</h1>
                    <div style="display: flex; gap: 24px; margin-top: 8px;">
                        <span class="status status-running">{{.RunningTasks}} Running</span>
                        <span class="status status-failed">{{.FailedTasks}} Failed</span>
                        <span class="status status-completed">{{.CompletedTasks}} Completed</span>
                    </div>
                </div>

                <div class="table-section" style="margin-top: 32px;">
                    <table>
                        <thead>
                            <tr>
                                <th>TASK ID</th>
                                <th>CAPABILITY</th>
                                <th>FROM AGENT</th>
                                <th>TO AGENT</th>
                                <th>STATUS</th>
                                <th>CREATED</th>
                            </tr>
                        </thead>
                        <tbody>
                            {{range .Tasks}}
                            <tr>
                                <td style="font-family: monospace;">{{truncateID .ID}}</td>
                                <td><span class="capability-tag">{{.Capability}}</span></td>
                                <td>{{truncateID .FromAgent}}</td>
                                <td>{{truncateID .ToAgent}}</td>
                                <td><span class="status status-{{taskStatusColor .Status}}">{{.Status}}</span></td>
                                <td>{{formatTime .CreatedAt}}</td>
                            </tr>
                            {{else}}
                            <tr>
                                <td colspan="6">
                                    <div class="empty-state">
                                        <div class="empty-state-title">No tasks yet</div>
                                        <div class="empty-state-description">Send a task using POST /api/v1/route or /api/v1/tasks</div>
                                    </div>
                                </td>
                            </tr>
                            {{end}}
                        </tbody>
                    </table>
                </div>
            </div>
        </main>
    </div>
    <script src="/static/app.js"></script>
</body>
</html>
`

const discoveryTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Beacon - Discovery</title>
    <link rel="stylesheet" href="/static/style.css">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
</head>
<body>
    <div class="app-container">
        ` + sidebarHTML + `

        <main class="main-content">
            <header class="header">
                <div class="header-left">
                    <span class="header-brand">Beacon</span>
                </div>
                <div class="header-right">
                    <a href="/tasks" class="btn btn-primary">Send Task</a>
                </div>
            </header>

            <div class="page-content">
                <div class="page-header">
                    <h1 class="page-title">Discovery</h1>
                    <p class="page-description">Find agents by their capabilities.</p>
                </div>

                <form action="/discovery" method="GET" style="margin: 32px 0;">
                    <div class="search-box" style="width: 100%; max-width: none;">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <circle cx="11" cy="11" r="8"></circle>
                            <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
                        </svg>
                        <input type="text" name="capability" placeholder="Enter capability name (e.g. echo, code-generation)..." value="{{.Filter}}">
                        <button type="submit" class="btn btn-secondary" style="margin-left: 16px;">Search</button>
                    </div>
                </form>

                {{if .Filter}}
                <div style="margin-bottom: 24px;">
                    <span style="color: var(--text-muted);">Showing agents with capability:</span>
                    <span class="capability-tag" style="margin-left: 8px;">{{.Filter}}</span>
                    <a href="/discovery" style="margin-left: 16px; color: var(--text-muted);">Clear</a>
                </div>
                {{end}}

                <div class="discovery-grid">
                    {{range .Agents}}
                    <div class="discovery-card">
                        <div class="discovery-card-header">
                            <div class="discovery-icon">
                                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                    <circle cx="12" cy="12" r="10"></circle>
                                    <circle cx="12" cy="12" r="4"></circle>
                                </svg>
                            </div>
                            <span class="status status-{{statusColor .Status}}">{{.Status}}</span>
                        </div>
                        <div class="discovery-name">{{.Name}}</div>
                        <div class="discovery-description">{{if .Description}}{{.Description}}{{else}}No description{{end}}</div>
                        <div class="capability-tags" style="margin-bottom: 16px;">
                            {{range .Capabilities}}
                            <span class="capability-tag">{{.Name}}</span>
                            {{end}}
                        </div>
                        <div style="font-size: 12px; color: var(--text-muted); margin-bottom: 16px;">
                            {{.Endpoint.Protocol}}://{{.Endpoint.Host}}:{{.Endpoint.Port}}
                        </div>
                        <div class="discovery-actions">
                            <a href="/agents/{{.ID}}" class="btn btn-secondary">View Details</a>
                        </div>
                    </div>
                    {{else}}
                    <div class="discovery-card" style="grid-column: span 3;">
                        <div class="empty-state">
                            {{if .Filter}}
                            <div class="empty-state-title">No agents found with capability "{{.Filter}}"</div>
                            <div class="empty-state-description">Try searching for a different capability</div>
                            {{else}}
                            <div class="empty-state-title">No agents registered</div>
                            <div class="empty-state-description">Register agents to discover them by capability</div>
                            {{end}}
                        </div>
                    </div>
                    {{end}}
                </div>
            </div>

            <div class="status-bar">
                <div class="status-item">API: /api/v1/discover?capability=name</div>
            </div>
        </main>
    </div>
    <script src="/static/app.js"></script>
</body>
</html>
`

const healthTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Beacon - Health</title>
    <link rel="stylesheet" href="/static/style.css">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
</head>
<body>
    <div class="app-container">
        ` + sidebarHTML + `

        <main class="main-content">
            <header class="header">
                <div class="header-left">
                    <span class="header-brand">Beacon</span>
                </div>
                <div class="header-right">
                    <a href="/tasks" class="btn btn-primary">Send Task</a>
                </div>
            </header>

            <div class="page-content">
                <div class="page-header">
                    <h1 class="page-title">Health Monitor</h1>
                    <p class="page-description">Monitor agent health via heartbeat status. Agents send heartbeats every 10 seconds and are marked unhealthy after 30 seconds of inactivity.</p>
                </div>

                <div class="table-section" style="margin-top: 32px;">
                    <table>
                        <thead>
                            <tr>
                                <th>AGENT</th>
                                <th>STATUS</th>
                                <th>LAST HEARTBEAT</th>
                                <th>REGISTERED</th>
                                <th>ENDPOINT</th>
                            </tr>
                        </thead>
                        <tbody>
                            {{range .Agents}}
                            <tr>
                                <td>
                                    <div class="agent-info">
                                        <div class="agent-icon" style="background: {{if eq .Status "healthy"}}rgba(34, 197, 94, 0.2){{else}}rgba(239, 68, 68, 0.2){{end}}">
                                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="{{if eq .Status "healthy"}}#22C55E{{else}}#EF4444{{end}}" stroke-width="2">
                                                <path d="M22 12h-4l-3 9L9 3l-3 9H2"></path>
                                            </svg>
                                        </div>
                                        <div>
                                            <div class="agent-name">{{.Name}}</div>
                                            <div class="agent-id">{{truncateID .ID}}</div>
                                        </div>
                                    </div>
                                </td>
                                <td><span class="status status-{{statusColor .Status}}">{{.Status}}</span></td>
                                <td>{{formatTime .LastHeartbeat}}</td>
                                <td>{{formatTime .RegisteredAt}}</td>
                                <td>{{.Endpoint.Host}}:{{.Endpoint.Port}}</td>
                            </tr>
                            {{else}}
                            <tr>
                                <td colspan="5">
                                    <div class="empty-state">
                                        <div class="empty-state-title">No agents to monitor</div>
                                        <div class="empty-state-description">Health status appears when agents register and send heartbeats</div>
                                    </div>
                                </td>
                            </tr>
                            {{end}}
                        </tbody>
                    </table>
                </div>

                <div style="margin-top: 32px; padding: 24px; background: var(--bg-card); border: 1px solid var(--border-color); border-radius: 12px;">
                    <h3 style="margin-bottom: 16px;">Heartbeat API</h3>
                    <p style="color: var(--text-secondary); margin-bottom: 16px;">Agents should POST to keep themselves healthy:</p>
                    <code style="display: block; background: var(--bg-tertiary); padding: 16px; border-radius: 8px; font-size: 13px;">
                        POST /api/v1/heartbeat<br>
                        {"agent_id": "your-agent-id"}
                    </code>
                </div>
            </div>
        </main>
    </div>
    <script src="/static/app.js"></script>
</body>
</html>
`

const statsTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Beacon - Stats</title>
    <link rel="stylesheet" href="/static/style.css">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
</head>
<body>
    <div class="app-container">
        ` + sidebarHTML + `

        <main class="main-content">
            <header class="header">
                <div class="header-left">
                    <span class="header-brand">Beacon</span>
                </div>
                <div class="header-right">
                    <a href="/tasks" class="btn btn-primary">Send Task</a>
                </div>
            </header>

            <div class="page-content">
                <div class="page-header">
                    <h1 class="page-title">Statistics</h1>
                    <p class="page-description">Platform metrics and usage statistics.</p>
                </div>

                <div class="stats-grid" style="grid-template-columns: repeat(4, 1fr); margin-top: 32px;">
                    <div class="stat-card">
                        <div class="label">TOTAL AGENTS</div>
                        <div class="value">{{len .Agents}}</div>
                    </div>
                    <div class="stat-card">
                        <div class="label">TOTAL TASKS</div>
                        <div class="value">{{len .Tasks}}</div>
                    </div>
                    <div class="stat-card">
                        <div class="label">gRPC PORT</div>
                        <div class="value">50051</div>
                    </div>
                    <div class="stat-card">
                        <div class="label">HTTP PORT</div>
                        <div class="value">8080</div>
                    </div>
                </div>

                <div style="margin-top: 32px; display: grid; grid-template-columns: 1fr 1fr; gap: 24px;">
                    <div style="padding: 24px; background: var(--bg-card); border: 1px solid var(--border-color); border-radius: 12px;">
                        <h3 style="margin-bottom: 16px;">HTTP API Endpoints</h3>
                        <div style="font-family: monospace; font-size: 13px; color: var(--text-secondary); line-height: 2;">
                            POST /api/v1/agents - Register agent<br>
                            GET /api/v1/agents - List agents<br>
                            GET /api/v1/agents/:id - Get agent<br>
                            DELETE /api/v1/agents/:id - Deregister<br>
                            POST /api/v1/heartbeat - Send heartbeat<br>
                            GET /api/v1/discover - Find by capability<br>
                            POST /api/v1/tasks - Create task<br>
                            POST /api/v1/route - Route to capable agent<br>
                            GET /api/v1/tasks - List tasks<br>
                            PATCH /api/v1/tasks/:id - Update task
                        </div>
                    </div>
                    <div style="padding: 24px; background: var(--bg-card); border: 1px solid var(--border-color); border-radius: 12px;">
                        <h3 style="margin-bottom: 16px;">gRPC Services</h3>
                        <div style="font-family: monospace; font-size: 13px; color: var(--text-secondary); line-height: 2;">
                            a2a.RegistryService<br>
                            - Register<br>
                            - Deregister<br>
                            - GetAgent<br>
                            - ListAgents<br>
                            - Discover<br>
                            - Heartbeat<br>
                            - Watch (streaming)<br><br>
                            a2a.BrokerService<br>
                            - SendTask<br>
                            - RouteTask<br>
                            - GetTask<br>
                            - ListTasks<br>
                            - Subscribe (streaming)
                        </div>
                    </div>
                </div>
            </div>
        </main>
    </div>
    <script src="/static/app.js"></script>
</body>
</html>
`
