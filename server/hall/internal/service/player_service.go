package service

import (
	"context"
	"errors"

	"github.com/fixkme/gokit/errs"
	"github.com/fixkme/gokit/framework/core"
	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/othello/server/common"
	"github.com/fixkme/othello/server/common/values"
	"github.com/fixkme/othello/server/hall/internal/system"
	"github.com/fixkme/othello/server/pb/datas"
	"github.com/fixkme/othello/server/pb/game"
	"github.com/fixkme/othello/server/pb/hall"
)

func (s *Service) Login(ctx context.Context, in *hall.CLogin) (*hall.SLogin, error) {
	mlog.Debugf("handler CLogin:%v", in)

	rc, err := getRpcContext(ctx)
	if err != nil {
		return nil, err
	}
	md := rc.ReqMd

	acct := in.Account
	if acct == "" {
		return nil, errors.New("account is empty")
	}
	p := system.Global.GetPlayerByAccount(acct)
	if p == nil {
		pinfo := &datas.PBPlayerInfo{
			Account: acct,
		}
		p, err = system.Global.CreatePlayer(acct, pinfo)
		if err != nil {
			return nil, err
		}
	}

	// 记录gateId
	p.SetGateId(md.GetStr(values.Rpc_GateId))
	// 设置session id
	replyMd := &rpc.Meta{}
	replyMd.SetInt(values.Rpc_PlayerId, p.Id())
	if v, _ := p.GetInTables(int64(datas.PlayType_PT_Common)); v != nil {
		replyMd.SetInt(values.Rpc_GameId, v.GetGameId())
	}
	rc.ReplyMd = replyMd

	resp := &hall.SLogin{
		PlayerData: p.ToPB(),
		ServerTz:   28800000,
	}
	return resp, nil
}

func (s *Service) EnterGame(ctx context.Context, in *hall.CEnterGame) (resp *hall.SEnterGame, err error) {
	var tb *datas.PBTableLocation
	defer func() {
		if err == nil {
			rc, _err := getRpcContext(ctx)
			if _err != nil {
				mlog.Errorf("EnterGame getRpcContext error %v", _err)
			} else {
				rc.ReplyMd.SetInt(values.Rpc_GameId, tb.GetGameId())
			}
		}
	}()
	resp = &hall.SEnterGame{}
	p, err := getPlayerWithContext(ctx)
	if err != nil {
		return nil, err
	}

	// 正在匹配
	tb = system.Global.GetMatchingTable(p.Id())
	if tb != nil {
		return
	}

	pt := datas.PlayType_PT_Unknown
	if mtb, _ := p.GetInTables(int64(pt)); mtb != nil {
		tb = mtb.ToPB()
		// 恢复局内
		enterReq := &game.CEnterGame{
			PlayerId: p.Id(),
			TableId:  tb.GetTableId(),
		}
		enterResp := &game.SEnterGame{}
		gameService := system.GameService(tb.GetGameId())
		meta := common.WarpMeta(p.Id(), p.GateId)
		err = core.Rpc.SyncCall(gameService, enterReq, enterResp, 0, meta)
		if err != nil {
			if cerr, ok := err.(errs.CodeError); ok && cerr.Code() == errs.ErrCode_Unknown {
				mlog.Errorf("EnterGame game node failed, pid:%d, err:%v", p.Id(), err)
				// 不在game节点
				p.RemoveInTables(int64(pt))
			} else {
				return nil, err
			}
		} else {
			resp.TableInfo = enterResp.TableInfo
			return resp, nil
		}
	}

	// 匹配成功
	tb = system.Global.GetMatchTable()
	if tb != nil {
		p1 := system.Global.GetPlayer(tb.Player1)
		p2 := p
		enterReq := &game.CCreateRoom{
			TableId: tb.GetTableId(),
			Players: []*datas.PBPlayerInfo{p1.GetModelPlayerInfo().ToPB(), p2.GetModelPlayerInfo().ToPB()},
			GateIds: map[int64]string{p1.Id(): p1.GateId, p2.Id(): p2.GateId},
		}
		enterResp := &game.SCreateRoom{}
		gameService := system.GameService(tb.GetGameId())
		err = core.Rpc.SyncCall(gameService, enterReq, enterResp, 0)
		if err != nil {
			return nil, err
		}
		system.Global.PlayerMatchSucceed(p1, p2, tb)
		//resp.TableInfo = enterResp.TableInfo
		return resp, nil
	}
	// 创建
	tb, err = system.Global.CreateMatchTable(p)
	if err != nil {
		return nil, err
	}

	//resp.TableInfo = enterResp.TableInfo
	return resp, nil
}

func (s *Service) LeaveGame(ctx context.Context, in *hall.CLeaveGame) (*hall.SLeaveGame, error) {
	p, err := getPlayerWithContext(ctx)
	if err != nil {
		return nil, err
	}
	err = system.Global.PlayerLeaveGame(p)
	return &hall.SLeaveGame{}, err
}

func (s *Service) PlayerOffline(ctx context.Context, in *hall.CPlayerOffline) (*hall.SPlayerOffline, error) {
	p := system.Global.GetPlayer(in.PlayerId)
	if p != nil {
		system.Global.RemoveMatchPlayer(p)
	}
	return nil, nil
}
