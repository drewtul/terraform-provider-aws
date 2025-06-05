// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// templateConfigurationBlock returns a schema.ListNestedBlock for template_configuration
// that can be reused across resources
func templateConfigurationBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[promptTemplateConfigurationModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"chat": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[chatPromptTemplateConfigurationModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
						listvalidator.ExactlyOneOf(
							path.MatchRelative().AtParent().AtName("chat"),
							path.MatchRelative().AtParent().AtName("text"),
						),
					},
					NestedObject: schema.NestedBlockObject{
						Blocks: map[string]schema.Block{
							"input_variable": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[promptInputVariableModel](ctx),
								Validators: []validator.List{
									listvalidator.SizeBetween(0, 20),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										names.AttrName: schema.StringAttribute{
											Required: true,
										},
									},
								},
							},
							names.AttrMessage: schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[messageModel](ctx),
								Validators: []validator.List{
									listvalidator.IsRequired(),
									listvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										names.AttrRole: schema.StringAttribute{
											CustomType: fwtypes.StringEnumType[awstypes.ConversationRole](),
											Required:   true,
										},
									},
									Blocks: map[string]schema.Block{
										names.AttrContent: schema.ListNestedBlock{
											CustomType: fwtypes.NewListNestedObjectTypeOf[contentBlockModel](ctx),
											Validators: []validator.List{
												listvalidator.SizeAtMost(1),
											},
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"text": schema.StringAttribute{
														Optional: true,
													},
												},
												Blocks: map[string]schema.Block{
													"cache_point": schema.ListNestedBlock{
														CustomType: fwtypes.NewListNestedObjectTypeOf[cachePointBlockModel](ctx),
														Validators: []validator.List{
															listvalidator.SizeAtMost(1),
															listvalidator.ExactlyOneOf(
																path.MatchRelative().AtParent().AtName("cache_point"),
																path.MatchRelative().AtParent().AtName("text"),
															),
														},
														NestedObject: schema.NestedBlockObject{
															Attributes: map[string]schema.Attribute{
																names.AttrType: schema.StringAttribute{
																	CustomType: fwtypes.StringEnumType[awstypes.CachePointType](),
																	Required:   true,
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
							"system": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[systemContentBlockModel](ctx),
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"text": schema.StringAttribute{
											Optional: true,
										},
									},
									Blocks: map[string]schema.Block{
										"cache_point": schema.ListNestedBlock{
											CustomType: fwtypes.NewListNestedObjectTypeOf[cachePointBlockModel](ctx),
											Validators: []validator.List{
												listvalidator.ExactlyOneOf(
													path.MatchRelative().AtParent().AtName("cache_point"),
													path.MatchRelative().AtParent().AtName("text"),
												),
												listvalidator.SizeAtMost(1),
											},
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													names.AttrType: schema.StringAttribute{
														CustomType: fwtypes.StringEnumType[awstypes.CachePointType](),
														Required:   true,
													},
												},
											},
										},
									},
								},
							},
							"tool_configuration": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[toolConfigurationModel](ctx),
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
								NestedObject: schema.NestedBlockObject{
									Blocks: map[string]schema.Block{
										"tool": schema.ListNestedBlock{
											CustomType: fwtypes.NewListNestedObjectTypeOf[toolModel](ctx),
											NestedObject: schema.NestedBlockObject{
												Blocks: map[string]schema.Block{
													"cache_point": schema.ListNestedBlock{
														CustomType: fwtypes.NewListNestedObjectTypeOf[cachePointBlockModel](ctx),
														Validators: []validator.List{
															listvalidator.SizeAtMost(1),
															listvalidator.ExactlyOneOf(
																path.MatchRelative().AtParent().AtName("cache_point"),
																path.MatchRelative().AtParent().AtName("tool_spec"),
															),
														},
														NestedObject: schema.NestedBlockObject{
															Attributes: map[string]schema.Attribute{
																names.AttrType: schema.StringAttribute{
																	CustomType: fwtypes.StringEnumType[awstypes.CachePointType](),
																	Required:   true,
																},
															},
														},
													},
													"tool_spec": schema.ListNestedBlock{
														CustomType: fwtypes.NewListNestedObjectTypeOf[toolSpecificationModel](ctx),
														Validators: []validator.List{
															listvalidator.SizeAtMost(1),
														},
														NestedObject: schema.NestedBlockObject{
															Attributes: map[string]schema.Attribute{
																names.AttrDescription: schema.StringAttribute{
																	Optional: true,
																},
																names.AttrName: schema.StringAttribute{
																	Required: true,
																},
															},
															Blocks: map[string]schema.Block{
																"input_schema": schema.ListNestedBlock{
																	CustomType: fwtypes.NewListNestedObjectTypeOf[toolInputSchemaModel](ctx),
																	Validators: []validator.List{
																		listvalidator.SizeAtMost(1),
																	},
																	NestedObject: schema.NestedBlockObject{
																		Attributes: map[string]schema.Attribute{
																			names.AttrJSON: schema.StringAttribute{
																				CustomType: jsontypes.NormalizedType{},
																				Optional:   true,
																				Validators: []validator.String{
																					stringvalidator.ExactlyOneOf(
																						path.MatchRelative().AtParent().AtName(names.AttrJSON),
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
											},
										},
										"tool_choice": schema.ListNestedBlock{
											CustomType: fwtypes.NewListNestedObjectTypeOf[toolChoiceModel](ctx),
											Validators: []validator.List{
												listvalidator.SizeAtMost(1),
											},
											NestedObject: schema.NestedBlockObject{
												Blocks: map[string]schema.Block{
													"any": schema.ListNestedBlock{
														CustomType: fwtypes.NewListNestedObjectTypeOf[anyToolChoiceModel](ctx),
														Validators: []validator.List{
															listvalidator.SizeAtMost(1),
															listvalidator.ExactlyOneOf(
																path.MatchRelative().AtParent().AtName("any"),
																path.MatchRelative().AtParent().AtName("auto"),
																path.MatchRelative().AtParent().AtName("tool"),
															),
														},
													},
													"auto": schema.ListNestedBlock{
														CustomType: fwtypes.NewListNestedObjectTypeOf[autoToolChoiceModel](ctx),
														Validators: []validator.List{
															listvalidator.SizeAtMost(1),
														},
													},
													"tool": schema.ListNestedBlock{
														CustomType: fwtypes.NewListNestedObjectTypeOf[specificToolChoiceModel](ctx),
														Validators: []validator.List{
															listvalidator.SizeAtMost(1),
														},
														NestedObject: schema.NestedBlockObject{
															Attributes: map[string]schema.Attribute{
																names.AttrName: schema.StringAttribute{
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
					},
				},
				"text": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[textPromptTemplateConfigurationModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"text": schema.StringAttribute{
								Required: true,
							},
						},
						Blocks: map[string]schema.Block{
							"cache_point": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[cachePointModel](ctx),
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										names.AttrType: schema.StringAttribute{
											CustomType: fwtypes.StringEnumType[awstypes.CachePointType](),
											Required:   true,
										},
									},
								},
							},
							"input_variable": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[promptInputVariableModel](ctx),
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										names.AttrName: schema.StringAttribute{
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
	}
}

// inferenceConfigurationBlock returns a schema.ListNestedBlock for inference_configuration
// that can be reused across resources
func inferenceConfigurationBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[promptInferenceConfigurationModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"text": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[promptModelInferenceConfigurationModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
						listvalidator.ExactlyOneOf(
							path.MatchRelative().AtParent().AtName("text"),
						),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"max_tokens": schema.Int32Attribute{
								Optional: true,
							},
							"stop_sequences": schema.ListAttribute{
								CustomType:  fwtypes.ListOfStringType,
								ElementType: types.StringType,
								Optional:    true,
							},
							"temperature": schema.Float32Attribute{
								Optional: true,
							},
							"top_p": schema.Float32Attribute{
								Optional: true,
							},
						},
					},
				},
			},
		},
	}
}
