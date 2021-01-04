package elasticbeanstalk

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/cloudquery/cloudquery/providers/common"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

type Environment struct {
	ID                           uint `gorm:"primarykey"`
	AccountID                    string
	Region                       string
	AbortableOperationInProgress *bool
	ApplicationName              *string
	CNAME                        *string
	DateCreated                  *time.Time
	DateUpdated                  *time.Time
	Description                  *string
	EndpointURL                  *string
	EnvironmentArn               *string
	EnvironmentId                *string
	EnvironmentLinks             []*EnvironmentLink `gorm:"constraint:OnDelete:CASCADE;"`
	EnvironmentName              *string
	Health                       *string
	HealthStatus                 *string
	OperationsRole               *string
	PlatformArn                  *string
	Resources                    *EnvironmentResource `gorm:"constraint:OnDelete:CASCADE;"`
	SolutionStackName            *string
	Status                       *string
	TemplateName                 *string
	Tier                         *elasticbeanstalk.EnvironmentTier `gorm:"embedded;embeddedPrefix:tier_"`
	VersionLabel                 *string
}

func (Environment) TableName() string {
	return "aws_elasticbeanstalk_environments"
}

type EnvironmentLink struct {
	ID              uint `gorm:"primarykey"`
	EnvironmentID   uint
	EnvironmentName *string
	LinkName        *string
}

func (EnvironmentLink) TableName() string {
	return "aws_elasticbeanstalk_environment_links"
}

type EnvironmentResource struct {
	ID            uint `gorm:"primarykey"`
	EnvironmentID uint
	LoadBalancer  *EnvironmentLoadBalancer `gorm:"constraint:OnDelete:CASCADE;"`
}

func (EnvironmentResource) TableName() string {
	return "aws_elasticbeanstalk_environment_resources"
}

type EnvironmentLoadBalancer struct {
	ID                    uint `gorm:"primarykey"`
	EnvironmentResourceID uint
	Domain                *string
	Listeners             []*EnvironmentListener `gorm:"constraint:OnDelete:CASCADE;"`
	LoadBalancerName      *string
}

func (EnvironmentLoadBalancer) TableName() string {
	return "aws_elasticbeanstalk_environment_load_balancers"
}

type EnvironmentListener struct {
	ID                        uint `gorm:"primarykey"`
	EnvironmentLoadBalancerID uint
	Port                      *int64
	Protocol                  *string
}

func (EnvironmentListener) TableName() string {
	return "aws_elasticbeanstalk_environment_listeners"
}

func (c *Client) transformEnvironmentLink(value *elasticbeanstalk.EnvironmentLink) *EnvironmentLink {
	return &EnvironmentLink{
		EnvironmentName: value.EnvironmentName,
		LinkName:        value.LinkName,
	}
}

func (c *Client) transformEnvironmentDescriptionEnvironmentLinks(values []*elasticbeanstalk.EnvironmentLink) []*EnvironmentLink {
	var tValues []*EnvironmentLink
	for _, v := range values {
		tValues = append(tValues, c.transformEnvironmentLink(v))
	}
	return tValues
}

func (c *Client) transformEnvironmentListener(value *elasticbeanstalk.Listener) *EnvironmentListener {
	return &EnvironmentListener{
		Port:     value.Port,
		Protocol: value.Protocol,
	}
}

func (c *Client) transformEnvironmentListeners(values []*elasticbeanstalk.Listener) []*EnvironmentListener {
	var tValues []*EnvironmentListener
	for _, v := range values {
		tValues = append(tValues, c.transformEnvironmentListener(v))
	}
	return tValues
}

func (c *Client) transformEnvironmentLoadBalancer(value *elasticbeanstalk.LoadBalancerDescription) *EnvironmentLoadBalancer {
	return &EnvironmentLoadBalancer{
		Domain:           value.Domain,
		Listeners:        c.transformEnvironmentListeners(value.Listeners),
		LoadBalancerName: value.LoadBalancerName,
	}
}

func (c *Client) transformEnvironmentResources(value *elasticbeanstalk.EnvironmentResourcesDescription) *EnvironmentResource {
	return &EnvironmentResource{
		LoadBalancer: c.transformEnvironmentLoadBalancer(value.LoadBalancer),
	}
}

func (c *Client) transformEnvironment(value *elasticbeanstalk.EnvironmentDescription) *Environment {
	res := Environment{
		Region:                       c.region,
		AccountID:                    c.accountID,
		AbortableOperationInProgress: value.AbortableOperationInProgress,
		ApplicationName:              value.ApplicationName,
		CNAME:                        value.CNAME,
		DateCreated:                  value.DateCreated,
		DateUpdated:                  value.DateUpdated,
		Description:                  value.Description,
		EndpointURL:                  value.EndpointURL,
		EnvironmentArn:               value.EnvironmentArn,
		EnvironmentId:                value.EnvironmentId,
		EnvironmentName:              value.EnvironmentName,
		Health:                       value.Health,
		HealthStatus:                 value.HealthStatus,
		OperationsRole:               value.OperationsRole,
		PlatformArn:                  value.PlatformArn,
		SolutionStackName:            value.SolutionStackName,
		Status:                       value.Status,
		TemplateName:                 value.TemplateName,
		Tier:                         value.Tier,
		VersionLabel:                 value.VersionLabel,
	}

	if value.EnvironmentLinks != nil {
		res.EnvironmentLinks = c.transformEnvironmentDescriptionEnvironmentLinks(value.EnvironmentLinks)
	}

	if value.Resources != nil {
		res.Resources = c.transformEnvironmentResources(value.Resources)
	}

	return &res
}

func (c *Client) transformEnvironments(values []*elasticbeanstalk.EnvironmentDescription) []*Environment {
	var tValues []*Environment
	for _, v := range values {
		tValues = append(tValues, c.transformEnvironment(v))
	}
	return tValues
}

func MigrateEnvironments(db *gorm.DB) error {
	return db.AutoMigrate(
		&Environment{},
		&EnvironmentLink{},
		&EnvironmentResource{},
		&EnvironmentLoadBalancer{},
		&EnvironmentListener{},
	)
}

func (c *Client) environments(gConfig interface{}) error {
	var config elasticbeanstalk.DescribeEnvironmentsInput
	err := mapstructure.Decode(gConfig, &config)
	if err != nil {
		return err
	}

	for {
		output, err := c.svc.DescribeEnvironments(&config)
		if err != nil {
			return err
		}
		c.db.Where("region = ?", c.region).Where("account_id = ?", c.accountID).Delete(&Environment{})
		common.ChunkedCreate(c.db, c.transformEnvironments(output.Environments))
		c.log.Info("Fetched resources", zap.String("resource", "elasticbeanstalk.environments"), zap.Int("count", len(output.Environments)))
		if aws.StringValue(output.NextToken) == "" {
			break
		}
		config.NextToken = output.NextToken
	}
	return nil
}
