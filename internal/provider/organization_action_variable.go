package provider

import (
	"context"
	"fmt"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

type OrganizationActionVariable struct{}

type OrganizationActionVariableArgs struct {
	ActionVariableArgs

	Organization string `pulumi:"organization"`
}

type OrganizationActionVariableState struct {
	OrganizationActionVariableArgs
	OwnerID int64 `pulumi:"ownerId"`
}

func (v *OrganizationActionVariable) Annotate(a infer.Annotator) {
	a.Describe(v, "A Forgejo Actions variable for an organization.")
	a.SetToken("index", "OrganizationActionVariable")
}

func (a *OrganizationActionVariableArgs) Annotate(ann infer.Annotator) {
	annotateActionVariableArgs(&a.ActionVariableArgs, ann)
	ann.Describe(&a.Organization, "Organization name.")
}

func (OrganizationActionVariable) Create(ctx context.Context, req infer.CreateRequest[OrganizationActionVariableArgs]) (infer.CreateResponse[OrganizationActionVariableState], error) {
	if req.DryRun {
		return infer.CreateResponse[OrganizationActionVariableState]{ID: organizationActionVariableID(req.Inputs.Organization, req.Inputs.Name), Output: organizationActionVariableStateFromArgs(req.Inputs, 0)}, nil
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.CreateResponse[OrganizationActionVariableState]{}, err
	}

	_, err = client.CreateOrgActionVariable(req.Inputs.Organization, forgejo.CreateVariableOption{Name: req.Inputs.Name, Data: req.Inputs.Value})
	if err != nil {
		return infer.CreateResponse[OrganizationActionVariableState]{}, err
	}
	variable, _, err := client.GetOrgActionVariable(req.Inputs.Organization, req.Inputs.Name)
	if err != nil {
		return infer.CreateResponse[OrganizationActionVariableState]{}, err
	}

	state := organizationActionVariableStateFromAPI(req.Inputs.Organization, variable)
	return infer.CreateResponse[OrganizationActionVariableState]{ID: organizationActionVariableID(state.Organization, state.Name), Output: state}, nil
}

func (OrganizationActionVariable) Read(ctx context.Context, req infer.ReadRequest[OrganizationActionVariableArgs, OrganizationActionVariableState]) (infer.ReadResponse[OrganizationActionVariableArgs, OrganizationActionVariableState], error) {
	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.ReadResponse[OrganizationActionVariableArgs, OrganizationActionVariableState]{}, err
	}

	variable, resp, err := client.GetOrgActionVariable(req.State.Organization, req.State.Name)
	if isNotFound(resp) {
		return infer.ReadResponse[OrganizationActionVariableArgs, OrganizationActionVariableState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[OrganizationActionVariableArgs, OrganizationActionVariableState]{}, err
	}

	state := organizationActionVariableStateFromAPI(req.State.Organization, variable)
	return infer.ReadResponse[OrganizationActionVariableArgs, OrganizationActionVariableState]{ID: organizationActionVariableID(state.Organization, state.Name), Inputs: state.OrganizationActionVariableArgs, State: state}, nil
}

func (OrganizationActionVariable) Diff(_ context.Context, req infer.DiffRequest[OrganizationActionVariableArgs, OrganizationActionVariableState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	addReplaceDiff(diff, "organization", req.State.Organization != req.Inputs.Organization)
	addReplaceDiff(diff, "name", req.State.Name != req.Inputs.Name)
	addUpdateDiff(diff, "value", req.State.Value != req.Inputs.Value)
	return p.DiffResponse{HasChanges: len(diff) > 0, DeleteBeforeReplace: true, DetailedDiff: diff}, nil
}

func (OrganizationActionVariable) Update(ctx context.Context, req infer.UpdateRequest[OrganizationActionVariableArgs, OrganizationActionVariableState]) (infer.UpdateResponse[OrganizationActionVariableState], error) {
	if req.DryRun {
		return infer.UpdateResponse[OrganizationActionVariableState]{Output: organizationActionVariableStateFromArgs(req.Inputs, req.State.OwnerID)}, nil
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.UpdateResponse[OrganizationActionVariableState]{}, err
	}
	_, err = client.UpdateOrgActionVariable(req.State.Organization, req.State.Name, forgejo.CreateVariableOption{Name: req.Inputs.Name, Data: req.Inputs.Value})
	if err != nil {
		return infer.UpdateResponse[OrganizationActionVariableState]{}, err
	}
	variable, _, err := client.GetOrgActionVariable(req.Inputs.Organization, req.Inputs.Name)
	if err != nil {
		return infer.UpdateResponse[OrganizationActionVariableState]{}, err
	}

	return infer.UpdateResponse[OrganizationActionVariableState]{Output: organizationActionVariableStateFromAPI(req.Inputs.Organization, variable)}, nil
}

func (OrganizationActionVariable) Delete(ctx context.Context, req infer.DeleteRequest[OrganizationActionVariableState]) (infer.DeleteResponse, error) {
	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.DeleteResponse{}, err
	}

	resp, err := client.DeleteOrgActionVariable(req.State.Organization, req.State.Name)
	if isNotFound(resp) {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, err
}

func organizationActionVariableStateFromAPI(org string, variable *forgejo.ActionVariable) OrganizationActionVariableState {
	return organizationActionVariableStateFromArgs(OrganizationActionVariableArgs{ActionVariableArgs: ActionVariableArgs{Name: variable.Name, Value: variable.Data}, Organization: org}, variable.OwnerID)
}

func organizationActionVariableStateFromArgs(args OrganizationActionVariableArgs, ownerID int64) OrganizationActionVariableState {
	return OrganizationActionVariableState{OrganizationActionVariableArgs: args, OwnerID: ownerID}
}

func organizationActionVariableID(org, name string) string {
	return fmt.Sprintf("%s/%s", org, name)
}
