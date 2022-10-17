package user

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gov4git/gov4git/lib/files"
	"github.com/gov4git/gov4git/lib/git"
	"github.com/gov4git/gov4git/proto/govproto"
)

type SetIn struct {
	Name            string `json:"name"`             // community unique handle for this user
	Key             string `json:"key"`              // user property key
	Value           string `json:"value"`            // user property value
	CommunityBranch string `json:"community_branch"` // branch in community repo where user will be added
}

type SetOut struct{}

func (x GovUserService) Set(ctx context.Context, in *SetIn) (*SetOut, error) {
	// clone community repo locally
	community, err := git.MakeLocalInCtx(ctx, "community")
	if err != nil {
		return nil, err
	}
	if err := community.CloneBranch(ctx, x.GovConfig.CommunityURL, in.CommunityBranch); err != nil {
		return nil, err
	}
	// make changes to repo
	if err := x.SetLocal(ctx, community, in.Name, in.Key, in.Value); err != nil {
		return nil, err
	}
	// push to origin
	if err := community.PushUpstream(ctx); err != nil {
		return nil, err
	}
	return &SetOut{}, nil
}

func (x GovUserService) SetLocal(ctx context.Context, community git.Local, name string, key string, value string) error {
	if err := x.SetLocalStageOnly(ctx, community, name, key, value); err != nil {
		return err
	}
	// commit changes
	if err := community.Commitf(ctx, "gov: change property %v of user %v", key, name); err != nil {
		return err
	}
	return nil
}

// XXX: sanitize key
// XXX: prevent overwrite
func (x GovUserService) SetLocalStageOnly(ctx context.Context, community git.Local, name string, key string, value string) error {
	propFile := filepath.Join(govproto.GovUsersDir, name, govproto.GovUserMetaDirbase, key)
	// write user file
	stage := files.ByteFiles{
		files.ByteFile{Path: propFile, Bytes: []byte(value)},
	}
	if err := community.Dir().WriteByteFiles(stage); err != nil {
		return err
	}
	// stage changes
	if err := community.Add(ctx, stage.Paths()); err != nil {
		return err
	}
	return nil
}

func (x GovUserService) SetFloat64Local(ctx context.Context, community git.Local, name string, key string, value float64) error {
	return x.SetLocal(ctx, community, name, key, fmt.Sprintf("%v", value))
}

func (x GovUserService) SetFloat64LocalStageOnly(ctx context.Context, community git.Local, name string, key string, value float64) error {
	return x.SetLocalStageOnly(ctx, community, name, key, fmt.Sprintf("%v", value))
}
