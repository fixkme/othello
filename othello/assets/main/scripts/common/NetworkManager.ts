import { _decorator, Component } from 'cc';
import { GlobalWebSocket } from './WebSocket';
import { CLogin } from '../pb/game/player';
import { PlayerInfo } from '../pb/datas/player_data';
const { ccclass } = _decorator;

@ccclass('NetworkManager')
export class NetworkManager extends Component {
    private static _instance: NetworkManager;
    private _ws: GlobalWebSocket = GlobalWebSocket.getInstance();

    private readonly WS_URL = "ws://127.0.0.1:7070/ws";
    private account: string = "";
    public playerInfo: PlayerInfo = null;
    public isLogined: boolean = false;

    public static getInstance(): NetworkManager {
        return NetworkManager._instance;
    }

    onLoad() {
        if (NetworkManager._instance) {
            this.destroy();
            return;
        }

        NetworkManager._instance = this;

        // 测试代码， 随机生成账号
        const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
        let result = '';
        for (let i = 0; i < length; i++) {
            result += chars.charAt(Math.floor(Math.random() * chars.length));
        }
        this.account = "acc_"  + result

        // 设置为常驻节点
        if (!this.node._persistNode) {
            this.node._persistNode = true;
        }

        // 初始化WebSocket连接
        this._initializeWebSocket();
    }

    private _initializeWebSocket(): void {
        // 配置重连
        this._ws.setReconnectConfig(3, 5000); // 最多重试3次，间隔5秒

        // 建立连接
        const url = this.WS_URL + "?x-account=" + this.account; 
        this._ws.init(url);

        // 注册事件监听
        this._registerEventHandlers();
    }

    private _registerEventHandlers(): void {
        this._ws.on('open', this._onConnected, this);
        this._ws.on('close', this._onDisconnected, this);
        this._ws.on('error', this._onError, this);
        this._ws.on('message', this._onUnknownMessage, this);

        // 注册游戏特定事件
        this._ws.on('game.SLogin', this._onLoginResponse, this);
    }

    private _unregisterEventHandlers(): void {
        this._ws.off('open', this._onConnected, this);
        this._ws.off('close', this._onDisconnected, this);
        this._ws.off('error', this._onError, this);
        this._ws.off('message', this._onUnknownMessage, this);
    }

    private _onConnected(event: Event): void {
        console.log('[Network] Connected to server');

        // 登录
        const clogin = CLogin.create()
        clogin.account = this.account
        let ok = GlobalWebSocket.getInstance().sendMessage(CLogin, clogin)
        if (ok) {
            console.log('[Network] succeed to send clogin message');
        } else {
            console.log('[Network] failed to send clogin message');
        }
    }

    private _onDisconnected(event: CloseEvent): void {
        console.log('[Network] Disconnected from server');
    }

    private _onError(event: Event): void {
        console.error('[Network] Error:', event);
    }

    private _onUnknownMessage(data: any): void {
        console.warn('[Network] unknown Message:', data);
    }

    private _onLoginResponse(data: any): void {
        const v = data as PlayerInfo
        this.playerInfo = v
        this.isLogined = true
        console.log('[Network] Login succeed, playerInfo:', v);
    }

    onDestroy() {
        this._unregisterEventHandlers();
        this._ws.close();
        NetworkManager._instance = null;
    }

    /**
     * 获取当前连接状态
     */
    public getConnectionStatus(): boolean {
        return this._ws.isConnected();
    }
}