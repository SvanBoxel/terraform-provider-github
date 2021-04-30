package github

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceGithubActionsAllowed() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubActionsAllowedRead,
		Schema: map[string]*schema.Schema{
			"github_allowed": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"verified_allowed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"patterns_allowed": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceGithubActionsAllowedRead(d *schema.ResourceData, meta interface{}) error {
	owner := meta.(*Owner).name
	log.Printf("[INFO] Refreshing GitHub Actions Public Key from: %s", owner)

	client := meta.(*Owner).v3client
	ctx := context.Background()
	actionsAllowed, _, err := client.Organizations.GetActionsAllowed(ctx, owner)

	if err != nil {
		return err
	}
	log.Println(actionsAllowed)

	d.SetId(fmt.Sprintf("%s/github-allowed-action", owner))
	d.Set("github_allowed", actionsAllowed.GetGithubOwnedAllowed())
	d.Set("verified_allowed", actionsAllowed.GetVerifiedAllowed())
	d.Set("patterns_allowed", actionsAllowed.PatternsAllowed)

	return nil
}
