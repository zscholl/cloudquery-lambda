package directconnect

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/cloudquery/cloudquery/providers/common"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Gateway struct {
	ID                        uint `gorm:"primarykey"`
	AccountID                 string
	Region                    string
	AmazonSideAsn             *int64
	DirectConnectGatewayId    *string
	DirectConnectGatewayName  *string
	DirectConnectGatewayState *string
	OwnerAccount              *string
	StateChangeError          *string
}

func (Gateway) TableName() string {
	return "aws_directconnect_gateways"
}

func (c *Client) transformGateway(value *directconnect.Gateway) *Gateway {
	return &Gateway{
		Region:                    c.region,
		AccountID:                 c.accountID,
		AmazonSideAsn:             value.AmazonSideAsn,
		DirectConnectGatewayId:    value.DirectConnectGatewayId,
		DirectConnectGatewayName:  value.DirectConnectGatewayName,
		DirectConnectGatewayState: value.DirectConnectGatewayState,
		OwnerAccount:              value.OwnerAccount,
		StateChangeError:          value.StateChangeError,
	}
}

func (c *Client) transformGateways(values []*directconnect.Gateway) []*Gateway {
	var tValues []*Gateway
	for _, v := range values {
		tValues = append(tValues, c.transformGateway(v))
	}
	return tValues
}

func MigrateGateways(db *gorm.DB) error {
	return db.AutoMigrate(
		&Gateway{},
	)
}

func (c *Client) gateways(gConfig interface{}) error {
	var config directconnect.DescribeDirectConnectGatewaysInput
	err := mapstructure.Decode(gConfig, &config)
	if err != nil {
		return err
	}

	for {
		output, err := c.svc.DescribeDirectConnectGateways(&config)
		if err != nil {
			return err
		}
		c.db.Where("region = ?", c.region).Where("account_id = ?", c.accountID).Delete(&Gateway{})
		common.ChunkedCreate(c.db, c.transformGateways(output.DirectConnectGateways))
		c.log.Info("Fetched resources", zap.String("resource", "directconnect.gateways"), zap.Int("count", len(output.DirectConnectGateways)))
		if aws.StringValue(output.NextToken) == "" {
			break
		}
		config.NextToken = output.NextToken
	}
	return nil
}
