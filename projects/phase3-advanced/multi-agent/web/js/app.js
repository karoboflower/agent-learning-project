// 应用主逻辑
class MultiAgentApp {
    constructor() {
        this.agents = new Map();
        this.tasks = new Map();
        this.results = new Map();
        this.currentTab = 'agents';
        this.apiBaseUrl = window.location.origin + '/api';
        this.wsUrl = `ws://${window.location.host}/ws?agent_id=web-client`;
    }

    // 初始化应用
    init() {
        console.log('Initializing Multi-Agent App...');

        // 初始化WebSocket
        this.initWebSocket();

        // 初始化UI
        this.initUI();

        // 加载初始数据
        this.loadInitialData();
    }

    // 初始化WebSocket
    initWebSocket() {
        window.wsClient = new WebSocketClient(this.wsUrl);

        // 注册消息处理器
        wsClient.on('AGENT_REGISTERED', (msg) => this.handleAgentRegistered(msg));
        wsClient.on('AGENT_STATUS_UPDATE', (msg) => this.handleAgentStatusUpdate(msg));
        wsClient.on('TASK_CREATED', (msg) => this.handleTaskCreated(msg));
        wsClient.on('TASK_ASSIGNED', (msg) => this.handleTaskAssigned(msg));
        wsClient.on('TASK_STATUS_UPDATE', (msg) => this.handleTaskStatusUpdate(msg));
        wsClient.on('RESULT_SUBMITTED', (msg) => this.handleResultSubmitted(msg));
        wsClient.on('RESULT_AGGREGATED', (msg) => this.handleResultAggregated(msg));

        // 连接
        wsClient.connect();
    }

