---
subcategory: "Bedrock Agents"
layout: "aws"
page_title: "AWS: aws_bedrockagent_flow"
description: |-
  Terraform resource for managing an AWS Bedrock Agents Flow.
---
# Resource: aws_bedrockagent_flow

Terraform resource for managing an AWS Bedrock Agents Flow. Amazon Bedrock Flows enable you to create multi-step workflows that orchestrate interactions between foundation models (FMs), agent actions, and knowledge bases.

## Example Usage

### Basic Usage

```terraform
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["bedrock.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "example" {
  name               = "bedrock-flow-role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_bedrockagent_flow" "example" {
  name               = "example-flow"
  description        = "Example Bedrock Flow"
  execution_role_arn = aws_iam_role.example.arn
}
```

### With Flow Definition

```terraform
resource "aws_bedrockagent_flow" "example" {
  name               = "example-flow"
  description        = "Example Bedrock Flow with definition"
  execution_role_arn = aws_iam_role.example.arn

  definition {
    nodes {
      name = "input"
      type = "INPUT"
      
      configuration {
        input {}
      }
      
      outputs {
        name = "output"
        type = "TEXT"
      }
    }
    
    nodes {
      name = "output"
      type = "OUTPUT"
      
      configuration {
        output {}
      }
      
      inputs {
        name = "input"
        type = "TEXT"
      }
    }
    
    connections {
      name   = "connection1"
      source = "input"
      target = "output"
      type   = "DATA"
      
      configuration {
        data {
          source_output = "output"
          target_input  = "input"
        }
      }
    }
  }
}
```

### With Customer Encryption Key

