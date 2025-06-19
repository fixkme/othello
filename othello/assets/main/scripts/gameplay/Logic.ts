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
    static readonly weights: number[][] = [
        [  8,  1,  5,  1,  1,  5,  1,  8],
        [  1, -5,  1,  1,  1,  1, -5,  1],
        [  5,  1,  5,  2,  2,  5,  1,  5],
        [  1,  1,  2,  1,  1,  2,  1,  1],
        [  1,  1,  2,  1,  1,  2,  1,  1],
        [  5,  1,  5,  2,  2,  5,  1,  5],
        [  1, -5,  1,  1,  1,  1, -5,  1],
        [  8,  1,  5,  1,  1,  5,  1,  8]
    ];

    chesses: PiecesType[][];
    previous: PiecesType[][];

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
    }

    reset() {
        for (let i = 0; i < Logic.rowSize; i++) {
            for (let j = 0; j < Logic.colSize; j++) {
                this.chesses[i][j] = PiecesType.NONE
            }
        }
        this.blackCount = 0
        this.whiteCount = 0
        this.operator = PiecesType.BLACK
    }

    addPiece(i: number, j: number, t: PiecesType) {
        //console.log(`addPiece ${t} : (${i}, ${j})`);
        this.chesses[i][j] = t
        this.addPieceCount(t, 1)
    }

    removePiece(i: number, j: number) {
        const t = this.chesses[i][j]
        this.chesses[i][j] = PiecesType.NONE
        this.addPieceCount(t, -1)
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
        const op = -t;
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

    canPlace(t: PiecesType): boolean {
        for (let i = 0; i < Logic.rowSize; i++) {
            for (let j = 0; j < Logic.colSize; j++) {
                if (this.canPlacePiece(i, j, t)){
                    return true
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
        const op = -t;
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

    getBestLocation(t: PiecesType, difficulty: number): Vec2 {
        let ret = this.getMax(t, difficulty)
        if (ret) {
            return new Vec2(ret[0], ret[1]);
        } else {
            return null
        }
    }

    getMax(t: PiecesType, deep: number): [number,number,number] {
        if (deep == 0) {
            return null
        }
        let op = -t
        let max = -65535;
        let loc: Vec2 = null;
        for (let i = 0; i < Logic.rowSize; i++) {
            for (let j = 0; j < Logic.colSize; j++) {
                if (!this.canPlaceLocation(i, j)) {
                    continue
                }
                this.addPiece(i, j, t)
                const list = this.reverse(i, j, t) 
                let num = list.length;
                if (num > 0) {
                    if (loc == null) {
                        loc = new Vec2(i, j)
                    }
                    let a = (this.blackCount+this.whiteCount)/64
                    num = a * num + (1-a) * Logic.weights[i][j] 
                    if (this.checkEnd()) {
                        num = 10000
                    } else {
                        let opRet = this.getMax(op, deep-1)
                        if (opRet) {
                            num = num - opRet[2]
                        }
                    }
                    if (num > max) {
                        max = num
                        loc.x = i
                        loc.y = j
                    }
                }
                // 还原
                this.removePiece(i, j)
                if (list.length > 0) {
                    for (let k = 0; k < list.length; k++) {
                        let item = list[k]
                        this.chesses[item.i][item.j] = op
                    }
                    this.addPieceCount(t, -list.length)
                    this.addPieceCount(op, list.length)
                }
            }
        }
        if (loc == null) {
            return null
        }
        return [loc.x, loc.y, max]
    }

    changeOperator() {
        this.operator = -this.operator;
    }

    setOperator(t: PiecesType) {
        this.operator = t;
    }

    isOperator(t: PiecesType): boolean {
        return this.operator == t;
    }

    reverse(i: number, j: number, t: PiecesType): Pieces[] {
        let list: Pieces[] = [];
        const op = -t;
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
                        list.push(new Pieces(x, y, op))
                        x -= di;
                        y -= dj;
                    }
                    break
                }
            }
        }
        if (list.length == 0) {
            return []
        }
        this.addPieceCount(t, list.length)
        this.addPieceCount(op, -list.length)
        return list
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
        for (let i = 0; i < Logic.rowSize; i++) {
            for (let j = 0; j < Logic.colSize; j++) {
                this.previous[i][j] = this.chesses[i][j];
            }
        }
    }

    undo(): {removes: Pieces[], changes: Pieces[]} {
        if (!this.previous || this.previous.length == 0) {
            return {removes: [], changes: []};
        }
        let {removes, changes} = {  removes: [], changes: []};
        this.blackCount = 0;
        this.whiteCount = 0;
        for (let i = 0; i < Logic.rowSize; i++) {
            for (let j = 0; j < Logic.colSize; j++) {
                if (this.chesses[i][j] != PiecesType.NONE && this.previous[i][j] == PiecesType.NONE) {
                    removes.push(new Pieces(i, j, this.chesses[i][j]));
                } else {
                    changes.push(new Pieces(i, j, this.previous[i][j]))
                }
                this.chesses[i][j] = this.previous[i][j];
                if (this.chesses[i][j] == PiecesType.BLACK) {
                    this.blackCount++;
                } else if (this.chesses[i][j] == PiecesType.WHITE) {
                    this.whiteCount++;
                }
            }
        }
        this.previous = [];
        return {removes, changes}
    }
     
}

