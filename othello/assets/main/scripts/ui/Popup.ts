import { _decorator, Button, Component, view, UITransform, director, Node} from 'cc';
const { ccclass, property } = _decorator;

@ccclass('Popup')
export class Popup extends Component {
    @property(Button)
    private Mask: Button = null;

    show(content: Node) {
        this.node.addChild(content);
        director.getScene().getChildByName('Canvas').addChild(this.node)
    }

    start() {
        this.Mask.node.getComponent(UITransform).setContentSize(view.getVisibleSize())
        this.Mask.node.on(Button.EventType.CLICK, this.onMaskClick, this);
    }

    onMaskClick() {
        this.node.destroy();
    }
}

