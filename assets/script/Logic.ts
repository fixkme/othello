import { _decorator } from 'cc';
const { ccclass, property } = _decorator;

export enum PiecesType {
    NONE = 0,
    BLACK = -1,
    WHITE = 1,
}

export class Pieces {
    i: number;
    j: number;
    t: PiecesType;
    constructor (i: number, j: number, t: PiecesType) {
        this.i = i;
        this.j = j;
        this.t = t;
    }
}

@ccclass('Logic')
export class Logic {
    static readonly rowSize: number = 8;
    static readonly colSize: number = 8;
    chesses: PiecesType[][];
    previous: PiecesType[][];
    changes: Pieces[];
    operator: boolean = false; //黑：false 白：true
    constructor( ) {
        this.chesses = [
            [0, 0, 0, 0, 0, 0, 0, 0],
            [0, 0, 0, 0, 0, 0, 0, 0],
            [0, 0, 0, 0, 0, 0, 0, 0],
            [0, 0, 0, 0, 0, 0, 0, 0],
            [0, 0, 0, 0, 0, 0, 0, 0],
            [0, 0, 0, 0, 0, 0, 0, 0],
            [0, 0, 0, 0, 0, 0, 0, 0],
            [0, 0, 0, 0, 0, 0, 0, 0]
        ];
        this.previous = [
            [0, 0, 0, 0, 0, 0, 0, 0],
            [0, 0, 0, 0, 0, 0, 0, 0],
            [0, 0, 0, 0, 0, 0, 0, 0],
            [0, 0, 0, 0, 0, 0, 0, 0],
            [0, 0, 0, 0, 0, 0, 0, 0],
            [0, 0, 0, 0, 0, 0, 0, 0],
            [0, 0, 0, 0, 0, 0, 0, 0],
            [0, 0, 0, 0, 0, 0, 0, 0]
        ];
        this.changes = [];
    }

    placePiece(i: number, j: number): boolean {
        return true;
    }

    record() {
        this.previous = [];
        for (let i = 0; i < Logic.rowSize; i++) {
            this.previous[i] = [];
            for (let j = 0; j < Logic.colSize; j++) {
                this.previous[i][j] = this.chesses[i][j];
            }
        }
    }

    undo() {
        for (let i = 0; i < Logic.rowSize; i++) {
            for (let j = 0; j < Logic.colSize; j++) {
                this.chesses[i][j] = this.previous[i][j];
            }
        }
        this.operator = !this.operator;
    }
}

