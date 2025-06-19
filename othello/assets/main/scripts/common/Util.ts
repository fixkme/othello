import { director, Node } from "cc";

export function createPopup(content: Node) {
    const scene = director.getScene();
    const canvas = scene.getChildByName('Canvas');

    const popup = new Node("Popup")
    
}
