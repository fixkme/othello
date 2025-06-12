import { _decorator, Component } from 'cc';
import { GlobalWebSocket } from './WebSocket';
const { ccclass } = _decorator;

@ccclass('NetworkManager')
export class NetworkManager extends Component {
    private static _instance: NetworkManager;
    private _ws: GlobalWebSocket = GlobalWebSocket.getInstance();

    // 配置项
    private readonly WS_URL = "ws://127.0.0.1:2333/ws";
    private readonly HEARTBEAT_INTERVAL = 20000; // 20秒
    private readonly HEARTBEAT_TIMEOUT = 40000; // 40秒

    public static getInstance(): NetworkManager {
        return NetworkManager._instance;
    }

    onLoad() {
        if (NetworkManager._instance) {
            this.destroy();
            return;
        }
        
        NetworkManager._instance = this;
        
        // 设置为常驻节点
        if (!this.node._persistNode) {
            this.node._persistNode = true;
        }
        
        // 初始化WebSocket连接
        this._initializeWebSocket();
    }

    private _initializeWebSocket(): void {
        // 配置心跳
        this._ws.setHeartbeatConfig(
            this.HEARTBEAT_INTERVAL, 
            this.HEARTBEAT_TIMEOUT
        );
        
        // 配置重连
        this._ws.setReconnectConfig(10, 5000); // 最多重试10次，间隔5秒
        
        // 建立连接
        this._ws.init(this.WS_URL);
        
        // 注册事件监听
        this._registerEventHandlers();
    }

    private _registerEventHandlers(): void {
        this._ws.on('open', this._onConnected, this);
        this._ws.on('close', this._onDisconnected, this);
        this._ws.on('error', this._onError, this);
        this._ws.on('message', this._onMessage, this);
        
        // 注册游戏特定事件
        this._ws.on('player_update', this._onPlayerUpdate, this);
        this._ws.on('game_state', this._onGameStateUpdate, this);
    }

    private _unregisterEventHandlers(): void {
        this._ws.off('open', this._onConnected, this);
        this._ws.off('close', this._onDisconnected, this);
        this._ws.off('error', this._onError, this);
        this._ws.off('message', this._onMessage, this);
        
        this._ws.off('player_update', this._onPlayerUpdate, this);
        this._ws.off('game_state', this._onGameStateUpdate, this);
    }

    private _onConnected(event: Event): void {
        console.log('[Network] Connected to server');
        
        // 发送认证信息
        this._ws.send({
            cmd: 'auth',
            token: 'player-token',
            version: '1.0.0'
        });
    }

    private _onDisconnected(event: CloseEvent): void {
        console.log('[Network] Disconnected from server');
    }

    private _onError(event: Event): void {
        console.error('[Network] Error:', event);
    }

    private _onMessage(data: any): void {
        console.log('[Network] Message:', data);
    }

    private _onPlayerUpdate(data: any): void {
        console.log('[Network] Player update:', data);
        // 更新玩家状态...
    }

    private _onGameStateUpdate(data: any): void {
        console.log('[Network] Game state update:', data);
        // 更新游戏状态...
    }

    onDestroy() {
        this._unregisterEventHandlers();
        this._ws.close();
        NetworkManager._instance = null;
    }

    // 公共方法 ------------------------------------------------------

    /**
     * 发送玩家移动指令
     */
    public sendPlayerMove(x: number, y: number): void {
        this._ws.send({
            cmd: 'move',
            x: x,
            y: y
        });
    }

    /**
     * 发送准备状态
     */
    public sendReadyState(isReady: boolean): void {
        this._ws.send({
            cmd: 'ready',
            state: isReady
        });
    }

    /**
     * 获取当前连接状态
     */
    public getConnectionStatus(): boolean {
        return this._ws.isConnected();
    }
}