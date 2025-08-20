import { _decorator, Component, Sprite, SpriteFrame, Graphics, 
    EventTouch, Node, UITransform, Vec3, v2, Vec2, Label, Button,
    Prefab, instantiate, view, Animation,
    AudioSource} from 'cc';
import { Logic, PiecesType, Pieces } from './Logic';
import { GameType } from '../common/ConstValue';
const { ccclass, property } = _decorator;


@ccclass('Pane')
export class Pane extends Component {
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
    
    private gameType: GameType = GameType.ROBOT; 
    private selfPieceType: PiecesType = PiecesType.BLACK; //玩家自己的
    private opponentPieceType: PiecesType = PiecesType.WHITE; //对手的
    // 人机难度
    private robotDifficulty: number = 1;
    private robotContinue = false;

    @property(SpriteFrame)
    private blackSprite: SpriteFrame | null = null;
    @property(SpriteFrame)
    private whiteSprite: SpriteFrame | null = null;

    @property(Label)
    private blackScore: Label | null = null;
    @property(Label)
    private whiteScore: Label | null = null;

    @property(Button)
    private buttonRenew: Button = null!;
    @property(Button)
    private buttonRetract: Button = null!;

    @property(Prefab)
    private prefabPiece: Prefab = null!;
    @property(Prefab)
    private prefabEndWidget: Prefab = null!;

    onLoad () {
    }

    start() {
        this.initGame();
        // 绑定事件
        this.node.on(Node.EventType.TOUCH_END, this.onTouchEnd, this);
        this.buttonRenew.node.on(Button.EventType.CLICK, this.onButtonRenewClick, this);
        this.buttonRetract.node.on(Button.EventType.CLICK, this.onButtonRetractClick, this);
    }

    update(deltaTime: number) {
        if (this.gameType == GameType.ROBOT && this.robotContinue) {
            this.processRobot();
        }
    }

    initData(gameType: number, data: any) {
        // console.log("gameType:", gameType, data);
        this.gameType = gameType;
        switch (gameType) {
            case GameType.ROBOT:
                this.selfPieceType = data.playerPieceType;
                this.robotDifficulty = data.difficulty;
                break;
        }
        
    }
    
    initGame() {
        this.drawPane();
        this.logic = new Logic();
        this.placePiece(3, 3, PiecesType.BLACK, false);
        this.placePiece(3, 4, PiecesType.WHITE, false);
        this.placePiece(4, 3, PiecesType.WHITE, false);
        this.placePiece(4, 4, PiecesType.BLACK, false);
        //this.node.getComponent(AudioSource).play();
        switch(this.gameType) {
            case GameType.ROBOT:
                //console.log("selfPieceType:", this.selfPieceType, this.robotDifficulty);
                if (this.selfPieceType == PiecesType.WHITE) {
                    this.scheduleOnce(function () {
                        this.robotContinue = true;
                    }, 3);
                }
                break;
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
        this.logic.changeOperator()

        this.logic.record();
        if (this.putPiece(i, j, selfPieceType)) {
            return;
        }

        // 机器人
        this.processRobot();
    }

    processRobot() {
        const difficulty = this.robotDifficulty;
        const robotPieceType = -this.selfPieceType;
        const op = -robotPieceType;
        let loc = this.logic.getBestLocation(robotPieceType, difficulty)
        if (loc) {
            console.log(`getBestLocation max : (${loc.x}, ${loc.y})`);
            this.scheduleOnce(function () {
                if (this.putPiece(loc.x, loc.y, robotPieceType)) {
                    return;
                }
                this.logic.changeOperator()
                if (!this.logic.canPlace(op)) {
                    this.logic.changeOperator()
                    this.robotContinue = true;
                } else {
                    this.robotContinue = false;
                }
            }, 0.5);
        } else {
            console.log(`getBestLocation max : null`);
            if (this.logic.canPlace(op)) {
                this.logic.changeOperator()
            } else {
                this.gameEnd();
            }
        }
    }

    putPiece(i: number, j: number, t: PiecesType): boolean {
        this.placePiece(i, j, t, true);
        const list = this.logic.reverse(i, j, t)
        list.forEach(element => {
            this.reversePiece(element.i, element.j, element.t);
        });
        this.blackScore.string = this.logic.blackCount.toString()
        this.whiteScore.string = this.logic.whiteCount.toString()
        return this.checkGameEnd();
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
            this.gameEnd()
            return true;
        }
    }

    gameEnd() {
        const winStr = this.logic.blackCount > this.logic.whiteCount ? "黑方胜" : (this.logic.blackCount < this.logic.whiteCount ? "白方胜" : "平局");
        console.log(`winer: (${winStr})`);
        const endWidget = instantiate(this.prefabEndWidget)
        endWidget.name = "endWidget";
        endWidget.position.set(0, 120);
        endWidget.getChildByName("Mask").getComponent(UITransform).setContentSize(view.getVisibleSize())
        endWidget.getChildByName("Label").getComponent(Label).string = winStr;
        endWidget.getChildByName("Button").getComponent(Button).node.on(Button.EventType.CLICK, () => {
            endWidget.destroy();
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
        this.robotContinue = false;
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
        this.robotContinue = false;
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

}

