import { _decorator, Component, Prefab, director, Button, assetManager, AssetManager, instantiate } from 'cc';
import { RobotMenu } from './RobotMenu';
import { GameType, PkgNames, SecneName } from '../common/ConstValue';
import { Popup } from './Popup';
import { GameOnline } from '../gameplay/GameOnline';
import { NetworkManager } from '../common/NetworkManager';
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
    public robotScenePreloaded: boolean = false;
    public onlineScenePreloaded: boolean = false;
    public readonly gameRobotScene: string = SecneName.GameRobot;
    public readonly gameOnlineScene: string = SecneName.GameOnline;


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
            bundle.preloadScene(this.gameRobotScene, 
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
                    this.robotScenePreloaded = true;
                    console.log('game场景预加载完成');
                }
            );
            // 预加载场景
            bundle.preloadScene(this.gameOnlineScene, 
                (completed, total) => {
                    // 可以在这里更新加载进度条
                    // const progress = (completed / total) * 100;
                    // console.log(`预加载进度: ${progress.toFixed(2)}%`);
                },
                (err) => {
                    if (err) {
                        console.error('game联机场景预加载失败:', err);
                        return;
                    }
                    this.onlineScenePreloaded = true;
                    console.log('game联机场景预加载完成');
                }
            );
        })
    }
    
    start() {
        this.btnStartRobot.node.on(Button.EventType.CLICK, this.onBtnStartRobotClick, this);
        this.btnStartPlayer.node.on(Button.EventType.CLICK, this.onBtnStartPlayerClick, this);
        this.btnStartFriend.node.on(Button.EventType.CLICK, this.onBtnStartFriendClick, this);
    }

    // 人机对战
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

    // 联机对战
    onBtnStartPlayerClick() {
        //是否已经登录
        if (!NetworkManager.getInstance().isLogined) {
            console.log("未登录不能进入联机玩法");
            return;
        }
        //进入联机场景
        if (!this.onlineScenePreloaded) {
            console.warn('game联机场景尚未预加载完成');
            return;
        }
        // 加载并运行场景
        this.targetBundle.loadScene(this.gameOnlineScene, (err, scene) => {
            if (err) {
                console.error('game联机场景加载失败:', err);
                return;
            }
            director.runScene(scene, () => {
                const pane = scene.scene.getChildByPath("Canvas/Pane")
                pane.getComponent(GameOnline).initData(GameType.PLAYER, null)
            });
        });
    }

    // 好友对战
    onBtnStartFriendClick() {

    }
}

