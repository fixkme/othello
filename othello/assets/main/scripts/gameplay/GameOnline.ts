import { _decorator, Component, Sprite, SpriteFrame, Graphics, 
    EventTouch, Node, UITransform, Vec3, v2, Vec2, Label, Button,
    Prefab, instantiate, view, Animation, AudioSource, assetManager, director} from 'cc';
import { Logic, PiecesType, Pieces } from './Logic';
import { GlobalWebSocket } from '../common/WebSocket';
import { CEnterGame, CLeaveGame, CPlacePiece, PGameResult, PPlacePiece,
    PPlayerEnterGame, SEnterGame } from '../pb/game/player';
import { PlayerInfo, TableInfo } from '../pb/datas/player_data';
import { NetworkManager } from '../common/NetworkManager';
import { PkgNames, SecneName } from '../common/ConstValue';
const { ccclass, property } = _decorator;


@ccclass('GameOnline')
export class GameOnline extends Component {
    // 格子的行数
    readonly rows:  number = 8;
    // 格子的列数
    readonly cols: number = 8;
    // 每个格子的宽度
    cellWidth: number = 80;
    // 每个格子的高度
    cellHeight: number = 80;
    
    logic: Logic;
    pieces: Node[][]= [
        [null,null,null,null,null,null,null,null],
        [null,null,null,null,null,null,null,null],
        [null,null,null,null,null,null,null,null],
        [null,null,null,null,null,null,null,null],
        [null,null,null,null,null,null,null,null],
        [null,null,null,null,null,null,null,null],
        [null,null,null,null,null,null,null,null],
        [null,null,null,null,null,null,null,null],
    ];

    private selfPieceType: PiecesType = PiecesType.BLACK; //玩家自己的
    private opponentPieceType: PiecesType = PiecesType.WHITE; //对手的
    private selfPlayer: PlayerInfo = null; //自己信息
    private oppoPlayer: PlayerInfo = null; //对手信息

    @property(SpriteFrame)
    private blackSprite: SpriteFrame | null = null;
    @property(SpriteFrame)
    private whiteSprite: SpriteFrame | null = null;

    @property(Label)
    private blackScore: Label | null = null;
    @property(Label)
    private whiteScore: Label | null = null;

    @property(Label)
    private blackPlayerName: Label | null = null;
    @property(Label)
    private whitePlayerName: Label | null = null;

    @property(Button)
    private buttonRenew: Button = null!;
    @property(Button)
    private buttonRetract: Button = null!;
    @property(Button)
    private buttonReturn: Button = null!;

    @property(Prefab)
    private prefabPiece: Prefab = null!;
    @property(Prefab)
    private prefabEndWidget: Prefab = null!;

    private _ws: GlobalWebSocket = null;

    onLoad () {
        this._ws = GlobalWebSocket.getInstance();
    }

    start() {
        this.drawPane();
        // 注册消息事件
        // 自己进入房间
        this._ws.on(SEnterGame.$type, this.onSelfEnterGame, this)
        // 其他玩家进入
        this._ws.on(PPlayerEnterGame.$type, this.onOtherEnterGame, this)
        // 棋子落下
        this._ws.on(PPlacePiece.$type, this.onPlacePieceMsg, this)
        // 游戏结束
        this._ws.on(PGameResult.$type, this.onGameResult, this)

        // 发送进入游戏的消息
        if (this._ws.isConnected()) {
            const cEnterGame = CEnterGame.create();
            if (this._ws.sendMessage(CEnterGame, cEnterGame)) {
                console.log("发送进入游戏消息成功");
            }
        } else {
            console.warn("WebSocket not connected.");
        }

        this.node.on(Node.EventType.TOUCH_END, this.onTouchEnd, this);
        this.buttonRenew.node.on(Button.EventType.CLICK, this.onButtonRenewClick, this);
        this.buttonRetract.node.on(Button.EventType.CLICK, this.onButtonRetractClick, this);
        this.buttonReturn.node.on(Button.EventType.CLICK, this.onButtonReturnClick, this);
    }

    onDestroy() {
        this._ws.off(SEnterGame.$type, this.onSelfEnterGame, this)
        this._ws.off(PPlayerEnterGame.$type, this.onOtherEnterGame, this)
        this._ws.off(PPlacePiece.$type, this.onPlacePieceMsg, this)
        this._ws.off(PGameResult.$type, this.onGameResult, this)
    }

    update(deltaTime: number) {
        
    }

    initData(gameType: number, data: any) {
        
    }
    
