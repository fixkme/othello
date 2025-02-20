import { _decorator, Component, UITransform, Node, EventTouch } from 'cc';
const { ccclass, property } = _decorator;

@ccclass('Test')
export class Test extends Component {
    protected onLoad(): void {
        console.log('bg loaded!');
        const canvas = this.node.getParent();
        if (canvas) {
            // 获取设计分辨率的宽度和高度
            const designWidth = canvas.getComponent(UITransform).width
            const designHeight = canvas.getComponent(UITransform).height
            console.log(`设计分辨率宽度: ${designWidth}, 高度: ${designHeight}`);
        } else {
            console.log('Canvas not found!');
        }
    }

    start() {
        //this.node.on(Node.EventType.TOUCH_END, this.onTouchEnd, this);
    }

    update(deltaTime: number) {

    }

    onTouchEnd(event: EventTouch) {
        const touchLocation = event.getLocation();
        console.log(`Touch end location: (${touchLocation.x}, ${touchLocation.y})`);
    }
}

