// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rekognition_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfrekognition "github.com/hashicorp/terraform-provider-aws/internal/service/rekognition"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRekognitionProject_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_project.test"
	feature := "CONTENT_MODERATION"
	autoUpdate := "ENABLED"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RekognitionEndpointID)
			testAccProjectPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RekognitionServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx, feature, rName),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_contentModeration(rName, autoUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "rekognition", regexache.MustCompile(`project/`+rName+`/\d+$`)),
					resource.TestCheckResourceAttr(resourceName, "auto_update", autoUpdate),
					resource.TestCheckResourceAttr(resourceName, "feature", feature),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRekognitionProject_ContentModeration(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_project.test"
	feature := "CONTENT_MODERATION"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RekognitionEndpointID)
			testAccProjectPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RekognitionServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_contentModeration(rName+"-1", "ENABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrID, rName+"-1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName+"-1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "rekognition", regexache.MustCompile(`project/`+rName+`-1/\d+$`)),
					resource.TestCheckResourceAttr(resourceName, "auto_update", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "feature", feature),
				),
			},
			{
				Config: testAccProjectConfig_contentModeration(rName+"-2", "DISABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrID, rName+"-2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName+"-2"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "rekognition", regexache.MustCompile(`project/`+rName+`-2/\d+$`)),
					resource.TestCheckResourceAttr(resourceName, "auto_update", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "feature", feature),
				),
			},
		},
	})
}

func TestAccRekognitionProject_CustomLabels(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_project.test"
	feature := "CUSTOM_LABELS"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RekognitionEndpointID)
			testAccProjectPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RekognitionServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx, feature, rName),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_customLabels(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "rekognition", regexache.MustCompile(`project/`+rName+`/\d+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrID, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "feature", feature),
					resource.TestCheckNoResourceAttr(resourceName, "auto_update"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRekognitionProject_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_project.test"
	feature := "CONTENT_MODERATION"
	autoUpdate := "ENABLED"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RekognitionEndpointID)
			testAccProjectPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RekognitionServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx, feature, rName),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_contentModeration(rName, autoUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfrekognition.ResourceProject, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRekognitionProject_tags(t *testing.T) {
	ctx := acctest.Context(t)

	rProjectId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_project.test"
	feature := "CUSTOM_LABELS"

	tags1 := `
  tags = {
    key1 = "value1"
  }
`
	tags2 := `
  tags = {
    key1 = "value1"
    key2 = "value2"
  }
`
	tags3 := `
  tags = {
    key2 = "value2"
  }
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RekognitionEndpointID)
			testAccProjectPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RekognitionServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx, feature, rProjectId),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_tags(rProjectId, tags1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccProjectConfig_tags(rProjectId, tags2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccProjectConfig_tags(rProjectId, tags3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckProjectExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Rekognition, create.ErrActionCheckingExistence, tfrekognition.ResNameProject, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Rekognition, create.ErrActionCheckingExistence, tfrekognition.ResNameProject, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RekognitionClient(ctx)
		_, err := tfrekognition.FindProjectByName(ctx, conn, rs.Primary.ID, "")

		if err != nil {
			return create.Error(names.Rekognition, create.ErrActionCheckingExistence, tfrekognition.ResNameProject, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckProjectDestroy(ctx context.Context, feature string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RekognitionClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rekognition_project" {
				continue
			}

			_, err := tfrekognition.FindProjectByName(ctx, conn, name, awstypes.CustomizationFeature(feature))
			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.Rekognition, create.ErrActionCheckingDestroyed, tfrekognition.ResNameProject, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccProjectPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RekognitionClient(ctx)

	input := &rekognition.DescribeProjectsInput{}
	_, err := conn.DescribeProjects(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccProjectConfig_contentModeration(rName string, autoUpdate string) string {
	return fmt.Sprintf(`
resource "aws_rekognition_project" "test" {
  name        = %[1]q
  auto_update = %[2]q
  feature     = "CONTENT_MODERATION"
}
`, rName, autoUpdate)
}

// auto-update not supported for custom_labels
func testAccProjectConfig_customLabels(rName string) string {
	return fmt.Sprintf(`
resource "aws_rekognition_project" "test" {
  name    = %[1]q
  feature = "CUSTOM_LABELS"
}
`, rName)
}

func testAccProjectConfig_tags(rName, tags string) string {
	return fmt.Sprintf(`
resource "aws_rekognition_project" "test" {
  name    = %[1]q
  feature = "CUSTOM_LABELS"

%[2]s
}
`, rName, tags)
}
