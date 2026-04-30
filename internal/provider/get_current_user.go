package provider

import (
	"context"

	"github.com/pulumi/pulumi-go-provider/infer"
)

type GetCurrentUser struct{}

type GetCurrentUserArgs struct{}

type GetCurrentUserResult struct {
	UserID   int64  `pulumi:"userId"`
	Login    string `pulumi:"login"`
	FullName string `pulumi:"fullName"`
	Email    string `pulumi:"email"`
	IsAdmin  bool   `pulumi:"isAdmin"`
}

func (g *GetCurrentUser) Annotate(a infer.Annotator) {
	a.Describe(g, "Returns the authenticated Forgejo user.")
	a.SetToken("index", "getCurrentUser")
}

func (GetCurrentUser) Invoke(ctx context.Context, _ infer.FunctionRequest[GetCurrentUserArgs]) (infer.FunctionResponse[GetCurrentUserResult], error) {
	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.FunctionResponse[GetCurrentUserResult]{}, err
	}

	user, _, err := client.GetMyUserInfo()
	if err != nil {
		return infer.FunctionResponse[GetCurrentUserResult]{}, err
	}

	return infer.FunctionResponse[GetCurrentUserResult]{Output: GetCurrentUserResult{
		UserID:   user.ID,
		Login:    user.UserName,
		FullName: user.FullName,
		Email:    user.Email,
		IsAdmin:  user.IsAdmin,
	}}, nil
}
