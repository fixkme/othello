import { _decorator, Vec2 } from 'cc';
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
    static readonly dirs: number[][] = [
        [1, 0], [-1, 0], //左右
        [0, -1], [0, 1], //上下
        [-1, -1], [1, 1], //左斜
        [1, -1], [-1, 1], //右斜
    ];
    chesses: PiecesType[][];
    previous: PiecesType[][];
    changes: Pieces[];

    blackCount: number = 0;
    whiteCount: number = 0;
    operator: PiecesType = PiecesType.BLACK;
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

    addPiece(i: number, j: number, t: PiecesType) {
        console.log(`addPiece ${t} : (${i}, ${j})`);
        this.chesses[i][j] = t
        this.addPieceCount(t, 1)
    }

    addPieceCount(t: PiecesType, add: number) {
        if (t == PiecesType.BLACK) {
            this.blackCount += add;
        } else if (t == PiecesType.WHITE) {
            this.whiteCount += add;
        }
    }

    canPlacePiece(i: number, j: number, t: PiecesType): boolean {
        if (!this.canPlaceLocation(i, j)) {
            return false;
        }
        const op = t == PiecesType.BLACK ? PiecesType.WHITE : PiecesType.BLACK;
        for (let d = 0; d < Logic.dirs.length; d++) {
            let di = Logic.dirs[d][0];
            let dj = Logic.dirs[d][1];
            let x = i + di
            let y = j + dj;
            while (this.canLocate(x, y) && this.chesses[x][y] != PiecesType.NONE) {
                if (this.chesses[x][y] == op) {
                    x += di;
                    y += dj;
                } else {
                    if (x == i+di && y == j+dj) {
                        break;
                    } else {
                        return true;
                    }
                }
            }
        }
        return false;
    }

    getCanPlacePieceCount(i: number, j: number, t: PiecesType): number {
        if (!this.canPlaceLocation(i, j)) {
            return 0;
        }
        let total = 0;
        const op = t == PiecesType.BLACK ? PiecesType.WHITE : PiecesType.BLACK;
        for (let d = 0; d < Logic.dirs.length; d++) {
            let di = Logic.dirs[d][0];
            let dj = Logic.dirs[d][1];
            let x = i + di;
            let y = j + dj;
            let getNum = 0;
            while (this.canLocate(x, y) && this.chesses[x][y] != PiecesType.NONE) {
                if (this.chesses[x][y] == op) {
                    getNum++;
                    x += di;
                    y += dj;
                } else {
                    if (!(x == i+di && y == j+dj)) {
                        total += getNum;
                    }
                    break;
                }
            }
        }
        return total;
    }

    changeOperator() {
        this.operator = this.operator == PiecesType.BLACK ? PiecesType.WHITE : PiecesType.BLACK;
    }

    isOperator(t: PiecesType): boolean {
        return this.operator == t;
    }

    reverse(i: number, j: number, t: PiecesType) {
        this.changes = [];
        const op = t == PiecesType.BLACK ? PiecesType.WHITE : PiecesType.BLACK;
        for (let d = 0; d < Logic.dirs.length; d++) {
            let di = Logic.dirs[d][0];
            let dj = Logic.dirs[d][1];
            let x = i + di
            let y = j + dj;
            while (this.canLocate(x,y) && this.chesses[x][y] != PiecesType.NONE) {
                if (this.chesses[x][y] == op) {
                    x += di;
                    y += dj;
                } else {
                    x -= di;
                    y -= dj;
                    while(!(x == i && y == j)) {
                        this.chesses[x][y] = t;
                        this.changes.push(new Pieces(x, y, t))
                        x -= di;
                        y -= dj;
                    }
                    break
                }
            }
        }
        this.addPieceCount(t, this.changes.length)
        this.addPieceCount(op, -this.changes.length)
    }

    getBestLocation(t: PiecesType): Vec2 {
        let maxCount = 0;
        let loc = new Vec2();
        for (let i = 0; i < Logic.rowSize; i++) {
            for (let j = 0; j < Logic.colSize; j++) {
                let num = this.getCanPlacePieceCount(i, j, t)
                if (num > maxCount) {
                    maxCount = num
                    loc.x = i
                    loc.y = j
                    //console.log(`max ${maxCount}: (${loc.x}, ${loc.y})`);
                }
            }
        }
        if (maxCount > 0) {
            return loc
        } else {
            return null
        }
    }

    checkEnd(): boolean {
        return this.blackCount == 0 || this.whiteCount == 0 || this.blackCount + this.whiteCount==64
    }

    canPlaceLocation(i: number, j: number): boolean {
        return this.canLocate(i,j) && this.chesses[i][j] == PiecesType.NONE;
    }

    canLocate(i: number, j: number): boolean {
        return 0 <= i && i < Logic.rowSize && 0 <= j && j < Logic.colSize
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
        this.changeOperator();
    }
     
}

