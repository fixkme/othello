import { _decorator, EventTarget } from 'cc';
const { ccclass } = _decorator;

type WebSocketEvent = 'open' | 'close' | 'error' | 'message' | 'raw_message';

/**
 * 全局WebSocket管理类 (单例模式)
 * @version 1.0.0
 * @description 
 * - 支持自动重连
 * - 心跳检测机制
 * - 消息队列缓存
 * - 二进制数据支持
 * - 完善的事件系统
 */
export class GlobalWebSocket {
    private static _instance: GlobalWebSocket | null = null;
    private _socket: WebSocket | null = null;
    private _isConnected: boolean = false;
    private _url: string = "";
    private _eventTarget: EventTarget = new EventTarget();
    
    // 重连配置
    private _reconnectAttempts: number = 0;
    private _maxReconnectAttempts: number = 5;
    private _reconnectInterval: number = 3000;
    
    // 心跳配置
    private _heartbeatInterval: number = 30000;
    private _heartbeatTimer: number | null = null;
    private _lastPongTime: number = 0;
    private _heartbeatTimeout: number = 60000;
    
    // 消息队列
    private _messageQueue: Array<{data: any, tries: number}> = [];
    private _maxRetryCount: number = 3;
    
    // 二进制配置
    private _binaryType: BinaryType = "arraybuffer";

    /**
     * 获取单例实例
     */
    public static getInstance(): GlobalWebSocket {
        if (!this._instance) {
            this._instance = new GlobalWebSocket();
        }
        return this._instance;
    }

    private constructor() {
        // 私有构造函数确保单例
    }

    /**
     * 初始化WebSocket连接
     * @param url WebSocket服务器地址
     * @param binaryType 二进制消息类型，默认"arraybuffer"
     */
    public init(url: string, binaryType: BinaryType = "arraybuffer"): void {
        this._url = url;
        this._binaryType = binaryType;
        this.connect();
    }

    private connect(): void {
        this.close(); // 关闭现有连接

        console.log(`[WebSocket] Connecting to: ${this._url}`);
        try {
            this._socket = new WebSocket(this._url);
            this._socket.binaryType = this._binaryType;
            this._socket.onopen = this._onOpen.bind(this);
            this._socket.onmessage = this._onMessage.bind(this);
            this._socket.onerror = this._onError.bind(this);
            this._socket.onclose = this._onClose.bind(this);
        } catch (e) {
            console.error('[WebSocket] Connection error:', e);
            this._scheduleReconnect();
        }
    }

    private _onOpen(event: Event): void {
        console.log('[WebSocket] Connected');
        this._isConnected = true;
        this._reconnectAttempts = 0;
        this._lastPongTime = Date.now();
        
        this._startHeartbeat();
        this._flushMessageQueue();
        
        this._eventTarget.emit('open', event);
    }

    private _onMessage(event: MessageEvent): void {
        let data = event.data;
        try {
            // 处理二进制数据
            if (data instanceof ArrayBuffer) {
                this._processBinaryData(data);
                return;
            }
            
            // 处理文本数据
            if (typeof data === 'string') {
                data = JSON.parse(data);
                
                // 心跳响应处理
                if (data.cmd === 'pong') {
                    this._lastPongTime = Date.now();
                    return;
                }
            }
            
            // 触发特定命令事件
            if (data.cmd) {
                this._eventTarget.emit(data.cmd, data);
            }
            
            // 触发通用消息事件
            this._eventTarget.emit('message', data);
        } catch (e) {
            console.error('[WebSocket] Message parse error:', e);
            this._eventTarget.emit('raw_message', event.data);
        }
    }

    private _processBinaryData(buffer: ArrayBuffer): void {
        try {
            const decoder = new TextDecoder()
            const content = decoder.decode(buffer);
            console.log('[WebSocket] recv Binary data:', content);
            // const view = new DataView(buffer);
            // const msgType = view.getInt32(0);
            
            // switch (msgType) {
            //     case 1: // 示例: 位置更新
            //         const x = view.getFloat32(4);
            //         const y = view.getFloat32(8);
            //         this._eventTarget.emit('binary_position', {x, y});
            //         break;
            //     default:
            //         this._eventTarget.emit('binary_data', buffer);
            // }
        } catch (e) {
            console.error('[WebSocket] Binary data process error:', e);
        }
    }

    private _onError(event: Event): void {
        console.error('[WebSocket] Error:', event);
        this._eventTarget.emit('error', event);
    }

    private _onClose(event: CloseEvent): void {
        console.log('[WebSocket] Closed:', event);
        this._isConnected = false;
        this._stopHeartbeat();
        this._eventTarget.emit('close', event);

        this._scheduleReconnect();
    }