    initGame(tb: TableInfo) {
        if (!tb) {
            console.log("initGame: table info is null");
            return;
        }
        this.selfPlayer = NetworkManager.getInstance().playerInfo
        console.log("initGame: selfPlayer: ", this.selfPlayer);
        console.log("initGame: selfPlayer: ", this.selfPlayer.id, tb.ownerPlayer.id);
        
        if (this.selfPlayer.id == tb.ownerPlayer.id) {
            this.selfPieceType = tb.ownerPlayer.playPieceType;
            if (tb.oppoPlayer) {
                this.oppoPlayer = tb.oppoPlayer;
                this.opponentPieceType = tb.oppoPlayer.playPieceType;
            }
        } else {
            this.selfPieceType = tb.oppoPlayer.playPieceType;
            this.oppoPlayer = tb.ownerPlayer;
            this.opponentPieceType = tb.ownerPlayer.playPieceType;
        }
        // 展示玩家信息
        this.showPlayerInfo();

        this.logic = new Logic();
        this.logic.setOperator(tb.turn);
        tb.pieces.forEach(p => {
            this.placePiece(p.x, p.y, p.color, false);
        });
    }

    showPlayerInfo() {
        if (this.selfPieceType === PiecesType.BLACK) {
            this.blackPlayerName.string = this.selfPlayer.name
            if (this.oppoPlayer) {
                this.whitePlayerName.string = this.oppoPlayer.name
            }
        } else if (this.selfPieceType === PiecesType.WHITE) {
            this.whitePlayerName.string = this.selfPlayer.name
            if (this.oppoPlayer) {
                this.blackPlayerName.string = this.oppoPlayer.name
            }
        }
    }

    drawPane() {
        // 获取节点的宽度和高度
        const uiTransform = this.node.getComponent(UITransform);
        const cwidth = uiTransform.width;
        const cheight = uiTransform.height;
        console.log(`pane size: (${cwidth}, ${cheight})`);

        this.cellWidth = cwidth / this.cols;
        this.cellHeight = cheight / this.rows;

        let graphics = this.node.getComponent(Graphics);
        // 绘制矩形
        graphics.rect(-cwidth / 2, -cheight / 2, cwidth, cheight);
        // 填充矩形
        graphics.fill();
        graphics.stroke();
        // 绘制水平线条
        for (let i = 1; i < this.rows; i++) {
            let y = i * this.cellHeight-cheight/2;
            graphics.moveTo(-cwidth/2, y);
            graphics.lineTo(cwidth/2, y);
        } 
        // 绘制垂直线条
        for (let j = 1; j < this.cols; j++) {
            let x = j * this.cellWidth-cwidth/2;
            graphics.moveTo(x, -cheight/2);
            graphics.lineTo(x, cheight/2);
        }
        // 绘制线条
        graphics.stroke();
    }

    // i:row[0,7], j:col[0,7]
    placePiece(i: number, j: number, t: PiecesType, sound: boolean) {
        let x = (j - 4 + 0.5) * this.cellWidth
        let y = (3 - i + 0.5) * this.cellHeight
        let node: Node = instantiate(this.prefabPiece);
        
        node.setPosition(x, y);
        switch (t) {
            case PiecesType.BLACK:
                node.getComponent(Sprite).spriteFrame = this.blackSprite;
                break;
            case PiecesType.WHITE:
                node.getComponent(Sprite).spriteFrame = this.whiteSprite;
                break;
            default:
                node.destroy();
                return;
        }
        this.node.addChild(node);
        this.pieces[i][j] = node
        if (sound) {
            let audio =  node.getComponent(AudioSource)
            audio.play();
        }
        this.logic.addPiece(i, j, t)
        this.blackScore.string = this.logic.blackCount.toString()
        this.whiteScore.string = this.logic.whiteCount.toString()
    }

    onTouchEnd(event: EventTouch) {
        let loc = this.getInput(event);
        const i = loc.x
        const j = loc.y

        const selfPieceType = this.selfPieceType
        if (!this.logic.isOperator(selfPieceType) ||
            !this.logic.canPlacePiece(i, j, selfPieceType)) {
            return;
        }
        // 发送落子请求
        const cplacePiece = CPlacePiece.create();
        cplacePiece.x = i;
        cplacePiece.y = j;
        cplacePiece.pieceType = selfPieceType;
        this._ws.sendMessage(CPlacePiece, cplacePiece);
    }

    onSelfEnterGame(data: any) {
        const senterGame = data as SEnterGame
        this.initGame(senterGame.tableInfo); 
    }

    onOtherEnterGame(data: any) {
        const pmsg = data as PPlayerEnterGame;
        this.oppoPlayer = pmsg.playerInfo;
        this.opponentPieceType = pmsg.playerInfo.playPieceType;
        this.showPlayerInfo();
    }

    onPlacePieceMsg(data: any) {
        const pmsg = data as PPlacePiece;
        this.putPiece(pmsg.x, pmsg.y, pmsg.pieceType);
        this.logic.setOperator(pmsg.operatePiece);
    }

