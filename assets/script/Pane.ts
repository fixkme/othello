import { _decorator, Component, Sprite, SpriteFrame, Graphics, EventTouch, Node, UITransform, Vec3, v2, Vec2} from 'cc';
import { Logic, PiecesType, Pieces } from './Logic';
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

    @property(SpriteFrame)
    private blackSprite: SpriteFrame | null = null;

    @property(SpriteFrame)
    private whiteSprite: SpriteFrame | null = null;

    onLoad () {
        // resources.load('imgs/black_piece', SpriteFrame, (err, spriteFrame: SpriteFrame) => {
        //     if (err) {
        //         console.error('Failed to load sprite frame:', err);
        //         return;
        //     }
        //     // 将加载好的 SpriteFrame 存储在类的属性中
        //     this.blackSprite = spriteFrame;
        // });
    }

    start() {
        this.initGame();
        // 监听触摸结束事件
        this.node.on(Node.EventType.TOUCH_END, this.onTouchEnd, this);
    }

    update(deltaTime: number) {
    }
    
    initGame() {
        this.drawPane();
        this.logic = new Logic();
        this.placePiece(3, 3, PiecesType.BLACK);
        this.placePiece(3, 4, PiecesType.WHITE);
        this.placePiece(4, 3, PiecesType.WHITE);
        this.placePiece(4, 4, PiecesType.BLACK);
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
    placePiece(i: number, j: number, t: PiecesType) {
        let x = (j - 4 + 0.5) * this.cellWidth
        let y = (3 - i + 0.5) * this.cellHeight
        let name = `pieces_${i}_${j}`
        let node: Node = new Node(name);
        node.setPosition(x, y);
        let sprite = node.addComponent(Sprite);
        switch (t) {
            case PiecesType.BLACK:
                sprite.spriteFrame = this.blackSprite;
                break;
            case PiecesType.WHITE:
                sprite.spriteFrame = this.whiteSprite;
                break;
            default:
                return;
        }
        let ui =  node.getComponent(UITransform)
        ui.width = 80
        ui.height = 80
        this.node.addChild(node);
        this.pieces[i][j] = node
        this.logic.addPiece(i, j, t)
    }

    onTouchEnd(event: EventTouch) {
        let loc = this.getInput(event);
        const i = loc.x
        const j = loc.y
        if (this.logic.chesses[i][j] != PiecesType.NONE) {
            return;
        }


        if (!this.logic.isOperator(PiecesType.BLACK) ||
            !this.logic.canPlacePiece(i, j, PiecesType.BLACK)) {
            return;
        }
        this.logic.changeOperator()

        if (this.putPiece(i, j, PiecesType.BLACK)) {
            return;
        }

        // 机器人
        loc = this.logic.getBestLocation(PiecesType.WHITE)
        if (loc) {
            console.log(`getBestLocation max : (${loc.x}, ${loc.y})`);
            this.scheduleOnce(function () {
                if (this.putPiece(loc.x, loc.y, PiecesType.WHITE)) {
                    return;
                }
                this.logic.changeOperator()
            }, 1.5);
        }
        
    }

    putPiece(i: number, j: number, t: PiecesType): boolean {
        this.placePiece(i, j, t);
        this.logic.reverse(i, j, t)
        this.logic.changes.forEach(element => {
            this.reversePiece(element.i, element.j, element.t);
        });
        this.logic.changes = [];
        return this.checkGameEnd();
    }

    reversePiece(i: number, j: number, t: PiecesType) {
        let node = this.pieces[i][j];
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
        let ui = node.getComponent(UITransform)
        ui.width = 80;
        ui.height = 80;
    }

    checkGameEnd(): boolean {
        if (this.logic.checkEnd()) {
            this.node.off(Node.EventType.TOUCH_END, this.onTouchEnd, this);
            this.gameEnd()
            return true;
        }
    }

    gameEnd() {
        const win = this.logic.blackCount > this.logic.whiteCount ? PiecesType.BLACK : PiecesType.WHITE;
        console.log(`winer: (${win})`);
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

    // i:row[0,7], j:col[0,7]
    drawPieces(i: number, j: number, type: PiecesType) {
        let x = (j-4+0.5)*this.cellWidth
        let y = (3-i+0.5)*this.cellHeight
        let graphics = this.node.getComponent(Graphics);
       switch (type) {
            case PiecesType.BLACK:
                graphics.fillColor.fromHEX('#000000'); break;
            case PiecesType.WHITE:
                graphics.fillColor.fromHEX('#ffffff'); break;
            default:
                graphics.fillColor.fromHEX('#C2914A7A'); break;
       }
        const radius = 32;
        // 绘制圆形路径
        graphics.arc(x, y, radius, 0, 2 * Math.PI, false);
        // 填充圆形
        graphics.fill();
    }
}