    // 初始化UI
    initUI() {
        // 标签页切换
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const tab = e.target.dataset.tab;
                this.switchTab(tab);
            });
        });

        // 注册Agent按钮
        document.getElementById('registerAgentBtn').addEventListener('click', () => {
            this.showModal('registerAgentModal');
        });

        // 提交注册
        document.getElementById('submitRegisterBtn').addEventListener('click', () => {
            this.submitAgentRegistration();
        });

        // 取消注册
        document.getElementById('cancelRegisterBtn').addEventListener('click', () => {
            this.hideModal('registerAgentModal');
        });

        // 刷新Agent
        document.getElementById('refreshAgentsBtn').addEventListener('click', () => {
            this.refreshAgents();
        });

        // 创建任务按钮
        document.getElementById('createTaskBtn').addEventListener('click', () => {
            this.showModal('createTaskModal');
        });

        // 提交创建任务
        document.getElementById('submitCreateBtn').addEventListener('click', () => {
            this.submitTaskCreation();
        });

        // 取消创建任务
        document.getElementById('cancelCreateBtn').addEventListener('click', () => {
            this.hideModal('createTaskModal');
        });

        // 刷新任务
        document.getElementById('refreshTasksBtn').addEventListener('click', () => {
            this.refreshTasks();
        });

        // 刷新结果
        document.getElementById('refreshResultsBtn').addEventListener('click', () => {
            this.refreshResults();
        });

        // 导出结果
        document.getElementById('exportResultsBtn').addEventListener('click', () => {
            this.exportResults();
        });

        // 对比结果
        document.getElementById('compareBtn').addEventListener('click', () => {
            this.compareResults();
        });

        // 模态框关闭按钮
        document.querySelectorAll('.modal-close').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const modal = e.target.closest('.modal');
                this.hideModal(modal.id);
            });
        });

        // 点击模态框背景关闭
        document.querySelectorAll('.modal').forEach(modal => {
            modal.addEventListener('click', (e) => {
                if (e.target === modal) {
                    this.hideModal(modal.id);
                }
            });
        });
    }

    // 加载初始数据
    async loadInitialData() {
        await this.refreshAgents();
        await this.refreshTasks();
        await this.refreshResults();
    }

    // 切换标签页
    switchTab(tab) {
        // 更新按钮状态
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.classList.remove('active');
            if (btn.dataset.tab === tab) {
                btn.classList.add('active');
            }
        });

        // 更新内容显示
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.remove('active');
        });
        document.getElementById(tab).classList.add('active');

        this.currentTab = tab;
    }

    // 显示模态框
    showModal(modalId) {
        document.getElementById(modalId).classList.add('active');
    }

    // 隐藏模态框
    hideModal(modalId) {
        document.getElementById(modalId).classList.remove('active');

        // 清空表单
        const modal = document.getElementById(modalId);
        const form = modal.querySelector('form');
        if (form) {
            form.reset();
        }
    }

    // 提交Agent注册
    async submitAgentRegistration() {
        const form = document.getElementById('registerAgentForm');
        const formData = new FormData(form);

        const agent = {
            id: formData.get('id'),
            name: formData.get('name'),
            capabilities: formData.get('capabilities').split(',').map(s => s.trim()),
            max_tasks: parseInt(formData.get('maxTasks'))
        };

        try {
            const response = await fetch(`${this.apiBaseUrl}/agents`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(agent)
            });

            if (response.ok) {
                this.hideModal('registerAgentModal');
                this.showNotification('Agent注册成功', 'success');
                await this.refreshAgents();
            } else {
                const error = await response.text();
                this.showNotification(`注册失败: ${error}`, 'error');
            }
        } catch (error) {
            this.showNotification(`注册失败: ${error.message}`, 'error');
        }
    }

    // 刷新Agent列表
    async refreshAgents() {
        try {
            const response = await fetch(`${this.apiBaseUrl}/agents`);
            if (!response.ok) throw new Error('Failed to fetch agents');

            const agents = await response.json();
            this.agents.clear();
            agents.forEach(agent => this.agents.set(agent.id, agent));

            this.renderAgents();
            this.updateAgentStats();
        } catch (error) {
            console.error('Failed to refresh agents:', error);
        }
    }

    // 渲染Agent列表
    renderAgents() {
        const tbody = document.getElementById('agentTableBody');

        if (this.agents.size === 0) {
            tbody.innerHTML = '<tr class="empty-row"><td colspan="8">暂无Agent数据</td></tr>';
            return;
        }

        tbody.innerHTML = '';
        this.agents.forEach(agent => {
            const tr = document.createElement('tr');
            tr.innerHTML = `
                <td>${agent.id}</td>
                <td>${agent.name}</td>
                <td><span class="status-badge ${agent.status.toLowerCase()}">${agent.status}</span></td>
                <td>
                    <div class="capability-tags">
                        ${agent.capabilities.map(cap => `<span class="capability-tag">${cap}</span>`).join('')}
                    </div>
                </td>
                <td>${(agent.load * 100).toFixed(0)}%</td>
                <td>${agent.current_tasks} / ${agent.max_tasks}</td>
                <td>${this.formatTime(agent.last_heartbeat)}</td>
                <td>
                    <div class="action-buttons">
                        <button class="action-btn action-btn-info" onclick="app.viewAgentDetail('${agent.id}')">详情</button>
                        <button class="action-btn action-btn-danger" onclick="app.removeAgent('${agent.id}')">删除</button>
                    </div>
                </td>
            `;
            tbody.appendChild(tr);
        });
    }

    // 更新Agent统计
    updateAgentStats() {
        const total = this.agents.size;
        let active = 0, idle = 0, offline = 0;

        this.agents.forEach(agent => {
            if (agent.status === 'ACTIVE') active++;
            else if (agent.status === 'IDLE') idle++;
            else if (agent.status === 'OFFLINE') offline++;
        });

        document.getElementById('totalAgents').textContent = total;
        document.getElementById('activeAgents').textContent = active;
        document.getElementById('idleAgents').textContent = idle;
        document.getElementById('offlineAgents').textContent = offline;
    }

    // 提交任务创建
    async submitTaskCreation() {
        const form = document.getElementById('createTaskForm');
        const formData = new FormData(form);

        const task = {
            id: formData.get('id'),
            type: formData.get('type'),
            priority: parseInt(formData.get('priority')),
            description: formData.get('description'),
            capabilities: formData.get('capabilities') ?
                formData.get('capabilities').split(',').map(s => s.trim()) : []
        };

        try {
            const response = await fetch(`${this.apiBaseUrl}/tasks`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(task)
            });

            if (response.ok) {
                this.hideModal('createTaskModal');
                this.showNotification('任务创建成功', 'success');
                await this.refreshTasks();
            } else {
                const error = await response.text();
                this.showNotification(`创建失败: ${error}`, 'error');
            }
        } catch (error) {
            this.showNotification(`创建失败: ${error.message}`, 'error');
        }
    }

    // 刷新任务列表
    async refreshTasks() {
        try {
            const response = await fetch(`${this.apiBaseUrl}/tasks`);
            if (!response.ok) throw new Error('Failed to fetch tasks');

            const tasks = await response.json();
            this.tasks.clear();
            tasks.forEach(task => this.tasks.set(task.id, task));

            this.renderTasks();
            this.updateTaskStats();
            this.renderTaskAllocation();
        } catch (error) {
            console.error('Failed to refresh tasks:', error);
        }
    }

    // 渲染任务列表
    renderTasks() {
        const tbody = document.getElementById('taskTableBody');

        if (this.tasks.size === 0) {
            tbody.innerHTML = '<tr class="empty-row"><td colspan="8">暂无任务数据</td></tr>';
            return;
        }

        tbody.innerHTML = '';
        this.tasks.forEach(task => {
            const tr = document.createElement('tr');
            tr.innerHTML = `
                <td>${task.id}</td>
                <td>${task.type}</td>
                <td><span class="status-badge ${task.status.toLowerCase()}">${task.status}</span></td>
                <td>${task.priority}</td>
                <td>${task.assigned_to || '-'}</td>
                <td>
                    <div class="progress-bar">
                        <div class="progress-fill" style="width: ${task.progress || 0}%"></div>
                    </div>
                </td>
                <td>${this.formatTime(task.created_at)}</td>
                <td>
                    <div class="action-buttons">
                        <button class="action-btn action-btn-info" onclick="app.viewTaskDetail('${task.id}')">详情</button>
                    </div>
                </td>
            `;
            tbody.appendChild(tr);
        });
    }

    // 更新任务统计
    updateTaskStats() {
        const total = this.tasks.size;
        let pending = 0, running = 0, completed = 0;

        this.tasks.forEach(task => {
            if (task.status === 'PENDING') pending++;
            else if (task.status === 'RUNNING') running++;
            else if (task.status === 'COMPLETED') completed++;
        });

        document.getElementById('totalTasks').textContent = total;
        document.getElementById('pendingTasks').textContent = pending;
        document.getElementById('runningTasks').textContent = running;
        document.getElementById('completedTasks').textContent = completed;
    }

    // 渲染任务分配可视化
    renderTaskAllocation() {
        const container = document.getElementById('taskAllocationChart');

        // 统计每个Agent的任务数
        const allocation = new Map();
        this.tasks.forEach(task => {
            if (task.assigned_to) {
                allocation.set(task.assigned_to, (allocation.get(task.assigned_to) || 0) + 1);
            }
        });

        if (allocation.size === 0) {
            container.innerHTML = '<p class="empty-message">暂无任务分配数据</p>';
            return;
        }

        // 简单的柱状图
        let html = '<div style="display: flex; gap: 20px; align-items: flex-end; height: 150px;">';
        const maxTasks = Math.max(...allocation.values());

        allocation.forEach((count, agentId) => {
            const height = (count / maxTasks) * 100;
            html += `
                <div style="flex: 1; display: flex; flex-direction: column; align-items: center; gap: 10px;">
                    <div style="font-weight: 600; color: #409eff;">${count}</div>
                    <div style="width: 100%; height: ${height}%; background: linear-gradient(180deg, #667eea 0%, #764ba2 100%); border-radius: 4px;"></div>
                    <div style="font-size: 12px; color: #606266;">${agentId}</div>
                </div>
            `;
        });

        html += '</div>';
        container.innerHTML = html;
    }

    // 刷新结果列表
    async refreshResults() {
        try {
            const response = await fetch(`${this.apiBaseUrl}/results`);
            if (!response.ok) throw new Error('Failed to fetch results');

            const results = await response.json();
            this.results.clear();
            results.forEach(result => this.results.set(result.id, result));

            this.renderResults();
            this.updateResultStats();
            this.updateTaskSelector();
        } catch (error) {
            console.error('Failed to refresh results:', error);
        }
    }

    // 渲染结果列表
    renderResults() {
        const tbody = document.getElementById('resultTableBody');

        if (this.results.size === 0) {
            tbody.innerHTML = '<tr class="empty-row"><td colspan="8">暂无结果数据</td></tr>';
            return;
        }

        tbody.innerHTML = '';
        this.results.forEach(result => {
            const tr = document.createElement('tr');
            const confidence = result.data?.confidence || 0;
            tr.innerHTML = `
                <td>${result.id}</td>
                <td>${result.task_id}</td>
                <td>${result.agent_id}</td>
                <td><span class="status-badge ${result.status.toLowerCase()}">${result.status}</span></td>
                <td>${result.score.toFixed(1)}</td>
                <td>${(confidence * 100).toFixed(0)}%</td>
                <td>${this.formatTime(result.created_at)}</td>
                <td>
                    <div class="action-buttons">
                        <button class="action-btn action-btn-info" onclick="app.viewResultDetail('${result.id}')">详情</button>
                    </div>
                </td>
            `;
            tbody.appendChild(tr);
        });
    }

    // 更新结果统计
    updateResultStats() {
        const total = this.results.size;
        let validated = 0, aggregated = 0;
        let totalConfidence = 0;

        this.results.forEach(result => {
            if (result.status === 'VALIDATED') validated++;
            if (result.status === 'MERGED') aggregated++;
            totalConfidence += (result.data?.confidence || 0);
        });

        const avgConfidence = total > 0 ? (totalConfidence / total * 100).toFixed(0) : 0;

        document.getElementById('totalResults').textContent = total;
        document.getElementById('validatedResults').textContent = validated;
        document.getElementById('aggregatedResults').textContent = aggregated;
        document.getElementById('avgConfidence').textContent = avgConfidence + '%';
    }

    // 更新任务选择器
    updateTaskSelector() {
        const selector = document.getElementById('taskSelector');

        // 获取有结果的任务
        const tasksWithResults = new Set();
        this.results.forEach(result => {
            tasksWithResults.add(result.task_id);
        });

        selector.innerHTML = '<option value="">-- 选择任务 --</option>';
        tasksWithResults.forEach(taskId => {
            const option = document.createElement('option');
            option.value = taskId;
            option.textContent = taskId;
            selector.appendChild(option);
        });
    }

    // 对比结果
    async compareResults() {
        const taskId = document.getElementById('taskSelector').value;
        if (!taskId) {
            this.showNotification('请先选择任务', 'warning');
            return;
        }

        try {
            const response = await fetch(`${this.apiBaseUrl}/results/aggregate/${taskId}`);
            if (!response.ok) throw new Error('Failed to fetch aggregated result');

            const aggregated = await response.json();
            this.renderComparison(aggregated);
        } catch (error) {
            this.showNotification(`获取聚合结果失败: ${error.message}`, 'error');
        }
    }

    // 渲染结果对比
    renderComparison(aggregated) {
        const container = document.getElementById('comparisonView');

        if (!aggregated || !aggregated.results) {
            container.innerHTML = '<p class="empty-message">暂无结果数据</p>';
            return;
        }

        let html = `
            <div style="margin-bottom: 20px; padding: 15px; background: #e6f7ff; border-radius: 6px;">
                <div style="display: flex; justify-content: space-between; align-items: center;">
                    <div>
                        <strong>聚合策略:</strong> ${aggregated.strategy}<br>
                        <strong>置信度:</strong> ${(aggregated.confidence * 100).toFixed(0)}%
                    </div>
                    <div>
                        <strong>结果数:</strong> ${aggregated.results.length}<br>
                        <strong>冲突数:</strong> ${aggregated.conflicts?.length || 0}
                    </div>
                </div>
            </div>
            <div class="comparison-grid">
        `;

        aggregated.results.forEach(result => {
            html += `
                <div class="result-item">
                    <div class="result-item-header">
                        <span class="result-item-title">${result.agent_id}</span>
                        <span class="result-item-score">分数: ${result.score}</span>
                    </div>
                    <div class="result-item-body">
                        <pre>${JSON.stringify(result.data, null, 2)}</pre>
                    </div>
                </div>
            `;
        });

        html += '</div>';

        // 显示合并结果
        html += `
            <div style="margin-top: 20px; padding: 15px; background: #f0f9ff; border-radius: 6px;">
                <h4 style="margin-bottom: 10px;">合并后的结果:</h4>
                <pre style="background: #fff; padding: 15px; border-radius: 4px;">${JSON.stringify(aggregated.merged_data, null, 2)}</pre>
            </div>
        `;

        // 显示冲突
        if (aggregated.conflicts && aggregated.conflicts.length > 0) {
            html += `
                <div style="margin-top: 20px; padding: 15px; background: #fff2f0; border-radius: 6px;">
                    <h4 style="margin-bottom: 10px; color: #ff4d4f;">检测到的冲突:</h4>
            `;

            aggregated.conflicts.forEach((conflict, index) => {
                html += `
                    <div style="margin-bottom: 10px; padding: 10px; background: #fff; border-radius: 4px;">
                        <strong>冲突 ${index + 1}:</strong> ${conflict.field}<br>
                        <strong>值:</strong> ${conflict.values.join(', ')}<br>
                        <strong>解决方案:</strong> ${conflict.resolution}
                    </div>
                `;
            });

            html += '</div>';
        }

        container.innerHTML = html;
    }

    // 导出结果
    exportResults() {
        const results = Array.from(this.results.values());
        const json = JSON.stringify(results, null, 2);
        const blob = new Blob([json], { type: 'application/json' });
        const url = URL.createObjectURL(blob);

        const a = document.createElement('a');
        a.href = url;
        a.download = `results-${Date.now()}.json`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);

        this.showNotification('结果已导出', 'success');
    }

    // 查看Agent详情
    viewAgentDetail(agentId) {
        const agent = this.agents.get(agentId);
        if (!agent) return;

        document.getElementById('detailModalTitle').textContent = `Agent详情: ${agent.name}`;
        document.getElementById('detailModalContent').textContent = JSON.stringify(agent, null, 2);
        this.showModal('detailModal');
    }

    // 查看任务详情
    viewTaskDetail(taskId) {
        const task = this.tasks.get(taskId);
        if (!task) return;

        document.getElementById('detailModalTitle').textContent = `任务详情: ${task.id}`;
        document.getElementById('detailModalContent').textContent = JSON.stringify(task, null, 2);
        this.showModal('detailModal');
    }

    // 查看结果详情
    viewResultDetail(resultId) {
        const result = this.results.get(resultId);
        if (!result) return;

        document.getElementById('detailModalTitle').textContent = `结果详情: ${result.id}`;
        document.getElementById('detailModalContent').textContent = JSON.stringify(result, null, 2);
        this.showModal('detailModal');
    }

    // 删除Agent
    async removeAgent(agentId) {
        if (!confirm(`确定要删除Agent ${agentId} 吗？`)) {
            return;
        }

        try {
            const response = await fetch(`${this.apiBaseUrl}/agents/${agentId}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                this.showNotification('Agent已删除', 'success');
                await this.refreshAgents();
            } else {
                const error = await response.text();
                this.showNotification(`删除失败: ${error}`, 'error');
            }
        } catch (error) {
            this.showNotification(`删除失败: ${error.message}`, 'error');
        }
    }

    // WebSocket消息处理器
    handleAgentRegistered(msg) {
        console.log('Agent registered:', msg);
        this.refreshAgents();
    }

    handleAgentStatusUpdate(msg) {
        console.log('Agent status updated:', msg);
        this.refreshAgents();
    }

    handleTaskCreated(msg) {
        console.log('Task created:', msg);
        this.refreshTasks();
    }

    handleTaskAssigned(msg) {
        console.log('Task assigned:', msg);
        this.refreshTasks();
    }

    handleTaskStatusUpdate(msg) {
        console.log('Task status updated:', msg);
        this.refreshTasks();
    }

    handleResultSubmitted(msg) {
        console.log('Result submitted:', msg);
        this.refreshResults();
    }

    handleResultAggregated(msg) {
        console.log('Result aggregated:', msg);
        this.refreshResults();
    }

    // 显示通知
    showNotification(message, type = 'info') {
        // 简单的alert，可以替换为更好的通知组件
        alert(message);
    }

    // 格式化时间
    formatTime(timestamp) {
        if (!timestamp) return '-';
        const date = new Date(timestamp);
        const now = new Date();
        const diff = Math.floor((now - date) / 1000);

        if (diff < 60) return `${diff}秒前`;
        if (diff < 3600) return `${Math.floor(diff / 60)}分钟前`;
        if (diff < 86400) return `${Math.floor(diff / 3600)}小时前`;

        return date.toLocaleDateString('zh-CN') + ' ' + date.toLocaleTimeString('zh-CN');
    }
}

// 初始化应用
const app = new MultiAgentApp();

// 页面加载完成后初始化
window.addEventListener('DOMContentLoaded', () => {
    app.init();
});

// 页面卸载时关闭WebSocket
window.addEventListener('beforeunload', () => {
    if (window.wsClient) {
        window.wsClient.close();
    }
});