    onGameResult(data: any) {
        const pmsg = data as PGameResult;
        console.log("game result:", pmsg);
        //this.checkGameEnd();
        this.node.off(Node.EventType.TOUCH_END, this.onTouchEnd, this);
        
        let t = pmsg.winnerPieceType;
        let winStr = "";
        if (pmsg.isGiveUp) {
            winStr = t == PiecesType.BLACK ? "白棋认输" : "黑棋认输";
        } else {
            winStr = t == PiecesType.BLACK ? "黑方胜" : (t == PiecesType.WHITE ? "白方胜" : "平局")
        }
        this.gameEnd(winStr)
    }

    putPiece(i: number, j: number, t: PiecesType) {
        this.placePiece(i, j, t, true);
        const list = this.logic.reverse(i, j, t)
        list.forEach(element => {
            this.reversePiece(element.i, element.j, element.t);
        });
        this.blackScore.string = this.logic.blackCount.toString()
        this.whiteScore.string = this.logic.whiteCount.toString() 
    }

    reversePiece(i: number, j: number, t: PiecesType) {
        const node = this.pieces[i][j];
        switch (t) {
            case PiecesType.BLACK:
                node.getComponent(Animation).play("reverse_black_piece");
                break;
            case PiecesType.WHITE:
                node.getComponent(Animation).play("reverse_white_piece");
                break;
            default:
                return;
        }
    }

    changePieceSprite(i: number, j: number, t: PiecesType) {
        const node = this.pieces[i][j];
        switch (t) {
            case PiecesType.BLACK:
                node.getComponent(Sprite).spriteFrame = this.blackSprite;
                break;
            case PiecesType.WHITE:
                node.getComponent(Sprite).spriteFrame = this.whiteSprite;
                break;
            default:
                return;
        }
    }

    checkGameEnd(): boolean {
        if (this.logic.checkEnd()) {
            this.node.off(Node.EventType.TOUCH_END, this.onTouchEnd, this);
            const winStr = this.logic.blackCount > this.logic.whiteCount ? "黑方胜" : "白方胜";
            this.gameEnd(winStr)
            return true;
        }
    }

    gameEnd(winStr: string) {
        console.log(`winer: (${winStr})`);
        const endWidget = instantiate(this.prefabEndWidget)
        endWidget.name = "endWidget";
        endWidget.position.set(0, 120);
        endWidget.getChildByName("Mask").getComponent(UITransform).setContentSize(view.getVisibleSize())
        endWidget.getChildByName("Label").getComponent(Label).string = winStr;
        endWidget.getChildByName("Button").getComponent(Button).node.on(Button.EventType.CLICK, () => {
            endWidget.destroy();
            this.returnStartScene()
        }, this);
        this.node.parent.addChild(endWidget);
    }

    getInput(event: EventTouch): Vec2 {
        let uiLoc = event.getUILocation();
        const uiPosVec3 = new Vec3(uiLoc.x, uiLoc.y, 0);
        let localPos = this.node.getComponent(UITransform).convertToNodeSpaceAR(uiPosVec3);
        //console.log(`Touch local location: (${localPos.x}, ${localPos.y})`);
        let x = Math.floor(localPos.x / this.cellWidth);
        let y = Math.floor(localPos.y / this.cellHeight); 
        let i = 3-y
        let j = 4+x
        // console.log(`Touch index location: (${x}, ${y})`);
        // console.log(`place index location: (${i}, ${j})`);
        return v2(i, j);
    }

    // 重来
    onButtonRenewClick() {
        this.unscheduleAllCallbacks();
        for (let i = 0; i < Logic.rowSize; i++) {
            for (let j = 0; j < Logic.colSize; j++) {
                if (this.pieces[i][j]) {
                    this.pieces[i][j].destroy();
                    this.pieces[i][j] = null;
                }
            }
        }
        this.logic.reset();
        this.placePiece(3, 3, PiecesType.BLACK, false);
        this.placePiece(3, 4, PiecesType.WHITE, false);
        this.placePiece(4, 3, PiecesType.WHITE, false);
        this.placePiece(4, 4, PiecesType.BLACK, false);
        this.node.on(Node.EventType.TOUCH_END, this.onTouchEnd, this);
    }

    // 悔棋
    onButtonRetractClick() {
        if (this.logic.checkEnd()) {
            return;
        }
        this.unscheduleAllCallbacks();
        const {removes, changes} = this.logic.undo();
        removes.forEach(item => {
            let node = this.pieces[item.i][item.j]
            if (node) {
                node.destroy();
            }
        });
        changes.forEach(item => {
            this.changePieceSprite(item.i, item.j, item.t);
        });
        this.blackScore.string = this.logic.blackCount.toString()
        this.whiteScore.string = this.logic.whiteCount.toString()
        this.logic.setOperator(PiecesType.BLACK)
    }

    // 返回
    onButtonReturnClick() {
        this.returnStartScene()
    }

    returnStartScene() {
        director.loadScene(SecneName.Start, (err, scene) => {
            if (err) {
                console.error('start场景加载失败:', err);
                return;
            }
        })
    }

}

