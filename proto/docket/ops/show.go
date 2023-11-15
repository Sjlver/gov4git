package ops

import (
	"context"

	"github.com/gov4git/gov4git/proto/docket/policy"
	"github.com/gov4git/gov4git/proto/docket/schema"
	"github.com/gov4git/gov4git/proto/gov"
)

func ShowMotion(ctx context.Context, addr gov.Address, id schema.MotionID) schema.MotionView {

	return ShowMotion_Local(ctx, gov.Clone(ctx, addr), id)
}

func ShowMotion_Local(ctx context.Context, cloned gov.Cloned, id schema.MotionID) schema.MotionView {

	t := cloned.Tree()
	m := schema.MotionKV.Get(ctx, schema.MotionNS, t, id)

	p := policy.Get(ctx, m.Policy.String())
	pv := p.Show(ctx, cloned, m, policy.MotionPolicyNS(id))

	return schema.MotionView{
		Motion: m,
		Policy: pv,
	}
}
