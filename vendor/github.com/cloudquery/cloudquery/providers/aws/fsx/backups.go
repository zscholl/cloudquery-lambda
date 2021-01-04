package fsx

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/cloudquery/cloudquery/providers/common"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

type Backup struct {
	ID                   uint `gorm:"primarykey"`
	AccountID            string
	Region               string
	BackupId             *string
	CreationTime         *time.Time
	DirectoryInformation *fsx.ActiveDirectoryBackupAttributes `gorm:"embedded;embeddedPrefix:directory_information_"`
	FailureDetails       *fsx.BackupFailureDetails            `gorm:"embedded;embeddedPrefix:failure_details_"`
	KmsKeyId             *string
	Lifecycle            *string
	ProgressPercent      *int64
	ResourceARN          *string
	Tags                 []*BackupTag `gorm:"constraint:OnDelete:CASCADE;"`
	Type                 *string
}

func (Backup) TableName() string {
	return "aws_fsx_backups"
}

type BackupTag struct {
	ID       uint `gorm:"primarykey"`
	BackupID uint
	Key      *string
	Value    *string
}

func (BackupTag) TableName() string {
	return "aws_fsx_backup_tags"
}

func (c *Client) transformBackupTag(value *fsx.Tag) *BackupTag {
	return &BackupTag{
		Key:   value.Key,
		Value: value.Value,
	}
}

func (c *Client) transformBackupTags(values []*fsx.Tag) []*BackupTag {
	var tValues []*BackupTag
	for _, v := range values {
		tValues = append(tValues, c.transformBackupTag(v))
	}
	return tValues
}

func (c *Client) transformBackup(value *fsx.Backup) *Backup {
	return &Backup{
		Region:               c.region,
		AccountID:            c.accountID,
		BackupId:             value.BackupId,
		CreationTime:         value.CreationTime,
		DirectoryInformation: value.DirectoryInformation,
		FailureDetails:       value.FailureDetails,
		KmsKeyId:             value.KmsKeyId,
		Lifecycle:            value.Lifecycle,
		ProgressPercent:      value.ProgressPercent,
		ResourceARN:          value.ResourceARN,
		Tags:                 c.transformBackupTags(value.Tags),
		Type:                 value.Type,
	}
}

func (c *Client) transformBackups(values []*fsx.Backup) []*Backup {
	var tValues []*Backup
	for _, v := range values {
		tValues = append(tValues, c.transformBackup(v))
	}
	return tValues
}

func MigrateBackups(db *gorm.DB) error {
	return db.AutoMigrate(
		&Backup{},
		&BackupTag{},
	)
}

func (c *Client) backups(gConfig interface{}) error {
	var config fsx.DescribeBackupsInput
	err := mapstructure.Decode(gConfig, &config)
	if err != nil {
		return err
	}

	for {
		output, err := c.svc.DescribeBackups(&config)
		if err != nil {
			return err
		}
		c.db.Where("region = ?", c.region).Where("account_id = ?", c.accountID).Delete(&Backup{})
		common.ChunkedCreate(c.db, c.transformBackups(output.Backups))
		c.log.Info("Fetched resources", zap.String("resource", "fsx.backups"), zap.Int("count", len(output.Backups)))
		if aws.StringValue(output.NextToken) == "" {
			break
		}
		config.NextToken = output.NextToken
	}
	return nil
}
