import { _decorator, Button, Component, Label, Slider, director, Toggle} from 'cc';
import { StartGame } from './StartGame';
import { Pane } from '../gameplay/Pane';
import { GameType } from '../common/ConstValue';
const { ccclass, property } = _decorator;

@ccclass('RobotMenu')
export class RobotMenu extends Component {

    @property(Label)
    private difficultyLabel: Label = null;
    @property(Slider)
    private difficultySlider: Slider = null;

    @property(Button)
    private startButton: Button = null;

    private startGame: StartGame = null;
    private playerPiecesType: number = -1; // 默认黑色先手

    init(s : StartGame) {
        this.startGame = s
    }

    onLoad(): void {
        this.difficultySlider.progress = 0.;
        this.difficultyLabel.string = '1'
        this.difficultySlider.node.on('slide' , this.onSliderMoved.bind(this), this);
    }

    start() {
        this.startButton.node.on(Button.EventType.CLICK, this.onStartButtonClicked, this);
    }

    onSliderMoved(slider: Slider) {
        let d: number;
        const progress = slider.progress;
        if (progress <= 0.2) d = 1;
        else if (progress <= 0.4) d = 2;
        else if (progress <= 0.6) d = 3;
        else if (progress <= 0.8) d = 4;
        else d = 5;
        this.difficultyLabel.string =  d.toString();
    }

    onStartButtonClicked() {
        if (!this.startGame.robotScenePreloaded) {
            console.warn('game场景尚未预加载完成');
            return;
        }
        // 加载并运行场景
        this.startGame.targetBundle.loadScene(this.startGame.gameRobotScene, (err, scene) => {
            if (err) {
                console.error('game场景加载失败:', err);
                return;
            }
            const playerPiecesType = this.playerPiecesType;
            const difficulty = parseInt(this.difficultyLabel.string)
            director.runScene(scene, () => {
                const data = {
                    playerPieceType: playerPiecesType,
                    difficulty: difficulty,
                }
                const pane = scene.scene.getChildByPath("Canvas/Pane")
                pane.getComponent(Pane).initData(GameType.ROBOT, data)
            });
        });
    }

    onToggleContainerClick (toggle: Toggle) {
        if (toggle.isChecked) {
            if (toggle.node.name === 'Toggle1') {
                this.playerPiecesType = -1
            } else if (toggle.node.name === 'Toggle2') {
                this.playerPiecesType = 1
            }
        }
        //console.log(`触发了 ToggleContainer 事件，点了${toggle.node.name}的 Toggle`);
    }

    // onToggleClick (toggle: Toggle) {
    //     // console.log(`触发了 toggle 事件，当前 Toggle 状态为：${toggle.isChecked}`);
    // }
}

