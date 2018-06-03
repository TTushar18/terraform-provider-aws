package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsConfigAuthorization() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsConfigAuthorizationPut,
		Read:   resourceAwsConfigAuthorizationRead,
		Delete: resourceAwsConfigAuthorizationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsConfigAuthorizationPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn

	accountId := d.Get("account_id").(string)
	region := d.Get("region").(string)

	req := &configservice.PutAggregationAuthorizationInput{
		AuthorizedAccountId: aws.String(accountId),
		AuthorizedAwsRegion: aws.String(region),
	}

	_, err := conn.PutAggregationAuthorization(req)
	if err != nil {
		return fmt.Errorf("Error creating authorization: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", accountId, region))
	return resourceAwsConfigAuthorizationRead(d, meta)
}

func resourceAwsConfigAuthorizationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn

	accountId, region, err := resourceAwsConfigAuthorizationParseID(d.Id())
	if err != nil {
		return err
	}

	d.Set("account_id", accountId)
	d.Set("region", region)

	res, err := conn.DescribeAggregationAuthorizations(&configservice.DescribeAggregationAuthorizationsInput{})
	if err != nil {
		return fmt.Errorf("Error retrieving list of authorizations: %s", err)
	}

	// Check for existing authorization
	for _, auth := range res.AggregationAuthorizations {
		if accountId == *auth.AuthorizedAccountId && region == *auth.AuthorizedAwsRegion {
			d.Set("arn", auth.AggregationAuthorizationArn)
			return nil
		}
	}

	log.Printf("[WARN] Authorization not found, removing from state: %s", d.Id())
	d.SetId("")
	return nil
}

func resourceAwsConfigAuthorizationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn

	accountId, region, err := resourceAwsConfigAuthorizationParseID(d.Id())
	if err != nil {
		return err
	}

	req := &configservice.DeleteAggregationAuthorizationInput{
		AuthorizedAccountId: aws.String(accountId),
		AuthorizedAwsRegion: aws.String(region),
	}

	_, err = conn.DeleteAggregationAuthorization(req)
	if err != nil {
		return fmt.Errorf("Error deleting authorization: %s", err)
	}

	d.SetId("")
	return nil
}

func resourceAwsConfigAuthorizationParseID(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("Please make sure the ID is in the form account_id:region (i.e. 123456789012:us-east-1")
	}
	accountId := idParts[0]
	region := idParts[1]
	return accountId, region, nil
}
