import { _decorator, Component, Graphics, EventTouch, Node, UITransform, Vec3, Camera, v3 } from 'cc';
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

    onLoad () {
        
    }

    start() {
        this.drawPane();
        // 监听触摸结束事件
        this.node.on(Node.EventType.TOUCH_END, this.onTouchEnd, this);
    }

    update(deltaTime: number) {
        
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

    onTouchEnd(event: EventTouch) {
        let uiLoc = event.getUILocation();

        const uiPosVec3 = new Vec3(uiLoc.x, uiLoc.y, 0);
        let localPos = this.node.getComponent(UITransform).convertToNodeSpaceAR(uiPosVec3);
        console.log(`Touch local location: (${localPos.x}, ${localPos.y})`);

        let x = Math.floor(localPos.x / this.cellWidth);
        let y = Math.floor(localPos.y / this.cellHeight);
        x = (x+0.5)*this.cellWidth
        y = (y+0.5)*this.cellHeight
        console.log(`Touch location: (${x}, ${y})`);
        this.drawPieces(x, y);
    }

    drawPieces(x: number, y: number) {
        let graphics = this.node.getComponent(Graphics);
        graphics.fillColor.fromHEX('#000000');
        const radius = 40;
        // 绘制圆形路径
        graphics.arc(x, y, radius, 0, 2 * Math.PI, false);
        // 填充圆形
        graphics.fill();
    }
}

