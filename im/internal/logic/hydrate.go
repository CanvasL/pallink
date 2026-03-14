package logic

import (
	"context"

	"pallink/im/im"
	"pallink/user/userclient"
)

func hydrateConversations(ctx context.Context, userRpc userclient.User, list []*im.ConversationInfo) error {
	if len(list) == 0 {
		return nil
	}

	cache := make(map[uint64]*userclient.UserInfo)
	for _, item := range list {
		if item == nil || item.PeerUserId == 0 {
			continue
		}
		info, ok := cache[item.PeerUserId]
		if !ok {
			resp, err := userRpc.GetUserInfo(ctx, &userclient.GetUserInfoRequest{UserId: item.PeerUserId})
			if err != nil {
				return err
			}
			cache[item.PeerUserId] = resp
			info = resp
		}
		if info != nil {
			item.PeerNickname = info.Nickname
			item.PeerAvatar = info.Avatar
		}
	}
	return nil
}
