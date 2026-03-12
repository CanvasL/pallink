package logic

import (
	"context"

	"pallink/activity/activity"
	"pallink/user/userclient"
)

func hydrateActivityUsers(ctx context.Context, userRpc userclient.User, activities []*activity.ActivityInfo) error {
	if len(activities) == 0 {
		return nil
	}

	cache := make(map[uint64]*userclient.UserInfo)
	for _, item := range activities {
		if item == nil || item.CreatorId == 0 {
			continue
		}
		info, ok := cache[item.CreatorId]
		if !ok {
			resp, err := userRpc.GetUserInfo(ctx, &userclient.GetUserInfoRequest{UserId: item.CreatorId})
			if err != nil {
				return err
			}
			cache[item.CreatorId] = resp
			info = resp
		}
		if info != nil {
			item.CreatorName = info.Nickname
			item.CreatorAvatar = info.Avatar
		}
	}

	return nil
}

func hydrateParticipants(ctx context.Context, userRpc userclient.User, participants []*activity.ParticipantInfo) error {
	if len(participants) == 0 {
		return nil
	}

	cache := make(map[uint64]*userclient.UserInfo)
	for _, item := range participants {
		if item == nil || item.UserId == 0 {
			continue
		}
		info, ok := cache[item.UserId]
		if !ok {
			resp, err := userRpc.GetUserInfo(ctx, &userclient.GetUserInfoRequest{UserId: item.UserId})
			if err != nil {
				return err
			}
			cache[item.UserId] = resp
			info = resp
		}
		if info != nil {
			item.Nickname = info.Nickname
			item.Avatar = info.Avatar
		}
	}

	return nil
}