    private _scheduleReconnect(): void {
        if (this._reconnectAttempts < this._maxReconnectAttempts) {
            setTimeout(() => {
                this._reconnectAttempts++;
                console.log(`[WebSocket] Reconnecting... (${this._reconnectAttempts}/${this._maxReconnectAttempts})`);
                this.connect();
            }, this._reconnectInterval);
        } else {
            console.log('[WebSocket] Max reconnection attempts reached');
        }
    }

    private _startHeartbeat(): void {
        this._stopHeartbeat();
        
        this._heartbeatTimer = setInterval(() => {
            // 检查心跳超时
            if (Date.now() - this._lastPongTime > this._heartbeatTimeout) {
                // console.warn('[WebSocket] Heartbeat timeout, reconnecting...');
                // this.close();
                // this.connect();
                return;
            }
            
            // 发送心跳
            this.send({
                cmd: 'ping',
                timestamp: Date.now()
            });
        }, this._heartbeatInterval) as unknown as number;
    }

    private _stopHeartbeat(): void {
        if (this._heartbeatTimer) {
            clearInterval(this._heartbeatTimer);
            this._heartbeatTimer = null;
        }
    }

    private _flushMessageQueue(): void {
        const failedMessages: Array<{data: any, tries: number}> = [];
        
        while (this._messageQueue.length > 0) {
            const item = this._messageQueue.shift()!;
            item.tries++;
            
            if (!this.send(item.data, true)) {
                if (item.tries < this._maxRetryCount) {
                    failedMessages.push(item);
                } else {
                    console.warn('[WebSocket] Message max retry reached:', item.data);
                }
            }
        }
        
        this._messageQueue = failedMessages.concat(this._messageQueue);
    }

    /**
     * 发送消息
     * @param data 要发送的数据
     * @param force 是否强制发送（不加入队列）
     * @returns 是否发送成功
     */
    public send(data: any, force: boolean = false): boolean {
        if (!this._isConnected && !force) {
            // 未连接时存入队列
            if (this._messageQueue.length < 100) {
                this._messageQueue.push({data, tries: 0});
            } else {
                console.warn('[WebSocket] Message queue is full');
            }
            return false;
        }
        
        try {
            if (typeof data === 'string') {
                this._socket?.send(data);
            } else {
                const jsonString = JSON.stringify(data);
                const encoder = new TextEncoder();
                const arrayBuffer = encoder.encode(jsonString).buffer; 
                this._socket?.send(arrayBuffer);
            }
            return true;
        } catch (e) {
            console.error('[WebSocket] Send error:', e);
            return false;
        }
    }

    /**
     * 发送二进制消息
     * @param data 二进制数据
     * @returns 是否发送成功
     */
    public sendBinary(data: ArrayBuffer): boolean {
        if (!this._isConnected) {
            return false;
        }
        
        try {
            this._socket?.send(data);
            return true;
        } catch (e) {
            console.error('[WebSocket] Send binary error:', e);
            return false;
        }
    }

    /**
     * 添加事件监听
     * @param type 事件类型
     * @param callback 回调函数
     * @param target 回调上下文
     */
    public on(type: string, callback: (arg?: any) => void, target?: any): void {
        this._eventTarget.on(type, callback, target);
    }

    /**
     * 移除事件监听
     * @param type 事件类型
     * @param callback 回调函数
     * @param target 回调上下文
     */
    public off(type: string, callback?: (arg?: any) => void, target?: any): void {
        this._eventTarget.off(type, callback, target);
    }

    /**
     * 一次性事件监听
     * @param type 事件类型
     * @param callback 回调函数
     * @param target 回调上下文
     */
    public once(type: string, callback: (arg?: any) => void, target?: any): void {
        this._eventTarget.once(type, callback, target);
    }

    /**
     * 关闭连接
     */
    public close(): void {
        if (this._socket) {
            this._socket.close();
            this._socket = null;
        }
        this._isConnected = false;
        this._stopHeartbeat();
        this._reconnectAttempts = this._maxReconnectAttempts; // 停止自动重连
    }

    /**
     * 获取连接状态
     * @returns 是否已连接
     */
    public isConnected(): boolean {
        return this._isConnected;
    }

    /**
     * 设置心跳配置
     * @param interval 心跳间隔(ms)
     * @param timeout 心跳超时时间(ms)
     */
    public setHeartbeatConfig(interval: number, timeout: number): void {
        this._heartbeatInterval = interval;
        this._heartbeatTimeout = timeout;
    }

    /**
     * 设置重连配置
     * @param maxAttempts 最大重试次数
     * @param interval 重试间隔(ms)
     */
    public setReconnectConfig(maxAttempts: number, interval: number): void {
        this._maxReconnectAttempts = maxAttempts;
        this._reconnectInterval = interval;
    }

    /**
     * 销毁单例实例
     */
    public static destroy(): void {
        if (this._instance) {
            this._instance.close();
            this._instance = null;
        }
    }
}