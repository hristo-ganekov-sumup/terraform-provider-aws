package aws

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEc2TransitGatewayRouteTable() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEc2TransitGatewayRouteTableRead,

		Schema: map[string]*schema.Schema{
			"index": {
				Type:     schema.TypeInt,
				Default:  0,
				Optional: true,
			},
			"index_last": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"default_association_route_table": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"default_propagation_route_table": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"filter": dataSourceFiltersSchema(),
			"id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"transit_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsEc2TransitGatewayRouteTableRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeTransitGatewayRouteTablesInput{}

	index := d.Get("index").(int)

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = buildAwsDataSourceFilters(v.(*schema.Set))
	}

	if v, ok := d.GetOk("id"); ok {
		input.TransitGatewayRouteTableIds = []*string{aws.String(v.(string))}
	}

	log.Printf("[DEBUG] Reading EC2 Transit Gateways: %s", input)
	output, err := conn.DescribeTransitGatewayRouteTables(input)

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Route Table: %s", err)
	}

	if output == nil || len(output.TransitGatewayRouteTables) == 0 {
		return errors.New("error reading EC2 Transit Gateway Route Table: no results found")
	}

	if index > len(output.TransitGatewayRouteTables) {
		return errors.New("Index out of range")
	}

	transitGatewayRouteTable := output.TransitGatewayRouteTables[index]

	if transitGatewayRouteTable == nil {
		return errors.New("error reading EC2 Transit Gateway Route Table: empty result")
	}

	d.Set("default_association_route_table", aws.BoolValue(transitGatewayRouteTable.DefaultAssociationRouteTable))
	d.Set("default_propagation_route_table", aws.BoolValue(transitGatewayRouteTable.DefaultPropagationRouteTable))
	d.Set("index_last", len(output.TransitGatewayRouteTables))

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(transitGatewayRouteTable.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("transit_gateway_id", aws.StringValue(transitGatewayRouteTable.TransitGatewayId))

	d.SetId(aws.StringValue(transitGatewayRouteTable.TransitGatewayRouteTableId))

	return nil
}
