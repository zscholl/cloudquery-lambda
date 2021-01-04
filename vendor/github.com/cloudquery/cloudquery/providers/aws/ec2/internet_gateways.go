package ec2

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudquery/cloudquery/providers/common"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type InternetGateway struct {
	ID                uint `gorm:"primarykey"`
	AccountID         string
	Region            string
	Attachments       []*InternetGatewayAttachment `gorm:"constraint:OnDelete:CASCADE;"`
	InternetGatewayId *string
	OwnerId           *string
	Tags              []*InternetGatewayTag `gorm:"constraint:OnDelete:CASCADE;"`
}

func (InternetGateway) TableName() string {
	return "aws_ec2_internet_gateways"
}

type InternetGatewayAttachment struct {
	ID                uint `gorm:"primarykey"`
	InternetGatewayID uint
	State             *string
	VpcId             *string
}

func (InternetGatewayAttachment) TableName() string {
	return "aws_ec2_internet_gateway_attachments"
}

type InternetGatewayTag struct {
	ID                uint `gorm:"primarykey"`
	InternetGatewayID uint
	Key               *string
	Value             *string
}

func (InternetGatewayTag) TableName() string {
	return "aws_ec2_internet_gateway_tags"
}

func (c *Client) transformInternetGatewayAttachment(value *ec2.InternetGatewayAttachment) *InternetGatewayAttachment {
	return &InternetGatewayAttachment{
		State: value.State,
		VpcId: value.VpcId,
	}
}

func (c *Client) transformInternetGatewayAttachments(values []*ec2.InternetGatewayAttachment) []*InternetGatewayAttachment {
	var tValues []*InternetGatewayAttachment
	for _, v := range values {
		tValues = append(tValues, c.transformInternetGatewayAttachment(v))
	}
	return tValues
}

func (c *Client) transformInternetGatewayTag(value *ec2.Tag) *InternetGatewayTag {
	return &InternetGatewayTag{
		Key:   value.Key,
		Value: value.Value,
	}
}

func (c *Client) transformInternetGatewayTags(values []*ec2.Tag) []*InternetGatewayTag {
	var tValues []*InternetGatewayTag
	for _, v := range values {
		tValues = append(tValues, c.transformInternetGatewayTag(v))
	}
	return tValues
}

func (c *Client) transformInternetGateway(value *ec2.InternetGateway) *InternetGateway {
	return &InternetGateway{
		Region:            c.region,
		AccountID:         c.accountID,
		Attachments:       c.transformInternetGatewayAttachments(value.Attachments),
		InternetGatewayId: value.InternetGatewayId,
		OwnerId:           value.OwnerId,
		Tags:              c.transformInternetGatewayTags(value.Tags),
	}
}

func (c *Client) transformInternetGateways(values []*ec2.InternetGateway) []*InternetGateway {
	var tValues []*InternetGateway
	for _, v := range values {
		tValues = append(tValues, c.transformInternetGateway(v))
	}
	return tValues
}

func MigrateInternetGateways(db *gorm.DB) error {
	return db.AutoMigrate(
		&InternetGateway{},
		&InternetGatewayAttachment{},
		&InternetGatewayTag{},
	)
}

func (c *Client) internetGateways(gConfig interface{}) error {
	var config ec2.DescribeInternetGatewaysInput
	err := mapstructure.Decode(gConfig, &config)
	if err != nil {
		return err
	}

	for {
		output, err := c.svc.DescribeInternetGateways(&config)
		if err != nil {
			return err
		}
		c.db.Where("region = ?", c.region).Where("account_id = ?", c.accountID).Delete(&InternetGateway{})
		common.ChunkedCreate(c.db, c.transformInternetGateways(output.InternetGateways))
		c.log.Info("Fetched resources", zap.Int("count", len(output.InternetGateways)))
		if aws.StringValue(output.NextToken) == "" {
			break
		}
		config.NextToken = output.NextToken
	}
	return nil
}
