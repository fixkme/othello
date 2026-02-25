import { director, Node } from "cc";

export function createPopup(content: Node) {
    const scene = director.getScene();
    const canvas = scene.getChildByName('Canvas');

    const popup = new Node("Popup")
    
}

// 获取浏览器tab页面的唯一标识，刷新页面不变
export function getBrowserTabId(): string {
  const key = '__TAB_ID__';
  let id = sessionStorage.getItem(key);
  if (!id || id.length == 0) {
    // 现代浏览器可用 crypto.randomUUID()
    //id = (crypto as any).randomUUID?.() ?? (Date.now() + '-' + Math.random());
    id = getRandomAccount()
    sessionStorage.setItem(key, id);
  }
  return id;
}

// 生成随机字符串账号
export function getRandomAccount(): string {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let result = '';
    let acc_len = 16;
    for (let i = 0; i < acc_len; i++) {
        result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return "acc_" + result
}