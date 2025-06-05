// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
)

// Exports for use in tests only.
var (
	ResourceAgent                         = newAgentResource
	ResourceAgentActionGroup              = newAgentActionGroupResource
	ResourceAgentAlias                    = newAgentAliasResource
	ResourceAgentCollaborator             = newAgentCollaboratorResource
	ResourceAgentKnowledgeBaseAssociation = newAgentKnowledgeBaseAssociationResource
	ResourceDataSource                    = newDataSourceResource
	ResourceFlow                          = newResourceFlow
	ResourceKnowledgeBase                 = newKnowledgeBaseResource
	ResourcePrompt                        = newPromptResource

	FindAgentByID                                  = findAgentByID
	FindAgentActionGroupByThreePartKey             = findAgentActionGroupByThreePartKey
	FindAgentAliasByTwoPartKey                     = findAgentAliasByTwoPartKey
	FindAgentCollaboratorByThreePartKey            = findAgentCollaboratorByThreePartKey
	FindAgentKnowledgeBaseAssociationByThreePartID = findAgentKnowledgeBaseAssociationByThreePartKey
	FindDataSourceByTwoPartKey                     = findDataSourceByTwoPartKey
	FindFlowByID                                   = findFlowByID
	FindKnowledgeBaseByID                          = findKnowledgeBaseByID
	FindPromptByID                                 = findPromptByID
	
	StatusFlow      = statusFlow
	WaitFlowCreated = waitFlowCreated
	WaitFlowUpdated = waitFlowUpdated
	WaitFlowDeleted = waitFlowDeleted
)
