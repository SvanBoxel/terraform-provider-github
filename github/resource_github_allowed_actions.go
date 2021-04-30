package github

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/go-github/v35/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceGithubAllowedActions() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubAllowedActionsCreateOrUpdateOrDelete,
		Read:   resourceGithubAllowedActionsRead,
		Update: resourceGithubAllowedActionsCreateOrUpdateOrDelete,
		Delete: resourceGithubAllowedActionsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"enabled_repositories": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateValueFunc([]string{"all", "selected", "disabled"}),
				// ForceNew:     true,
			},
			"allowed_actions": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateValueFunc([]string{"all", "local_only", "selected"}),
				ForceNew:     true,
			},
			"github_allowed": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"verified_allowed": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"patterns_allowed": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceGithubAllowedActionsCreateOrUpdateOrDelete(
	d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)

	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name

	enabledRepositories := d.Get("enabled_repositories").(string)
	allowedActions := d.Get("allowed_actions").(string)

	githubAllowed := d.Get("github_allowed").(bool)
	verifiedAllowed := d.Get("verified_allowed").(bool)
	patternsAllowed := strings.Split(d.Get("patterns_allowed").(string), ",")
	ctx := context.Background()

	if allowedActions != "" && enabledRepositories == "selected" {
		return fmt.Errorf("Cannot set allowed actions if Actions is only enabled for enabled repositories. This is a limitation in the Terraform GitHub Provider.")
	} else if allowedActions != "" && enabledRepositories == "disabled" {
		return fmt.Errorf("Cannot set allowed actions if Actions is disabled for repositories")
	}

	log.Printf("[DEBUG] Setting Actions permissions for %s", orgName)

	client.Organizations.EditActionsPermissions(ctx,
		orgName,
		github.ActionsPermissions{
			EnabledRepositories: &enabledRepositories,
			AllowedActions:      &allowedActions,
		})

	log.Printf("[DEBUG] Setting allowed Actions for %s", orgName)
	actionsAllowed, _, err := client.Organizations.EditActionsAllowed(ctx,
		orgName,
		github.ActionsAllowed{
			GithubOwnedAllowed: &githubAllowed,
			VerifiedAllowed:    &verifiedAllowed,
			PatternsAllowed:    patternsAllowed,
		})

	actionsAllowedAsString := actionsAllowed.String()
	log.Printf(actionsAllowedAsString)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s/github-allowed-action", orgName))
	return resourceGithubOrganizationProjectRead(d, meta)
}

func resourceGithubAllowedActionsRead(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	orgName := strings.Split(d.Id(), "/")[0]

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	// if !d.IsNewResource() {
	// 	ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	// }

	log.Printf("[DEBUG] Reading organization allowed settings for %s", orgName)
	actionsPermissions, resp, err := client.Organizations.GetActionsPermissions(ctx, orgName)
	actionsAllowed, resp, err := client.Organizations.GetActionsAllowed(ctx, orgName)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[WARN] Removing organization allowed settings for %s", orgName)
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("etag", resp.Header.Get("ETag"))
	d.Set("enabled_repositories", actionsPermissions.GetEnabledRepositories())
	d.Set("allowed_actions", actionsPermissions.GetAllowedActions())
	d.Set("github_allowed", actionsAllowed.GetGithubOwnedAllowed())
	d.Set("verified_allowed", actionsAllowed.GetVerifiedAllowed())
	d.Set("patterns_allowed", actionsAllowed.PatternsAllowed)

	return nil
}

func resourceGithubAllowedActionsDelete(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	orgName := strings.Split(d.Id(), "/")[0]
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	log.Printf("[DEBUG] Resetting organization allowed actions settings for %s", orgName)

	enabledRepositories := "all"
	_, _, err = client.Organizations.EditActionsPermissions(ctx,
		orgName,
		github.ActionsPermissions{
			EnabledRepositories: &enabledRepositories,
		})

	if err != nil {
		return err
	}
	return nil
}
