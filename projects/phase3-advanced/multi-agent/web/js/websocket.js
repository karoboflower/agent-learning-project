// WebSocket客户端管理
class WebSocketClient {
    constructor(url) {
        this.url = url;
        this.ws = null;
        this.reconnectInterval = 3000;
        this.reconnectTimer = null;
        this.messageHandlers = new Map();
        this.isIntentionalClose = false;
    }

    // 连接WebSocket
    connect() {
        try {
            this.ws = new WebSocket(this.url);
            this.setupEventHandlers();
        } catch (error) {
            console.error('WebSocket connection failed:', error);
            this.scheduleReconnect();
        }
    }

    // 设置事件处理器
    setupEventHandlers() {
        this.ws.onopen = () => {
            console.log('WebSocket connected');
            this.updateStatus('connected');
            this.clearReconnectTimer();

            // 发送连接消息
            this.send({
                message_id: this.generateId(),
                type: 'CLIENT_CONNECT',
                from: 'web-client',
                to: 'server',
                timestamp: new Date().toISOString(),
                payload: {
                    client_type: 'web',
                    user_agent: navigator.userAgent
                }
            });
        };

        this.ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                this.handleMessage(message);
            } catch (error) {
                console.error('Failed to parse message:', error);
            }
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.updateStatus('error');
        };

        this.ws.onclose = (event) => {
            console.log('WebSocket closed:', event.code, event.reason);
            this.updateStatus('disconnected');

            if (!this.isIntentionalClose) {
                this.scheduleReconnect();
            }
        };
    }

    // 处理接收到的消息
    handleMessage(message) {
        console.log('Received message:', message);

        const type = message.type;
        if (this.messageHandlers.has(type)) {
            const handlers = this.messageHandlers.get(type);
            handlers.forEach(handler => {
                try {
                    handler(message);
                } catch (error) {
                    console.error(`Handler error for ${type}:`, error);
                }
            });
        }

        // 触发通用消息事件
        if (this.messageHandlers.has('*')) {
            const handlers = this.messageHandlers.get('*');
            handlers.forEach(handler => {
                try {
                    handler(message);
                } catch (error) {
                    console.error('Generic handler error:', error);
                }
            });
        }
    }

    // 注册消息处理器
    on(type, handler) {
        if (!this.messageHandlers.has(type)) {
            this.messageHandlers.set(type, []);
        }
        this.messageHandlers.get(type).push(handler);
    }

    // 移除消息处理器
    off(type, handler) {
        if (!this.messageHandlers.has(type)) {
            return;
        }
        const handlers = this.messageHandlers.get(type);
        const index = handlers.indexOf(handler);
        if (index > -1) {
            handlers.splice(index, 1);
        }
    }

    // 发送消息
    send(message) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            // 确保消息包含必需字段
            if (!message.message_id) {
                message.message_id = this.generateId();
            }
            if (!message.timestamp) {
                message.timestamp = new Date().toISOString();
            }

            this.ws.send(JSON.stringify(message));
            console.log('Sent message:', message);
            return true;
        } else {
            console.error('WebSocket is not connected');
            return false;
        }
    }

    // 关闭连接
    close() {
        this.isIntentionalClose = true;
        this.clearReconnectTimer();
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }

    // 计划重连
    scheduleReconnect() {
        this.clearReconnectTimer();
        console.log(`Reconnecting in ${this.reconnectInterval}ms...`);
        this.reconnectTimer = setTimeout(() => {
            console.log('Attempting to reconnect...');
            this.connect();
        }, this.reconnectInterval);
    }

    // 清除重连计时器
    clearReconnectTimer() {
        if (this.reconnectTimer) {
            clearTimeout(this.reconnectTimer);
            this.reconnectTimer = null;
        }
    }

    // 更新���接状态显示
    updateStatus(status) {
        const statusIndicator = document.getElementById('wsStatus');
        const statusText = statusIndicator.querySelector('.status-text');

        statusIndicator.classList.remove('connected', 'disconnected', 'error');

        switch (status) {
            case 'connected':
                statusIndicator.classList.add('connected');
                statusText.textContent = '已连接';
                break;
            case 'disconnected':
                statusIndicator.classList.add('disconnected');
                statusText.textContent = '未连接';
                break;
            case 'error':
                statusIndicator.classList.add('disconnected');
                statusText.textContent = '连接错误';
                break;
        }
    }

    // 生成唯一ID
    generateId() {
        return 'msg-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
    }

    // 获取连接状态
    isConnected() {
        return this.ws && this.ws.readyState === WebSocket.OPEN;
    }
}

// 导出实例
window.wsClient = null;
