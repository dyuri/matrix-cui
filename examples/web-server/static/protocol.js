/**
 * MatrixProtocol - WebSocket client for Matrix CUI protocol
 */
class MatrixProtocol {
    constructor(wsUrl) {
        this.wsUrl = wsUrl;
        this.ws = null;
        this.pendingCommands = new Map();
        this.eventHandlers = new Map();
        this.connected = false;
        this.reconnectDelay = 1000;
        this.maxReconnectDelay = 30000;
    }

    connect() {
        return new Promise((resolve, reject) => {
            try {
                this.ws = new WebSocket(this.wsUrl);

                this.ws.onopen = () => {
                    console.log('WebSocket connected');
                    this.connected = true;
                    this.reconnectDelay = 1000;
                    this.triggerEvent('connected');
                    resolve();
                };

                this.ws.onclose = () => {
                    console.log('WebSocket closed');
                    this.connected = false;
                    this.triggerEvent('disconnected');

                    // Auto-reconnect
                    setTimeout(() => this.connect(), this.reconnectDelay);
                    this.reconnectDelay = Math.min(this.reconnectDelay * 2, this.maxReconnectDelay);
                };

                this.ws.onerror = (error) => {
                    console.error('WebSocket error:', error);
                    this.triggerEvent('error', { error });
                    reject(error);
                };

                this.ws.onmessage = (event) => {
                    try {
                        const envelope = JSON.parse(event.data);
                        this.handleMessage(envelope);
                    } catch (error) {
                        console.error('Failed to parse message:', error);
                    }
                };
            } catch (error) {
                reject(error);
            }
        });
    }

    disconnect() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }

    handleMessage(envelope) {
        switch (envelope.type) {
            case 'response':
                this.handleResponse(envelope);
                break;
            case 'event':
                this.handleEvent(envelope);
                break;
            default:
                console.warn('Unknown message type:', envelope.type);
        }
    }

    handleResponse(envelope) {
        const { id, payload } = envelope;

        // Check for special server-initiated responses
        if (id === 'snapshot') {
            this.triggerEvent('snapshot', payload);
            return;
        }
        if (id === 'delta') {
            this.triggerEvent('delta', payload);
            return;
        }

        // Match response to pending command
        const pending = this.pendingCommands.get(id);
        if (pending) {
            this.pendingCommands.delete(id);
            if (payload.ok) {
                pending.resolve(payload);
            } else {
                pending.reject(new Error(payload.error || 'Command failed'));
            }
        }
    }

    handleEvent(envelope) {
        const { payload } = envelope;

        switch (payload.event) {
            case 'key':
                this.triggerEvent('key', payload);
                break;
            case 'mouse':
                this.triggerEvent('mouse', payload);
                break;
            case 'resize':
                this.triggerEvent('resize', payload);
                break;
            default:
                console.warn('Unknown event type:', payload.event);
        }
    }

    sendCommand(cmd) {
        return new Promise((resolve, reject) => {
            if (!this.connected || !this.ws) {
                reject(new Error('Not connected'));
                return;
            }

            const id = this.generateUUID();
            const envelope = {
                type: 'command',
                id: id,
                payload: cmd
            };

            // Store pending command
            this.pendingCommands.set(id, { resolve, reject });

            // Send command
            try {
                this.ws.send(JSON.stringify(envelope));
            } catch (error) {
                this.pendingCommands.delete(id);
                reject(error);
            }

            // Timeout after 5 seconds
            setTimeout(() => {
                if (this.pendingCommands.has(id)) {
                    this.pendingCommands.delete(id);
                    reject(new Error('Command timeout'));
                }
            }, 5000);
        });
    }

    on(eventType, handler) {
        if (!this.eventHandlers.has(eventType)) {
            this.eventHandlers.set(eventType, []);
        }
        this.eventHandlers.get(eventType).push(handler);
    }

    off(eventType, handler) {
        if (this.eventHandlers.has(eventType)) {
            const handlers = this.eventHandlers.get(eventType);
            const index = handlers.indexOf(handler);
            if (index !== -1) {
                handlers.splice(index, 1);
            }
        }
    }

    triggerEvent(eventType, data) {
        if (this.eventHandlers.has(eventType)) {
            for (const handler of this.eventHandlers.get(eventType)) {
                try {
                    handler(data);
                } catch (error) {
                    console.error(`Error in ${eventType} handler:`, error);
                }
            }
        }
    }

    generateUUID() {
        return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
            const r = Math.random() * 16 | 0;
            const v = c === 'x' ? r : (r & 0x3 | 0x8);
            return v.toString(16);
        });
    }

    // Command helpers
    async put(x, y, cell) {
        return this.sendCommand({
            cmd: 'put',
            x: x,
            y: y,
            cell: cell
        });
    }

    async get(x, y) {
        return this.sendCommand({
            cmd: 'get',
            x: x,
            y: y
        });
    }

    async clear() {
        return this.sendCommand({ cmd: 'clear' });
    }

    async fill(cell) {
        return this.sendCommand({
            cmd: 'fill',
            cell: cell
        });
    }

    async putString(x, y, text, cell) {
        return this.sendCommand({
            cmd: 'putString',
            x: x,
            y: y,
            text: text,
            cell: cell
        });
    }

    async getSize() {
        return this.sendCommand({ cmd: 'getSize' });
    }

    // Send event to server (fire-and-forget, no response expected)
    sendEvent(eventPayload) {
        if (!this.connected || !this.ws) {
            console.warn('Cannot send event: not connected');
            return;
        }

        const envelope = {
            type: 'event',
            payload: eventPayload
        };

        try {
            this.ws.send(JSON.stringify(envelope));
        } catch (error) {
            console.error('Failed to send event:', error);
        }
    }
}