```terraform
resource "aws_kms_key" "example" {
  description             = "KMS key for Bedrock Flow"
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_bedrockagent_flow" "example" {
  name                      = "example-flow"
  description               = "Example Bedrock Flow with customer encryption key"
  execution_role_arn        = aws_iam_role.example.arn
  customer_encryption_key_arn = aws_kms_key.example.arn
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the flow.
* `execution_role_arn` - (Required) ARN of the IAM role that allows the flow to access AWS resources.

The following arguments are optional:

* `description` - (Optional) Description of the flow.
* `customer_encryption_key_arn` - (Optional) ARN of the KMS key used to encrypt the flow. If not specified, AWS owned key is used.
* `prepare_flow` - (Optional) Whether to wait for the flow to be prepared. Default is `true`.
* `definition` - (Optional) Configuration that defines the flow. See [Definition](#definition) below for details.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Definition

The `definition` block supports the following:

* `nodes` - (Required) List of nodes in the flow. See [Nodes](#nodes) below for details.
* `connections` - (Required) List of connections between nodes in the flow. See [Connections](#connections) below for details.

#### Nodes

The `nodes` block supports the following:

* `name` - (Required) Name of the node.
* `type` - (Required) Type of the node. Valid values are `INPUT`, `OUTPUT`, `CONDITION`, `LAMBDA`, `PROMPT`, `KNOWLEDGE_BASE`, `AGENT`, and `RETRIEVAL`.
* `configuration` - (Required) Configuration for the node. The configuration depends on the node type. See [Node Configuration](#node-configuration) below for details.
* `inputs` - (Optional) List of inputs for the node. See [Node Inputs](#node-inputs) below for details.
* `outputs` - (Optional) List of outputs for the node. See [Node Outputs](#node-outputs) below for details.

##### Node Configuration

The `configuration` block supports the following:

* `input` - (Optional) Configuration for an INPUT node.
* `output` - (Optional) Configuration for an OUTPUT node.
* `condition` - (Optional) Configuration for a CONDITION node. See [Condition Configuration](#condition-configuration) below for details.
* `lambda` - (Optional) Configuration for a LAMBDA node. See [Lambda Configuration](#lambda-configuration) below for details.
* `prompt` - (Optional) Configuration for a PROMPT node. See [Prompt Configuration](#prompt-configuration) below for details.
* `knowledge_base` - (Optional) Configuration for a KNOWLEDGE_BASE node. See [Knowledge Base Configuration](#knowledge-base-configuration) below for details.
* `agent` - (Optional) Configuration for an AGENT node. See [Agent Configuration](#agent-configuration) below for details.
* `retrieval` - (Optional) Configuration for a RETRIEVAL node. See [Retrieval Configuration](#retrieval-configuration) below for details.

###### Condition Configuration

The `condition` block supports the following:

* `conditions` - (Required) List of conditions. See [Flow Condition](#flow-condition) below for details.

###### Prompt Configuration

The `prompt` block supports the following:

* `source_configuration` - (Required) Configuration for the prompt source. See [Prompt Source Configuration](#prompt-source-configuration) below for details.
* `guardrail_configuration` - (Optional) Configuration for the guardrail. See [Guardrail Configuration](#guardrail-configuration) below for details.

###### Prompt Source Configuration

The `source_configuration` block supports the following:

* `inline` - (Optional) Configuration for an inline prompt. See [Inline Prompt Configuration](#inline-prompt-configuration) below for details.
* `resource` - (Optional) Configuration for a prompt resource. See [Prompt Resource Configuration](#prompt-resource-configuration) below for details.

###### Inline Prompt Configuration

The `inline` block supports the following:

* `model_id` - (Required) ID of the model to use.
* `template_type` - (Required) Type of the template. Valid values are `TEXT` and `CHAT`.
* `additional_model_request_fields` - (Optional) Additional fields to include in the model request.
* `template_configuration` - (Required) Configuration for the template. See [Template Configuration](#template-configuration) below for details.
* `inference_configuration` - (Optional) Configuration for inference. See [Inference Configuration](#inference-configuration) below for details.

###### Prompt Resource Configuration

The `resource` block supports the following:

* `prompt_arn` - (Required) ARN of the prompt resource.

### Template Configuration

The `template_configuration` block supports the following:

* `chat` - (Optional) Configuration for a chat template. See [Chat Template Configuration](#chat-template-configuration) below for details.
* `text` - (Optional) Configuration for a text template. See [Text Template Configuration](#text-template-configuration) below for details.

#### Chat Template Configuration

The `chat` block supports the following:

* `input_variable` - (Optional) List of input variables for the template. See [Input Variable](#input-variable) below for details.
* `message` - (Required) List of messages in the chat. See [Message](#message) below for details.
* `system` - (Optional) System content for the chat. See [System Content](#system-content) below for details.
* `tool_configuration` - (Optional) Configuration for tools. See [Tool Configuration](#tool-configuration) below for details.

##### Input Variable

The `input_variable` block supports the following:

* `name` - (Required) Name of the input variable.

##### Message

The `message` block supports the following:

* `role` - (Required) Role of the message sender. Valid values are `USER` and `ASSISTANT`.
* `content` - (Required) Content of the message. See [Content Block](#content-block) below for details.

###### Content Block

The `content` block supports the following:

* `text` - (Optional) Text content of the message.
* `cache_point` - (Optional) Cache point configuration. See [Cache Point](#cache-point) below for details.

##### System Content

The `system` block supports the following:

* `text` - (Optional) Text content of the system message.
* `cache_point` - (Optional) Cache point configuration. See [Cache Point](#cache-point) below for details.

##### Tool Configuration

The `tool_configuration` block supports the following:

* `tool` - (Optional) List of tools. See [Tool](#tool) below for details.
* `tool_choice` - (Optional) Tool choice configuration. See [Tool Choice](#tool-choice) below for details.

###### Tool

The `tool` block supports the following:

* `cache_point` - (Optional) Cache point configuration. See [Cache Point](#cache-point) below for details.
* `tool_spec` - (Optional) Tool specification. See [Tool Specification](#tool-specification) below for details.

###### Tool Choice

The `tool_choice` block supports the following:

* `any` - (Optional) Any tool choice configuration.
* `auto` - (Optional) Auto tool choice configuration.
* `tool` - (Optional) Specific tool choice configuration. See [Specific Tool Choice](#specific-tool-choice) below for details.

###### Specific Tool Choice

The `tool` block supports the following:

* `name` - (Required) Name of the tool to use.

###### Tool Specification

The `tool_spec` block supports the following:

* `name` - (Required) Name of the tool.
* `description` - (Optional) Description of the tool.
* `input_schema` - (Optional) Input schema for the tool. See [Tool Input Schema](#tool-input-schema) below for details.

###### Tool Input Schema

The `input_schema` block supports the following:

* `json` - (Required) JSON schema for the tool input.

#### Text Template Configuration

The `text` block supports the following:

* `text` - (Required) Text template.
* `cache_point` - (Optional) Cache point configuration. See [Cache Point](#cache-point) below for details.
* `input_variable` - (Optional) List of input variables for the template. See [Input Variable](#input-variable) above for details.

### Cache Point

The `cache_point` block supports the following:

* `type` - (Required) Type of cache point. Valid values are `PROMPT_TEMPLATE` and `PROMPT_TEMPLATE_VARIABLE`.

###### Knowledge Base Configuration

The `knowledge_base` block supports the following:

* `knowledge_base_id` - (Required) ID of the knowledge base.
* `model_id` - (Optional) ID of the model to use.
* `number_of_results` - (Optional) Number of results to return.
* `guardrail_configuration` - (Optional) Configuration for the guardrail. See [Guardrail Configuration](#guardrail-configuration) below for details.
* `inference_configuration` - (Optional) Configuration for inference. See [Inference Configuration](#inference-configuration) below for details.
* `orchestration_configuration` - (Optional) Configuration for orchestration. See [Orchestration Configuration](#orchestration-configuration) below for details.
* `prompt_template` - (Optional) Template for the prompt. See [Knowledge Base Prompt Template](#knowledge-base-prompt-template) below for details.
* `reranking_configuration` - (Optional) Configuration for reranking. See [Reranking Configuration](#reranking-configuration) below for details.

###### Agent Configuration

The `agent` block supports the following:

* `agent_alias_arn` - (Required) ARN of the agent alias.

###### Retrieval Configuration

The `retrieval` block supports the following:

* `service_configuration` - (Required) Configuration for the retrieval service. See [Retrieval Service Configuration](#retrieval-service-configuration) below for details.

##### Node Inputs

The `inputs` block supports the following:

* `name` - (Required) Name of the input.
* `type` - (Required) Data type of the input. Valid values are `TEXT`, `JSON`, `IMAGE`, and `BINARY`.
* `category` - (Optional) Category of the input. Valid values are `REQUIRED`, `OPTIONAL`, and `STATIC`.
* `expression` - (Optional) Expression to evaluate for the input.

##### Node Outputs

The `outputs` block supports the following:

* `name` - (Required) Name of the output.
* `type` - (Required) Data type of the output. Valid values are `TEXT`, `JSON`, `IMAGE`, and `BINARY`.

#### Connections

The `connections` block supports the following:

* `name` - (Required) Name of the connection.
* `source` - (Required) Name of the source node.
* `target` - (Required) Name of the target node.
* `type` - (Required) Type of the connection. Valid values are `DATA` and `CONDITIONAL`.
* `configuration` - (Required) Configuration for the connection. See [Connection Configuration](#connection-configuration) below for details.

##### Connection Configuration

The `configuration` block supports the following:

* `data` - (Optional) Configuration for a DATA connection. See [Data Connection Configuration](#data-connection-configuration) below for details.
* `conditional` - (Optional) Configuration for a CONDITIONAL connection. See [Conditional Connection Configuration](#conditional-connection-configuration) below for details.

###### Data Connection Configuration

The `data` block supports the following:

* `source_output` - (Required) Name of the output from the source node.
* `target_input` - (Required) Name of the input to the target node.

###### Conditional Connection Configuration

The `conditional` block supports the following:

* `condition` - (Required) Name of the condition.

### Flow Condition

The `conditions` block supports the following:

* `name` - (Required) Name of the condition.
* `expression` - (Optional) Expression to evaluate for the condition.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the flow.
* `id` - ID of the flow.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock Agents Flow using the flow ID. For example:

```terraform
import {
  to = aws_bedrockagent_flow.example
  id = "flow-id-12345678"
}
```

Using `terraform import`, import Bedrock Agents Flow using the flow ID. For example:

```console
% terraform import aws_bedrockagent_flow.example flow-id-12345678
```
