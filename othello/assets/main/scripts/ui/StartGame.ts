import { _decorator, Component, Prefab, director, Button, assetManager, AssetManager, instantiate } from 'cc';
import { RobotMenu } from './RobotMenu';
import { PkgNames } from '../common/ConstValue';
import { Popup } from './Popup';
const { ccclass, property } = _decorator;

@ccclass('StartGame')
export class StartGame extends Component {

    @property(Button)
    private btnStartRobot: Button = null;
    @property(Button)
    private btnStartPlayer: Button = null;
    @property(Button)
    private btnStartFriend: Button = null;

    @property(Prefab)
    private robotMenuPrefab: Prefab = null;

    public targetBundle: AssetManager.Bundle = null;
    public isScenePreloaded: boolean = false;
    public readonly gameScene: string = "game_robot";

    protected onLoad(): void {
        assetManager.loadBundle(PkgNames.Common, (err, bundle) => { 
            if (err) {
                console.error(PkgNames.Common, 'Bundle加载失败:', err); 
            }
        });
        assetManager.loadBundle(PkgNames.Game, (err, bundle) => { 
            if (err) {
                console.error(PkgNames.Game, 'Bundle加载失败:', err);
                return;
            }
            this.targetBundle = bundle;
            // 预加载场景
            bundle.preloadScene(this.gameScene, 
                (completed, total) => {
                    // 可以在这里更新加载进度条
                    // const progress = (completed / total) * 100;
                    // console.log(`预加载进度: ${progress.toFixed(2)}%`);
                },
                (err) => {
                    if (err) {
                        console.error('game场景预加载失败:', err);
                        return;
                    }
                    this.isScenePreloaded = true;
                    console.log('game场景预加载完成');
                }
            );
        })
    }
    
    start() {
        this.btnStartRobot.node.on(Button.EventType.CLICK, this.onBtnStartRobotClick, this);
    }

    onBtnStartRobotClick() {
        const bundle = assetManager.getBundle(PkgNames.Common)
        bundle.load("prefabs/Popup", (err, asset) => {
            if (err) {
                console.error('Popup 加载Prefab失败:', err);
                return;
            }
            const popup = instantiate(asset as Prefab)
            const menu = instantiate(this.robotMenuPrefab)
            menu.getComponent(RobotMenu).init(this)
            popup.getComponent(Popup).show(menu)
        })
    }
}

