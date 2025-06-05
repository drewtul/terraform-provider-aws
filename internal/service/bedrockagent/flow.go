// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_bedrockagent_flow", name="Flow")
// @Tags(identifierAttribute="arn")
func newResourceFlow(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceFlow{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameFlow = "Flow"
)

type resourceFlow struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceFlow) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"execution_role_arn": schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.ARNType,
			},
			"customer_encryption_key_arn": schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.ARNType,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"definition": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[flowDefinitionModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"nodes": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[flowNodeModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Required: true,
									},
									"type": schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.FlowNodeType](),
									},
								},
								Blocks: map[string]schema.Block{
									"configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[flowNodeConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"prompt": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[promptFlowNodeConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"source_configuration": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[promptFlowNodeSourceConfigurationModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Blocks: map[string]schema.Block{
																		"inline": schema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[promptFlowNodeInlineConfigurationModel](ctx),
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"model_id": schema.StringAttribute{
																						Required: true,
																					},
																					"template_type": schema.StringAttribute{
																						Required:   true,
																						CustomType: fwtypes.StringEnumType[awstypes.PromptTemplateType](),
																					},
																					"additional_model_request_fields": schema.StringAttribute{
																						Optional: true,
																					},
																				},
																				Blocks: map[string]schema.Block{
																					"template_configuration":  templateConfigurationBlock(ctx),
																					"inference_configuration": inferenceConfigurationBlock(ctx),
																				},
																			},
																		},
																		"resource": schema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[promptFlowNodeResourceConfigurationModel](ctx),
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"prompt_arn": schema.StringAttribute{
																						Required: true,
																					},
																				},
																			},
																		},
																	},
																	Validators: []validator.Object{
																		objectvalidator.ExactlyOneOf(
																			path.MatchRelative().AtParent().AtName("inline"),
																			path.MatchRelative().AtParent().AtName("resource"),
																		),
																	},
																},
															},
															"guardrail_configuration": schema.ListNestedBlock{
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"guardrail_identifier": schema.StringAttribute{
																			Required: true,
																		},
																		"guardrail_version": schema.StringAttribute{
																			Optional: true,
																		},
																	},
																},
															},
														},
													},
												},
												"lambda": schema.ListNestedBlock{
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"lambda_arn": schema.StringAttribute{
																Required: true,
															},
														},
													},
												},
												"knowledge_base": schema.ListNestedBlock{
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"knowledge_base_id": schema.StringAttribute{
																Required: true,
															},
															"model_id": schema.StringAttribute{
																Optional: true,
															},
															"number_of_results": schema.Int64Attribute{
																Optional: true,
															},
														},
														Blocks: map[string]schema.Block{
															"guardrail_configuration": schema.ListNestedBlock{
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"guardrail_identifier": schema.StringAttribute{
																			Required: true,
																		},
																		"guardrail_version": schema.StringAttribute{
																			Optional: true,
																		},
																	},
																},
															},
															"inference_configuration": schema.ListNestedBlock{
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Blocks: map[string]schema.Block{
																		"text": schema.ListNestedBlock{
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"max_tokens": schema.Int64Attribute{
																						Optional: true,
																					},
																					"stop_sequences": schema.ListAttribute{
																						ElementType: types.StringType,
																						Optional:    true,
																					},
																					"temperature": schema.Float64Attribute{
																						Optional: true,
																					},
																					"top_p": schema.Float64Attribute{
																						Optional: true,
																					},
																				},
																			},
																		},
																	},
																},
															},
															"orchestration_configuration": schema.ListNestedBlock{
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"additional_model_request_fields": schema.StringAttribute{
																			Optional: true,
																		},
																	},
																	Blocks: map[string]schema.Block{
																		"inference_config": schema.ListNestedBlock{
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Blocks: map[string]schema.Block{
																					"text": schema.ListNestedBlock{
																						Validators: []validator.List{
																							listvalidator.SizeAtMost(1),
																						},
																						NestedObject: schema.NestedBlockObject{
																							Attributes: map[string]schema.Attribute{
																								"max_tokens": schema.Int64Attribute{
																									Optional: true,
																								},
																								"stop_sequences": schema.ListAttribute{
																									ElementType: types.StringType,
																									Optional:    true,
																								},
																								"temperature": schema.Float64Attribute{
																									Optional: true,
																								},
																								"top_p": schema.Float64Attribute{
																									Optional: true,
																								},
																							},
																						},
																					},
																				},
																			},
																		},
																		"performance_config": schema.ListNestedBlock{
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"latency": schema.StringAttribute{
																						Required:   true,
																						CustomType: fwtypes.StringEnumType[awstypes.PerformanceConfigLatency](),
																					},
																				},
																			},
																		},
																		"prompt_template": schema.ListNestedBlock{
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"text_prompt_template": schema.StringAttribute{
																						Required: true,
																					},
																				},
																			},
																		},
																	},
																},
															},
															"prompt_template": schema.ListNestedBlock{
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"text_prompt_template": schema.StringAttribute{
																			Required: true,
																		},
																	},
																},
															},
															"reranking_configuration": schema.ListNestedBlock{
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"type": schema.StringAttribute{
																			Required:   true,
																			CustomType: fwtypes.StringEnumType[awstypes.VectorSearchRerankingConfigurationType](),
																		},
																	},
																	Blocks: map[string]schema.Block{
																		"bedrock_reranking_configuration": schema.ListNestedBlock{
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"number_of_reranked_results": schema.Int64Attribute{
																						Optional: true,
																					},
																				},
																				Blocks: map[string]schema.Block{
																					"model_configuration": schema.ListNestedBlock{
																						Validators: []validator.List{
																							listvalidator.SizeAtMost(1),
																						},
																						NestedObject: schema.NestedBlockObject{
																							// Add attributes based on VectorSearchBedrockRerankingModelConfiguration
																						},
																					},
																					"metadata_configuration": schema.ListNestedBlock{
																						Validators: []validator.List{
																							listvalidator.SizeAtMost(1),
																						},
																						NestedObject: schema.NestedBlockObject{
																							// Add attributes based on MetadataConfigurationForReranking
																						},
																					},
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
												"condition": schema.ListNestedBlock{
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"conditions": schema.ListNestedBlock{
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"name": schema.StringAttribute{
																			Required: true,
																		},
																		"expression": schema.StringAttribute{
																			Optional: true,
																		},
																	},
																},
															},
														},
													},
												},
												"agent": schema.ListNestedBlock{
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"agent_alias_arn": schema.StringAttribute{
																Required: true,
															},
														},
													},
												},
												"input": schema.ListNestedBlock{
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														// InputFlowNodeConfiguration has no fields in the API
													},
												},
												"output": schema.ListNestedBlock{
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														// OutputFlowNodeConfiguration has no fields in the API
													},
												},
												"retrieval": schema.ListNestedBlock{
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"service_configuration": schema.ListNestedBlock{
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Blocks: map[string]schema.Block{
																		"s3": schema.ListNestedBlock{
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"bucket_name": schema.StringAttribute{
																						Required: true,
																					},
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
											Validators: []validator.Object{
												objectvalidator.ExactlyOneOf(
													path.MatchRelative().AtParent().AtName("prompt"),
													path.MatchRelative().AtParent().AtName("lambda"),
													path.MatchRelative().AtParent().AtName("knowledge_base"),
													path.MatchRelative().AtParent().AtName("condition"),
													path.MatchRelative().AtParent().AtName("agent"),
													path.MatchRelative().AtParent().AtName("input"),
													path.MatchRelative().AtParent().AtName("output"),
													path.MatchRelative().AtParent().AtName("retrieval"),
												),
											},
										},
									},
									"inputs": schema.ListNestedBlock{
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"name": schema.StringAttribute{
													Required: true,
												},
												"type": schema.StringAttribute{
													Required:   true,
													CustomType: fwtypes.StringEnumType[awstypes.FlowNodeIODataType](),
												},
												"category": schema.StringAttribute{
													Optional:   true,
													CustomType: fwtypes.StringEnumType[awstypes.FlowNodeInputCategory](),
												},
												"expression": schema.StringAttribute{
													Optional: true,
												},
											},
										},
									},
									"outputs": schema.ListNestedBlock{
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"name": schema.StringAttribute{
													Required: true,
												},
												"type": schema.StringAttribute{
													Required:   true,
													CustomType: fwtypes.StringEnumType[awstypes.FlowNodeIODataType](),
												},
											},
										},
									},
								},
							},
						},
						"connections": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Required: true,
									},
									"source": schema.StringAttribute{
										Required: true,
									},
									"target": schema.StringAttribute{
										Required: true,
									},
									"type": schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.FlowConnectionType](),
									},
								},
								Blocks: map[string]schema.Block{
									"configuration": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"conditional": schema.ListNestedBlock{
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"condition": schema.StringAttribute{
																Required: true,
															},
														},
													},
												},
												"data": schema.ListNestedBlock{
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"source_output": schema.StringAttribute{
																Required: true,
															},
															"target_input": schema.StringAttribute{
																Required: true,
															},
														},
													},
												},
											},
											Validators: []validator.Object{
												objectvalidator.ExactlyOneOf(
													path.MatchRelative().AtParent().AtName("conditional"),
													path.MatchRelative().AtParent().AtName("data"),
												),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceFlow) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var plan resourceFlowModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input bedrockagent.CreateFlowInput

	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateFlow(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionCreating, ResNameFlow, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionCreating, ResNameFlow, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitFlowCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionWaitingForCreation, ResNameFlow, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceFlow) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var state resourceFlowModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findFlowByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionReading, ResNameFlow, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceFlow) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var plan, state resourceFlowModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input bedrockagent.UpdateFlowInput
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateFlow(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionUpdating, ResNameFlow, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionUpdating, ResNameFlow, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitFlowUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionWaitingForUpdate, ResNameFlow, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceFlow) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var state resourceFlowModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrockagent.DeleteFlowInput{
		FlowIdentifier: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteFlow(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionDeleting, ResNameFlow, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitFlowDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionWaitingForDeletion, ResNameFlow, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceFlow) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

const (
	statusChangePending = "PENDING"
	statusDeleting      = "DELETING"
	statusNormal        = "ACTIVE"
	statusUpdated       = "ACTIVE"
)

func waitFlowCreated(ctx context.Context, conn *bedrockagent.Client, id string, timeout time.Duration) (*bedrockagent.GetFlowOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(statusNormal),
		Refresh:                   statusFlow(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagent.GetFlowOutput); ok {
		return out, err
	}

	return nil, err
}

func waitFlowUpdated(ctx context.Context, conn *bedrockagent.Client, id string, timeout time.Duration) (*bedrockagent.GetFlowOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(statusChangePending),
		Target:                    enum.Slice(statusUpdated),
		Refresh:                   statusFlow(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagent.GetFlowOutput); ok {
		return out, err
	}

	return nil, err
}

func waitFlowDeleted(ctx context.Context, conn *bedrockagent.Client, id string, timeout time.Duration) (*bedrockagent.GetFlowOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(statusDeleting, statusNormal),
		Target:  []string{},
		Refresh: statusFlow(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagent.GetFlowOutput); ok {
		return out, err
	}

	return nil, err
}

func statusFlow(ctx context.Context, conn *bedrockagent.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findFlowByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findFlowByID(ctx context.Context, conn *bedrockagent.Client, id string) (*bedrockagent.GetFlowOutput, error) {
	input := bedrockagent.GetFlowInput{
		FlowIdentifier: aws.String(id),
	}

	out, err := conn.GetFlow(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out, nil
}

type resourceFlowModel struct {
	ARN                      types.String                                         `tfsdk:"arn"`
	Definition               fwtypes.ListNestedObjectValueOf[flowDefinitionModel] `tfsdk:"definition"`
	Description              types.String                                         `tfsdk:"description"`
	ID                       types.String                                         `tfsdk:"id"`
	Name                     types.String                                         `tfsdk:"name"`
	ExecutionRoleArn         fwtypes.ARN                                          `tfsdk:"execution_role_arn"`
	CustomerEncryptionKeyArn fwtypes.ARN                                          `tfsdk:"customer_encryption_key_arn"`
	Tags                     tftags.Map                                           `tfsdk:"tags"`
	TagsAll                  tftags.Map                                           `tfsdk:"tags_all"`
	Timeouts                 timeouts.Value                                       `tfsdk:"timeouts"`
}

type flowDefinitionModel struct {
	Nodes       fwtypes.ListNestedObjectValueOf[flowNodeModel]       `tfsdk:"nodes"`
	Connections fwtypes.ListNestedObjectValueOf[flowConnectionModel] `tfsdk:"connections"`
}

type flowNodeModel struct {
	Name          types.String                                                `tfsdk:"name"`
	Type          fwtypes.StringEnum[awstypes.FlowNodeType]                   `tfsdk:"type"`
	Configuration fwtypes.ListNestedObjectValueOf[flowNodeConfigurationModel] `tfsdk:"configuration"`
	Inputs        fwtypes.ListNestedObjectValueOf[flowNodeInputModel]         `tfsdk:"inputs"`
	Outputs       fwtypes.ListNestedObjectValueOf[flowNodeOutputModel]        `tfsdk:"outputs"`
}

// Models for FlowNodeConfiguration implementations
type promptFlowNodeConfigurationModel struct {
	SourceConfiguration    fwtypes.ListNestedObjectValueOf[promptFlowNodeSourceConfigurationModel] `tfsdk:"source_configuration"`
	GuardrailConfiguration fwtypes.ListNestedObjectValueOf[flowGuardrailConfigurationModel]        `tfsdk:"guardrail_configuration"`
}

// Models for PromptFlowNodeSourceConfiguration implementations
type promptFlowNodeInlineConfigurationModel struct {
	ModelId                      types.String                                                       `tfsdk:"model_id"`
	TemplateConfiguration        fwtypes.ListNestedObjectValueOf[promptTemplateConfigurationModel]  `tfsdk:"template_configuration"`
	TemplateType                 fwtypes.StringEnum[awstypes.PromptTemplateType]                    `tfsdk:"template_type"`
	AdditionalModelRequestFields types.String                                                       `tfsdk:"additional_model_request_fields"`
	InferenceConfiguration       fwtypes.ListNestedObjectValueOf[promptInferenceConfigurationModel] `tfsdk:"inference_configuration"`
}

type promptFlowNodeResourceConfigurationModel struct {
	PromptArn types.String `tfsdk:"prompt_arn"`
}

type promptFlowNodeSourceConfigurationModel struct {
	Inline   fwtypes.ListNestedObjectValueOf[promptFlowNodeInlineConfigurationModel]   `tfsdk:"inline"`
	Resource fwtypes.ListNestedObjectValueOf[promptFlowNodeResourceConfigurationModel] `tfsdk:"resource"`
}

var (
	_ fwflex.Expander  = promptFlowNodeSourceConfigurationModel{}
	_ fwflex.Flattener = &promptFlowNodeSourceConfigurationModel{}
)

func (m promptFlowNodeSourceConfigurationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.Inline.IsNull():
		inlineData, d := m.Inline.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.PromptFlowNodeSourceConfigurationMemberInline
		diags.Append(fwflex.Expand(ctx, inlineData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags

	case !m.Resource.IsNull():
		resourceData, d := m.Resource.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.PromptFlowNodeSourceConfigurationMemberResource
		diags.Append(fwflex.Expand(ctx, resourceData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

func (m *promptFlowNodeSourceConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.PromptFlowNodeSourceConfigurationMemberInline:
		var model promptFlowNodeInlineConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Inline = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	case awstypes.PromptFlowNodeSourceConfigurationMemberResource:
		var model promptFlowNodeResourceConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Resource = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	default:
		diags.AddError("Unsupported configuration type", fmt.Sprintf("Unsupported configuration type: %T", v))
		return diags
	}
}

type flowGuardrailConfigurationModel struct {
	GuardrailIdentifier types.String `tfsdk:"guardrail_identifier"`
	GuardrailVersion    types.String `tfsdk:"guardrail_version"`
}

type lambdaFunctionFlowNodeConfigurationModel struct {
	LambdaArn types.String `tfsdk:"lambda_arn"`
}

type knowledgeBaseFlowNodeConfigurationModel struct {
	KnowledgeBaseId            types.String                                                           `tfsdk:"knowledge_base_id"`
	GuardrailConfiguration     fwtypes.ListNestedObjectValueOf[flowGuardrailConfigurationModel]       `tfsdk:"guardrail_configuration"`
	InferenceConfiguration     fwtypes.ListNestedObjectValueOf[promptInferenceConfigurationModel]     `tfsdk:"inference_configuration"`
	ModelId                    types.String                                                           `tfsdk:"model_id"`
	NumberOfResults            types.Int64                                                            `tfsdk:"number_of_results"`
	OrchestrationConfiguration fwtypes.ListNestedObjectValueOf[knowledgeBaseOrchestrationConfigModel] `tfsdk:"orchestration_configuration"`
	PromptTemplate             fwtypes.ListNestedObjectValueOf[knowledgeBasePromptTemplateModel]      `tfsdk:"prompt_template"`
	RerankingConfiguration     fwtypes.ListNestedObjectValueOf[vectorSearchRerankingConfigModel]      `tfsdk:"reranking_configuration"`
}

// Models for PromptInferenceConfiguration implementations are defined in prompt.go

type knowledgeBaseOrchestrationConfigModel struct {
	AdditionalModelRequestFields types.String                                                       `tfsdk:"additional_model_request_fields"`
	InferenceConfig              fwtypes.ListNestedObjectValueOf[promptInferenceConfigurationModel] `tfsdk:"inference_config"`
	PerformanceConfig            fwtypes.ListNestedObjectValueOf[performanceConfigurationModel]     `tfsdk:"performance_config"`
	PromptTemplate               fwtypes.ListNestedObjectValueOf[knowledgeBasePromptTemplateModel]  `tfsdk:"prompt_template"`
}

type performanceConfigurationModel struct {
	Latency fwtypes.StringEnum[awstypes.PerformanceConfigLatency] `tfsdk:"latency"`
}

type knowledgeBasePromptTemplateModel struct {
	TextPromptTemplate types.String `tfsdk:"text_prompt_template"`
}

type vectorSearchRerankingConfigModel struct {
	Type                          fwtypes.StringEnum[awstypes.VectorSearchRerankingConfigurationType]      `tfsdk:"type"`
	BedrockRerankingConfiguration fwtypes.ListNestedObjectValueOf[vectorSearchBedrockRerankingConfigModel] `tfsdk:"bedrock_reranking_configuration"`
}

type vectorSearchBedrockRerankingConfigModel struct {
	ModelConfiguration      fwtypes.ListNestedObjectValueOf[vectorSearchBedrockRerankingModelConfigModel] `tfsdk:"model_configuration"`
	MetadataConfiguration   fwtypes.ListNestedObjectValueOf[metadataConfigurationForRerankingModel]       `tfsdk:"metadata_configuration"`
	NumberOfRerankedResults types.Int64                                                                   `tfsdk:"number_of_reranked_results"`
}

type vectorSearchBedrockRerankingModelConfigModel struct {
	// Add fields based on VectorSearchBedrockRerankingModelConfiguration
}

type metadataConfigurationForRerankingModel struct {
	// Add fields based on MetadataConfigurationForReranking
}

type conditionFlowNodeConfigurationModel struct {
	Conditions fwtypes.ListNestedObjectValueOf[flowConditionModel] `tfsdk:"conditions"`
}

type flowConditionModel struct {
	Name       types.String `tfsdk:"name"`
	Expression types.String `tfsdk:"expression"`
}

type agentFlowNodeConfigurationModel struct {
	AgentAliasArn types.String `tfsdk:"agent_alias_arn"`
}

type inputFlowNodeConfigurationModel struct {
	// InputFlowNodeConfiguration has no fields in the API
}

type outputFlowNodeConfigurationModel struct {
	// OutputFlowNodeConfiguration has no fields in the API
}

type retrievalFlowNodeConfigurationModel struct {
	ServiceConfiguration fwtypes.ListNestedObjectValueOf[retrievalFlowNodeServiceConfigurationModel] `tfsdk:"service_configuration"`
}

type retrievalFlowNodeServiceConfigurationModel struct {
	S3 fwtypes.ListNestedObjectValueOf[retrievalFlowNodeS3ConfigurationModel] `tfsdk:"s3"`
}

var (
	_ fwflex.Expander  = retrievalFlowNodeServiceConfigurationModel{}
	_ fwflex.Flattener = &retrievalFlowNodeServiceConfigurationModel{}
)

func (m retrievalFlowNodeServiceConfigurationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.S3.IsNull():
		s3Data, d := m.S3.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.RetrievalFlowNodeServiceConfigurationMemberS3
		diags.Append(fwflex.Expand(ctx, s3Data, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

func (m *retrievalFlowNodeServiceConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.RetrievalFlowNodeServiceConfigurationMemberS3:
		var model retrievalFlowNodeS3ConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.S3 = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	default:
		diags.AddError("Unsupported configuration type", fmt.Sprintf("Unsupported configuration type: %T", v))
		return diags
	}
}

type retrievalFlowNodeS3ConfigurationModel struct {
	BucketName types.String `tfsdk:"bucket_name"`
}

// Container model for FlowNodeConfiguration
type flowNodeConfigurationModel struct {
	Prompt        fwtypes.ListNestedObjectValueOf[promptFlowNodeConfigurationModel]         `tfsdk:"prompt"`
	Lambda        fwtypes.ListNestedObjectValueOf[lambdaFunctionFlowNodeConfigurationModel] `tfsdk:"lambda"`
	KnowledgeBase fwtypes.ListNestedObjectValueOf[knowledgeBaseFlowNodeConfigurationModel]  `tfsdk:"knowledge_base"`
	Condition     fwtypes.ListNestedObjectValueOf[conditionFlowNodeConfigurationModel]      `tfsdk:"condition"`
	Agent         fwtypes.ListNestedObjectValueOf[agentFlowNodeConfigurationModel]          `tfsdk:"agent"`
	Input         fwtypes.ListNestedObjectValueOf[inputFlowNodeConfigurationModel]          `tfsdk:"input"`
	Output        fwtypes.ListNestedObjectValueOf[outputFlowNodeConfigurationModel]         `tfsdk:"output"`
	Retrieval     fwtypes.ListNestedObjectValueOf[retrievalFlowNodeConfigurationModel]      `tfsdk:"retrieval"`
}

var (
	_ fwflex.Expander  = flowNodeConfigurationModel{}
	_ fwflex.Flattener = &flowNodeConfigurationModel{}
)

func (m flowNodeConfigurationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.Prompt.IsNull():
		promptData, d := m.Prompt.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowNodeConfigurationMemberPrompt
		diags.Append(fwflex.Expand(ctx, promptData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags

	case !m.Lambda.IsNull():
		lambdaData, d := m.Lambda.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowNodeConfigurationMemberLambdaFunction
		diags.Append(fwflex.Expand(ctx, lambdaData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags

	case !m.KnowledgeBase.IsNull():
		kbData, d := m.KnowledgeBase.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowNodeConfigurationMemberKnowledgeBase
		diags.Append(fwflex.Expand(ctx, kbData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags

	case !m.Condition.IsNull():
		conditionData, d := m.Condition.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowNodeConfigurationMemberCondition
		diags.Append(fwflex.Expand(ctx, conditionData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags

	case !m.Agent.IsNull():
		agentData, d := m.Agent.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowNodeConfigurationMemberAgent
		diags.Append(fwflex.Expand(ctx, agentData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags

	case !m.Input.IsNull():
		inputData, d := m.Input.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowNodeConfigurationMemberInput
		diags.Append(fwflex.Expand(ctx, inputData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags

	case !m.Output.IsNull():
		outputData, d := m.Output.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowNodeConfigurationMemberOutput
		diags.Append(fwflex.Expand(ctx, outputData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags

	case !m.Retrieval.IsNull():
		retrievalData, d := m.Retrieval.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowNodeConfigurationMemberRetrieval
		diags.Append(fwflex.Expand(ctx, retrievalData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

func (m *flowNodeConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.FlowNodeConfigurationMemberPrompt:
		var model promptFlowNodeConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Prompt = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	case awstypes.FlowNodeConfigurationMemberLambdaFunction:
		var model lambdaFunctionFlowNodeConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Lambda = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	case awstypes.FlowNodeConfigurationMemberKnowledgeBase:
		var model knowledgeBaseFlowNodeConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.KnowledgeBase = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	case awstypes.FlowNodeConfigurationMemberCondition:
		var model conditionFlowNodeConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Condition = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	case awstypes.FlowNodeConfigurationMemberAgent:
		var model agentFlowNodeConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Agent = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	case awstypes.FlowNodeConfigurationMemberInput:
		var model inputFlowNodeConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Input = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	case awstypes.FlowNodeConfigurationMemberOutput:
		var model outputFlowNodeConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Output = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	case awstypes.FlowNodeConfigurationMemberRetrieval:
		var model retrievalFlowNodeConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Retrieval = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	default:
		diags.AddError("Unsupported configuration type", fmt.Sprintf("Unsupported configuration type: %T", v))
		return diags
	}
}

type flowNodeInputModel struct {
	Name       types.String                                       `tfsdk:"name"`
	Type       fwtypes.StringEnum[awstypes.FlowNodeIODataType]    `tfsdk:"type"`
	Category   fwtypes.StringEnum[awstypes.FlowNodeInputCategory] `tfsdk:"category"`
	Expression types.String                                       `tfsdk:"expression"`
}

type flowNodeOutputModel struct {
	Name types.String                                    `tfsdk:"name"`
	Type fwtypes.StringEnum[awstypes.FlowNodeIODataType] `tfsdk:"type"`
}

type flowConnectionModel struct {
	Name          types.String                                                      `tfsdk:"name"`
	Source        types.String                                                      `tfsdk:"source"`
	Target        types.String                                                      `tfsdk:"target"`
	Type          fwtypes.StringEnum[awstypes.FlowConnectionType]                   `tfsdk:"type"`
	Configuration fwtypes.ListNestedObjectValueOf[flowConnectionConfigurationModel] `tfsdk:"configuration"`
}

// Models for FlowConnectionConfiguration implementations
type flowConditionalConnectionConfigurationModel struct {
	Condition types.String `tfsdk:"condition"`
}

type flowDataConnectionConfigurationModel struct {
	SourceOutput types.String `tfsdk:"source_output"`
	TargetInput  types.String `tfsdk:"target_input"`
}

// Container model for FlowConnectionConfiguration
type flowConnectionConfigurationModel struct {
	Conditional fwtypes.ListNestedObjectValueOf[flowConditionalConnectionConfigurationModel] `tfsdk:"conditional"`
	Data        fwtypes.ListNestedObjectValueOf[flowDataConnectionConfigurationModel]        `tfsdk:"data"`
}

var (
	_ fwflex.Expander  = flowConnectionConfigurationModel{}
	_ fwflex.Flattener = &flowConnectionConfigurationModel{}
)

func (m flowConnectionConfigurationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.Conditional.IsNull():
		conditionalData, d := m.Conditional.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowConnectionConfigurationMemberConditional
		diags.Append(fwflex.Expand(ctx, conditionalData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags

	case !m.Data.IsNull():
		dataData, d := m.Data.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowConnectionConfigurationMemberData
		diags.Append(fwflex.Expand(ctx, dataData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

func (m *flowConnectionConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.FlowConnectionConfigurationMemberConditional:
		var model flowConditionalConnectionConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Conditional = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	case awstypes.FlowConnectionConfigurationMemberData:
		var model flowDataConnectionConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Data = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	default:
		diags.AddError("Unsupported configuration type", fmt.Sprintf("Unsupported configuration type: %T", v))
		return diags
	}
}
