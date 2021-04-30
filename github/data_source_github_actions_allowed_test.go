package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccGithubActionsAllowedDataSource(t *testing.T) {

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("queries a repository milestone", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_actions_allowed" "test" {
				enabled_repositories = "all"
				allowed_actions      = "all"
			  github_allowed       = true
				verified_allowed     = true
				patterns_allowed     = "foo/%s"
			}

			data "github_actions_allowed" "test" {
			}

		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet("data.github_actions_allowed.test", "github_allowed"),
			resource.TestCheckResourceAttrSet("data.github_actions_allowed.test", "verified_allowed"),
			resource.TestCheckResourceAttrSet("data.github_actions_allowed.test", "patterns_allowed"),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  check,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})
	})

}
